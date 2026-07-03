package simulation

import (
	"testing"
)

func TestDecisionFrameIsValueType(t *testing.T) {
	// DecisionFrame should be constructible as a plain value — no required methods.
	frame := DecisionFrame{
		DroneID: 42,
		Tick:    100,
		Steps: []DecisionStep{
			{Kind: "action", Name: "HARVEST", Result: "true"},
			{Kind: "condition", Name: "SENSE_RESOURCE()", Result: "true", Taken: true},
		},
	}

	if frame.DroneID != 42 || frame.Tick != 100 {
		t.Errorf("unexpected frame metadata: %+v", frame)
	}
	if len(frame.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(frame.Steps))
	}
	if frame.Steps[0].Kind != "action" {
		t.Errorf("first step should be action, got %s", frame.Steps[0].Kind)
	}
	if !frame.Steps[1].Taken {
		t.Error("second step should be taken=true")
	}
}

func TestDecisionStepTakenOnlyMeaningfulForConditions(t *testing.T) {
	// Action steps should have Taken=false (zero value) by convention.
	action := DecisionStep{Kind: "action", Name: "HARVEST", Result: "true"}
	if action.Taken {
		t.Error("action step should default Taken=false")
	}
}
