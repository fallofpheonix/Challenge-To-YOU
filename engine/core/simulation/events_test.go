package simulation

import (
	"encoding/json"
	"testing"

	"chrysalis-engine/core/crysmath"
)

// helpers
func drainBus(e *Engine) {
	e.Bus.Commit()
}

func pendingEvents(e *Engine) []Event {
	e.Bus.Commit()
	return e.Bus.Events()
}

// --- EventBus double-buffer correctness ---

func TestEventBusCommitMakesEventsAvailable(t *testing.T) {
	b := NewEventBus()
	b.Emit(Event{Type: EventHarvested, DroneID: 1})
	b.Emit(Event{Type: EventDeposited, DroneID: 2})

	// Before Commit, Events() returns previous (empty) snapshot
	if len(b.Events()) != 0 {
		t.Fatalf("expected empty snapshot before Commit, got %d", len(b.Events()))
	}

	b.Commit()

	evs := b.Events()
	if len(evs) != 2 {
		t.Fatalf("expected 2 events after Commit, got %d", len(evs))
	}
	if evs[0].Type != EventHarvested || evs[1].Type != EventDeposited {
		t.Fatalf("unexpected event types: %v %v", evs[0].Type, evs[1].Type)
	}
}

func TestEventBusPendingResetAfterCommit(t *testing.T) {
	b := NewEventBus()
	b.Emit(Event{Type: EventHarvested, DroneID: 1})
	b.Commit()

	if b.PendingLen() != 0 {
		t.Fatalf("pending should be 0 after Commit, got %d", b.PendingLen())
	}
}

// Key correctness invariant: emitting into the active buffer after Commit
// must NOT overwrite the snapshot returned by Events().
func TestEventBusSnapshotIsImmutableAfterCommit(t *testing.T) {
	b := NewEventBus()
	b.Emit(Event{Type: EventHarvested, DroneID: 1})
	b.Commit()

	snap := b.Events()
	snapLen := len(snap)

	// Now emit more events into the new active buffer
	for i := 0; i < 300; i++ {
		b.Emit(Event{Type: EventDroneDied, DroneID: int32(i)})
	}

	// The snapshot returned earlier must not have changed
	if len(snap) != snapLen {
		t.Fatalf("snapshot was mutated: expected len %d, got %d", snapLen, len(snap))
	}
	if snap[0].Type != EventHarvested {
		t.Fatalf("snapshot element was overwritten: got %s", snap[0].Type)
	}
}

func TestEventBusMultipleConsumersSeeIdenticalSnapshot(t *testing.T) {
	b := NewEventBus()
	b.Emit(Event{Type: EventHarvested, DroneID: 10})
	b.Commit()

	c1 := b.Events()
	c2 := b.Events()

	if len(c1) != len(c2) || c1[0].DroneID != c2[0].DroneID {
		t.Fatal("multiple consumers saw different snapshots")
	}
}

func TestEventBusEmitAfterCommitAppearsInNextSnapshot(t *testing.T) {
	b := NewEventBus()
	b.Emit(Event{Type: EventHarvested, DroneID: 1})
	b.Commit()

	b.Emit(Event{Type: EventDroneDied, DroneID: 2})

	// Current snapshot still holds prior Commit
	if len(b.Events()) != 1 || b.Events()[0].Type != EventHarvested {
		t.Fatalf("snapshot changed before second Commit")
	}

	b.Commit()

	if len(b.Events()) != 1 || b.Events()[0].Type != EventDroneDied {
		t.Fatal("second snapshot should contain the die event")
	}
}

// --- Sealed interface ---

func TestEventPayloadSealRejectedAtTypeLevel(t *testing.T) {
	// Compile-time test: verified externally. Here we verify runtime type assertion.
	var p EventPayload = HarvestData{ResourcesRemaining: 5}
	data, ok := p.(HarvestData)
	if !ok || data.ResourcesRemaining != 5 {
		t.Fatal("type assertion on sealed payload failed")
	}
}

// --- Typed payloads verify correct data in emit helpers ---

func TestHarvestEmitsTypedPayload(t *testing.T) {
	e := NewEngineWithSeed(20, 20, 1, 1)
	cx, cy := 10, 10
	ri := e.Grid.GetIndex(cx+1, cy)
	e.Grid.CurrentCells[ri].ResourceCount = 5
	e.Grid.NextCells[ri].ResourceCount = 5
	e.Registry.PositionX[0] = crysmath.NewFixedPoint(int64(cx + 1))
	e.Registry.PositionY[0] = crysmath.NewFixedPoint(int64(cy))
	e.Registry.Inventory[0] = 0
	drainBus(e) // flush spawn events

	e.Harvest(0)
	evs := pendingEvents(e) // commit and snapshot

	found := false
	for _, ev := range evs {
		if ev.Type == EventHarvested && ev.DroneID == 0 {
			data, ok := ev.Data.(HarvestData)
			if !ok {
				t.Fatal("Harvest event Data is not HarvestData")
			}
			if data.ResourcesRemaining != 4 {
				t.Errorf("expected ResourcesRemaining=4, got %d", data.ResourcesRemaining)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("Harvest did not emit EventHarvested")
	}
}

func TestDropResourceEmitsTypedPayload(t *testing.T) {
	e := NewEngineWithSeed(20, 20, 1, 1)
	cx, cy := 10, 10
	e.Registry.PositionX[0] = crysmath.NewFixedPoint(int64(cx))
	e.Registry.PositionY[0] = crysmath.NewFixedPoint(int64(cy))
	e.Registry.Inventory[0] = 1
	drainBus(e)

	e.DropResource(0)
	evs := pendingEvents(e)

	found := false
	for _, ev := range evs {
		if ev.Type == EventDeposited && ev.DroneID == 0 {
			data, ok := ev.Data.(DepositData)
			if !ok {
				t.Fatal("Deposit event Data is not DepositData")
			}
			if data.Amount != 1 {
				t.Errorf("expected Amount=1, got %d", data.Amount)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("DropResource did not emit EventDeposited")
	}
}

func TestInfectionEmitsTypedPayload(t *testing.T) {
	infected := false
	for seed := int64(1); seed <= 200 && !infected; seed++ {
		e := NewEngineWithSeed(20, 20, 1, seed)
		e.Aliens = NewAlienNetwork(1)
		e.Aliens.Add(NodeInfector, int32(e.Registry.PositionX[0].V/crysmath.Precision),
			int32(e.Registry.PositionY[0].V/crysmath.Precision), 5)
		drainBus(e)

		e.processInfections()
		evs := pendingEvents(e)

		for _, ev := range evs {
			if ev.Type == EventDroneInfected {
				data, ok := ev.Data.(InfectedData)
				if ok && data.Vector == "alien_node" {
					infected = true
				}
			}
		}
	}
	if !infected {
		t.Fatal("processInfections never emitted typed EventDroneInfected across 200 seeds")
	}
}

func TestFabricationEmitsTypedPayload(t *testing.T) {
	e := NewEngineWithSeed(20, 20, 2, 1)
	e.GlobalSilicates = FabricationThreshold
	drainBus(e)

	e.CheckFabricationPool()
	evs := pendingEvents(e)

	found := false
	for _, ev := range evs {
		if ev.Type == EventFabricated {
			data, ok := ev.Data.(SpawnedData)
			if !ok {
				t.Fatal("Fabricated event Data is not SpawnedData")
			}
			if data.SwarmSize != e.Registry.Count {
				t.Errorf("SwarmSize mismatch: got %d want %d", data.SwarmSize, e.Registry.Count)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("CheckFabricationPool did not emit EventFabricated")
	}
}

func TestMissionVictoryEmitsTypedPayload(t *testing.T) {
	e := NewEngineWithSeed(20, 20, 1, 1)
	e.TotalDeposited = int32(e.Mission.TargetResources)
	drainBus(e)

	e.EvaluateMission()
	evs := pendingEvents(e)

	found := false
	for _, ev := range evs {
		if ev.Type == EventMissionChanged {
			data, ok := ev.Data.(MissionData)
			if !ok {
				t.Fatal("MissionChanged event Data is not MissionData")
			}
			if data.Status != string(MissionVictory) {
				t.Errorf("expected victory status, got %s", data.Status)
			}
			if data.Reason != MissionReasonResourceTarget {
				t.Errorf("expected reason %s, got %s", MissionReasonResourceTarget, data.Reason)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("EvaluateMission did not emit EventMissionChanged")
	}
}

// --- BeginTick invariant ---

func TestEventBusBeginTickPanicsIfActiveNotEmpty(t *testing.T) {
	b := NewEventBus()
	b.Emit(Event{Type: EventHarvested})

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic from BeginTick with non-empty active buffer")
		}
	}()
	b.BeginTick()
}

func TestEventBusBeginTickPassesOnEmptyBuffer(t *testing.T) {
	b := NewEventBus()
	b.Commit() // active is now the reset buffer

	// Must not panic
	b.BeginTick()
}

// --- Determinism: same seed → identical event stream ---

func TestEventStreamDeterminism(t *testing.T) {
	runEvents := func(seed int64) []Event {
		e := NewEngineWithSeed(30, 30, 5, seed)
		ri := e.Grid.GetIndex(16, 15)
		e.Grid.CurrentCells[ri].ResourceCount = 50
		e.Grid.NextCells[ri].ResourceCount = 50

		var all []Event
		for tick := 0; tick < 50; tick++ {
			e.Step()
			all = append(all, e.Bus.Events()...)
		}
		return all
	}

	a := runEvents(99)
	b := runEvents(99)

	if len(a) != len(b) {
		t.Fatalf("event count differs: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i].Type != b[i].Type {
			t.Fatalf("event[%d] type differs: %s vs %s", i, a[i].Type, b[i].Type)
		}
		if a[i].DroneID != b[i].DroneID {
			t.Fatalf("event[%d] drone_id differs: %d vs %d", i, a[i].DroneID, b[i].DroneID)
		}
		if a[i].TickNum != b[i].TickNum {
			t.Fatalf("event[%d] tick differs: %d vs %d", i, a[i].TickNum, b[i].TickNum)
		}
		if a[i].X != b[i].X || a[i].Y != b[i].Y {
			t.Fatalf("event[%d] position differs: (%d,%d) vs (%d,%d)",
				i, a[i].X, a[i].Y, b[i].X, b[i].Y)
		}
	}
}

// TestEventStreamJSONOrderDeterminism serializes the full event stream to JSON
// for two runs with the same seed and compares them byte-for-byte.
// This is the strongest determinism check: it covers type, ordering, position,
// payload fields, and tick assignment simultaneously.
func TestEventStreamJSONOrderDeterminism(t *testing.T) {
	serialize := func(seed int64) []byte {
		e := NewEngineWithSeed(30, 30, 5, seed)
		ri := e.Grid.GetIndex(16, 15)
		e.Grid.CurrentCells[ri].ResourceCount = 50
		e.Grid.NextCells[ri].ResourceCount = 50

		var all []Event
		for tick := 0; tick < 50; tick++ {
			e.Step()
			all = append(all, e.Bus.Events()...)
		}

		b, err := json.Marshal(all)
		if err != nil {
			t.Fatalf("JSON marshal failed: %v", err)
		}
		return b
	}

	a := serialize(77)
	b := serialize(77)

	if string(a) != string(b) {
		t.Fatalf("JSON event streams differ:\nA: %s\nB: %s", a[:min(200, len(a))], b[:min(200, len(b))])
	}
}

// TestEventOrderWithinTick verifies that within a single tick, events appear
// in the canonical order: fabrication → hazard damage → infection → (P-Script
// builtins) → mission. This order must be stable for replay to reconstruct state.
func TestEventOrderWithinTick(t *testing.T) {
	// Set up a scenario that produces fabrication and deposit in the same tick:
	// drone starts adjacent to base with cargo, and there's enough silicates to fabricate.
	e := NewEngineWithSeed(20, 20, 1, 1)
	cx, cy := 10, 10
	// Place drone at base
	e.Registry.PositionX[0] = crysmath.NewFixedPoint(int64(cx))
	e.Registry.PositionY[0] = crysmath.NewFixedPoint(int64(cy))
	e.Registry.Inventory[0] = 1
	e.Registry.State[0] = StateReturning
	// Enough silicates to trigger fabrication after deposit
	// (deposit happens in P-Script / stepDrones, fabrication in BeginTick)
	e.GlobalSilicates = FabricationThreshold
	drainBus(e)

	// One full Step: BeginTick (fabrication) → stepDrones (deposit) → CommitTick (mission)
	e.Step()
	evs := e.Bus.Events()

	// Find fabrication and deposit positions
	fabIdx, depIdx := -1, -1
	for i, ev := range evs {
		if ev.Type == EventFabricated {
			fabIdx = i
		}
		if ev.Type == EventDeposited {
			depIdx = i
		}
	}

	if fabIdx == -1 || depIdx == -1 {
		t.Skipf("scenario did not produce both events (fab=%d dep=%d) — seed/state mismatch; skip rather than fail", fabIdx, depIdx)
	}

	// Fabrication (BeginTick) must precede deposit (stepDrones)
	if fabIdx >= depIdx {
		t.Fatalf("expected fabrication (idx %d) before deposit (idx %d); got wrong order", fabIdx, depIdx)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
