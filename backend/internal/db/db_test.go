package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteIntegration(t *testing.T) {
	// Create a temp database path
	tempDir, err := os.MkdirTemp("", "challenge_db_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test_challenge.db")

	// 1. Test InitDB
	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer CloseDB()

	// 2. Test GetOrCreateProfile defaults
	profile, err := GetOrCreateProfile()
	if err != nil {
		t.Fatalf("GetOrCreateProfile failed: %v", err)
	}
	if profile.Name != "Intruder" {
		t.Errorf("Expected profile name 'Intruder', got: %s", profile.Name)
	}
	if profile.Luck != 1.0 {
		t.Errorf("Expected profile luck 1.0, got: %f", profile.Luck)
	}
	if profile.Reputation != 0 {
		t.Errorf("Expected profile reputation 0, got: %d", profile.Reputation)
	}
	if !profile.IsParadigmUnlocked("MAGITECH") {
		t.Error("Expected MAGITECH to be unlocked by default")
	}
	if profile.IsParadigmUnlocked("CYBERPUNK") {
		t.Error("Expected CYBERPUNK to be locked by default")
	}

	// 3. Test SaveProfile mutations
	profile.Luck = 1.25
	profile.Reputation = 60
	profile.UnlockedParadigms = "MAGITECH,CYBERPUNK"
	if err := SaveProfile(profile); err != nil {
		t.Fatalf("SaveProfile failed: %v", err)
	}

	// Retrieve again and verify
	mutated, err := GetOrCreateProfile()
	if err != nil {
		t.Fatalf("Retrieving mutated profile failed: %v", err)
	}
	if mutated.Luck != 1.25 {
		t.Errorf("Expected luck 1.25, got: %f", mutated.Luck)
	}
	if mutated.Reputation != 60 {
		t.Errorf("Expected reputation 60, got: %d", mutated.Reputation)
	}
	if !mutated.IsParadigmUnlocked("CYBERPUNK") {
		t.Error("Expected CYBERPUNK to be unlocked now")
	}

	// 4. Test RecordToken duplicate check
	token := "TEST_LOGOS_TOKEN_XYZ"
	isNew, err := RecordToken(token)
	if err != nil {
		t.Fatalf("RecordToken failed: %v", err)
	}
	if !isNew {
		t.Error("Expected token to be newly registered")
	}

	isNewAgain, err := RecordToken(token)
	if err != nil {
		t.Fatalf("RecordToken duplicate check failed: %v", err)
	}
	if isNewAgain {
		t.Error("Expected token to be marked as duplicate (not new)")
	}
}
