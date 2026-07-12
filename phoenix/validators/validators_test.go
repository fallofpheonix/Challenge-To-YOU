package validators

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateChallengeContent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "phoenix_validators_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	engine := NewValidationEngine(tempDir)

	// Validate empty directory should pass (no files)
	err = engine.ValidateChallengeContent()
	if err != nil {
		t.Errorf("expected no error on empty challenges dir, got %v", err)
	}

	// Create valid challenge
	chalDir := filepath.Join(tempDir, "challenges")
	_ = os.MkdirAll(chalDir, 0755)
	validChal := `{"id": "M-01", "skill_type": "write_from_spec"}`
	_ = os.WriteFile(filepath.Join(chalDir, "chal1.json"), []byte(validChal), 0644)

	err = engine.ValidateChallengeContent()
	if err != nil {
		t.Errorf("expected valid challenge to pass validation, got %v", err)
	}

	// Create duplicate ID challenge
	dupChal := `{"id": "M-01", "skill_type": "optimize"}`
	_ = os.WriteFile(filepath.Join(chalDir, "chal2.json"), []byte(dupChal), 0644)

	err = engine.ValidateChallengeContent()
	if err == nil {
		t.Error("expected duplicate challenge ID to cause validation error")
	}
}
