package simulation

import "testing"

// TestSetStateRoundTrip verifies that GetState → SetState produces identical
// WorldHash on a fresh engine. This is the core replay correctness guarantee.
func TestSetStateRoundTrip(t *testing.T) {
	const seed = 42

	// Run a reference engine for 30 ticks.
	ref := NewEngineWithSeed(20, 20, 8, seed)
	for i := 0; i < 30; i++ {
		ref.BeginTick()
		ref.CommitTick()
	}
	wantHash := ref.WorldHash()
	snapshot := ref.GetState()

	// Restore into a fresh engine and verify hash matches.
	restored := NewEngineWithSeed(20, 20, 8, seed) // same capacity, different state
	restored.SetState(snapshot)
	gotHash := restored.WorldHash()

	if wantHash != gotHash {
		t.Fatalf("WorldHash after SetState mismatch: want %016x got %016x", wantHash, gotHash)
	}
}

// TestSetStateResumesSimulation verifies that SetState from a mid-run snapshot and
// then continuing to simulate produces the same output as an uninterrupted run.
// This is what replay reconstruction depends on.
func TestSetStateResumesSimulation(t *testing.T) {
	const seed = 99
	const checkpointAt = 20
	const resumeFor = 10

	// Control engine: runs uninterrupted for checkpointAt+resumeFor ticks.
	control := NewEngineWithSeed(20, 20, 8, seed)
	for i := 0; i < checkpointAt+resumeFor; i++ {
		control.BeginTick()
		control.CommitTick()
	}

	// Checkpointed engine: runs to checkpointAt, snapshots, then resumes.
	chk := NewEngineWithSeed(20, 20, 8, seed)
	for i := 0; i < checkpointAt; i++ {
		chk.BeginTick()
		chk.CommitTick()
	}
	snap := chk.GetState()

	// Restore from snapshot into a new engine and run for resumeFor more ticks.
	resumed := NewEngineWithSeed(20, 20, 8, seed)
	resumed.SetState(snap)
	for i := 0; i < resumeFor; i++ {
		resumed.BeginTick()
		resumed.CommitTick()
	}

	if control.WorldHash() != resumed.WorldHash() {
		t.Fatalf("resumed simulation diverged: control hash %016x resumed hash %016x",
			control.WorldHash(), resumed.WorldHash())
	}
}

// TestWorldHashDeterminism verifies that two engines with the same seed always
// produce identical WorldHash at every tick. If this fails, the simulation itself
// is non-deterministic — which would invalidate replay entirely.
func TestWorldHashDeterminism(t *testing.T) {
	const seed = 7
	a := NewEngineWithSeed(20, 20, 8, seed)
	b := NewEngineWithSeed(20, 20, 8, seed)

	for tick := 1; tick <= 50; tick++ {
		a.BeginTick()
		a.CommitTick()
		b.BeginTick()
		b.CommitTick()

		if a.WorldHash() != b.WorldHash() {
			t.Fatalf("WorldHash diverged at tick %d", tick)
		}
	}
}

// TestRNGRestoreProducesIdenticalSequence verifies that restoreDetRNG(seed, calls)
// produces the same future draws as the original RNG at the same call count.
func TestRNGRestoreProducesIdenticalSequence(t *testing.T) {
	const seed = 1234
	const advanceCalls = 77

	// Advance original
	orig := newDetRNG(seed)
	for i := 0; i < advanceCalls; i++ {
		orig.Intn(100)
	}
	_, calls := orig.snapshot()

	// Restore at same position
	rest := restoreDetRNG(seed, calls)

	// Next 20 draws must be identical
	for i := 0; i < 20; i++ {
		a := orig.Intn(100)
		b := rest.Intn(100)
		if a != b {
			t.Fatalf("RNG draw %d diverged after restore: orig=%d restored=%d", i, a, b)
		}
	}
}

// --- WorldHash coverage tests ---
// Each test mutates exactly one canonical field and verifies the hash changes.
// If a test fails it means WorldHash has a coverage gap for that field.

func TestWorldHashCoversGrid(t *testing.T) {
	e := NewEngineWithSeed(10, 10, 2, 1)
	base := e.WorldHash()
	e.Grid.CurrentCells[0].HomePheromone = 99
	if e.WorldHash() == base {
		t.Fatal("WorldHash did not change after grid mutation")
	}
}

func TestWorldHashCoversDroneCompromised(t *testing.T) {
	e := NewEngineWithSeed(10, 10, 2, 1)
	base := e.WorldHash()
	e.Registry.Compromised[0] = true
	if e.WorldHash() == base {
		t.Fatal("WorldHash did not change after Compromised mutation")
	}
}

func TestWorldHashCoversHazards(t *testing.T) {
	e := NewEngineWithSeed(10, 10, 2, 1)
	base := e.WorldHash()
	e.Hazards.Intensity[0] += 1 // Hazard in slot 0 is active (added in NewEngineWithSeed)
	if e.WorldHash() == base {
		t.Fatal("WorldHash did not change after hazard intensity mutation")
	}
}

func TestWorldHashCoversAliens(t *testing.T) {
	e := NewEngineWithSeed(10, 10, 2, 1)
	base := e.WorldHash()
	e.Aliens.Radius[0]++ // Alien in slot 0 is active
	if e.WorldHash() == base {
		t.Fatal("WorldHash did not change after alien radius mutation")
	}
}

func TestWorldHashCoversMissionStatus(t *testing.T) {
	e := NewEngineWithSeed(10, 10, 2, 1)
	base := e.WorldHash()
	e.Mission.Status = MissionVictory
	if e.WorldHash() == base {
		t.Fatal("WorldHash did not change after mission status mutation")
	}
}

func TestWorldHashCoversRNGPosition(t *testing.T) {
	e := NewEngineWithSeed(10, 10, 2, 1)
	base := e.WorldHash()
	e.rng.Intn(2) // advance RNG by one draw
	if e.WorldHash() == base {
		t.Fatal("WorldHash did not change after RNG advancement")
	}
}

func TestWorldHashCoversEconomy(t *testing.T) {
	e := NewEngineWithSeed(10, 10, 2, 1)
	base := e.WorldHash()
	e.TotalDeposited++
	if e.WorldHash() == base {
		t.Fatal("WorldHash did not change after TotalDeposited mutation")
	}
}

// TestSetStatePreservesRNGPosition verifies that after SetState, the engine's RNG
// produces the same sequence it would have at that tick in the original run.
// Non-determinism in RNG restoration causes gradual replay divergence.
func TestSetStatePreservesRNGPosition(t *testing.T) {
	const seed = 55

	ref := NewEngineWithSeed(20, 20, 4, seed)
	for i := 0; i < 15; i++ {
		ref.BeginTick()
		ref.CommitTick()
	}
	snap := ref.GetState()

	// Restore and record the next 5 MoveRandom draws
	clone := NewEngineWithSeed(20, 20, 4, seed)
	clone.SetState(snap)

	refDraws := make([]int, 5)
	cloneDraws := make([]int, 5)
	for i := range refDraws {
		refDraws[i] = ref.rng.Intn(100)
		cloneDraws[i] = clone.rng.Intn(100)
	}

	for i := range refDraws {
		if refDraws[i] != cloneDraws[i] {
			t.Fatalf("RNG draw %d after SetState: ref=%d clone=%d", i, refDraws[i], cloneDraws[i])
		}
	}
}
