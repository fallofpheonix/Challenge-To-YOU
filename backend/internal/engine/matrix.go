package engine

import (
	"encoding/json"
	"fmt"
)

// Paradigm represents the thematic universe (Magitech, Cyberpunk, Cosmic)
type Paradigm string

const (
	Magitech  Paradigm = "MAGITECH"
	Cyberpunk Paradigm = "CYBERPUNK"
	Cosmic    Paradigm = "COSMIC"
)

// StructuralValue represents any state primitive within the simulation
type StructuralValue interface{}

// ParadigmState tracks the current metaphysical variables of the level
type ParadigmState map[string]StructuralValue

// AxiomaticEffect defines the transformation applied to the state if conditions clear
type AxiomaticEffect struct {
	TargetStateKey string          `json:"target_state_key"`
	MutationValue  StructuralValue `json:"mutation_value"`
	LogosCipher    string          `json:"logos_cipher,omitempty"` // Spawned passcode if applicable
}

// DemiurgicCondition enforces logical checks before an effect ripples through reality
type DemiurgicCondition struct {
	StateKey string `json:"state_key"`
	Operator string `json:"operator"` // "EQUALS", "GREATER_THAN", "LESS_THAN", "NOT"
	Value    string `json:"value"`
}

// UnsanctionedGlitch defines the emergent path/loophole the player is attempting to force open
type UnsanctionedGlitch struct {
	ID         string               `json:"id"`
	InputEvent string               `json:"input_event"`
	Conditions []DemiurgicCondition `json:"conditions"`
	Effects    []AxiomaticEffect    `json:"effects"`
}

// AxiomaticFabric is our universal logic graph manager
type AxiomaticFabric struct {
	CurrentParadigm Paradigm                       `json:"paradigm"`
	State           ParadigmState                  `json:"state"`
	Glitches        map[string]*UnsanctionedGlitch `json:"glitches"`
	WinConditionKey string                         `json:"win_condition_key"`
	WinConditionVal StructuralValue                `json:"win_condition_val"`
	ArchonVigilance float64                        `json:"archon_vigilance"`
}

// NewAxiomaticFabric instantiates an empty reality-matrix container
func NewAxiomaticFabric(paradigm Paradigm, winKey string, winVal StructuralValue) *AxiomaticFabric {
	return &AxiomaticFabric{
		CurrentParadigm: paradigm,
		State:           make(ParadigmState),
		Glitches:        make(map[string]*UnsanctionedGlitch),
		WinConditionKey: winKey,
		WinConditionVal: winVal,
		ArchonVigilance: 0.0,
	}
}

// RegisterGlitch adds an unscripted solution path to the fabric
func (af *AxiomaticFabric) RegisterGlitch(glitch *UnsanctionedGlitch) {
	af.Glitches[glitch.InputEvent] = glitch
}

// SetState initializes or updates a state variable in the fabric
func (af *AxiomaticFabric) SetState(key string, value StructuralValue) {
	af.State[key] = value
}

// GetState retrieves the current value of a state variable
func (af *AxiomaticFabric) GetState(key string) (StructuralValue, bool) {
	val, exists := af.State[key]
	return val, exists
}

// ToJSON serializes the entire fabric state to JSON (for transmission to Godot)
func (af *AxiomaticFabric) ToJSON() ([]byte, error) {
	return json.Marshal(af)
}

// FromJSON deserializes a fabric state from JSON (for loading saved states)
func (af *AxiomaticFabric) FromJSON(data []byte) error {
	return json.Unmarshal(data, af)
}

// ResetVigilance resets the Archon's detection meter (for testing or level resets)
func (af *AxiomaticFabric) ResetVigilance() {
	af.ArchonVigilance = 0.0
}

// String returns a human-readable representation of the fabric state
func (af *AxiomaticFabric) String() string {
	return fmt.Sprintf("Fabric{Paradigm: %s, State: %v, Glitches: %d, Vigilance: %.2f}",
		af.CurrentParadigm, af.State, len(af.Glitches), af.ArchonVigilance)
}
