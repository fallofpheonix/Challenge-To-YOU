package engine

import (
	"challenge-to-you/backend/internal/sandbox"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// ─── Schema ─────────────────────────────────────────────────────────────────

// SubChallengeRef declares a sub-challenge participation in a composite,
// including which key its result is published under for the combiner.
type SubChallengeRef struct {
	ID        string `json:"id"`
	OutputKey string `json:"output_key"` // key under which this sub-puzzle emits its result
}

// CombineRule describes how sub-challenge results are evaluated together.
// Three deterministic types — no Goja scripting at the combiner layer.
type CombineRule struct {
	// Type is one of: "concat" | "all_state_match" | "pipe"
	Type string `json:"type"`

	// concat: ordered list of output_keys to join as strings, compared against Expected.
	Order    []string `json:"order,omitempty"`
	Expected string   `json:"expected,omitempty"`

	// all_state_match: a map of state_key → required value.
	// All keys must hold their required value after sub-challenges complete.
	RequiredState map[string]interface{} `json:"required_state,omitempty"`

	// pipe: name of the state key the upstream output is injected as
	// into the next sub-challenge's sandbox initial state.
	InjectAs string `json:"inject_as,omitempty"`
}

// CompositeChallenge is a meta-challenge that chains N sub-challenges.
// Dependencies being empty/omitted gives simultaneous (parallel) mode.
// Populating dependencies gives sequential (dependency-gated) mode.
type CompositeChallenge struct {
	ID            string            `json:"id"`
	Paradigm      Paradigm          `json:"paradigm"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	LogosToken    string            `json:"logos_token"`
	SkillType     string            `json:"skill_type"` // always "composite"
	SubChallenges []SubChallengeRef `json:"sub_challenges"`

	// Dependencies maps sub-challenge ID → list of sub-challenge IDs that must
	// complete before it is available. Empty map = all run simultaneously.
	Dependencies map[string][]string `json:"dependencies,omitempty"`

	Combine      CombineRule  `json:"combine"`
	WinCondition WinCondition `json:"win_condition"`
}

// LoadComposite reads and parses a composite challenge JSON file.
func LoadComposite(path string) (*CompositeChallenge, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read composite: %w", err)
	}
	var c CompositeChallenge
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse composite: %w", err)
	}
	if c.ID == "" {
		return nil, fmt.Errorf("composite missing id")
	}
	if c.SkillType != "composite" {
		return nil, fmt.Errorf("composite %s has wrong skill_type=%q, want \"composite\"", c.ID, c.SkillType)
	}
	if len(c.SubChallenges) < 2 {
		return nil, fmt.Errorf("composite %s needs at least 2 sub_challenges", c.ID)
	}
	return &c, nil
}

// ─── Session ─────────────────────────────────────────────────────────────────

// SubResult holds the normalised outcome of one completed sub-challenge.
// Both ExecuteScript (sandbox) and EvaluateAnswer (recognize) normalise into this.
type SubResult struct {
	OutputKey    string                 // which composite output slot this fills
	Value        interface{}            // the emitted value (answer string, state value, etc.)
	MutatedState map[string]interface{} // full mutated state, used by all_state_match
	Complete     bool
}

// CompositeSession tracks partial completion state across sub-challenges
// within one composite run.
type CompositeSession struct {
	composite *CompositeChallenge
	// results holds completed SubResults keyed by sub-challenge ID
	results map[string]*SubResult
	// sharedState is the merged state across all sub-challenges, used by all_state_match
	sharedState map[string]interface{}
}

// NewCompositeSession initialises a fresh session for the given composite.
func NewCompositeSession(c *CompositeChallenge) *CompositeSession {
	return &CompositeSession{
		composite:   c,
		results:     make(map[string]*SubResult),
		sharedState: make(map[string]interface{}),
	}
}

// Available returns the IDs of sub-challenges whose dependency gates are open
// (i.e. all prerequisites are already complete).
func (s *CompositeSession) Available() []string {
	var avail []string
	for _, sc := range s.composite.SubChallenges {
		if _, done := s.results[sc.ID]; done {
			continue // already complete
		}
		deps := s.composite.Dependencies[sc.ID]
		gated := false
		for _, depID := range deps {
			if _, ok := s.results[depID]; !ok {
				gated = true
				break
			}
		}
		if !gated {
			avail = append(avail, sc.ID)
		}
	}
	return avail
}

// Complete returns true once every sub-challenge has a result.
func (s *CompositeSession) Complete() bool {
	return len(s.results) == len(s.composite.SubChallenges)
}

// RecordResult registers a sub-challenge's result into the session.
// It merges the mutated state into the shared state pool used by all_state_match.
func (s *CompositeSession) RecordResult(subID string, result *SubResult) error {
	if _, exists := s.results[subID]; exists {
		return fmt.Errorf("sub-challenge %s already has a result recorded", subID)
	}
	// Verify subID belongs to this composite
	found := false
	for _, sc := range s.composite.SubChallenges {
		if sc.ID == subID {
			result.OutputKey = sc.OutputKey
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("sub-challenge %s is not part of composite %s", subID, s.composite.ID)
	}
	s.results[subID] = result
	for k, v := range result.MutatedState {
		s.sharedState[k] = v
	}
	return nil
}

// PipeContext returns the augmented initial state for a pipe-mode sub-challenge.
// It injects the upstream sub-challenge's output value under the configured key.
// Returns an error if the upstream result is not yet recorded.
func (s *CompositeSession) PipeContext(subID string, baseState map[string]interface{}) (map[string]interface{}, error) {
	if s.composite.Combine.Type != "pipe" {
		return baseState, nil
	}
	deps := s.composite.Dependencies[subID]
	if len(deps) == 0 {
		return baseState, nil
	}
	upstreamID := deps[0] // pipe only uses the direct upstream
	upstream, ok := s.results[upstreamID]
	if !ok {
		return nil, fmt.Errorf("pipe: upstream sub-challenge %s not yet complete", upstreamID)
	}
	augmented := make(map[string]interface{}, len(baseState)+1)
	for k, v := range baseState {
		augmented[k] = v
	}
	augmented[s.composite.Combine.InjectAs] = upstream.Value
	return augmented, nil
}

// ─── Combiner ────────────────────────────────────────────────────────────────

// RunCombiner evaluates the composite's combination rule against the session's
// completed sub-results. Returns (true, nil) on success, (false, err) on failure.
// This is purely deterministic — no Goja sandbox is involved.
func (s *CompositeSession) RunCombiner() (bool, error) {
	if !s.Complete() {
		return false, fmt.Errorf("combiner called before all sub-challenges are complete (%d/%d done)",
			len(s.results), len(s.composite.SubChallenges))
	}

	rule := s.composite.Combine

	switch rule.Type {
	case "concat":
		return s.runConcat(rule)
	case "all_state_match":
		return s.runAllStateMatch(rule)
	case "pipe":
		return s.runPipe(rule)
	default:
		return false, fmt.Errorf("unknown combiner type: %q", rule.Type)
	}
}

// runConcat joins sub-results in the declared order and compares to Expected.
func (s *CompositeSession) runConcat(rule CombineRule) (bool, error) {
	if len(rule.Order) == 0 {
		return false, fmt.Errorf("concat combiner requires non-empty order")
	}
	if rule.Expected == "" {
		return false, fmt.Errorf("concat combiner requires expected")
	}

	// Build a lookup from output_key → result
	byKey := make(map[string]interface{})
	for _, r := range s.results {
		byKey[r.OutputKey] = r.Value
	}

	var parts []string
	for _, key := range rule.Order {
		val, ok := byKey[key]
		if !ok {
			return false, fmt.Errorf("concat: output_key %q not found in sub-results", key)
		}
		parts = append(parts, fmt.Sprintf("%v", val))
	}

	actual := strings.Join(parts, "")
	if actual != rule.Expected {
		return false, fmt.Errorf("concat mismatch: got %q, want %q", actual, rule.Expected)
	}
	return true, nil
}

// runAllStateMatch checks that all required state keys hold their required values.
func (s *CompositeSession) runAllStateMatch(rule CombineRule) (bool, error) {
	if len(rule.RequiredState) == 0 {
		return false, fmt.Errorf("all_state_match combiner requires non-empty required_state")
	}
	for key, requiredVal := range rule.RequiredState {
		actual, ok := s.sharedState[key]
		if !ok {
			return false, fmt.Errorf("all_state_match: state key %q not found in merged state", key)
		}
		if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", requiredVal) {
			return false, fmt.Errorf("all_state_match: key %q = %v, want %v", key, actual, requiredVal)
		}
	}
	return true, nil
}

// runPipe verifies that the final downstream sub-challenge completed successfully.
// The actual data injection was done via PipeContext before ExecuteScript was called —
// so by the time we reach the combiner, the downstream result either exists or doesn't.
func (s *CompositeSession) runPipe(rule CombineRule) (bool, error) {
	if rule.InjectAs == "" {
		return false, fmt.Errorf("pipe combiner requires inject_as")
	}
	// The last sub-challenge in dependency order is the terminal one.
	// It must be complete (guaranteed by Complete() check above).
	// Its mutated state drives the win condition — check it.
	winKey := s.composite.WinCondition.TargetStateKey
	winVal := fmt.Sprintf("%v", s.composite.WinCondition.ExpectedValue)
	actual, ok := s.sharedState[winKey]
	if !ok {
		return false, fmt.Errorf("pipe: win state key %q not found in shared state", winKey)
	}
	if fmt.Sprintf("%v", actual) != winVal {
		return false, fmt.Errorf("pipe: win condition not met: %q = %v, want %v", winKey, actual, winVal)
	}
	return true, nil
}

// ─── Sub-result helpers ───────────────────────────────────────────────────────

// SubResultFromScript builds a SubResult from an ExecuteScript call.
// outputStateKey is the state key whose mutated value becomes the emitted Value.
func SubResultFromScript(outputKey, outputStateKey string, mutatedState map[string]interface{}) *SubResult {
	return &SubResult{
		OutputKey:    outputKey,
		Value:        mutatedState[outputStateKey],
		MutatedState: mutatedState,
		Complete:     true,
	}
}

// SubResultFromAnswer builds a SubResult from an EvaluateAnswer call (recognize).
// The answer string itself is the emitted Value.
func SubResultFromAnswer(outputKey, answer string, mutatedState map[string]interface{}) *SubResult {
	return &SubResult{
		OutputKey:    outputKey,
		Value:        answer,
		MutatedState: mutatedState,
		Complete:     true,
	}
}

// ExecuteSubChallenge is a convenience orchestrator that:
//  1. Loads the sub-challenge definition.
//  2. Routes to EvaluateAnswer (recognize) or ExecuteScript (all other types).
//  3. Returns a SubResult ready to hand to CompositeSession.RecordResult.
func ExecuteSubChallenge(def *ChallengeDefinition, playerInput string, state map[string]interface{}, outputStateKey string) (*SubResult, error) {
	outputKey := "" // will be set by RecordResult from the SubChallengeRef
	switch def.SkillType {
	case "recognize":
		mutated, err := def.EvaluateAnswer(playerInput, state)
		if err != nil {
			return nil, err
		}
		return SubResultFromAnswer(outputKey, playerInput, mutated), nil
	default:
		res, err := def.ExecuteScript(playerInput, state)
		if err != nil {
			return nil, err
		}
		return SubResultFromScript(outputKey, outputStateKey, res.MutatedState), nil
	}
}

// ExecuteSubChallengeWithTimeout is ExecuteSubChallenge with an explicit sandbox timeout.
func ExecuteSubChallengeWithTimeout(def *ChallengeDefinition, playerInput string, state map[string]interface{}, outputStateKey string, timeout time.Duration) (*SubResult, error) {
	if def.SkillType == "recognize" {
		return ExecuteSubChallenge(def, playerInput, state, outputStateKey)
	}
	fullCode := playerInput + "\n" + def.ValidationScript
	res, err := sandbox.Execute(fullCode, state, timeout)
	if err != nil {
		return nil, err
	}
	outputKey := ""
	return SubResultFromScript(outputKey, outputStateKey, res.MutatedState), nil
}
