package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const vigilanceThreshold = 1.0 - 1e-9 // Tolerance for floating point comparison

// TriggerOntologicalShift forces an event through the fabric to assess state mutations
// Returns: (passcode, levelComplete, error)
func (af *AxiomaticFabric) TriggerOntologicalShift(eventID string) (string, bool, error) {
	return af.TriggerOntologicalShiftWithPayload(eventID, "")
}

// TriggerOntologicalShiftWithPayload processes an event with dynamic string input for payload-based exploits
func (af *AxiomaticFabric) TriggerOntologicalShiftWithPayload(eventID string, payload string) (string, bool, error) {
	af.MoveCount++
	af.ArchonVigilance += 0.10 // System friction baseline increase
	if af.ArchonVigilance >= vigilanceThreshold {
		return "", false, fmt.Errorf("ONTOLOGICAL_PURGE: The Vigilant Archon has terminated execution")
	}

	// Set dynamic input payload variables
	af.State["input_payload_len"] = len(payload)
	af.State["input_payload"] = payload

	if af.StartTime.IsZero() {
		af.StartTime = time.Now()
	}

	// Calculate time since last event if LastEventTime is set
	if !af.LastEventTime.IsZero() {
		af.State["time_since_last_event_ms"] = float64(time.Since(af.LastEventTime).Milliseconds())
	} else {
		af.State["time_since_last_event_ms"] = 0.0
	}
	af.LastEventTime = time.Now()

	glitch, exists := af.Glitches[eventID]
	if !exists {
		return "", false, nil // Event dissipates harmlessly into ambient chaos
	}

	// Validate conditions
	allConditionsMet := true
	for _, cond := range glitch.Conditions {
		if !af.evaluateCondition(cond) {
			allConditionsMet = false
			break
		}
	}

	var generatedCipher string

	if allConditionsMet {
		// Apply primary effects
		for _, effect := range glitch.Effects {
			af.State[effect.TargetStateKey] = effect.MutationValue
			if effect.LogosCipher != "" {
				generatedCipher = effect.LogosCipher
			}
		}
	} else if len(glitch.FallbackEffects) > 0 {
		// Apply fallback effects (entropy, penalties, etc.)
		for _, effect := range glitch.FallbackEffects {
			af.State[effect.TargetStateKey] = effect.MutationValue
		}
	}

	// Check if level win state is achieved
	levelComplete := false
	if currentVal, tracked := af.State[af.WinConditionKey]; tracked {
		if fmt.Sprintf("%v", currentVal) == fmt.Sprintf("%v", af.WinConditionVal) {
			levelComplete = true
			// WallClockTimeSecs measures total real-world duration, primarily a UX/scoring metric.
			af.WallClockTimeSecs = time.Since(af.StartTime).Seconds()
		}
	}

	return generatedCipher, levelComplete, nil
}

// evaluateCondition checks if a single condition is satisfied against current state
func (af *AxiomaticFabric) evaluateCondition(cond DemiurgicCondition) bool {
	stateKey := cond.StateKey
	if strings.HasPrefix(stateKey, "time_since_last_event_ms") {
		stateKey = "time_since_last_event_ms"
	}
	if strings.HasPrefix(stateKey, "input_payload_len") {
		stateKey = "input_payload_len"
	}

	currentVal, exists := af.State[stateKey]
	if !exists {
		return false
	}

	currStr := fmt.Sprintf("%v", currentVal)

	switch cond.Operator {
	case "EQUALS":
		return currStr == cond.Value
	case "NOT":
		return currStr != cond.Value
	case "GREATER_THAN":
		cNum, err1 := strconv.ParseFloat(currStr, 64)
		vNum, err2 := strconv.ParseFloat(cond.Value, 64)
		if err1 == nil && err2 == nil {
			return cNum > vNum
		}
	case "LESS_THAN":
		cNum, err1 := strconv.ParseFloat(currStr, 64)
		vNum, err2 := strconv.ParseFloat(cond.Value, 64)
		if err1 == nil && err2 == nil {
			return cNum < vNum
		}
	}
	return false
}

// EvaluateAllGlitches checks which glitches could be triggered given current state
// Returns list of triggerable event IDs
func (af *AxiomaticFabric) EvaluateAllGlitches() []string {
	var triggerable []string

	for eventID, glitch := range af.Glitches {
		allConditionsMet := true
		for _, cond := range glitch.Conditions {
			if !af.evaluateCondition(cond) {
				allConditionsMet = false
				break
			}
		}
		if allConditionsMet {
			triggerable = append(triggerable, eventID)
		}
	}

	return triggerable
}

// CheckWinCondition verifies if the current state satisfies the win condition
func (af *AxiomaticFabric) CheckWinCondition() bool {
	currentVal, tracked := af.State[af.WinConditionKey]
	if !tracked {
		return false
	}
	return fmt.Sprintf("%v", currentVal) == fmt.Sprintf("%v", af.WinConditionVal)
}

// GetArchonStatus returns the current vigilance level as a percentage
func (af *AxiomaticFabric) GetArchonStatus() float64 {
	return af.ArchonVigilance * 100
}

// IsPurged returns true if the Archon has terminated execution
func (af *AxiomaticFabric) IsPurged() bool {
	return af.ArchonVigilance >= vigilanceThreshold
}
