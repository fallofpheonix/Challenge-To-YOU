package missionengine

import "fmt"

// ObjectiveEngine handles objective evaluation and tracking
type ObjectiveEngine struct {
	evaluators map[string]ObjectiveEvaluator
}

// ObjectiveEvaluator evaluates whether an objective is met
type ObjectiveEvaluator interface {
	Evaluate(session *MissionSession, objective *Objective, event ObjectiveEvent) bool
}

// ObjectiveEvent represents an event that might complete an objective
type ObjectiveEvent struct {
	Type   string                 // "challenge_completed", "item_found", "npc_talked", "area_explored", "trigger"
	Target string                 // The target ID (challenge_id, item_id, etc.)
	Player string                 // Player ID
	Data   map[string]interface{} // Additional data
}

// NewObjectiveEngine creates a new objective engine
func NewObjectiveEngine() *ObjectiveEngine {
	engine := &ObjectiveEngine{
		evaluators: make(map[string]ObjectiveEvaluator),
	}

	// Register built-in evaluators
	engine.RegisterEvaluator("complete_challenge", &ChallengeCompleteEvaluator{})
	engine.RegisterEvaluator("find_item", &ItemFoundEvaluator{})
	engine.RegisterEvaluator("talk_to_npc", &NPCTalkEvaluator{})
	engine.RegisterEvaluator("explore_area", &AreaExploreEvaluator{})
	engine.RegisterEvaluator("trigger_event", &TriggerEventEvaluator{})
	engine.RegisterEvaluator("state_match", &StateMatchEvaluator{})

	return engine
}

// RegisterEvaluator adds a custom evaluator
func (oe *ObjectiveEngine) RegisterEvaluator(objectiveType string, evaluator ObjectiveEvaluator) {
	oe.evaluators[objectiveType] = evaluator
}

// EvaluateObjective checks if an objective is completed by an event
func (oe *ObjectiveEngine) EvaluateObjective(session *MissionSession, objective *Objective, event ObjectiveEvent) bool {
	// Check if objective is already completed
	status := session.ObjectiveStatuses[objective.ID]
	if status == ObjectiveStatusCompleted {
		return false
	}

	// Check if objective type matches event type
	if objective.Type != event.Type {
		return false
	}

	// Get evaluator
	evaluator, ok := oe.evaluators[objective.Type]
	if !ok {
		// Default: check if target matches
		return objective.Target == event.Target
	}

	return evaluator.Evaluate(session, objective, event)
}

// ChallengeCompleteEvaluator evaluates challenge completion objectives
type ChallengeCompleteEvaluator struct{}

func (e *ChallengeCompleteEvaluator) Evaluate(session *MissionSession, objective *Objective, event ObjectiveEvent) bool {
	return objective.Target == event.Target
}

// ItemFoundEvaluator evaluates item discovery objectives
type ItemFoundEvaluator struct{}

func (e *ItemFoundEvaluator) Evaluate(session *MissionSession, objective *Objective, event ObjectiveEvent) bool {
	return objective.Target == event.Target
}

// NPCTalkEvaluator evaluates NPC conversation objectives
type NPCTalkEvaluator struct{}

func (e *NPCTalkEvaluator) Evaluate(session *MissionSession, objective *Objective, event ObjectiveEvent) bool {
	return objective.Target == event.Target
}

// AreaExploreEvaluator evaluates area exploration objectives
type AreaExploreEvaluator struct{}

func (e *AreaExploreEvaluator) Evaluate(session *MissionSession, objective *Objective, event ObjectiveEvent) bool {
	return objective.Target == event.Target
}

// TriggerEventEvaluator evaluates custom trigger objectives
type TriggerEventEvaluator struct{}

func (e *TriggerEventEvaluator) Evaluate(session *MissionSession, objective *Objective, event ObjectiveEvent) bool {
	return objective.Target == event.Target
}

// StateMatchEvaluator evaluates objectives based on state
type StateMatchEvaluator struct{}

func (e *StateMatchEvaluator) Evaluate(session *MissionSession, objective *Objective, event ObjectiveEvent) bool {
	for key, requiredValue := range objective.RequiresState {
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

// ProcessEvent processes an objective event against all active objectives
func (oe *ObjectiveEngine) ProcessEvent(session *MissionSession, event ObjectiveEvent) []string {
	var completedIDs []string

	for _, objective := range session.Mission.Objectives {
		status := session.ObjectiveStatuses[objective.ID]
		if status != ObjectiveStatusPending && status != ObjectiveStatusActive {
			continue
		}

		if oe.EvaluateObjective(session, &objective, event) {
			completedIDs = append(completedIDs, objective.ID)
		}
	}

	return completedIDs
}

// GetObjectiveProgress returns progress info for an objective
func (oe *ObjectiveEngine) GetObjectiveProgress(session *MissionSession, objectiveID string) map[string]interface{} {
	for _, obj := range session.Mission.Objectives {
		if obj.ID == objectiveID {
			status := session.ObjectiveStatuses[objectiveID]
			return map[string]interface{}{
				"id":          obj.ID,
				"type":        obj.Type,
				"target":      obj.Target,
				"description": obj.Description,
				"status":      status.String(),
				"is_optional": obj.IsOptional,
				"is_required": obj.IsRequired,
			}
		}
	}
	return nil
}

// GetRequiredObjectives returns all required objectives
func (oe *ObjectiveEngine) GetRequiredObjectives(mission *Mission) []Objective {
	var required []Objective
	for _, obj := range mission.Objectives {
		if obj.IsRequired || !obj.IsOptional {
			required = append(required, obj)
		}
	}
	return required
}

// GetOptionalObjectives returns all optional objectives
func (oe *ObjectiveEngine) GetOptionalObjectives(mission *Mission) []Objective {
	var optional []Objective
	for _, obj := range mission.Objectives {
		if obj.IsOptional {
			optional = append(optional, obj)
		}
	}
	return optional
}
