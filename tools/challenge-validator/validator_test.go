package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChallengeValidator(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "challenge_validator_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create valid challenge
	validChal := `{
		"id": "cosmic_01_airlock",
		"paradigm": "COSMIC",
		"name": "The Quantum Airlock",
		"description": "Open stuck airlock",
		"initial_state": {"airlock_open": false},
		"flaws": [{
			"id": "trigger_siren",
			"trigger_event": "trigger_siren",
			"name": "Trigger siren"
		}],
		"win_condition": {
			"target_state_key": "airlock_open",
			"expected_value": true
		}
	}`
	_ = os.WriteFile(filepath.Join(tempDir, "chal1.json"), []byte(validChal), 0644)

	// Create valid pack
	validPack := `{
		"version": 1,
		"era": "cosmic",
		"tier": 1,
		"name": "Superposition Sabotage",
		"challenges": [{
			"id": "cosmic_01_airlock",
			"file": "chal1.json",
			"difficulty": 0.5
		}]
	}`
	_ = os.WriteFile(filepath.Join(tempDir, "pack.json"), []byte(validPack), 0644)

	validator := NewChallengeValidator(tempDir)
	report, errScan := validator.Validate()
	if errScan != nil {
		t.Fatalf("Validate failed: %v", errScan)
	}

	if !report.Valid {
		t.Errorf("expected valid challenge configuration to pass, got errors: %v", report.Errors)
	}
}
