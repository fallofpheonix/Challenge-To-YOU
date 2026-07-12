package engine

import (
	"os"
	"path/filepath"
	"testing"
)

// ─── LoadComposite ────────────────────────────────────────────────────────────

func TestLoadComposite_Concat(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "composite_01_concat.json")
	c, err := LoadComposite(path)
	if err != nil {
		t.Fatalf("LoadComposite failed: %v", err)
	}
	if c.ID != "composite_01_vault_breach" {
		t.Errorf("unexpected id: %s", c.ID)
	}
	if c.Combine.Type != "concat" {
		t.Errorf("expected combiner type concat, got %s", c.Combine.Type)
	}
	if len(c.SubChallenges) != 2 {
		t.Errorf("expected 2 sub_challenges, got %d", len(c.SubChallenges))
	}
}

func TestLoadComposite_RejectsBadSkillType(t *testing.T) {
	// Build a minimal in-memory composite with wrong skill_type — use LoadComposite
	// indirectly by testing the validation path via a bad JSON file we inline.
	badPath := filepath.Join(t.TempDir(), "bad.json")
	badJSON := `{
		"id": "bad_composite",
		"skill_type": "optimize",
		"sub_challenges": [
			{"id": "a", "output_key": "x"},
			{"id": "b", "output_key": "y"}
		],
		"combine": {"type": "concat"},
		"win_condition": {"target_state_key": "done", "expected_value": true}
	}`
	if err := writeTestFile(badPath, badJSON); err != nil {
		t.Fatalf("setup: %v", err)
	}
	_, err := LoadComposite(badPath)
	if err == nil {
		t.Fatal("expected error for wrong skill_type, got nil")
	}
}

func TestLoadComposite_RejectsTooFewSubChallenges(t *testing.T) {
	onePath := filepath.Join(t.TempDir(), "one.json")
	oneJSON := `{
		"id": "one_sub",
		"skill_type": "composite",
		"sub_challenges": [{"id": "a", "output_key": "x"}],
		"combine": {"type": "concat"},
		"win_condition": {"target_state_key": "done", "expected_value": true}
	}`
	if err := writeTestFile(onePath, oneJSON); err != nil {
		t.Fatalf("setup: %v", err)
	}
	_, err := LoadComposite(onePath)
	if err == nil {
		t.Fatal("expected error for single sub_challenge, got nil")
	}
}

// ─── Dependency gating ────────────────────────────────────────────────────────

func TestDependencyGating_SimultaneousMode(t *testing.T) {
	// No dependencies → all sub-challenges available from the start.
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "composite_01_concat.json")
	c, _ := LoadComposite(path)
	session := NewCompositeSession(c)

	avail := session.Available()
	if len(avail) != 2 {
		t.Errorf("expected 2 available sub-challenges in simultaneous mode, got %d: %v", len(avail), avail)
	}
}

func TestDependencyGating_SequentialMode_BlocksDownstream(t *testing.T) {
	// pipe composite: cyberpunk_08_spec depends on cyberpunk_09_recognize.
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "composite_03_pipe.json")
	c, err := LoadComposite(path)
	if err != nil {
		t.Fatalf("LoadComposite: %v", err)
	}
	session := NewCompositeSession(c)

	// Before completing recognize, only recognize should be available.
	avail := session.Available()
	if len(avail) != 1 || avail[0] != "cyberpunk_09_recognize" {
		t.Errorf("expected only cyberpunk_09_recognize available, got %v", avail)
	}

	// Record recognize complete.
	err = session.RecordResult("cyberpunk_09_recognize", &SubResult{
		OutputKey:    "cipher_key",
		Value:        "4",
		MutatedState: map[string]interface{}{"system_firewall_disabled": true},
		Complete:     true,
	})
	if err != nil {
		t.Fatalf("RecordResult: %v", err)
	}

	// Now spec should be unlocked.
	avail = session.Available()
	if len(avail) != 1 || avail[0] != "cyberpunk_08_spec" {
		t.Errorf("expected cyberpunk_08_spec available after upstream completes, got %v", avail)
	}
}

func TestDependencyGating_CompleteOnlyWhenAllDone(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "composite_01_concat.json")
	c, _ := LoadComposite(path)
	session := NewCompositeSession(c)

	if session.Complete() {
		t.Fatal("expected session to be incomplete initially")
	}

	session.RecordResult("cyberpunk_07_optimize", &SubResult{OutputKey: "fib_result", Value: "55", MutatedState: map[string]interface{}{"security_bypass": true}, Complete: true})
	if session.Complete() {
		t.Fatal("expected session incomplete with only one sub-challenge done")
	}

	session.RecordResult("cyberpunk_09_recognize", &SubResult{OutputKey: "vuln_line", Value: "4", MutatedState: map[string]interface{}{"system_firewall_disabled": true}, Complete: true})
	if !session.Complete() {
		t.Fatal("expected session complete after both sub-challenges done")
	}
}

// ─── Concat combiner ──────────────────────────────────────────────────────────

func TestConcatCombiner_Passes(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "composite_01_concat.json")
	c, _ := LoadComposite(path)
	session := NewCompositeSession(c)

	// Optimize emits fib(10)=55, Recognize emits vuln line=4 → "554"
	session.RecordResult("cyberpunk_07_optimize", &SubResult{OutputKey: "fib_result", Value: "55", MutatedState: map[string]interface{}{"security_bypass": true}, Complete: true})
	session.RecordResult("cyberpunk_09_recognize", &SubResult{OutputKey: "vuln_line", Value: "4", MutatedState: map[string]interface{}{"system_firewall_disabled": true}, Complete: true})

	ok, err := session.RunCombiner()
	if !ok || err != nil {
		t.Errorf("expected concat combiner to pass, got ok=%v err=%v", ok, err)
	}
}

func TestConcatCombiner_FailsOnWrongValue(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "composite_01_concat.json")
	c, _ := LoadComposite(path)
	session := NewCompositeSession(c)

	session.RecordResult("cyberpunk_07_optimize", &SubResult{OutputKey: "fib_result", Value: "99", MutatedState: map[string]interface{}{}, Complete: true})
	session.RecordResult("cyberpunk_09_recognize", &SubResult{OutputKey: "vuln_line", Value: "4", MutatedState: map[string]interface{}{}, Complete: true})

	ok, err := session.RunCombiner()
	if ok {
		t.Fatal("expected concat combiner to fail with wrong value, but it passed")
	}
	if err == nil {
		t.Fatal("expected error from concat combiner, got nil")
	}
}

func TestConcatCombiner_FailsIfIncomplete(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "composite_01_concat.json")
	c, _ := LoadComposite(path)
	session := NewCompositeSession(c)

	// Only one sub-challenge recorded
	session.RecordResult("cyberpunk_07_optimize", &SubResult{OutputKey: "fib_result", Value: "55", MutatedState: map[string]interface{}{}, Complete: true})

	_, err := session.RunCombiner()
	if err == nil {
		t.Fatal("expected error when combiner called before all sub-challenges complete")
	}
}

// ─── AllStateMatch combiner ───────────────────────────────────────────────────

func TestAllStateMatchCombiner_Passes(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "composite_02_state.json")
	c, err := LoadComposite(path)
	if err != nil {
		t.Fatalf("LoadComposite: %v", err)
	}
	session := NewCompositeSession(c)

	session.RecordResult("cyberpunk_07_optimize", &SubResult{OutputKey: "optimize_done", Value: true, MutatedState: map[string]interface{}{"security_bypass": true, "cpu_overload": false}, Complete: true})
	session.RecordResult("cyberpunk_08_spec", &SubResult{OutputKey: "spec_done", Value: true, MutatedState: map[string]interface{}{"access_granted": true, "packet_decoded": true}, Complete: true})

	ok, err := session.RunCombiner()
	if !ok || err != nil {
		t.Errorf("expected all_state_match to pass, got ok=%v err=%v", ok, err)
	}
}

func TestAllStateMatchCombiner_FailsOnPartialState(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "composite_02_state.json")
	c, _ := LoadComposite(path)
	session := NewCompositeSession(c)

	// Only optimize done — access_granted still false in shared state
	session.RecordResult("cyberpunk_07_optimize", &SubResult{OutputKey: "optimize_done", Value: true, MutatedState: map[string]interface{}{"security_bypass": true}, Complete: true})
	session.RecordResult("cyberpunk_08_spec", &SubResult{OutputKey: "spec_done", Value: true, MutatedState: map[string]interface{}{"access_granted": false}, Complete: true})

	ok, err := session.RunCombiner()
	if ok {
		t.Fatal("expected all_state_match to fail when access_granted=false")
	}
	if err == nil {
		t.Fatal("expected error from all_state_match combiner")
	}
}

// ─── Pipe combiner ────────────────────────────────────────────────────────────

func TestPipeCombiner_InjectsUpstreamOutput(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "composite_03_pipe.json")
	c, err := LoadComposite(path)
	if err != nil {
		t.Fatalf("LoadComposite: %v", err)
	}
	session := NewCompositeSession(c)

	// Complete upstream (recognize)
	session.RecordResult("cyberpunk_09_recognize", &SubResult{
		OutputKey:    "cipher_key",
		Value:        "4",
		MutatedState: map[string]interface{}{"system_firewall_disabled": true},
		Complete:     true,
	})

	// Verify PipeContext injects correctly
	baseState := map[string]interface{}{"packet_decoded": false, "access_granted": false}
	augmented, err := session.PipeContext("cyberpunk_08_spec", baseState)
	if err != nil {
		t.Fatalf("PipeContext returned error: %v", err)
	}
	if augmented["input_data"] != "4" {
		t.Errorf("expected input_data=4 injected into downstream context, got %v", augmented["input_data"])
	}
	// Base state keys must also be preserved
	if augmented["access_granted"] != false {
		t.Errorf("PipeContext should preserve base state keys")
	}
}

func TestPipeCombiner_BlocksContextBeforeUpstreamComplete(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "composite_03_pipe.json")
	c, _ := LoadComposite(path)
	session := NewCompositeSession(c)

	// Upstream not yet complete — PipeContext must return error
	baseState := map[string]interface{}{}
	_, err := session.PipeContext("cyberpunk_08_spec", baseState)
	if err == nil {
		t.Fatal("expected PipeContext to fail when upstream is not yet complete")
	}
}

func TestPipeCombiner_PassesWhenDownstreamSucceeds(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "composite_03_pipe.json")
	c, _ := LoadComposite(path)
	session := NewCompositeSession(c)

	session.RecordResult("cyberpunk_09_recognize", &SubResult{
		OutputKey:    "cipher_key",
		Value:        "4",
		MutatedState: map[string]interface{}{"system_firewall_disabled": true},
		Complete:     true,
	})
	session.RecordResult("cyberpunk_08_spec", &SubResult{
		OutputKey:    "decoded_data",
		Value:        "AAABBCCCC",
		MutatedState: map[string]interface{}{"access_granted": true, "packet_decoded": true},
		Complete:     true,
	})

	ok, err := session.RunCombiner()
	if !ok || err != nil {
		t.Errorf("expected pipe combiner to pass, got ok=%v err=%v", ok, err)
	}
}

func TestPipeCombiner_FailsWhenWinConditionNotMet(t *testing.T) {
	path := filepath.Join("..", "..", "challenges", "cyberpunk_tier1", "composite_03_pipe.json")
	c, _ := LoadComposite(path)
	session := NewCompositeSession(c)

	session.RecordResult("cyberpunk_09_recognize", &SubResult{
		OutputKey: "cipher_key", Value: "4",
		MutatedState: map[string]interface{}{"system_firewall_disabled": true}, Complete: true,
	})
	// Downstream spec fails — access_granted stays false
	session.RecordResult("cyberpunk_08_spec", &SubResult{
		OutputKey: "decoded_data", Value: "",
		MutatedState: map[string]interface{}{"access_granted": false, "packet_decoded": false}, Complete: true,
	})

	ok, err := session.RunCombiner()
	if ok {
		t.Fatal("expected pipe combiner to fail when downstream win condition not met")
	}
	if err == nil {
		t.Fatal("expected error from pipe combiner")
	}
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
