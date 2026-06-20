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
	if e.TotalDeposited != 1 {
		t.Fatalf("total deposited was not tracked: got %d", e.TotalDeposited)
	}
}

func TestMissionVictoryPrecedesSameTickDefeat(t *testing.T) {
	e := NewEngineWithSeed(9, 9, 2, 13)
	e.Mission.TargetResources = 1
	e.Mission.InfectionLossThreshold = 0.50
	e.TotalDeposited = 1
	e.Registry.Compromised[0] = true
	e.Tick = e.Mission.MaxTicks

	e.EvaluateMission()

	if e.Mission.Status != MissionVictory || e.Mission.Reason != MissionReasonResourceTarget {
		t.Fatalf("mission did not prefer victory: status=%s reason=%s",
			e.Mission.Status, e.Mission.Reason)
	}
}

func TestMissionDoesNotInstantLoseToInitialInfection(t *testing.T) {
	e := NewEngineWithSeed(9, 9, 2, 17)
	e.Mission.InfectionLossThreshold = 0.50
	e.Registry.Compromised[0] = true

	e.EvaluateMission()

	if e.Mission.Status != MissionRunning {
		t.Fatalf("mission lost at tick zero: status=%s reason=%s",
			e.Mission.Status, e.Mission.Reason)
	}

	e.Tick = 1
	e.EvaluateMission()
	if e.Mission.Status != MissionDefeat || e.Mission.Reason != MissionReasonInfectionExceeded {
		t.Fatalf("mission did not lose after tick zero: status=%s reason=%s",
			e.Mission.Status, e.Mission.Reason)
	}
}

func TestAdjacentCarrierReturnsAndDeposits(t *testing.T) {
	e := NewEngineWithSeed(9, 9, 1, 19)
	e.Hazards = NewHazardSystem(0)
	e.Aliens = NewAlienNetwork(0)
	e.Registry.PositionX[0] = crysmath.NewFixedPoint(5)
	e.Registry.PositionY[0] = crysmath.NewFixedPoint(4)
	e.Registry.Inventory[0] = 1
	e.Registry.State[0] = StateReturning

	e.BeginTick()
	e.MoveTowardsHome(0)
	e.CommitTick()

	x := int(e.Registry.PositionX[0].V / crysmath.Precision)
	y := int(e.Registry.PositionY[0].V / crysmath.Precision)
	if x != 4 || y != 4 {
		t.Fatalf("adjacent carrier did not move to base: got (%d,%d)", x, y)
	}

	e.BeginTick()
	e.DropResource(0)
	e.CommitTick()

	if e.TotalDeposited != 1 || e.GlobalSilicates != 1 || e.Registry.Inventory[0] != 0 {
		t.Fatalf("adjacent carrier did not deposit: total=%d colony=%d cargo=%d",
			e.TotalDeposited, e.GlobalSilicates, e.Registry.Inventory[0])
	}
}

func TestLoadedDroneReturnsWithinBoundWithoutHomeTrail(t *testing.T) {
	cases := []struct {
		name string
		x    int
		y    int
	}{
		{name: "straight_distance_5", x: 55, y: 50},
		{name: "straight_distance_10", x: 60, y: 50},
		{name: "diagonal_distance_5", x: 55, y: 55},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEngineWithSeed(100, 100, 1, 23)
			e.Hazards = NewHazardSystem(0)
			e.Aliens = NewAlienNetwork(0)
			e.Registry.PositionX[0] = crysmath.NewFixedPoint(int64(tc.x))
			e.Registry.PositionY[0] = crysmath.NewFixedPoint(int64(tc.y))
			e.Registry.Inventory[0] = 1
			e.Registry.State[0] = StateReturning

			baseX, baseY := e.Grid.Width/2, e.Grid.Height/2
			distance := chebyshevDistance(tc.x, tc.y, baseX, baseY)
			maxTicks := 2*distance + 4

			for tick := 0; tick < maxTicks && e.TotalDeposited == 0; tick++ {
				e.BeginTick()
				e.MoveTowardsHome(0)
				e.CommitTick()
			}

			if e.TotalDeposited != 1 {
				x := int(e.Registry.PositionX[0].V / crysmath.Precision)
				y := int(e.Registry.PositionY[0].V / crysmath.Precision)
				t.Fatalf("loaded drone did not deposit within %d ticks: start=(%d,%d) end=(%d,%d) cargo=%d",
					maxTicks, tc.x, tc.y, x, y, e.Registry.Inventory[0])
			}
		})
	}
}

func TestReturningDroneIgnoresHomeTrailThatDoesNotImproveDistance(t *testing.T) {
	e := NewEngineWithSeed(100, 100, 1, 29)
	e.Hazards = NewHazardSystem(0)
	e.Aliens = NewAlienNetwork(0)
	e.Registry.PositionX[0] = crysmath.NewFixedPoint(60)
	e.Registry.PositionY[0] = crysmath.NewFixedPoint(50)
	e.Registry.Inventory[0] = 1
	e.Registry.State[0] = StateReturning

	awayIdx := e.Grid.GetIndex(61, 50)
	e.Grid.CurrentCells[awayIdx].HomePheromone = MaxPheromone
	e.Grid.NextCells[awayIdx].HomePheromone = MaxPheromone

	e.BeginTick()
	e.MoveTowardsHome(0)
	e.CommitTick()

	x := int(e.Registry.PositionX[0].V / crysmath.Precision)
	y := int(e.Registry.PositionY[0].V / crysmath.Precision)
	if x != 59 || y != 50 {
		t.Fatalf("returning drone followed non-improving stale trail: got (%d,%d)", x, y)
	}
}
