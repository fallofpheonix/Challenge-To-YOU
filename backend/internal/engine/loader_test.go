package engine

import (
	"testing"
)

func TestLoadFabric(t *testing.T) {
	fabric, err := LoadFabric(challengePath())
	if err != nil {
		t.Fatalf("LoadFabric: %v", err)
	}

	if fabric == nil {
		t.Fatal("expected fabric, got nil")
	}

	if fabric.CurrentParadigm != "MAGITECH" {
		t.Errorf("expected MAGITECH paradigm, got %s", fabric.CurrentParadigm)
	}

	if fabric.WinConditionKey != "ward_sealed" {
		t.Errorf("expected win condition key ward_sealed, got %s", fabric.WinConditionKey)
	}

	if len(fabric.Glitches) != 3 {
		t.Errorf("expected 3 glitches, got %d", len(fabric.Glitches))
	}

	// Verify initial state
	if fabric.State["binding_active"] != false {
		t.Errorf("expected binding_active=false, got %v", fabric.State["binding_active"])
	}
	if fabric.State["ward_sealed"] != true {
		t.Errorf("expected ward_sealed=true, got %v", fabric.State["ward_sealed"])
	}

	// Verify triggerable glitches
	triggerable := fabric.EvaluateAllGlitches()
	if len(triggerable) != 1 || triggerable[0] != "invoke_binding" {
		t.Errorf("expected triggerable [invoke_binding], got %v", triggerable)
	}
}

func TestLoadFabricSolutionPath(t *testing.T) {
	fabric, err := LoadFabric(challengePath())
	if err != nil {
		t.Fatalf("LoadFabric: %v", err)
	}

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
