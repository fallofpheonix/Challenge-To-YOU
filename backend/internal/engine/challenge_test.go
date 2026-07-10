package engine

import (
	"path/filepath"
	"runtime"
	"testing"
)

func challengePath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "challenges", "magitech_01.json")
}

func TestLoadMagitechChallenge(t *testing.T) {
	def, err := LoadChallenge(challengePath())
	if err != nil {
		t.Fatalf("LoadChallenge: %v", err)
	}
	if def.ID != "magitech_01_breach" {
		t.Errorf("expected magitech_01_breach, got %s", def.ID)
	}
	if len(def.Flaws) != 3 {
		t.Errorf("expected 3 flaws, got %d", len(def.Flaws))
	}
	if def.Paradigm != "MAGITECH" {
		t.Errorf("expected MAGITECH paradigm, got %s", def.Paradigm)
	}
	if def.WinCondition.TargetStateKey != "ward_sealed" {
		t.Errorf("expected ward_sealed win condition, got %s", def.WinCondition.TargetStateKey)
	}
}

func TestMagitechChallengeSolutionPath(t *testing.T) {
	def, err := LoadChallenge(challengePath())
	if err != nil {
		t.Fatalf("LoadChallenge: %v", err)
	}
	fabric := def.BuildFabric()

	// Correct sequence: invoke_binding -> surge_mana -> trigger_release
	steps := []string{
		"invoke_binding",
		"surge_mana",
		"trigger_release",
	}
	var cipher string
	var complete bool
	for _, event := range steps {
		cipher, complete, err = fabric.TriggerOntologicalShift(event)
		if err != nil {
			t.Fatalf("step %s failed: %v", event, err)
		}
	}
	if !complete {
		t.Error("expected win after solution path")
	}
	if cipher != "LOGOS_MGT_77F_BREACH" {
		t.Errorf("expected LOGOS_MGT_77F_BREACH, got %s", cipher)
	}
}

func TestMagitechChallengeWrongOrder(t *testing.T) {
	def, err := LoadChallenge(challengePath())
	if err != nil {
		t.Fatalf("LoadChallenge: %v", err)
	}
	fabric := def.BuildFabric()

	// Wrong order: trigger_release first (should fail - conditions not met)
	cipher, complete, err := fabric.TriggerOntologicalShift("trigger_release")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if complete {
		t.Error("release before setup should not win")
	}
	if cipher != "" {
		t.Errorf("expected no cipher, got %s", cipher)
	}

	// Wrong order: surge_mana before binding (should fail - fallback entropy)
	cipher, complete, err = fabric.TriggerOntologicalShift("surge_mana")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if complete {
		t.Error("surge before binding should not win")
	}
	if cipher != "" {
		t.Errorf("expected no cipher, got %s", cipher)
	}
	// Check entropy increased
	entVal := fabric.State["entropy"]
	if entVal == nil {
		t.Errorf("entropy not set")
	} else {
		entFloat, ok := entVal.(float64)
		if !ok || entFloat != 10 {
			t.Errorf("expected entropy 10 from fallback, got %v (type %T)", entVal, entVal)
		}
	}
}