package engine

import (
	"path/filepath"
	"runtime"
	"testing"
)

func packPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "challenges", "magitech_tier1")
}

func TestLoadPack(t *testing.T) {
	pack, err := LoadPack(packPath())
	if err != nil {
		t.Fatalf("LoadPack failed: %v", err)
	}

	if pack.Era != "magitech" {
		t.Errorf("expected era magitech, got %s", pack.Era)
	}
	if pack.Tier != 1 {
		t.Errorf("expected tier 1, got %d", pack.Tier)
	}
	if len(pack.Challenges) != 7 {
		t.Errorf("expected 7 challenges, got %d", len(pack.Challenges))
	}

	expectedIDs := []string{
		"magitech_01_breach",
		"magitech_02_centrifuge",
		"magitech_03_vault",
		"magitech_04_golem",
		"magitech_05_grimoire",
		"magitech_06_astrolabe",
		"magitech_07_loom",
	}
	for i, ref := range pack.Challenges {
		if ref.ID != expectedIDs[i] {
			t.Errorf("challenge %d: expected %s, got %s", i, expectedIDs[i], ref.ID)
		}
	}
}

func TestChallengePackRuntime(t *testing.T) {
	pack, err := LoadPack(packPath())
	if err != nil {
		t.Fatalf("LoadPack failed: %v", err)
	}

	runtimePack, err := NewChallengePack(pack, packPath())
	if err != nil {
		t.Fatalf("NewChallengePack failed: %v", err)
	}

	// Test GetChallenge
	challenge, ok := runtimePack.GetChallenge("magitech_01_breach")
	if !ok {
		t.Fatal("magitech_01_breach not found")
	}
	if challenge.Paradigm != Magitech {
		t.Errorf("expected Magitech paradigm, got %s", challenge.Paradigm)
	}

	// Test GetFabric returns fresh copy
	fabric1, ok := runtimePack.GetFabric("magitech_01_breach")
	if !ok {
		t.Fatal("fabric not found")
	}
	fabric1.SetState("test_key", "test_val")

	fabric2, ok := runtimePack.GetFabric("magitech_01_breach")
	if !ok {
		t.Fatal("fabric not found second time")
	}
	if _, exists := fabric2.State["test_key"]; exists {
		t.Error("GetFabric should return independent copies")
	}

	// Test available challenges (all 7 should be available since no prerequisites)
	completed := map[string]bool{}
	available := runtimePack.GetAvailableChallenges(completed)
	if len(available) != 7 {
		t.Errorf("expected 7 available, got %d: %v", len(available), available)
	}
}
