package engine

import (
	"testing"
)

func TestMagitechToCyberpunkConfluence(t *testing.T) {
	// Level Goal: Destabilize the reality shield (reality_stable == false)
	fabric := NewAxiomaticFabric(Magitech, "reality_stable", false)

	// Set baseline environment variables
	fabric.State["ambient_humidity"] = 85.0
	fabric.State["reality_stable"] = true

	// Craft a glitch rule representing our "Frankenstein" layer exploit
	// Attaching an Ice Rune near an unstable heat source collapses the field
	glitch := &UnsanctionedGlitch{
		ID:         "GLITCH_ICE_FIRE_COLLISION",
		InputEvent: "TRIGGER_ICE_RUNE",
		Conditions: []DemiurgicCondition{
			{StateKey: "ambient_humidity", Operator: "GREATER_THAN", Value: "80"},
		},
		Effects: []AxiomaticEffect{
			{TargetStateKey: "reality_stable", MutationValue: false, LogosCipher: "GLITCH_FOUND_EIDOLON_BLINDED"},
		},
	}
	fabric.Glitches[glitch.InputEvent] = glitch

	// Trigger execution frame
	cipher, complete, err := fabric.TriggerOntologicalShift("TRIGGER_ICE_RUNE")
	if err != nil {
		t.Fatalf("Engine failure during reality evaluation: %v", err)
	}

	if !complete {
		t.Errorf("Expected win condition met (reality_stable == false), logic engine failed mutation validation")
	}

	if cipher != "GLITCH_FOUND_EIDOLON_BLINDED" {
		t.Errorf("Failed to derive the proper Logos Token from the confluence sequence, got: %s", cipher)
	}
}

func TestArchonVigilanceIncreases(t *testing.T) {
	fabric := NewAxiomaticFabric(Cyberpunk, "system_compromised", true)
	fabric.State["firewall_active"] = true

	// Register a valid glitch
	glitch := &UnsanctionedGlitch{
		ID:         "GLITCH_BUFFER_OVERFLOW",
		InputEvent: "TRIGGER_BUFFER_OVERFLOW",
		Conditions: []DemiurgicCondition{
			{StateKey: "firewall_active", Operator: "EQUALS", Value: "true"},
		},
		Effects: []AxiomaticEffect{
			{TargetStateKey: "system_compromised", MutationValue: true},
		},
	}
	fabric.RegisterGlitch(glitch)

	// Initial vigilance should be 0
	if fabric.ArchonVigilance != 0.0 {
		t.Errorf("Expected initial vigilance 0.0, got %f", fabric.ArchonVigilance)
	}

	// Trigger multiple times
	for i := 0; i < 9; i++ {
		_, _, err := fabric.TriggerOntologicalShift("TRIGGER_BUFFER_OVERFLOW")
		if err != nil {
			t.Fatalf("Unexpected error on trigger %d: %v", i+1, err)
		}
	}

	// After 9 triggers, vigilance should be 0.9 (within floating point tolerance)
	if fabric.ArchonVigilance < 0.89 || fabric.ArchonVigilance > 0.91 {
		t.Errorf("Expected vigilance ~0.9 after 9 triggers, got %f", fabric.ArchonVigilance)
	}

	// 10th trigger should cause purge
	_, _, err := fabric.TriggerOntologicalShift("TRIGGER_BUFFER_OVERFLOW")
	if err == nil {
		t.Error("Expected purge error on 10th trigger, got nil")
	}

	if !fabric.IsPurged() {
		t.Error("Expected fabric to be purged after 10 triggers")
	}
}

func TestConditionOperators(t *testing.T) {
	fabric := NewAxiomaticFabric(Magitech, "test_win", true)

	// Test EQUALS
	fabric.SetState("test_key", "hello")
	cond := DemiurgicCondition{StateKey: "test_key", Operator: "EQUALS", Value: "hello"}
	if !fabric.evaluateCondition(cond) {
		t.Error("EQUALS condition failed")
	}

	// Test NOT
	cond = DemiurgicCondition{StateKey: "test_key", Operator: "NOT", Value: "world"}
	if !fabric.evaluateCondition(cond) {
		t.Error("NOT condition failed")
	}

	// Test GREATER_THAN
	fabric.SetState("numeric_key", 100.0)
	cond = DemiurgicCondition{StateKey: "numeric_key", Operator: "GREATER_THAN", Value: "50"}
	if !fabric.evaluateCondition(cond) {
		t.Error("GREATER_THAN condition failed")
	}

	// Test LESS_THAN
	cond = DemiurgicCondition{StateKey: "numeric_key", Operator: "LESS_THAN", Value: "150"}
	if !fabric.evaluateCondition(cond) {
		t.Error("LESS_THAN condition failed")
	}
}

func TestWinConditionCheck(t *testing.T) {
	fabric := NewAxiomaticFabric(Cosmic, "portal_open", true)

	// Initially not won
	if fabric.CheckWinCondition() {
		t.Error("Expected win condition not met initially")
	}

	// Set win state
	fabric.SetState("portal_open", true)
	if !fabric.CheckWinCondition() {
		t.Error("Expected win condition met after setting state")
	}
}

func TestEvaluateAllGlitches(t *testing.T) {
	fabric := NewAxiomaticFabric(Magitech, "test", "win")

	fabric.SetState("power_level", 50.0)
	fabric.SetState("mana_available", 30.0)

	// Glitch 1: Only power check
	glitch1 := &UnsanctionedGlitch{
		ID:         "GLITCH_1",
		InputEvent: "EVENT_1",
		Conditions: []DemiurgicCondition{
			{StateKey: "power_level", Operator: "GREATER_THAN", Value: "40"},
		},
	}
	fabric.RegisterGlitch(glitch1)

	// Glitch 2: Both checks
	glitch2 := &UnsanctionedGlitch{
		ID:         "GLITCH_2",
		InputEvent: "EVENT_2",
		Conditions: []DemiurgicCondition{
			{StateKey: "power_level", Operator: "GREATER_THAN", Value: "40"},
			{StateKey: "mana_available", Operator: "GREATER_THAN", Value: "25"},
		},
	}
	fabric.RegisterGlitch(glitch2)

	// Glitch 3: Fails power check
	glitch3 := &UnsanctionedGlitch{
		ID:         "GLITCH_3",
		InputEvent: "EVENT_3",
		Conditions: []DemiurgicCondition{
			{StateKey: "power_level", Operator: "GREATER_THAN", Value: "60"},
		},
	}
	fabric.RegisterGlitch(glitch3)

	triggerable := fabric.EvaluateAllGlitches()

	// Should be triggerable: EVENT_1, EVENT_2 (but not EVENT_3)
	if len(triggerable) != 2 {
		t.Errorf("Expected 2 triggerable glitches, got %d: %v", len(triggerable), triggerable)
	}
}

func TestJSONSerialization(t *testing.T) {
	fabric := NewAxiomaticFabric(Magitech, "test", "pass")
	fabric.SetState("key", "value")

	data, err := fabric.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize fabric: %v", err)
	}

	// Deserialize into new fabric
	fabric2 := NewAxiomaticFabric(Cyberpunk, "", nil)
	err = fabric2.FromJSON(data)
	if err != nil {
		t.Fatalf("Failed to deserialize fabric: %v", err)
	}

	if fabric2.CurrentParadigm != Magitech {
		t.Errorf("Expected paradigm MAGITECH, got %s", fabric2.CurrentParadigm)
	}

	val, exists := fabric2.GetState("key")
	if !exists || val != "value" {
		t.Errorf("Expected state key=value, got %v", val)
	}
}
