package missionengine

import (
	"fmt"
	"time"
)

// State transition errors
var (
	ErrInvalidTransition = fmt.Errorf("invalid state transition")
	ErrMissionNotActive  = fmt.Errorf("mission is not active")
	ErrObjectiveNotFound = fmt.Errorf("objective not found")
	ErrObjectiveComplete = fmt.Errorf("objective already completed")
	ErrMissionComplete   = fmt.Errorf("mission already completed")
)

// StateMachine manages mission state transitions
type StateMachine struct{}

// NewStateMachine creates a new state machine
func NewStateMachine() *StateMachine {
	return &StateMachine{}
}

// StartMission transitions a mission from Available to Active
func (sm *StateMachine) StartMission(session *MissionSession) error {
	if session.Status != MissionStatusAvailable {
		return fmt.Errorf("%w: can only start available missions, got %s", ErrInvalidTransition, session.Status)
	}

	session.Status = MissionStatusActive
	session.StartTime = time.Now()

	// Initialize all objectives to pending
	for _, obj := range session.Mission.Objectives {
		session.ObjectiveStatuses[obj.ID] = ObjectiveStatusPending
	}

	// Activate first objective if any
	if len(session.Mission.Objectives) > 0 {
		session.ObjectiveStatuses[session.Mission.Objectives[0].ID] = ObjectiveStatusActive
	}

	return nil
}

// CompleteObjective marks an objective as completed
func (sm *StateMachine) CompleteObjective(session *MissionSession, objectiveID string) error {
	if session.Status != MissionStatusActive {
		return fmt.Errorf("%w: mission must be active to complete objectives", ErrMissionNotActive)
	}

	// Find the objective
	found := false
	var completedObj *Objective
	for _, obj := range session.Mission.Objectives {
		if obj.ID == objectiveID {
			found = true
			completedObj = &obj
			break
		}
	}
	if !found {
		return fmt.Errorf("%w: %s", ErrObjectiveNotFound, objectiveID)
	}

	// Check if already completed
	status := session.ObjectiveStatuses[objectiveID]
	if status == ObjectiveStatusCompleted {
		return fmt.Errorf("%w: %s", ErrObjectiveComplete, objectiveID)
	}

	// Complete the objective
	session.ObjectiveStatuses[objectiveID] = ObjectiveStatusCompleted

	// Return the state changes that should be applied to the Fabric
	// The caller (GameLoop) is responsible for applying these to the Fabric
	if completedObj != nil && len(completedObj.SetsState) > 0 {
		// State changes are returned for the caller to apply to Fabric
		// We store them temporarily in FabricState for persistence
		if session.FabricState == nil {
			session.FabricState = make(map[string]interface{})
		}
		for k, v := range completedObj.SetsState {
			session.FabricState[k] = v
		}
	}

	// Activate next pending objective
	sm.activateNextObjective(session)

	// Check if mission is complete
	if session.IsComplete() {
		return sm.CompleteMission(session)
	}

	return nil
}

// FailObjective marks an objective as failed
func (sm *StateMachine) FailObjective(session *MissionSession, objectiveID string) error {
	if session.Status != MissionStatusActive {
		return fmt.Errorf("%w: mission must be active", ErrMissionNotActive)
	}

	session.ObjectiveStatuses[objectiveID] = ObjectiveStatusFailed
	session.FailedObjectives = append(session.FailedObjectives, objectiveID)

	// Check if any required objective failed
	for _, obj := range session.Mission.Objectives {
		if obj.ID == objectiveID && (obj.IsRequired || !obj.IsOptional) {
			return sm.FailMission(session, fmt.Sprintf("required objective %s failed", objectiveID))
		}
	}

	return nil
}

// CompleteMission transitions a mission to Completed
func (sm *StateMachine) CompleteMission(session *MissionSession) error {
	if session.Status == MissionStatusCompleted {
		return nil
	}

	if session.Status != MissionStatusActive {
		return fmt.Errorf("%w: can only complete active missions", ErrInvalidTransition)
	}

	if !session.IsComplete() {
		return nil
	}

	now := time.Now()
	session.Status = MissionStatusCompleted
	session.EndTime = &now

	return nil
}

// FailMission transitions a mission to Failed
func (sm *StateMachine) FailMission(session *MissionSession, reason string) error {
	if session.Status == MissionStatusCompleted || session.Status == MissionStatusFailed {
		return nil // Already in terminal state
	}

	now := time.Now()
	session.Status = MissionStatusFailed
	session.EndTime = &now

	// Store failure reason in FabricState for persistence
	if session.FabricState == nil {
		session.FabricState = make(map[string]interface{})
	}
	session.FabricState["failure_reason"] = reason

	return nil
}

// ExpireMission transitions a mission to Expired
func (sm *StateMachine) ExpireMission(session *MissionSession) error {
	if session.Status == MissionStatusCompleted || session.Status == MissionStatusFailed {
		return nil
	}

	now := time.Now()
	session.Status = MissionStatusExpired
	session.EndTime = &now

	return nil
}

// CancelMission transitions a mission back to Available
func (sm *StateMachine) CancelMission(session *MissionSession) error {
	if session.Status != MissionStatusActive {
		return fmt.Errorf("%w: can only cancel active missions", ErrInvalidTransition)
	}

	session.Status = MissionStatusAvailable
	session.CurrentStep = 0
	session.ObjectiveStatuses = make(map[string]ObjectiveStatus)
	session.FabricState = make(map[string]interface{})
	session.StartTime = time.Time{}
	session.EndTime = nil
	session.CompletedChallenges = nil
	session.FailedObjectives = nil

	return nil
}

// activateNextObjective finds and activates the next pending objective
func (sm *StateMachine) activateNextObjective(session *MissionSession) {
	for _, obj := range session.Mission.Objectives {
		status := session.ObjectiveStatuses[obj.ID]
		if status == ObjectiveStatusPending {
			// Check if prerequisites are met
			if sm.arePrereqsMet(session, obj) {
				session.ObjectiveStatuses[obj.ID] = ObjectiveStatusActive
				return
			}
		}
	}
}

// arePrereqsMet checks if an objective's prerequisites are satisfied
func (sm *StateMachine) arePrereqsMet(session *MissionSession, obj Objective) bool {
	for key, requiredValue := range obj.RequiresState {
		actualValue, ok := session.FabricState[key]
		if !ok {
			return false
		}
		if fmt.Sprintf("%v", actualValue) != fmt.Sprintf("%v", requiredValue) {
			return false
		}
	}
	return true
}

// GetStatus returns a summary of the mission status
func (sm *StateMachine) GetStatus(session *MissionSession) map[string]interface{} {
	completed := 0
	total := len(session.Mission.Objectives)
	for _, status := range session.ObjectiveStatuses {
		if status == ObjectiveStatusCompleted {
			completed++
		}
	}

	result := map[string]interface{}{
		"mission_id":       session.Mission.ID,
		"status":           session.Status.String(),
		"progress":         session.GetProgress(),
		"objectives_done":  completed,
		"objectives_total": total,
		"elapsed_time":     time.Since(session.StartTime).Seconds(),
	}

	if session.EndTime != nil {
		result["duration"] = session.EndTime.Sub(session.StartTime).Seconds()
	}

	return result
}
