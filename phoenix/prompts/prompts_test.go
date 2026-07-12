package prompts

import (
	"testing"
)

func TestExtractCodeBlock(t *testing.T) {
	text := "Some introductory text\n```go\nfunc main() {}\n```\nSome footer text"
	expected := "func main() {}"
	result := extractCodeBlock(text)
	if result != expected {
		t.Errorf("expected extracted code to be %q, got %q", expected, result)
	}
}

func TestLoadPromptsDefault(t *testing.T) {
	lib := LoadPrompts("/nonexistent_dir")
	if lib.PlannerPrompt == "" {
		t.Error("expected non-empty default planner prompt")
	}
	if lib.RepairPrompt == "" {
		t.Error("expected non-empty default repair prompt")
	}
}
