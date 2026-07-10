package engine

import (
	"math/rand"
	"testing"
)

func TestGenerateJunkFlaws(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	junk := GenerateJunkFlaws(Magitech, 3, rng)
	if len(junk) != 3 {
		t.Errorf("expected 3 junk flaws, got %d", len(junk))
	}

	for _, f := range junk {
		if f.ID == "" {
			t.Error("junk flaw missing ID")
		}
		if f.TriggerEvent == "" {
			t.Error("junk flaw missing TriggerEvent")
		}
		if f.Name == "" {
			t.Error("junk flaw missing Name")
		}
		if f.FlavorText == "" {
			t.Error("junk flaw missing FlavorText")
		}
		if f.Conditions == nil {
			t.Error("junk flaw Conditions should not be nil")
		}
		if f.Mutations == nil {
			t.Error("junk flaw Mutations should not be nil")
		}
		if f.FallbackMutations == nil {
			t.Error("junk flaw FallbackMutations should not be nil")
		}
		if _, hasEntropy := f.Mutations["entropy"]; !hasEntropy {
			t.Errorf("junk flaw %s missing entropy mutation", f.ID)
		}
	}
}

func TestGenerateJunkFlawsAllParadigms(t *testing.T) {
	rng := rand.New(rand.NewSource(123))

	for _, paradigm := range []Paradigm{Magitech, Cyberpunk, Cosmic} {
		junk := GenerateJunkFlaws(paradigm, 2, rng)
		if len(junk) != 2 {
			t.Errorf("%s: expected 2 junk flaws, got %d", paradigm, len(junk))
		}
		for _, f := range junk {
			if f.ID == "" || f.TriggerEvent == "" {
				t.Errorf("%s: junk flaw incomplete", paradigm)
			}
		}
	}
}

func TestHydrateChallenge(t *testing.T) {
	base := &ChallengeDefinition{
		ID:          "test_hydrate",
		Paradigm:    Cyberpunk,
		Name:        "Test Hydration",
		Description: "Base challenge",
		LogosToken:  "LOGOS_TEST",
		InitialState: ParadigmState{
			"target_locked": true,
			"entropy":       0,
		},
		Flaws: []Flaw{
			{
				ID:           "core_exploit",
				TriggerEvent: "trigger_exploit",
				Name:         "Core Exploit",
				FlavorText:   "The real solution.",
				Conditions:   ParadigmState{},
				Mutations:    ParadigmState{"target_locked": false},
			},
		},
		WinCondition: WinCondition{
			TargetStateKey: "target_locked",
			ExpectedValue:  false,
		},
	}

	cfg := HydratorConfig{
		Seed:           999,
		Luck:           0.0,
		BaseDefinition: base,
	}

	hydrated := HydrateChallenge(cfg)
	if hydrated == nil {
		t.Fatal("HydrateChallenge returned nil")
	}

	if len(hydrated.Flaws) < 2 {
		t.Errorf("expected at least 2 flaws (core + junk), got %d", len(hydrated.Flaws))
	}

	foundCore := false
	foundJunk := false
	for _, f := range hydrated.Flaws {
		if f.ID == "core_exploit" {
			foundCore = true
		}
		if len(f.ID) > 4 && f.ID[:4] == "junk" {
			foundJunk = true
		}
	}
	if !foundCore {
		t.Error("core exploit missing from hydrated flaws")
	}
	if !foundJunk {
		t.Error("no junk flaws found in hydrated challenge")
	}
}

func TestHydrateChallengeHighLuck(t *testing.T) {
	base := &ChallengeDefinition{
		ID:       "test_luck",
		Paradigm: Magitech,
		Flaws: []Flaw{
			{ID: "core", TriggerEvent: "win", Mutations: ParadigmState{"done": true}},
		},
		WinCondition: WinCondition{TargetStateKey: "done", ExpectedValue: true},
	}

	cfg := HydratorConfig{Seed: 1, Luck: 1.0, BaseDefinition: base}
	hydrated := HydrateChallenge(cfg)

	if len(hydrated.Flaws) > 2 {
		t.Errorf("high luck should add minimal junk, got %d flaws", len(hydrated.Flaws))
	}
}

func TestHydrateChallengeZeroLuck(t *testing.T) {
	base := &ChallengeDefinition{
		ID:       "test_zero_luck",
		Paradigm: Cosmic,
		Flaws: []Flaw{
			{ID: "core", TriggerEvent: "win", Mutations: ParadigmState{"done": true}},
		},
		WinCondition: WinCondition{TargetStateKey: "done", ExpectedValue: true},
	}

	cfg := HydratorConfig{Seed: 1, Luck: 0.0, BaseDefinition: base}
	hydrated := HydrateChallenge(cfg)

	if len(hydrated.Flaws) < 5 {
		t.Errorf("zero luck should add max junk, got %d flaws", len(hydrated.Flaws))
	}
}

func TestHydrateChallengeShufflesFlaws(t *testing.T) {
	base := &ChallengeDefinition{
		ID:       "test_shuffle",
		Paradigm: Cyberpunk,
		Flaws: []Flaw{
			{ID: "a", TriggerEvent: "a"},
			{ID: "b", TriggerEvent: "b"},
			{ID: "c", TriggerEvent: "c"},
		},
		WinCondition: WinCondition{TargetStateKey: "done", ExpectedValue: true},
	}

	cfg := HydratorConfig{Seed: 42, Luck: 0.0, BaseDefinition: base}
	hydrated := HydrateChallenge(cfg)

	firstIDs := []string{}
	for _, f := range hydrated.Flaws {
		firstIDs = append(firstIDs, f.ID)
	}

	cfg2 := HydratorConfig{Seed: 43, Luck: 0.0, BaseDefinition: base}
	hydrated2 := HydrateChallenge(cfg2)

	secondIDs := []string{}
	for _, f := range hydrated2.Flaws {
		secondIDs = append(secondIDs, f.ID)
	}

	same := true
	for i := range firstIDs {
		if firstIDs[i] != secondIDs[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("different seeds should produce different shuffle orders")
	}
}