package simulation

import (
	"testing"

	"chrysalis-engine/core/crysmath"
)

func TestEngineWithSameSeedIsDeterministic(t *testing.T) {
	a := NewEngineWithSeed(20, 20, 8, 42)
	b := NewEngineWithSeed(20, 20, 8, 42)

	for tick := 0; tick < 100; tick++ {
		a.Step()
		b.Step()
	}

	if a.Tick != b.Tick || a.GlobalSilicates != b.GlobalSilicates {
		t.Fatalf("scalar state diverged: a=(%d,%d) b=(%d,%d)",
			a.Tick, a.GlobalSilicates, b.Tick, b.GlobalSilicates)
	}
	for i := 0; i < a.Registry.Count; i++ {
		if a.Registry.PositionX[i] != b.Registry.PositionX[i] ||
			a.Registry.PositionY[i] != b.Registry.PositionY[i] ||
			a.Registry.Battery[i] != b.Registry.Battery[i] ||
			a.Registry.State[i] != b.Registry.State[i] ||
			a.Registry.Compromised[i] != b.Registry.Compromised[i] ||
			a.Registry.CorruptionFactor[i] != b.Registry.CorruptionFactor[i] {
			t.Fatalf("drone %d diverged for identical seed", i)
		}
	}
}

func TestBeginTickRunsHazardsAndFabrication(t *testing.T) {
	e := NewEngineWithSeed(20, 20, 1, 7)
	center := 10
	e.Hazards = NewHazardSystem(1)
	e.Hazards.Add(HazardMagnetic, int32(center), int32(center), 1, 100*crysmath.Precision)
	e.GlobalSilicates = FabricationThreshold

	e.BeginTick()

	if e.Registry.State[0] != StateInert || e.Registry.Battery[0] != 0 {
		t.Fatalf("hazard pass did not inert drone: state=%d battery=%d",
			e.Registry.State[0], e.Registry.Battery[0])
	}
	if e.Registry.Count != 2 || e.GlobalSilicates != 0 {
		t.Fatalf("fabrication pass failed: count=%d silicates=%d",
			e.Registry.Count, e.GlobalSilicates)
	}
}

func TestInertDroneCannotAct(t *testing.T) {
	e := NewEngineWithSeed(20, 20, 1, 9)
	e.Registry.State[0] = StateInert
	e.Registry.Battery[0] = 0
	beforeX := e.Registry.PositionX[0]
	beforeY := e.Registry.PositionY[0]

	e.MoveRandom(0)
	e.MoveTowardsResource(0)
	e.MoveTowardsHome(0)
	e.Harvest(0)
	e.DropResource(0)

	if e.Registry.PositionX[0] != beforeX || e.Registry.PositionY[0] != beforeY {
		t.Fatal("inert drone moved")
	}
}

func TestCargoLoopReturnsResourceToBase(t *testing.T) {
	e := NewEngineWithSeed(9, 9, 1, 11)
	baseX, baseY := 4, 4
	resourceX, resourceY := 5, 4
	resourceIdx := e.Grid.GetIndex(resourceX, resourceY)
	e.Grid.CurrentCells[resourceIdx].ResourceCount = 1
	e.Grid.NextCells[resourceIdx].ResourceCount = 1
	e.Registry.PositionX[0] = crysmath.NewFixedPoint(int64(resourceX))
	e.Registry.PositionY[0] = crysmath.NewFixedPoint(int64(resourceY))

	e.BeginTick()
	e.Harvest(0)
	if !e.SenseCargo(0) {
		t.Fatal("drone failed to harvest")
	}
	e.MoveTowardsHome(0)
	e.CommitTick()

	x := int(e.Registry.PositionX[0].V / crysmath.Precision)
	y := int(e.Registry.PositionY[0].V / crysmath.Precision)
	if x != baseX || y != baseY {
		t.Fatalf("drone did not follow home gradient: got (%d,%d)", x, y)
	}

	e.BeginTick()
	e.DropResource(0)
	e.CommitTick()
	if e.GlobalSilicates != 1 || e.SenseCargo(0) {
		t.Fatalf("resource was not deposited: colony=%d cargo=%v",
			e.GlobalSilicates, e.SenseCargo(0))
	}
}
