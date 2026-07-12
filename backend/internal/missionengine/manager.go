package missionengine

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"challenge-to-you/backend/internal/db"
	"challenge-to-you/backend/internal/eventbus"
)

// MissionManager manages all active mission sessions
type MissionManager struct {
	mu              sync.RWMutex
	registry        *MissionRegistry
	stateMachine    *StateMachine
	objectiveEngine *ObjectiveEngine
	bus             *eventbus.EventBus
	db              *db.DB

	// Active sessions per player
	sessions map[string]map[string]*MissionSession // playerID -> missionID -> session

	// Player progress
	completedMissions map[string]map[string]bool // playerID -> missionID -> completed
}

// NewMissionManager creates a new mission manager
func NewMissionManager(registry *MissionRegistry, bus *eventbus.EventBus, database *db.DB) *MissionManager {
	mm := &MissionManager{
		registry:          registry,
		stateMachine:      NewStateMachine(),
		objectiveEngine:   NewObjectiveEngine(),
		bus:               bus,
		db:                database,
		sessions:          make(map[string]map[string]*MissionSession),
		completedMissions: make(map[string]map[string]bool),
	}
	mm.loadPersistedState()
	return mm
}

// loadPersistedState restores mission completion and active sessions from the database
func (mm *MissionManager) loadPersistedState() {
	completed, err := mm.db.GetCompletedMissions()
	if err != nil {
		log.Printf("Warning: failed to load completed missions: %v", err)
		return
	}
	for _, cm := range completed {
		if mm.completedMissions["player"] == nil {
			mm.completedMissions["player"] = make(map[string]bool)
		}
		mm.completedMissions["player"][cm.MissionID] = true
	}

	// Restore active mission sessions (preserves fabric state across restarts / reconnects)
	activeMissions, err := mm.db.GetActiveMissions()
	if err != nil {
		log.Printf("Warning: failed to load active missions: %v", err)
		return
	}
	for _, am := range activeMissions {
		var sess MissionSession
		if err := json.Unmarshal([]byte(am.SessionJSON), &sess); err != nil {
			log.Printf("Warning: failed to deserialize active session %s: %v", am.MissionID, err)
			continue
		}
		// Backward compatibility: if FabricState is empty but State has data, migrate it
		if len(sess.FabricState) == 0 {
			// We need to check the raw JSON for legacy "state" field
			var raw map[string]interface{}
			if err := json.Unmarshal([]byte(am.SessionJSON), &raw); err == nil {
				if legacyState, ok := raw["state"].(map[string]interface{}); ok {
					sess.FabricState = legacyState
				}
			}
		}

		// Re-link the live Mission pointer from the registry (JSON stores data but not the pointer)
		if mission, ok := mm.registry.Get(am.MissionID); ok {
			sess.Mission = mission
		} else {
			log.Printf("Warning: mission %s not found in registry, skipping restore", am.MissionID)
			continue
		}
		if mm.sessions[am.PlayerID] == nil {
			mm.sessions[am.PlayerID] = make(map[string]*MissionSession)
		}
		mm.sessions[am.PlayerID][am.MissionID] = &sess
		log.Printf("Restored active mission session: player=%s mission=%s fabric_keys=%d",
			am.PlayerID, am.MissionID, len(sess.FabricState))
	}
}

// GetAvailableMissions returns missions a player can start
func (mm *MissionManager) GetAvailableMissions(playerID string) []*Mission {
	completed := mm.completedMissions[playerID]
	if completed == nil {
		completed = make(map[string]bool)
	}

	var available []*Mission
	for _, mission := range mm.registry.GetAll() {
		if CanStart(mission, completed) {
			// Check if already has active session
			if mm.GetActiveSession(playerID, mission.ID) == nil {
				available = append(available, mission)
			}
		}
	}
	return available
}

// StartMission begins a mission for a player
func (mm *MissionManager) StartMission(playerID string, missionID string) (*MissionSession, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Check if mission exists
	mission, ok := mm.registry.Get(missionID)
	if !ok {
		return nil, fmt.Errorf("mission %s not found", missionID)
	}

	// Check if already completed
	if mm.completedMissions[playerID] != nil && mm.completedMissions[playerID][missionID] {
		return nil, fmt.Errorf("mission %s already completed", missionID)
	}

	// Check if already has active session (internal, no lock)
	if playerSessions, ok := mm.sessions[playerID]; ok {
		if playerSessions[missionID] != nil {
			return nil, fmt.Errorf("mission %s already active", missionID)
		}
	}

	// Check requirements
	completed := mm.completedMissions[playerID]
	if completed == nil {
		completed = make(map[string]bool)
	}
	if !CanStart(mission, completed) {
		return nil, fmt.Errorf("mission %s requirements not met", missionID)
	}

	// Create session
	session := &MissionSession{
		Mission:           mission,
		PlayerID:          playerID,
		Status:            MissionStatusAvailable,
		CurrentStep:       0,
		ObjectiveStatuses: make(map[string]ObjectiveStatus),
		FabricState:       make(map[string]interface{}),
		StartTime:         time.Time{},
	}

	// Start the mission
	if err := mm.stateMachine.StartMission(session); err != nil {
		return nil, fmt.Errorf("failed to start mission: %w", err)
	}

	// Store session
	if mm.sessions[playerID] == nil {
		mm.sessions[playerID] = make(map[string]*MissionSession)
	}
	mm.sessions[playerID][missionID] = session

	// Persist to database
	if err := mm.db.SaveActiveMission(missionID, playerID, session); err != nil {
		log.Printf("Warning: failed to persist active mission %s: %v", missionID, err)
	}

	// Publish event
	mm.bus.Publish(eventbus.Event{
		Type:    eventbus.EventMissionStarted,
		Payload: session,
		Source:  "mission_manager",
	})

	return session, nil
}

// GetActiveSession returns the active session for a mission
func (mm *MissionManager) GetActiveSession(playerID string, missionID string) *MissionSession {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if playerSessions, ok := mm.sessions[playerID]; ok {
		return playerSessions[missionID]
	}
	return nil
}

// GetPlayerSessions returns all active sessions for a player
func (mm *MissionManager) GetPlayerSessions(playerID string) []*MissionSession {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var sessions []*MissionSession
	if playerSessions, ok := mm.sessions[playerID]; ok {
		for _, session := range playerSessions {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// ProcessEvent processes an objective event
func (mm *MissionManager) ProcessEvent(playerID string, event ObjectiveEvent) ([]string, error) {
	mm.mu.Lock()

	sessions := mm.sessions[playerID]
	if sessions == nil {
		mm.mu.Unlock()
		return nil, nil
	}

	var allCompleted []string
	var completedMissions []*MissionSession

	for _, session := range sessions {
		if session.Status != MissionStatusActive {
			continue
		}

		completedIDs := mm.objectiveEngine.ProcessEvent(session, event)

		for _, objID := range completedIDs {
			if err := mm.stateMachine.CompleteObjective(session, objID); err != nil {
				continue
			}
			allCompleted = append(allCompleted, objID)
		}

		if session.Status == MissionStatusCompleted {
			mm.handleMissionComplete(playerID, session)
			completedMissions = append(completedMissions, session)
		}
	}

	mm.mu.Unlock()

	// Publish events without holding the lock
	for _, session := range completedMissions {
		mm.publishMissionComplete(playerID, session)
	}

	return allCompleted, nil
}

// CompleteChallenge marks a challenge as completed for a mission
func (mm *MissionManager) CompleteChallenge(playerID string, missionID string, challengeID string) error {
	mm.mu.Lock()

	session, ok := mm.sessions[playerID][missionID]
	if !ok {
		mm.mu.Unlock()
		return fmt.Errorf("no active session for mission %s", missionID)
	}

	if session.Status != MissionStatusActive {
		mm.mu.Unlock()
		return fmt.Errorf("mission %s is not active", missionID)
	}

	// Add to completed challenges
	session.CompletedChallenges = append(session.CompletedChallenges, challengeID)

	// Process as objective event
	event := ObjectiveEvent{
		Type:   "complete_challenge",
		Target: challengeID,
		Player: playerID,
	}

	completedIDs := mm.objectiveEngine.ProcessEvent(session, event)
	for _, objID := range completedIDs {
		mm.stateMachine.CompleteObjective(session, objID)
	}

	missionCompleted := session.IsComplete()
	if missionCompleted {
		mm.handleMissionComplete(playerID, session)
	}
	mm.mu.Unlock()

	if missionCompleted {
		mm.publishMissionComplete(playerID, session)
	}

	return nil
}

// FailMission fails a mission
func (mm *MissionManager) FailMission(playerID string, missionID string, reason string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	session, ok := mm.sessions[playerID][missionID]
	if !ok {
		return fmt.Errorf("no active session for mission %s", missionID)
	}

	return mm.stateMachine.FailMission(session, reason)
}

// CancelMission cancels an active mission
func (mm *MissionManager) CancelMission(playerID string, missionID string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	session, ok := mm.sessions[playerID][missionID]
	if !ok {
		return fmt.Errorf("no active session for mission %s", missionID)
	}

	if err := mm.stateMachine.CancelMission(session); err != nil {
		return err
	}

	// Remove session
	delete(mm.sessions[playerID], missionID)

	return nil
}

// GetMissionProgress returns progress for a mission
func (mm *MissionManager) GetMissionProgress(playerID string, missionID string) map[string]interface{} {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	session, ok := mm.sessions[playerID][missionID]
	if !ok {
		return nil
	}

	return mm.stateMachine.GetStatus(session)
}

// IsMissionCompleted checks if a mission is completed
func (mm *MissionManager) IsMissionCompleted(playerID string, missionID string) bool {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if mm.completedMissions[playerID] != nil {
		return mm.completedMissions[playerID][missionID]
	}
	return false
}

// handleMissionComplete processes mission completion (caller must hold mm.mu)
func (mm *MissionManager) handleMissionComplete(playerID string, session *MissionSession) {
	if mm.completedMissions[playerID] == nil {
		mm.completedMissions[playerID] = make(map[string]bool)
	}
	mm.completedMissions[playerID][session.Mission.ID] = true

	// Remove active session
	delete(mm.sessions[playerID], session.Mission.ID)

	// Persist to database
	if _, err := mm.db.RecordMissionCompletion(session.Mission.ID); err != nil {
		log.Printf("Warning: failed to record mission completion %s: %v", session.Mission.ID, err)
	}
	if err := mm.db.RemoveActiveMission(session.Mission.ID); err != nil {
		log.Printf("Warning: failed to remove active mission %s from DB: %v", session.Mission.ID, err)
	}
	if session.Mission.Reward != nil && session.Mission.Reward.XP > 0 {
		if err := mm.db.AddXP(session.Mission.Reward.XP); err != nil {
			log.Printf("Warning: failed to add XP for mission %s: %v", session.Mission.ID, err)
		}
	}
}

// publishMissionComplete publishes reward and completion events (no lock held)
func (mm *MissionManager) publishMissionComplete(playerID string, session *MissionSession) {
	if session.Mission.Reward != nil {
		mm.bus.Publish(eventbus.Event{
			Type:    eventbus.EventXPEarned,
			Payload: session.Mission.Reward,
			Source:  "mission_manager",
		})
	}

	mm.bus.Publish(eventbus.Event{
		Type: eventbus.EventMissionCompleted,
		Payload: map[string]interface{}{
			"player_id":  playerID,
			"mission_id": session.Mission.ID,
			"reward":     session.Mission.Reward,
		},
		Source: "mission_manager",
	})
}
