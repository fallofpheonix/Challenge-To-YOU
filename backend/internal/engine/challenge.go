package engine

import (
	"challenge-to-you/backend/internal/sandbox"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// BrokenModule is display metadata for a challenge module the player reads.
type BrokenModule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	InputEvent  string `json:"input_event"`
}

// Flaw is a glitch/loophole the player can exploit.
// This maps to UnsanctionedGlitch internally.
type Flaw struct {
	ID                string        `json:"id"`
	TriggerEvent      string        `json:"trigger_event"`
	Name              string        `json:"name"`
	FlavorText        string        `json:"flavor_text"`
	Conditions        ParadigmState `json:"conditions"`
	Mutations         ParadigmState `json:"mutations"`
	FallbackMutations ParadigmState `json:"fallback_mutations"`
}

// WinCondition defines the victory state.
type WinCondition struct {
	TargetStateKey string          `json:"target_state_key"`
	ExpectedValue  StructuralValue `json:"expected_value"`
}

// ChallengeDefinition is a single playable level loaded from JSON.
// Uses the new data-driven format with "flaws" instead of "glitches".
type ChallengeDefinition struct {
	ID               string                 `json:"id"`
	Paradigm         Paradigm               `json:"paradigm"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	LogosToken       string                 `json:"logos_token"`
	ExploitClass     string                 `json:"exploit_class,omitempty"`
	TelemetryHooks   map[string]interface{} `json:"telemetry_hooks,omitempty"`
	InitialState     ParadigmState          `json:"initial_state"`
	Flaws            []Flaw                 `json:"flaws"`
	WinCondition     WinCondition           `json:"win_condition"`
	TemplateCode     string                 `json:"template_code,omitempty"`
	ValidationScript string                 `json:"validation_script,omitempty"`
	SkillType        string                 `json:"skill_type,omitempty"`
	// ExpectedAnswer is used for recognize-type challenges: grading short-circuits
	// to a direct comparison against this value — the sandbox never runs.
	ExpectedAnswer string `json:"expected_answer,omitempty"`
}

// EvaluateAnswer grades a player's submitted answer for recognize-type challenges.
// It does NOT spin up the Goja sandbox — it compares the answer directly against
// ExpectedAnswer, mutates state on match, and returns an error on mismatch.
// Call ExecuteScript instead for optimize / write_from_spec / modify challenges.
func (c *ChallengeDefinition) EvaluateAnswer(answer string, state map[string]interface{}) (map[string]interface{}, error) {
	if c.SkillType != "recognize" {
		return nil, fmt.Errorf("EvaluateAnswer called on non-recognize challenge (type=%q), use ExecuteScript instead", c.SkillType)
	}
	if c.ExpectedAnswer == "" {
		return nil, fmt.Errorf("challenge %s has skill_type=recognize but no expected_answer defined", c.ID)
	}
	if strings.TrimSpace(answer) != strings.TrimSpace(c.ExpectedAnswer) {
		return nil, fmt.Errorf("incorrect answer: got %q, want %q", answer, c.ExpectedAnswer)
	}
	// Clone state and apply the win condition mutation
	result := make(map[string]interface{}, len(state))
	for k, v := range state {
		result[k] = v
	}
	result[c.WinCondition.TargetStateKey] = c.WinCondition.ExpectedValue
	return result, nil
}

// ExecuteScript executes player submitted code concatenated with validation test suite inside the sandbox
func (c *ChallengeDefinition) ExecuteScript(playerCode string, state map[string]interface{}) (*sandbox.Result, error) {
	fullCode := playerCode + "\n" + c.ValidationScript
	return sandbox.Execute(fullCode, state, 2*time.Second)
}

// LoadChallenge reads and parses a challenge JSON file (new format).
func LoadChallenge(path string) (*ChallengeDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read challenge: %w", err)
	}
	var def ChallengeDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("parse challenge: %w", err)
	}
	if def.ID == "" {
		return nil, fmt.Errorf("challenge missing id")
	}
	if len(def.Flaws) == 0 {
		return nil, fmt.Errorf("challenge %s has no flaws", def.ID)
	}
	return &def, nil
}

// BuildFabric hydrates an AxiomaticFabric from this challenge definition.
func (c *ChallengeDefinition) BuildFabric() *AxiomaticFabric {
	fabric := NewAxiomaticFabric(c.Paradigm, c.WinCondition.TargetStateKey, c.WinCondition.ExpectedValue)
	for k, v := range c.InitialState {
		fabric.SetState(k, v)
	}
	for i := range c.Flaws {
		flaw := &c.Flaws[i]
		glitch := &UnsanctionedGlitch{
			ID:              flaw.ID,
			InputEvent:      flaw.TriggerEvent,
			Conditions:      flaw.ConditionsToConditions(),
			Effects:         flaw.MutationsToEffects(),
			FallbackEffects: flaw.FallbackMutationsToEffects(),
		}
		// Attach logos_token to the effect that matches the win condition
		for j := range glitch.Effects {
			if glitch.Effects[j].TargetStateKey == c.WinCondition.TargetStateKey {
				glitch.Effects[j].LogosCipher = c.LogosToken
			}
		}
		fabric.RegisterGlitch(glitch)
	}
	return fabric
}

// ConditionsToConditions converts Flaw.Conditions to DemiurgicCondition slice.
func (f *Flaw) ConditionsToConditions() []DemiurgicCondition {
	var conds []DemiurgicCondition
	for k, v := range f.Conditions {
		valStr := fmt.Sprintf("%v", v)
		operator := "EQUALS"
		value := valStr

		if str, ok := v.(string); ok {
			if strings.HasPrefix(str, "GREATER_THAN:") {
				operator = "GREATER_THAN"
				value = strings.TrimPrefix(str, "GREATER_THAN:")
			} else if strings.HasPrefix(str, "LESS_THAN:") {
				operator = "LESS_THAN"
				value = strings.TrimPrefix(str, "LESS_THAN:")
			} else if strings.HasPrefix(str, "NOT:") {
				operator = "NOT"
				value = strings.TrimPrefix(str, "NOT:")
			} else if strings.HasPrefix(str, "EQUALS:") {
				operator = "EQUALS"
				value = strings.TrimPrefix(str, "EQUALS:")
			}
		}

		conds = append(conds, DemiurgicCondition{
			StateKey: k,
			Operator: operator,
			Value:    value,
		})
	}
	return conds
}

// MutationsToEffects converts Flaw.Mutations to AxiomaticEffect slice.
func (f *Flaw) MutationsToEffects() []AxiomaticEffect {
	var effects []AxiomaticEffect
	for k, v := range f.Mutations {
		effects = append(effects, AxiomaticEffect{
			TargetStateKey: k,
			MutationValue:  v,
		})
	}
	return effects
}

// FallbackMutationsToEffects converts Flaw.FallbackMutations to AxiomaticEffect slice.
func (f *Flaw) FallbackMutationsToEffects() []AxiomaticEffect {
	var effects []AxiomaticEffect
	for k, v := range f.FallbackMutations {
		effects = append(effects, AxiomaticEffect{
			TargetStateKey: k,
			MutationValue:  v,
		})
	}
	return effects
}

// Snapshot is the wire format sent to the Godot client after each action.
type Snapshot struct {
	ChallengeID   string         `json:"challenge_id"`
	Paradigm      string         `json:"paradigm"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	SkillType     string         `json:"skill_type,omitempty"` // "modify" | "recognize" | "optimize" | "write_from_spec" | "composite"
	Modules       []BrokenModule `json:"modules"`
	State         ParadigmState  `json:"state"`
	Vigilance     float64        `json:"vigilance"`
	Triggerable   []string       `json:"triggerable"`
	LastCipher    string         `json:"last_cipher,omitempty"`
	LevelComplete bool           `json:"level_complete"`
	Message       string         `json:"message,omitempty"`
	ErrorMessage  string         `json:"error_message,omitempty"`
}

// NewSnapshot builds a client-facing snapshot from the current fabric state.
func (c *ChallengeDefinition) NewSnapshot(fabric *AxiomaticFabric, lastCipher, message string, levelComplete bool) Snapshot {
	return Snapshot{
		ChallengeID:   c.ID,
		Paradigm:      string(c.Paradigm),
		Title:         c.Name,
		Description:   c.Description,
		SkillType:     c.SkillType,
		Modules:       c.ToModules(),
		State:         fabric.State,
		Vigilance:     fabric.ArchonVigilance,
		Triggerable:   fabric.EvaluateAllGlitches(),
		LastCipher:    lastCipher,
		LevelComplete: levelComplete,
		Message:       message,
	}
}

func (c *ChallengeDefinition) ToModules() []BrokenModule {
	modules := make([]BrokenModule, len(c.Flaws))
	for i, f := range c.Flaws {
		modules[i] = BrokenModule{
			ID:          f.ID,
			Name:        f.Name,
			Description: f.FlavorText,
			InputEvent:  f.TriggerEvent,
		}
	}
	return modules
}
