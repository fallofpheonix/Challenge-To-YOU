package gameloop

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"challenge-to-you/backend/internal/engine"
	"challenge-to-you/backend/internal/eventbus"
	"challenge-to-you/backend/internal/missionengine"
)

type GameLoop struct {
	config GameLoopConfig

	fabric         *engine.AxiomaticFabric
	missionManager *missionengine.MissionManager
	missionSession *missionengine.MissionSession
	resourceMgr    *ResourceManager
	replayRecorder *ReplayRecorder
	telemetry      *TelemetryCollector

	currentTick int
	state       GameState

	eventChan chan GameEvent
	done      chan struct{}

	mu          sync.RWMutex
	tickResults []TickResult

	playerID  string
	missionID string
}

func NewGameLoop(
	fabric *engine.AxiomaticFabric,
	missionManager *missionengine.MissionManager,
	missionSession *missionengine.MissionSession,
	playerID string,
	missionID string,
	config GameLoopConfig,
	seed int64,
) *GameLoop {
	if config.TickInterval == 0 {
		config.TickInterval = 500 * time.Millisecond
	}

	return &GameLoop{
		config:         config,
		fabric:         fabric,
		missionManager: missionManager,
		missionSession: missionSession,
		resourceMgr:    NewResourceManager(),
		replayRecorder: NewReplayRecorder(seed),
		telemetry:      NewTelemetryCollector(1000),
		currentTick:    0,
		state:          StatePlaying,
		eventChan:      make(chan GameEvent, 100),
		done:           make(chan struct{}),
		playerID:       playerID,
		missionID:      missionID,
	}
}

func (gl *GameLoop) Start(ctx context.Context) <-chan TickResult {
	tickResults := make(chan TickResult, 100)

	go func() {
		defer close(tickResults)

		ticker := time.NewTicker(gl.config.TickInterval)
		defer ticker.Stop()

		tickChan := make(chan struct{}, 1)

		go func() {
			for {
				select {
				case <-gl.done:
					return
				case <-ticker.C:
					select {
					case tickChan <- struct{}{}:
					default:
					}
				}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				gl.shutdown()
				return
			case <-gl.done:
				return
			case ev := <-gl.eventChan:
				gl.replayRecorder.RecordEvent(ev)
				result := gl.processEvent(ev)
				if result != nil {
					tickResults <- *result
				}
			case <-tickChan:
				gl.currentTick++

				if gl.config.MaxTicks > 0 && gl.currentTick > gl.config.MaxTicks {
					result := gl.buildTickResult()
					tickResults <- result
					gl.shutdown()
					return
				}

				result := gl.executeTick()
				if result.Snapshot != nil {
					gl.recordTelemetry(result.Snapshot)
				}
				tickResults <- result
			}
		}
	}()

	return tickResults
}

func (gl *GameLoop) Stop() {
	select {
	case <-gl.done:
	default:
		close(gl.done)
	}
}

func (gl *GameLoop) shutdown() {
	select {
	case <-gl.done:
	default:
		close(gl.done)
	}
}

func (gl *GameLoop) Pause() {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	if gl.state == StatePlaying {
		gl.state = StatePaused
	}
}

func (gl *GameLoop) Resume() {
	gl.mu.Lock()
	defer gl.mu.Unlock()
	if gl.state == StatePaused {
		gl.state = StatePlaying
	}
}

func (gl *GameLoop) IsPaused() bool {
	gl.mu.RLock()
	defer gl.mu.RUnlock()
	return gl.state == StatePaused
}

func (gl *GameLoop) Events() chan<- GameEvent {
	return gl.eventChan
}

func (gl *GameLoop) Submit(ev GameEvent) {
	select {
	case gl.eventChan <- ev:
	default:
		log.Println("GameLoop: event channel full, dropping event")
	}
}

func (gl *GameLoop) State() GameState {
	gl.mu.RLock()
	defer gl.mu.RUnlock()
	return gl.state
}

func (gl *GameLoop) TickCount() int {
	gl.mu.RLock()
	defer gl.mu.RUnlock()
	return gl.currentTick
}

func (gl *GameLoop) Replay() *ReplayRecorder {
	return gl.replayRecorder
}

func (gl *GameLoop) Telemetry() *TelemetryCollector {
	return gl.telemetry
}

func (gl *GameLoop) Resources() *ResourceManager {
	return gl.resourceMgr
}

func (gl *GameLoop) Fabric() *engine.AxiomaticFabric {
	return gl.fabric
}

func (gl *GameLoop) MissionSession() *missionengine.MissionSession {
	return gl.missionSession
}

func (gl *GameLoop) Snapshot(message string) map[string]interface{} {
	return gl.buildSnapshot(message)
}

func (gl *GameLoop) MissionManager() *missionengine.MissionManager {
	return gl.missionManager
}

func (gl *GameLoop) processEvent(ev GameEvent) *TickResult {
	switch ev.Type {
	case EventPlayerInput:
		return gl.handlePlayerInput(ev)
	case EventSystem:
		return gl.handleSystemEvent(ev)
	default:
		return nil
	}
}

func (gl *GameLoop) handlePlayerInput(ev GameEvent) *TickResult {
	switch ev.Action {
	case "gather":
		resourceType := ResourceType(ev.Payload)
		result := gl.resourceMgr.Gather(gl.playerID, resourceType, 1)
		snapshot := gl.buildSnapshot(result.Message)
		return &TickResult{
			Tick:     gl.currentTick,
			State:    gl.state,
			Events:   []GameEvent{ev},
			Snapshot: snapshot,
		}

	case "deliver":
		resourceType := ResourceType(ev.Payload)
		result := gl.resourceMgr.Deliver(gl.playerID, resourceType, 1, "mission_target")
		snapshot := gl.buildSnapshot(result.Message)

		if result.Success && gl.missionSession != nil {
			objectiveEvent := missionengine.ObjectiveEvent{
				Type:   "deliver_item",
				Target: string(resourceType),
				Player: gl.playerID,
				Data:   map[string]interface{}{"amount": result.Removed},
			}
			completed, _ := gl.evaluateObjectives(objectiveEvent)
			if len(completed) > 0 {
				snapshot["objectives_completed"] = completed
			}
		}

		return &TickResult{
			Tick:     gl.currentTick,
			State:    gl.state,
			Events:   []GameEvent{ev},
			Snapshot: snapshot,
		}

	case "trigger_event":
		eventID := ev.Payload
		cipher, complete, err := gl.fabric.TriggerOntologicalShiftWithPayload(eventID, ev.Payload)
		if err != nil {
			snapshot := gl.buildSnapshot(fmt.Sprintf("Error: %v", err))
			snapshot["error"] = err.Error()
			if gl.fabric.IsPurged() {
				gl.state = StatePurged
				snapshot["game_over"] = true
				snapshot["game_state"] = "purged"
				if gl.config.EventBus != nil {
					gl.config.EventBus.Publish(eventbus.Event{
						Type: eventbus.EventPurge,
						Payload: map[string]interface{}{
							"reason": err.Error(),
							"tick":   gl.currentTick,
							"player": gl.playerID,
						},
						Source: "gameloop",
					})
				}
			}
			return &TickResult{
				Tick:     gl.currentTick,
				State:    gl.state,
				Events:   []GameEvent{ev},
				Snapshot: snapshot,
				Errors:   []string{err.Error()},
			}
		}

		snapshot := gl.buildSnapshot(fmt.Sprintf("Event %s processed", eventID))
		if complete {
			winMessage := fmt.Sprintf("Challenge complete! Cipher: %s", cipher)
			snapshot["level_complete"] = true
			snapshot["cipher"] = cipher
			snapshot["message"] = winMessage
			gl.state = StateWon

			if gl.missionSession != nil {
				_ = gl.missionManager.CompleteChallenge(gl.playerID, gl.missionID, eventID)
				if gl.missionManager.IsMissionCompleted(gl.playerID, gl.missionID) {
					snapshot["mission_complete"] = true
					snapshot["message"] = "MISSION COMPLETE! " + winMessage
				}
			}

			if gl.config.EventBus != nil {
				gl.config.EventBus.Publish(eventbus.Event{
					Type:    eventbus.EventChallengeCompleted,
					Payload: map[string]interface{}{"challenge_id": eventID, "tick": gl.currentTick, "player": gl.playerID},
					Source:  "gameloop",
				})
			}
		}

		return &TickResult{
			Tick:     gl.currentTick,
			State:    gl.state,
			Events:   []GameEvent{ev},
			Snapshot: snapshot,
		}

	case "submit_code":
		code := ev.Payload
		gl.fabric.SetState("code_submitted", true)

		if gl.missionSession != nil {
			objectiveEvent := missionengine.ObjectiveEvent{
				Type:   "code_submitted",
				Target: ev.Action,
				Player: gl.playerID,
				Data:   map[string]interface{}{"code_preview": code},
			}
			completed, _ := gl.evaluateObjectives(objectiveEvent)
			snapshot := gl.buildSnapshot("Code submitted")
			if len(completed) > 0 {
				snapshot["objectives_completed"] = completed
			}
			return &TickResult{
				Tick:     gl.currentTick,
				State:    gl.state,
				Events:   []GameEvent{ev},
				Snapshot: snapshot,
			}
		}

		return &TickResult{
			Tick:     gl.currentTick,
			State:    gl.state,
			Events:   []GameEvent{ev},
			Snapshot: gl.buildSnapshot("Code submitted"),
		}

	case "mending_protocol":
		entropyVal, _ := gl.fabric.GetState("entropy")
		var entropy int
		switch e := entropyVal.(type) {
		case int:
			entropy = e
		case float64:
			entropy = int(e)
		case float32:
			entropy = int(e)
		}
		newEntropy := entropy + 30
		gl.fabric.SetState("entropy", newEntropy)
		snapshot := gl.buildSnapshot("Mending protocol initiated. System entropy spiked by 30.")
		snapshot["entropy_spike"] = 30
		return &TickResult{
			Tick:     gl.currentTick,
			State:    gl.state,
			Events:   []GameEvent{ev},
			Snapshot: snapshot,
		}

	default:
		return &TickResult{
			Tick:     gl.currentTick,
			State:    gl.state,
			Events:   []GameEvent{ev},
			Snapshot: gl.buildSnapshot(fmt.Sprintf("Unknown action: %s", ev.Action)),
			Errors:   []string{fmt.Sprintf("unknown action: %s", ev.Action)},
		}
	}
}

func (gl *GameLoop) handleSystemEvent(ev GameEvent) *TickResult {
	switch ev.Action {
	case "start_mission":
		if gl.missionManager != nil && gl.missionSession == nil {
			session, err := gl.missionManager.StartMission(gl.playerID, gl.missionID)
			if err != nil {
				return &TickResult{
					Tick:     gl.currentTick,
					State:    gl.state,
					Snapshot: gl.buildSnapshot(fmt.Sprintf("Failed to start mission: %v", err)),
					Errors:   []string{err.Error()},
				}
			}
			gl.missionSession = session
			return &TickResult{
				Tick:     gl.currentTick,
				State:    gl.state,
				Snapshot: gl.buildSnapshot("Mission started"),
			}
		}
	case "fail_mission":
		if gl.missionManager != nil && gl.missionSession != nil {
			_ = gl.missionManager.FailMission(gl.playerID, gl.missionID, ev.Payload)
			gl.state = StateLost
			return &TickResult{
				Tick:     gl.currentTick,
				State:    gl.state,
				Snapshot: gl.buildSnapshot(fmt.Sprintf("Mission failed: %s", ev.Payload)),
			}
		}
	case "move_room":
		target := ev.Payload
		if target == "" {
			target = ev.Payload
		}
		gl.fabric.SetState("current_room", target)
		return &TickResult{
			Tick:     gl.currentTick,
			State:    gl.state,
			Snapshot: gl.buildSnapshot(fmt.Sprintf("Navigated to room: %s", target)),
		}
	case "inspect_object":
		objectID := ev.Payload
		msg := fmt.Sprintf("Inspecting object: %s.", objectID)
		if objectID == "scroll_rack" {
			gl.fabric.SetState("has_rune_shard", true)
			msg += " Found an ancient Rune Shard!"
		}
		return &TickResult{
			Tick:     gl.currentTick,
			State:    gl.state,
			Snapshot: gl.buildSnapshot(msg),
		}
	case "start_game":
		return &TickResult{
			Tick:     gl.currentTick,
			State:    gl.state,
			Snapshot: gl.buildSnapshot("Game started"),
		}
	case "stop_game":
		gl.Stop()
		return &TickResult{
			Tick:     gl.currentTick,
			State:    gl.state,
			Snapshot: gl.buildSnapshot("Game stopped"),
		}
	case "mission_status":
		return &TickResult{
			Tick:          gl.currentTick,
			State:         gl.state,
			Snapshot:      gl.buildSnapshot("Mission status"),
			MissionStatus: gl.getMissionStatus(),
		}
	case "resource_update":
		return &TickResult{
			Tick:     gl.currentTick,
			State:    gl.state,
			Snapshot: gl.buildSnapshot("Resource update"),
		}
	case "telemetry":
		return &TickResult{
			Tick:     gl.currentTick,
			State:    gl.state,
			Snapshot: gl.buildSnapshot("Telemetry data"),
		}
	}
	return nil
}

func (gl *GameLoop) executeTick() TickResult {
	if gl.state == StatePaused {
		return gl.buildTickResult()
	}
	if gl.state != StatePlaying {
		if gl.state == StatePaused {
			return TickResult{
				Tick:     gl.currentTick,
				State:    gl.state,
				Snapshot: gl.buildSnapshot("Game paused"),
			}
		}
		return TickResult{
			Tick:     gl.currentTick,
			State:    gl.state,
			Snapshot: gl.buildSnapshot(fmt.Sprintf("Game state: %s", gl.state)),
		}
	}

	if gl.fabric.IsPurged() {
		gl.state = StatePurged
		snapshot := gl.buildSnapshot("ONTOLOGICAL_PURGE: The Vigilant Archon has terminated execution")
		snapshot["game_over"] = true
		snapshot["game_state"] = "purged"

		gl.replayRecorder.RecordTick(gl.currentTick, gl.state, snapshot)

		if gl.config.EventBus != nil {
			gl.config.EventBus.Publish(eventbus.Event{
				Type:    eventbus.EventPurge,
				Payload: "The Vigilant Archon has terminated execution due to high entropy!",
				Source:  "gameloop",
			})
		}

		gl.syncFabricStateToSession()
		gl.evictFabric()

		return TickResult{
			Tick:     gl.currentTick,
			State:    gl.state,
			Snapshot: snapshot,
		}
	}

	entropyVal, hasEntropy := gl.fabric.GetState("entropy")
	var entropy float64
	if hasEntropy {
		switch e := entropyVal.(type) {
		case int:
			entropy = float64(e)
		case float64:
			entropy = e
		case float32:
			entropy = float64(e)
		}
	}

	if entropy > 0 {
		gl.fabric.ArchonVigilance += entropy * 0.002
		if gl.fabric.ArchonVigilance > 1.0 {
			gl.fabric.ArchonVigilance = 1.0
		}
	}

	if gl.missionSession != nil {
		if gl.missionSession.Status == missionengine.MissionStatusCompleted {
			gl.state = StateWon
			snapshot := gl.buildSnapshot("Mission completed successfully!")
			snapshot["game_state"] = "won"
			snapshot["mission_complete"] = true
			gl.replayRecorder.RecordTick(gl.currentTick, gl.state, snapshot)

			// Sync fabric state to mission session for persistence and evict fabric
			gl.syncFabricStateToSession()
			gl.evictFabric()

			return TickResult{
				Tick:     gl.currentTick,
				State:    gl.state,
				Snapshot: snapshot,
			}
		}

		if gl.missionSession.Status == missionengine.MissionStatusFailed {
			gl.state = StateLost
			reason := "unknown"
			if r, ok := gl.missionSession.FabricState["failure_reason"]; ok {
				reason = fmt.Sprintf("%v", r)
			}
			snapshot := gl.buildSnapshot(fmt.Sprintf("Mission failed: %s", reason))
			snapshot["game_state"] = "lost"
			snapshot["failure_reason"] = reason
			gl.replayRecorder.RecordTick(gl.currentTick, gl.state, snapshot)

			// Sync fabric state to mission session for persistence and evict fabric
			gl.syncFabricStateToSession()
			gl.evictFabric()

			return TickResult{
				Tick:     gl.currentTick,
				State:    gl.state,
				Snapshot: snapshot,
			}
		}
	}

	return gl.buildTickResult()
}

func (gl *GameLoop) buildTickResult() TickResult {
	// Sync fabric state to mission session periodically for persistence
	if gl.currentTick%50 == 0 {
		gl.syncFabricStateToSession()
	}

	message := fmt.Sprintf("Tick %d", gl.currentTick)
	if gl.fabric.ArchonVigilance > 0 {
		message = fmt.Sprintf("Vigilance: %.1f%%", gl.fabric.ArchonVigilance*100)
	}

	if gl.missionSession != nil {
		objectiveEvent := missionengine.ObjectiveEvent{
			Type:   "tick_evaluation",
			Target: "tick",
			Player: gl.playerID,
		}
		completed, _ := gl.evaluateObjectives(objectiveEvent)
		snapshot := gl.buildSnapshot(message)
		if len(completed) > 0 {
			snapshot["objectives_completed"] = completed
			// Sync after objective completion
			gl.syncFabricStateToSession()
		}
		return TickResult{
			Tick:          gl.currentTick,
			State:         gl.state,
			Snapshot:      snapshot,
			MissionStatus: gl.getMissionStatus(),
		}
	}

	return TickResult{
		Tick:     gl.currentTick,
		State:    gl.state,
		Snapshot: gl.buildSnapshot(message),
	}
}

func (gl *GameLoop) evaluateObjectives(event missionengine.ObjectiveEvent) ([]string, error) {
	if gl.missionManager == nil || gl.missionSession == nil {
		return nil, nil
	}
	completed, err := gl.missionManager.ProcessEvent(gl.playerID, event)
	return completed, err
}

func (gl *GameLoop) recordTelemetry(snapshot map[string]interface{}) {
	vigilance := gl.fabric.ArchonVigilance
	entropyVal, _ := gl.fabric.GetState("entropy")
	var entropy float64
	if entropyVal != nil {
		switch e := entropyVal.(type) {
		case int:
			entropy = float64(e)
		case float64:
			entropy = e
		case float32:
			entropy = float64(e)
		}
	}

	metrics := map[string]float64{}
	if gl.missionSession != nil {
		metrics["mission_progress"] = gl.missionSession.GetProgress()
	}
	inventory := gl.resourceMgr.GetInventory(gl.playerID)
	metrics["resources"] = float64(len(inventory.Resources))

	gl.telemetry.Record(gl.currentTick, gl.state, vigilance, entropy, snapshot, metrics)

	if gl.config.EventBus != nil {
		gl.config.EventBus.Publish(eventbus.Event{
			Type: eventbus.EventSnapshot,
			Payload: map[string]interface{}{
				"tick":      gl.currentTick,
				"vigilance": vigilance,
				"entropy":   entropy,
				"state":     gl.state.String(),
			},
			Source: "gameloop",
		})
	}
}

func (gl *GameLoop) buildSnapshot(message string) map[string]interface{} {
	snapshot := map[string]interface{}{
		"tick":        gl.currentTick,
		"game_state":  gl.state.String(),
		"vigilance":   gl.fabric.ArchonVigilance,
		"paradigm":    string(gl.fabric.CurrentParadigm),
		"state":       gl.fabric.GetStateMap(),
		"triggerable": gl.fabric.EvaluateAllGlitches(),
		"message":     message,
	}

	if gl.missionSession != nil {
		snapshot["mission_id"] = gl.missionSession.Mission.ID
		snapshot["mission_title"] = gl.missionSession.Mission.Title
		snapshot["mission_status"] = gl.missionSession.Status.String()
		snapshot["mission_progress"] = gl.missionSession.GetProgress()

		objectives := make([]map[string]interface{}, 0)
		for _, obj := range gl.missionSession.Mission.Objectives {
			status := gl.missionSession.ObjectiveStatuses[obj.ID]
			objectives = append(objectives, map[string]interface{}{
				"id":          obj.ID,
				"type":        obj.Type,
				"target":      obj.Target,
				"description": obj.Description,
				"status":      status.String(),
				"is_required": obj.IsRequired || !obj.IsOptional,
			})
		}
		snapshot["objectives"] = objectives
	}

	inventory := gl.resourceMgr.GetInventory(gl.playerID)
	snapshot["inventory"] = inventory.Resources
	snapshot["replay_frame"] = gl.currentTick

	gl.replayRecorder.RecordTick(gl.currentTick, gl.state, snapshot)

	return snapshot
}

func (gl *GameLoop) getMissionStatus() *MissionStatus {
	if gl.missionSession == nil {
		return nil
	}
	return &MissionStatus{
		ID:       gl.missionSession.Mission.ID,
		Status:   gl.missionSession.Status.String(),
		Progress: gl.missionSession.GetProgress(),
		Won:      gl.state == StateWon,
		Lost:     gl.state == StateLost || gl.state == StatePurged,
	}
}

// syncFabricStateToSession copies the current Fabric state to the MissionSession
// for persistence. This ensures the simulation state is saved.
func (gl *GameLoop) syncFabricStateToSession() {
	if gl.missionSession == nil || gl.fabric == nil {
		return
	}
	// Copy all fabric state to session for persistence
	if gl.missionSession.FabricState == nil {
		gl.missionSession.FabricState = make(map[string]interface{})
	}
	for k, v := range gl.fabric.GetStateMap() {
		gl.missionSession.FabricState[k] = v
	}
}

// evictFabric calls the configured eviction callback when the mission ends.
func (gl *GameLoop) evictFabric() {
	if gl.config.EvictionFn != nil && gl.config.SessionKey != "" {
		gl.config.EvictionFn(gl.config.SessionKey)
	}
}

func (gl *GameLoop) Reset(fabric *engine.AxiomaticFabric, session *missionengine.MissionSession, seed int64) {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	gl.fabric = fabric
	gl.missionSession = session
	gl.resourceMgr = NewResourceManager()
	gl.replayRecorder = NewReplayRecorder(seed)
	gl.telemetry = NewTelemetryCollector(1000)
	gl.currentTick = 0
	gl.state = StatePlaying
	gl.tickResults = nil
}
