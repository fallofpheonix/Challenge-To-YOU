package intelligence

import (
	"testing"
)

func TestDependencySelector(t *testing.T) {
	graph := NewRepositoryGraph(".")
	selector := NewDependencySelector(graph)

	tests, err := selector.SelectTargetTests("backend/main.go")
	if err != nil {
		t.Fatalf("failed to select target tests: %v", err)
	}

	if len(tests) == 0 {
		t.Error("expected non-empty test list")
	}
}

func TestSafetyEngine(t *testing.T) {
	se := NewSafetyEngine()

	// Valid patch
	validPatch := "diff --git a/backend/main.go b/backend/main.go\n+ // add comment"
	err := se.ValidatePatchSafety(validPatch)
	if err != nil {
		t.Errorf("expected valid patch to pass safety check, got %v", err)
	}

	// Protected file patch
	protectedPatch := "diff --git a/backend/go.mod b/backend/go.mod\n- go 1.21"
	err = se.ValidatePatchSafety(protectedPatch)
	if err == nil {
		t.Error("expected modification of protected go.mod to cause safety error")
	}

	// Protected API patch
	apiPatch := "diff --git a/backend/main.go b/backend/main.go\n- RaiseLifecycleEvent(event)"
	err = se.ValidatePatchSafety(apiPatch)
	if err == nil {
		t.Error("expected modification of protected RaiseLifecycleEvent to cause safety error")
	}
}
