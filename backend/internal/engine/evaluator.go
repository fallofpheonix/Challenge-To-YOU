package engine

import (
	"fmt"
	"strconv"
)

const vigilanceThreshold = 1.0 - 1e-9 // Tolerance for floating point comparison

// TriggerOntologicalShift forces an event through the fabric to assess state mutations
// Returns: (passcode, levelComplete, error)
func (af *AxiomaticFabric) TriggerOntologicalShift(eventID string) (string, bool, error) {
	af.ArchonVigilance += 0.10 // System friction baseline increase
	if af.ArchonVigilance >= vigilanceThreshold {
		return "", false, fmt.Errorf("ONTOLOGICAL_PURGE: The Vigilant Archon has terminated execution")
	}

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
		}
	}

	return generatedCipher, levelComplete, nil
}

// evaluateCondition checks if a single condition is satisfied against current state
func (af *AxiomaticFabric) evaluateCondition(cond DemiurgicCondition) bool {
	currentVal, exists := af.State[cond.StateKey]
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
