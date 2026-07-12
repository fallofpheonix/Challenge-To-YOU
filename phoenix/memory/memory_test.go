package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMemoryManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "phoenix_memory_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mgr := NewMemoryManager(tempDir)

	failure := &FailureRecord{
		Category:   "compilation",
		Diagnostic: "undefined: eventbus.Register",
		StackTrace: "main.go:20: undefined symbol",
	}

	err = mgr.SaveFailure(failure)
	if err != nil {
		t.Fatalf("failed to save failure: %v", err)
	}

	// Verify build_failures has file
	files, errRead := os.ReadDir(filepath.Join(tempDir, "build_failures"))
	if errRead != nil || len(files) != 1 {
		t.Errorf("expected 1 failure record, got %d", len(files))
	}

	repair := &RepairRecord{
		AssociatedFailure: "undefined: eventbus.Register",
		AppliedPatch:      "diff --git ...",
		TestsPassed:       true,
		Keywords:          []string{"undefined", "eventbus"},
	}

	err = mgr.SaveRepair(repair, true)
	if err != nil {
		t.Fatalf("failed to save repair: %v", err)
	}

	// Test SearchSimilar
	matches, errSearch := mgr.SearchSimilar("error occurred: undefined symbol eventbus")
	if errSearch != nil {
		t.Fatalf("SearchSimilar failed: %v", errSearch)
	}
	if len(matches) != 1 {
		t.Errorf("expected 1 search match, got %d", len(matches))
	}
}
