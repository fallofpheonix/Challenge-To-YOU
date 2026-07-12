package intelligence

import (
	"testing"
)

func TestRepairPlanner(t *testing.T) {
	planner := NewRepairPlanner()
	_, prompt, err := planner.GeneratePlan("undefined variables", "main.go")
	if err != nil {
		t.Fatalf("failed to generate plan: %v", err)
	}

	if len(prompt) == 0 {
		t.Error("expected non-empty prompt template")
	}

	mockJSON := `
Some chat wrapper prefix...
{
  "problem": "compiler error",
  "likely_root_causes": ["syntax"],
  "affected_files": ["main.go"],
  "tests": ["go test"],
  "risk": "LOW",
  "repair_plan": ["fix syntax"]
}
Some chat wrapper suffix...
`
	plan, errParse := planner.ParsePlanJSON(mockJSON)
	if errParse != nil {
		t.Fatalf("failed to parse JSON plan: %v", errParse)
	}

	if plan.Problem != "compiler error" {
		t.Errorf("expected problem 'compiler error', got %s", plan.Problem)
	}
}
