package generator

import (
	"testing"

	"challenge-to-you/backend/internal/engine"
)

func TestGeneratorDeterminism(t *testing.T) {
	seed := int64(12345)
	luck := 0.5
	paradigm := engine.Cyberpunk

	def1, err := GenerateChallenge(seed, luck, paradigm)
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}

	def2, err := GenerateChallenge(seed, luck, paradigm)
	if err != nil {
		t.Fatalf("Failed to generate second time: %v", err)
	}

	if def1.ID != def2.ID {
		t.Errorf("Mismatch in ID: %s vs %s", def1.ID, def2.ID)
	}
	if def1.Name != def2.Name {
		t.Errorf("Mismatch in Name: %s vs %s", def1.Name, def2.Name)
	}
	if len(def1.Flaws) != len(def2.Flaws) {
		t.Errorf("Mismatch in flaws count: %d vs %d", len(def1.Flaws), len(def2.Flaws))
	}
	if def1.LogosToken != def2.LogosToken {
		t.Errorf("Mismatch in LogosToken: %s vs %s", def1.LogosToken, def2.LogosToken)
	}
	if def1.WinCondition.TargetStateKey != def2.WinCondition.TargetStateKey {
		t.Errorf("Mismatch in win state target key")
	}
}

func TestGeneratorLuckScaling(t *testing.T) {
	// Luck 1.0 -> 0 noise flaws, total flaws = 3 (the 3 steps of golden path)
	defMax, err := GenerateChallenge(42, 1.0, engine.Magitech)
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}
	if len(defMax.Flaws) != 3 {
		t.Errorf("Expected exactly 3 flaws at luck 1.0 (no noise), got: %d", len(defMax.Flaws))
	}

	// Luck 0.0 -> 5 noise flaws, total flaws = 8
	defMin, err := GenerateChallenge(42, 0.0, engine.Magitech)
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}
	if len(defMin.Flaws) != 8 {
		t.Errorf("Expected exactly 8 flaws at luck 0.0 (5 noise), got: %d", len(defMin.Flaws))
	}
}

func TestGeneratorSolvability(t *testing.T) {
	paradigms := []engine.Paradigm{engine.Magitech, engine.Cyberpunk, engine.Cosmic}

	for _, paradigm := range paradigms {
		t.Run(string(paradigm), func(t *testing.T) {
			def, err := GenerateChallenge(999, 0.5, paradigm)
			if err != nil {
				t.Fatalf("Failed to generate: %v", err)
			}

			fabric := def.BuildFabric()

			// To solve a procedural level, we must identify the golden path trigger events.
			// Step 1: find the flaw that mutates the state1 key (which has no conditions)
			var step1Evt, step2Evt, step3Evt string
			var state1Key, state2Key, state3Key string

			// Scan the flaws to map out the state keys and their triggers
			for _, f := range def.Flaws {
				// Step 1 has no conditions
				if len(f.Conditions) == 0 && f.ID == "flaw_step1" {
					step1Evt = f.TriggerEvent
					for k := range f.Mutations {
						if k != "entropy" {
							state1Key = k
						}
					}
				}
			}

			if step1Evt == "" || state1Key == "" {
				t.Fatal("Could not identify Step 1 of Golden Path")
			}

			// Step 2 has a condition on state1Key
			for _, f := range def.Flaws {
				if f.ID == "flaw_step2" {
					if _, ok := f.Conditions[state1Key]; ok {
						step2Evt = f.TriggerEvent
						for k := range f.Mutations {
							if k != "entropy" {
								state2Key = k
							}
						}
					}
				}
			}

			if step2Evt == "" || state2Key == "" {
				t.Fatalf("Could not identify Step 2 of Golden Path. State1: %s", state1Key)
			}

			// Step 3 has a condition on state2Key
			for _, f := range def.Flaws {
				if f.ID == "flaw_step3" {
					if _, ok := f.Conditions[state2Key]; ok {
						step3Evt = f.TriggerEvent
						state3Key = def.WinCondition.TargetStateKey
					}
				}
			}

			if step3Evt == "" || state3Key == "" {
				t.Fatalf("Could not identify Step 3 of Golden Path. State2: %s", state2Key)
			}

			// Verify correct execution sequence
			var cipher string
			var complete bool

			// Trigger Step 1
			cipher, complete, err = fabric.TriggerOntologicalShift(step1Evt)
			if err != nil {
				t.Fatalf("Step 1 failed: %v", err)
			}
			if complete {
				t.Fatal("Level solved prematurely at step 1")
			}

			// Trigger Step 2
			cipher, complete, err = fabric.TriggerOntologicalShift(step2Evt)
			if err != nil {
				t.Fatalf("Step 2 failed: %v", err)
			}
			if complete {
				t.Fatal("Level solved prematurely at step 2")
			}

			// Trigger Step 3
			cipher, complete, err = fabric.TriggerOntologicalShift(step3Evt)
			if err != nil {
				t.Fatalf("Step 3 failed: %v", err)
			}
			if !complete {
				t.Fatal("Level was not solved after completing the Golden Path sequence")
			}

			if cipher != def.LogosToken {
				t.Errorf("Expected LogosToken %s, got: %s", def.LogosToken, cipher)
			}
		})
	}
}
