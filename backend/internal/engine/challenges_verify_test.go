package engine

import (
	"path/filepath"
	"runtime"
	"testing"
)

// Helper to get challenges folder path
func challengesRootPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "challenges")
}

func TestVerifyAllChallenges(t *testing.T) {
	root := challengesRootPath()

	// Map of challenge ID to its correct sequence of triggers that leads to a win
	solutions := map[string][]string{
		// Magitech
		"magitech_01_breach":     {"invoke_binding", "surge_mana", "trigger_release"},
		"magitech_02_centrifuge": {"overheat_centrifuge", "shatter_crystal", "trigger_vent"},
		"magitech_03_vault":      {"engage_anchor", "wait_for_swing", "disable_anchor"},
		"magitech_04_golem":      {"blind_sensor", "issue_command_defend", "purge_mana"},
		"magitech_05_grimoire":   {"project_illusion", "activate_silence_seal", "trigger_truth_ward"},
		"magitech_06_astrolabe":  {"engage_sun_lens", "force_moon_gear_slip", "align_void_prism"},
		"magitech_07_loom":       {"tangle_thread", "crank_tension", "snap_thread"},

		// Cyberpunk
		"cyberpunk_01_autodoc":  {"inject_adrenaline", "cut_power"},
		"cyberpunk_02_elevator": {"spoof_negative_mass", "deploy_brakes"},
		"cyberpunk_03_server":   {"send_bad_packet", "trigger_compactor"},
		"cyberpunk_04_barista":  {"fill_order_queue", "overheat_frother"},
		"cyberpunk_05_drone":    {"spoof_altimeter", "deploy_flotation"},
		"cyberpunk_06_traffic":  {"force_all_red", "trigger_shorted_button"},
		"cyberpunk_c42_thread_race": {"spawn_writer_thread", "race_second_thread"},

		// Cosmic
		"cosmic_01_airlock":     {"trigger_siren", "blind_observer", "materialize_door"},
		"cosmic_02_nav":         {"feed_zero_value", "trigger_abort"},
		"cosmic_03_stasis":      {"trigger_breach_alarm", "align_laser", "fire_laser"},
		"cosmic_04_winch":       {"spoof_high_reading", "overcrank_motor"},
		"cosmic_05_seed":        {"ping_soil_analyzer", "deploy_shield"},
		"cosmic_06_singularity": {"invert_polarities", "dump_mass"},
		"cosmic_07_relay":       {"max_amplifier", "overload_scrubber"},
		"cosmic_08_valve":       {"activate_dirty_sensor", "trigger_argon_flush"},
	}

	packs := []struct {
		name string
		era  string
		tier int
	}{
		{"magitech_tier1", "magitech", 1},
		{"cyberpunk_tier1", "cyberpunk", 1},
		{"cosmic_tier1", "cosmic", 1},
	}

	for _, p := range packs {
		packDir := filepath.Join(root, p.name)
		pack, err := LoadPack(packDir)
		if err != nil {
			t.Fatalf("Failed to load pack %s: %v", p.name, err)
		}

		if pack.Era != p.era {
			t.Errorf("Pack %s: expected era %s, got %s", p.name, p.era, pack.Era)
		}
		if pack.Tier != p.tier {
			t.Errorf("Pack %s: expected tier %d, got %d", p.name, p.tier, pack.Tier)
		}

		runtimePack, err := NewChallengePack(pack, packDir)
		if err != nil {
			t.Fatalf("Failed to create runtime pack %s: %v", p.name, err)
		}

		for _, ref := range pack.Challenges {
			t.Run(ref.ID, func(t *testing.T) {
				challenge, exists := runtimePack.GetChallenge(ref.ID)
				if !exists {
					t.Fatalf("Challenge %s not found in runtime pack", ref.ID)
				}

				// Check initial state
				fabric, exists := runtimePack.GetFabric(ref.ID)
				if !exists {
					t.Fatalf("Fabric for challenge %s not found", ref.ID)
				}

				if len(fabric.State) == 0 {
					t.Error("Initial state is empty")
				}

				// Check that the win condition targets a key in state
				if _, ok := fabric.State[challenge.WinCondition.TargetStateKey]; !ok {
					t.Errorf("Win condition targets state key %s which is not present in initial state", challenge.WinCondition.TargetStateKey)
				}

				// Verify out-of-order execution doesn't solve it
				solSteps, found := solutions[ref.ID]
				if !found {
					t.Fatalf("No solution steps documented for %s", ref.ID)
				}

				// Run steps in reverse order (except single-step or trivial ones)
				if len(solSteps) > 1 {
					fabricRev := NewAxiomaticFabric(fabric.CurrentParadigm, fabric.WinConditionKey, fabric.WinConditionVal)
					for k, v := range fabric.State {
						fabricRev.SetState(k, v)
					}
					for _, g := range fabric.Glitches {
						fabricRev.RegisterGlitch(g)
					}

					// Trigger the last step first
					_, completeRev, err := fabricRev.TriggerOntologicalShift(solSteps[len(solSteps)-1])
					if err != nil {
						t.Errorf("Unexpected error during reverse test: %v", err)
					}
					if completeRev {
						t.Error("Triggering last step first should not win the challenge")
					}
				}

				// Run correct steps in order
				var cipher string
				var complete bool
				for i, event := range solSteps {
					c, comp, err := fabric.TriggerOntologicalShift(event)
					if err != nil {
						t.Fatalf("Step %d (%s) failed with error: %v", i+1, event, err)
					}
					cipher = c
					complete = comp

					// Intermediate steps should not be complete
					if i < len(solSteps)-1 && complete {
						t.Fatalf("Challenge solved prematurely at step %d (%s)", i+1, event)
					}
				}

				if !complete {
					t.Error("Challenge not completed after running all solution steps")
				}

				if cipher == "" {
					t.Error("No cipher/logos token generated upon completion")
				}

				if cipher != challenge.LogosToken {
					t.Errorf("Expected logos token %s, got %s", challenge.LogosToken, cipher)
				}
			})
		}
	}
}
