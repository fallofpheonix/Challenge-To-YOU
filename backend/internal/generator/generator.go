package generator

import (
	"fmt"
	"math/rand"

	"challenge-to-you/backend/internal/engine"
)

// Vocabulary pools for procedural randomization
var magitechVocab = struct {
	Step1    []string
	Step2    []string
	Step3    []string
	State1   []string
	State2   []string
	State3   []string
	Flaw1    []string
	Flaw2    []string
	Flaw3    []string
	Flaw1Flv []string
	Flaw2Flv []string
	Flaw3Flv []string
	JunkEvt  []string
	JunkName []string
	JunkFlv  []string
}{
	Step1:    []string{"invoke_binding", "align_resonance", "stabilize_sigil"},
	Step2:    []string{"surge_mana", "excite_ether", "channel_leyline"},
	Step3:    []string{"trigger_release", "shatter_barrier", "discharge_node"},
	State1:   []string{"binding_active", "resonance_locked", "sigil_stable"},
	State2:   []string{"mana_critical", "ether_saturated", "leyline_pegged"},
	State3:   []string{"ward_sealed", "gate_locked", "runic_barrier_active"},
	Flaw1:    []string{"Cracked Binding Rune", "Fraying Resonance Sigil", "Unstable Focal Point"},
	Flaw2:    []string{"Leaking Mana Conduit", "Bleeding Ether Pipe", "Overloaded Leyline Junction"},
	Flaw3:    []string{"Fractured Release Seal", "Compromised Gate Trigger", "Failing Discharge Glyphs"},
	Flaw1Flv: []string{"The runic structure is compromised. Activating it forcefully bypasses safety interlocks.", "A localized leak allows manual injection of binding coefficients."},
	Flaw2Flv: []string{"Unregulated mana bleeds from this connection. Spikes the pressure if the interlock is active.", "Raw ether leaks. Feedback loop causes critical state if the sigil is stabilized."},
	Flaw3Flv: []string{"The release lock is jammed. Requires a critical mana spike to physically shatter the gate.", "The mechanical latch ignores normal inputs. A sudden ether discharge will force it open."},
	JunkEvt:  []string{"vent_coolant", "flush_exhaust", "realign_lenses", "ground_coils", "pulse_beacon"},
	JunkName: []string{"Coolant Vent Sigil", "Emergency Exhaust Valve", "Navigational Lens Array", "Grounding Copper Coil", "Arcane Signal Beacon"},
	JunkFlv:  []string{"Vents ambient steam. Does not affect the lock, but increases instability.", "Flashes warning glyphs. Harmless unless triggered during critical load.", "Realigns visual focus. Harmless but draws the Archon's attention."},
}

var cyberpunkVocab = struct {
	Step1    []string
	Step2    []string
	Step3    []string
	State1   []string
	State2   []string
	State3   []string
	Flaw1    []string
	Flaw2    []string
	Flaw3    []string
	Flaw1Flv []string
	Flaw2Flv []string
	Flaw3Flv []string
	JunkEvt  []string
	JunkName []string
	JunkFlv  []string
}{
	Step1:    []string{"bypass_firewall", "disable_proxy", "kill_daemon"},
	Step2:    []string{"spoof_credentials", "inject_token", "fake_handshake"},
	Step3:    []string{"grant_root", "spawn_shell", "force_override"},
	State1:   []string{"firewall_disabled", "proxy_offline", "daemon_terminated"},
	State2:   []string{"credentials_spoofed", "token_validated", "handshake_accepted"},
	State3:   []string{"root_access_granted", "shell_active", "system_compromised"},
	Flaw1:    []string{"Compromised Firewall Daemon", "Rerouted Proxy Connection", "Failing Access Daemon"},
	Flaw2:    []string{"Spoofed Biometric Injector", "Leaked Session Token Buffer", "Out-of-Order Handshake Routine"},
	Flaw3:    []string{"Unsanctioned Privilege Escalator", "Vulnerable Exec Shunt", "Unprotected Shell Port"},
	Flaw1Flv: []string{"A diagnostic port remains open. Bypassing it drops the local firewall security.", "Terminating this background daemon disables proxy routing checks."},
	Flaw2Flv: []string{"Allows injection of fake credential parameters. Succeeds if proxy routing is disabled.", "Overwriting the session buffer bypasses biometric checks if firewall is offline."},
	Flaw3Flv: []string{"An escalation exploit that grants root privileges if credentials are successfully spoofed.", "A debug shell that spawns with admin privileges when session token is valid."},
	JunkEvt:  []string{"flush_dns", "reboot_nic", "dump_syslog", "ping_loopback", "alloc_heap"},
	JunkName: []string{"DNS Cache Flusher", "NIC Controller Reset", "Syslog Buffer Dumper", "Loopback Diagnostic Ping", "Heap Memory Allocator"},
	JunkFlv:  []string{"Clears routing caches. Causes system friction and raises detection levels.", "Resets network interfaces, generating temporary noise flags.", "Dumps syslog records into active buffers, causing minor memory leakage."},
}

var cosmicVocab = struct {
	Step1    []string
	Step2    []string
	Step3    []string
	State1   []string
	State2   []string
	State3   []string
	Flaw1    []string
	Flaw2    []string
	Flaw3    []string
	Flaw1Flv []string
	Flaw2Flv []string
	Flaw3Flv []string
	JunkEvt  []string
	JunkName []string
	JunkFlv  []string
}{
	Step1:    []string{"stabilize_manifold", "anchor_gravity", "contain_singularity"},
	Step2:    []string{"excite_quantum_state", "align_phase", "coherent_flux"},
	Step3:    []string{"open_rift_door", "unseal_airlock", "unlock_hatch"},
	State1:   []string{"manifold_stable", "gravity_anchored", "singularity_contained"},
	State2:   []string{"quantum_excited", "phase_aligned", "flux_coherent"},
	State3:   []string{"airlock_open", "rift_unsealed", "hatch_unlocked"},
	Flaw1:    []string{"Miscalibrated Manifold Servo", "Leaking Gravity Anchor", "Failing Singularity Core"},
	Flaw2:    []string{"Acoustic Quantum Exciter", "Phase-Shift Synchronizer", "Coherent Flux Radiator"},
	Flaw3:    []string{"Emergency Superposition Airlock", "Rift Controller Gate", "Pneumatic Hatch Release"},
	Flaw1Flv: []string{"A feedback loop stabilizes gravity coefficients when triggered manually.", "Manually pulsing this coil contains the singularity's spatial distortion."},
	Flaw2Flv: []string{"Vibrates the alignment servo into excitation if gravity is contained.", "Synchronizes quantum phases when the containment field is stable."},
	Flaw3Flv: []string{"Venting system collapses superposition, forcing the door into a physical open state.", "Releases the pneumatic seal if quantum state is excited."},
	JunkEvt:  []string{"purge_coolant", "reboot_navigation", "cycle_compressor", "vent_nitrogen", "ping_telemetry"},
	JunkName: []string{"Coolant Purge Valve", "Nav-Computer Reboot Switch", "Air Compressor Cycler", "Nitrogen Vent Line", "Telemetry Beacon Ping"},
	JunkFlv:  []string{"Purges active nitrogen lines, introducing noise into structural sensors.", "Forces navigation recalculations, drawing attention from the system scanner.", "Cycles cargo seals, temporarily destabilizing pressure metrics."},
}

// GenerateChallenge procedurally builds a fully solvable level config with luck-based noise
func GenerateChallenge(seed int64, luck float64, paradigm engine.Paradigm) (*engine.ChallengeDefinition, error) {
	// Create seed-based local random generator.
	// #nosec G404 -- deterministic seeded RNG is required for reproducible challenges, not security.
	r := rand.New(rand.NewSource(seed))

	var title string
	var description string
	var logosToken string
	var step1Evt, step2Evt, step3Evt string
	var state1Key, state2Key, state3Key string
	var flaw1Name, flaw2Name, flaw3Name string
	var flaw1Flv, flaw2Flv, flaw3Flv string
	var junkEvtPool []string
	var junkNamePool []string
	var junkFlvPool []string

	// 1. Pick vocab and templates based on Paradigm
	switch paradigm {
	case engine.Magitech:
		title = fmt.Sprintf("Procedural Runic Gateway #%d", r.Intn(1000))
		description = "A runic gateway blocks the data-stream. Exploit its compromised state transitions to bypass it."
		logosToken = fmt.Sprintf("LOGOS_MGT_GEN_%d", r.Intn(99999))

		step1Evt = magitechVocab.Step1[r.Intn(len(magitechVocab.Step1))]
		step2Evt = magitechVocab.Step2[r.Intn(len(magitechVocab.Step2))]
		step3Evt = magitechVocab.Step3[r.Intn(len(magitechVocab.Step3))]

		state1Key = magitechVocab.State1[r.Intn(len(magitechVocab.State1))]
		state2Key = magitechVocab.State2[r.Intn(len(magitechVocab.State2))]
		state3Key = magitechVocab.State3[r.Intn(len(magitechVocab.State3))]

		flaw1Name = magitechVocab.Flaw1[r.Intn(len(magitechVocab.Flaw1))]
		flaw2Name = magitechVocab.Flaw2[r.Intn(len(magitechVocab.Flaw2))]
		flaw3Name = magitechVocab.Flaw3[r.Intn(len(magitechVocab.Flaw3))]

		flaw1Flv = magitechVocab.Flaw1Flv[r.Intn(len(magitechVocab.Flaw1Flv))]
		flaw2Flv = magitechVocab.Flaw2Flv[r.Intn(len(magitechVocab.Flaw2Flv))]
		flaw3Flv = magitechVocab.Flaw3Flv[r.Intn(len(magitechVocab.Flaw3Flv))]

		junkEvtPool = magitechVocab.JunkEvt
		junkNamePool = magitechVocab.JunkName
		junkFlvPool = magitechVocab.JunkFlv

	case engine.Cyberpunk:
		title = fmt.Sprintf("Procedural Override Subsystem #%d", r.Intn(1000))
		description = "Bypass corporate network components by triggering diagnostic overrides and buffer vulnerabilities."
		logosToken = fmt.Sprintf("LOGOS_CYB_GEN_%d", r.Intn(99999))

		step1Evt = cyberpunkVocab.Step1[r.Intn(len(cyberpunkVocab.Step1))]
		step2Evt = cyberpunkVocab.Step2[r.Intn(len(cyberpunkVocab.Step2))]
		step3Evt = cyberpunkVocab.Step3[r.Intn(len(cyberpunkVocab.Step3))]

		state1Key = cyberpunkVocab.State1[r.Intn(len(cyberpunkVocab.State1))]
		state2Key = cyberpunkVocab.State2[r.Intn(len(cyberpunkVocab.State2))]
		state3Key = cyberpunkVocab.State3[r.Intn(len(cyberpunkVocab.State3))]

		flaw1Name = cyberpunkVocab.Flaw1[r.Intn(len(cyberpunkVocab.Flaw1))]
		flaw2Name = cyberpunkVocab.Flaw2[r.Intn(len(cyberpunkVocab.Flaw2))]
		flaw3Name = cyberpunkVocab.Flaw3[r.Intn(len(cyberpunkVocab.Flaw3))]

		flaw1Flv = cyberpunkVocab.Flaw1Flv[r.Intn(len(cyberpunkVocab.Flaw1Flv))]
		flaw2Flv = cyberpunkVocab.Flaw2Flv[r.Intn(len(cyberpunkVocab.Flaw2Flv))]
		flaw3Flv = cyberpunkVocab.Flaw3Flv[r.Intn(len(cyberpunkVocab.Flaw3Flv))]

		junkEvtPool = cyberpunkVocab.JunkEvt
		junkNamePool = cyberpunkVocab.JunkName
		junkFlvPool = cyberpunkVocab.JunkFlv

	case engine.Cosmic:
		title = fmt.Sprintf("Procedural Superposition Matrix #%d", r.Intn(1000))
		description = "A containment field prevents physical bridge alignment. Access terminal overrides to collapse and bypass it."
		logosToken = fmt.Sprintf("LOGOS_CSM_GEN_%d", r.Intn(99999))

		step1Evt = cosmicVocab.Step1[r.Intn(len(cosmicVocab.Step1))]
		step2Evt = cosmicVocab.Step2[r.Intn(len(cosmicVocab.Step2))]
		step3Evt = cosmicVocab.Step3[r.Intn(len(cosmicVocab.Step3))]

		state1Key = cosmicVocab.State1[r.Intn(len(cosmicVocab.State1))]
		state2Key = cosmicVocab.State2[r.Intn(len(cosmicVocab.State2))]
		state3Key = cosmicVocab.State3[r.Intn(len(cosmicVocab.State3))]

		flaw1Name = cosmicVocab.Flaw1[r.Intn(len(cosmicVocab.Flaw1))]
		flaw2Name = cosmicVocab.Flaw2[r.Intn(len(cosmicVocab.Flaw2))]
		flaw3Name = cosmicVocab.Flaw3[r.Intn(len(cosmicVocab.Flaw3))]

		flaw1Flv = cosmicVocab.Flaw1Flv[r.Intn(len(cosmicVocab.Flaw1Flv))]
		flaw2Flv = cosmicVocab.Flaw2Flv[r.Intn(len(cosmicVocab.Flaw2Flv))]
		flaw3Flv = cosmicVocab.Flaw3Flv[r.Intn(len(cosmicVocab.Flaw3Flv))]

		junkEvtPool = cosmicVocab.JunkEvt
		junkNamePool = cosmicVocab.JunkName
		junkFlvPool = cosmicVocab.JunkFlv

	default:
		return nil, fmt.Errorf("unknown paradigm: %s", paradigm)
	}

	// 2. Build initial state map
	initialState := make(engine.ParadigmState)
	initialState[state1Key] = false
	initialState[state2Key] = false
	// Win target key starts as true (Magitech/Cosmic: sealed, locked) or false (Cyberpunk: granted)
	if paradigm == engine.Cyberpunk {
		initialState[state3Key] = false
	} else {
		initialState[state3Key] = true
	}
	initialState["entropy"] = 0

	// 3. Assemble Golden Path Flaws
	flaw1 := engine.Flaw{
		ID:           "flaw_step1",
		TriggerEvent: step1Evt,
		Name:         flaw1Name,
		FlavorText:   flaw1Flv,
		Conditions:   engine.ParadigmState{},
		Mutations: engine.ParadigmState{
			state1Key: true,
		},
		FallbackMutations: engine.ParadigmState{},
	}

	flaw2 := engine.Flaw{
		ID:           "flaw_step2",
		TriggerEvent: step2Evt,
		Name:         flaw2Name,
		FlavorText:   flaw2Flv,
		Conditions: engine.ParadigmState{
			state1Key: true,
		},
		Mutations: engine.ParadigmState{
			state2Key: true,
		},
		FallbackMutations: engine.ParadigmState{
			"entropy": 10,
		},
	}

	// Win condition value is false (unsealed/offline) or true (granted)
	winValue := false
	if paradigm == engine.Cyberpunk {
		winValue = true
	}

	flaw3 := engine.Flaw{
		ID:           "flaw_step3",
		TriggerEvent: step3Evt,
		Name:         flaw3Name,
		FlavorText:   flaw3Flv,
		Conditions: engine.ParadigmState{
			state2Key: true,
		},
		Mutations: engine.ParadigmState{
			state3Key: winValue,
		},
		FallbackMutations: engine.ParadigmState{
			"entropy": 25,
		},
	}

	flawsList := []engine.Flaw{flaw1, flaw2, flaw3}

	// 4. Generate Noise (Apophenic Cacophony) based on Luck
	// Max noise is 5 (at luck = 0.0), minimum is 0 (at luck = 1.0)
	noiseCount := int((1.0 - luck) * 5.0)
	if noiseCount < 0 {
		noiseCount = 0
	} else if noiseCount > 5 {
		noiseCount = 5
	}

	// Shuffle junk pools
	junkIndices := r.Perm(len(junkEvtPool))
	for i := 0; i < noiseCount && i < len(junkIndices); i++ {
		idx := junkIndices[i]
		junkEvt := fmt.Sprintf("%s_%d", junkEvtPool[idx], r.Intn(10))
		junkName := fmt.Sprintf("%s (0x%02X)", junkNamePool[idx], r.Intn(256))
		junkFlv := junkFlvPool[idx%len(junkFlvPool)]

		// Add a trap condition or mutation
		junkFlaw := engine.Flaw{
			ID:           fmt.Sprintf("junk_%d", i),
			TriggerEvent: junkEvt,
			Name:         junkName,
			FlavorText:   junkFlv + " WARNING: Triggers Archon static warning.",
			Conditions:   engine.ParadigmState{},
			Mutations: engine.ParadigmState{
				"entropy": 15,
			},
			FallbackMutations: engine.ParadigmState{},
		}
		flawsList = append(flawsList, junkFlaw)
	}

	// Shuffle the final flaws list so the Golden Path steps are randomized visually
	shuffledFlaws := make([]engine.Flaw, len(flawsList))
	perm := r.Perm(len(flawsList))
	for i, v := range perm {
		shuffledFlaws[v] = flawsList[i]
	}

	// Assemble Challenge
	def := &engine.ChallengeDefinition{
		ID:           fmt.Sprintf("proc_%s_%d", paradigm, seed),
		Paradigm:     paradigm,
		Name:         title,
		Description:  description,
		LogosToken:   logosToken,
		InitialState: initialState,
		Flaws:        shuffledFlaws,
		WinCondition: engine.WinCondition{
			TargetStateKey: state3Key,
			ExpectedValue:  winValue,
		},
	}

	return def, nil
}
