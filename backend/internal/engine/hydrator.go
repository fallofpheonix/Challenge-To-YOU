package engine

import (
	"math/rand"
	"time"
)

type JunkFlawTemplate struct {
	IDPrefix       string
	TriggerEvent   string
	Name           string
	FlavorText     string
	ConditionKeys  []string
	MutationKeys   map[string]any
	FallbackKeys   map[string]any
}

var JunkFlawCatalog = map[Paradigm][]JunkFlawTemplate{
	Magitech: {
		{
			IDPrefix:     "junk_ward_",
			TriggerEvent: "reinforce_ward",
			Name:         "Ward Reinforcement Protocol",
			FlavorText:   "The ward matrix accepts structural reinforcement commands. Initiating stabilization cycle consumes entropy reserves.",
			ConditionKeys: []string{},
			MutationKeys:  map[string]any{"entropy": 15},
			FallbackKeys:  map[string]any{"entropy": 20},
		},
		{
			IDPrefix:     "junk_mana_",
			TriggerEvent: "vent_mana",
			Name:         "Mana Vent Release",
			FlavorText:   "Emergency mana venting protocol. Releases excess etheric pressure at the cost of temporary instability.",
			ConditionKeys: []string{},
			MutationKeys:  map[string]any{"entropy": 10},
			FallbackKeys:  map[string]any{"entropy": 15},
		},
		{
			IDPrefix:     "junk_binding_",
			TriggerEvent: "recite_binding",
			Name:         "Binding Rune Recitation",
			FlavorText:   "Reciting the binding mantras realigns local reality but attracts minor attention from the Archon.",
			ConditionKeys: []string{},
			MutationKeys:  map[string]any{"entropy": 8},
			FallbackKeys:  map[string]any{"entropy": 12},
		},
		{
			IDPrefix:     "junk_glyph_",
			TriggerEvent: "trace_glyph",
			Name:         "Glyph Tracing Sequence",
			FlavorText:   "Manual glyph tracing to stabilize the weave. Requires focus; failure increases entropic bleed.",
			ConditionKeys: []string{},
			MutationKeys:  map[string]any{"entropy": 12},
			FallbackKeys:  map[string]any{"entropy": 18},
		},
		{
			IDPrefix:     "junk_ether_",
			TriggerEvent: "distill_ether",
			Name:         "Ether Distillation",
			FlavorText:   "Distills raw ether into usable mana. The process is inefficient and produces volatile byproducts.",
			ConditionKeys: []string{},
			MutationKeys:  map[string]any{"entropy": 20},
			FallbackKeys:  map[string]any{"entropy": 25},
		},
	},
	Cyberpunk: {
		{
			IDPrefix:     "junk_diag_",
			TriggerEvent: "run_diagnostic",
			Name:         "System Diagnostic",
			FlavorText:   "Full system diagnostic sweep. Logs extensive telemetry, increasing detection risk.",
			ConditionKeys: []string{},
			MutationKeys:  map[string]any{"entropy": 10},
			FallbackKeys:  map[string]any{"entropy": 15},
		},
		{
			IDPrefix:     "junk_firewall_",
			TriggerEvent: "ping_firewall",
			Name:         "Firewall Ping Test",
			FlavorText:   "Sends ICMP probes through the corporate firewall. Each ping increments the intrusion counter.",
			ConditionKeys: []string{},
			MutationKeys:  map[string]any{"entropy": 8},
			FallbackKeys:  map[string]any{"entropy": 12},
		},
		{
			IDPrefix:     "junk_cache_",
			TriggerEvent: "flush_cache",
			Name:         "Cache Flush",
			FlavorText:   "Flushes L3 cache to clear stale pointers. Triggers garbage collection spikes visible to monitors.",
			ConditionKeys: []string{},
			MutationKeys:  map[string]any{"entropy": 12},
			FallbackKeys:  map[string]any{"entropy": 18},
		},
		{
			IDPrefix:     "junk_handshake_",
			TriggerEvent: "force_handshake",
			Name:         "Force TLS Handshake",
			FlavorText:   "Renegotiates encrypted tunnel. Handshake timing side-channels leak entropy to passive listeners.",
			ConditionKeys: []string{},
			MutationKeys:  map[string]any{"entropy": 15},
			FallbackKeys:  map[string]any{"entropy": 20},
		},
		{
			IDPrefix:     "junk_log_",
			TriggerEvent: "rotate_logs",
			Name:         "Log Rotation",
			FlavorText:   "Rotates audit logs to hide access patterns. The rotation event itself is logged at a higher privilege level.",
			ConditionKeys: []string{},
			MutationKeys:  map[string]any{"entropy": 10},
			FallbackKeys:  map[string]any{"entropy": 15},
		},
	},
	Cosmic: {
		{
			IDPrefix:     "junk_observe_",
			TriggerEvent: "observe_void",
			Name:         "Void Observation",
			FlavorText:   "Direct observation of the void collapses local probability waves. Increases quantum decoherence.",
			ConditionKeys: []string{},
			MutationKeys:  map[string]any{"entropy": 12},
			FallbackKeys:  map[string]any{"entropy": 18},
		},
		{
			IDPrefix:     "junk_collapse_",
			TriggerEvent: "collapse_wave",
			Name:         "Wave Function Collapse",
			FlavorText:   "Forces a quantum state decision. The universe resists; entropy is the price.",
			ConditionKeys: []string{},
			MutationKeys:  map[string]any{"entropy": 15},
			FallbackKeys:  map[string]any{"entropy": 22},
		},
		{
			IDPrefix:     "junk_anchor_",
			TriggerEvent: "deploy_anchor",
			Name:         "Reality Anchor Deployment",
			FlavorText:   "Deploys a localized reality anchor to stabilize spacetime. The anchor's mass creates temporal drag.",
			ConditionKeys: []string{},
			MutationKeys:  map[string]any{"entropy": 10},
			FallbackKeys:  map[string]any{"entropy": 15},
		},
		{
			IDPrefix:     "junk_hawking_",
			TriggerEvent: "emit_hawking",
			Name:         "Hawking Radiation Emitter",
			FlavorText:   "Simulates blackbody radiation to mask the ship's thermal signature. Decoheres nearby quantum states.",
			ConditionKeys: []string{},
			MutationKeys:  map[string]any{"entropy": 18},
			FallbackKeys:  map[string]any{"entropy": 25},
		},
		{
			IDPrefix:     "junk_entangle_",
			TriggerEvent: "break_entanglement",
			Name:         "Entanglement Severance",
			FlavorText:   "Severs quantum entanglement pairs to prevent tracking. The backlash ripples through local causality.",
			ConditionKeys: []string{},
			MutationKeys:  map[string]any{"entropy": 20},
			FallbackKeys:  map[string]any{"entropy": 28},
		},
	},
}

func GenerateJunkFlaws(paradigm Paradigm, count int, rng *rand.Rand) []Flaw {
	templates := JunkFlawCatalog[paradigm]
	if len(templates) == 0 || count <= 0 {
		return nil
	}

	var junk []Flaw
	for i := 0; i < count; i++ {
		tmpl := templates[rng.Intn(len(templates))]
		id := tmpl.IDPrefix + randomHex(4, rng)

		cond := ParadigmState{}
		for _, k := range tmpl.ConditionKeys {
			cond[k] = true
		}

		mut := ParadigmState{}
		for k, v := range tmpl.MutationKeys {
			mut[k] = v
		}

		fallback := ParadigmState{}
		for k, v := range tmpl.FallbackKeys {
			fallback[k] = v
		}

		junk = append(junk, Flaw{
			ID:                id,
			TriggerEvent:      tmpl.TriggerEvent,
			Name:              tmpl.Name,
			FlavorText:        tmpl.FlavorText,
			Conditions:        cond,
			Mutations:         mut,
			FallbackMutations: fallback,
		})
	}
	return junk
}

func randomHex(n int, rng *rand.Rand) string {
	const hexChars = "0123456789abcdef"
	b := make([]byte, n)
	for i := range b {
		b[i] = hexChars[rng.Intn(len(hexChars))]
	}
	return string(b)
}

type HydratorConfig struct {
	Seed           int64
	Luck           float64
	BaseDefinition *ChallengeDefinition
}

func HydrateChallenge(cfg HydratorConfig) *ChallengeDefinition {
	if cfg.BaseDefinition == nil {
		return nil
	}

	var rng *rand.Rand
	if cfg.Seed != 0 {
		rng = rand.New(rand.NewSource(cfg.Seed))
	} else {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	luck := cfg.Luck
	if luck < 0 {
		luck = 0
	}
	if luck > 1 {
		luck = 1
	}

	def := *cfg.BaseDefinition

	maxJunk := 5
	minJunk := 1
	junkCount := minJunk + int((1.0-luck)*float64(maxJunk-minJunk))
	if junkCount > maxJunk {
		junkCount = maxJunk
	}

	junkFlaws := GenerateJunkFlaws(def.Paradigm, junkCount, rng)
	def.Flaws = append(def.Flaws, junkFlaws...)

	rng.Shuffle(len(def.Flaws), func(i, j int) {
		def.Flaws[i], def.Flaws[j] = def.Flaws[j], def.Flaws[i]
	})

	return &def
}