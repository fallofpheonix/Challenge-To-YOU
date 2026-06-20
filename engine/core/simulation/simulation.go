package simulation

import (
	"chrysalis-engine/core/crysmath"
	"fmt"
	"math/rand"
)

const (
	FabricationThreshold int32 = 5   // 5 silicates required to construct a new unit
	MaxSwarmCapacity     int   = 500 // Safety cap for the MVP engine pass
	DefaultWorldSeed     int64 = 1
)

type Engine struct {
	Grid            *Grid
	Registry        *SwarmRegistry
	Hazards         *HazardSystem
	Aliens          *AlienNetwork
	Mission         MissionState
	Tick            int64
	GlobalSilicates int32
	TotalDeposited  int32
	HistoricalTotal int32
	rng             *rand.Rand
}

func NewEngine(width, height int, droneCount int) *Engine {
	return NewEngineWithSeed(width, height, droneCount, DefaultWorldSeed)
}

func NewEngineWithSeed(width, height int, droneCount int, seed int64) *Engine {
	e := &Engine{
		Grid:     NewGrid(width, height),
		Registry: NewSwarmRegistry(droneCount),
		Hazards:  NewHazardSystem(10), // Support up to 10 active hazards
		Aliens:   NewAlienNetwork(5),  // Support up to 5 alien nodes
		Mission:  NewDefaultMissionState(),
		rng:      rand.New(rand.NewSource(seed)),
	}

	// Initialize base in the center
	centerX, centerY := width/2, height/2
	idx := e.Grid.GetIndex(centerX, centerY)
	e.Grid.CurrentCells[idx].IsBase = true
	e.Grid.NextCells[idx].IsBase = true
	// Seed initial home pheromone at base
	e.Grid.CurrentCells[idx].HomePheromone = MaxPheromone
	e.Grid.NextCells[idx].HomePheromone = MaxPheromone

	// Spawn drones at base
	for i := 0; i < droneCount; i++ {
		e.Registry.Spawn(centerX, centerY, 100*crysmath.Precision) // 100% battery
	}
	e.HistoricalTotal = int32(droneCount)

	// Add some initial hazards for testing
	e.Hazards.Add(HazardMagnetic, int32(centerX+20), int32(centerY+20), 15, 1*crysmath.Precision)

	// Add an initial alien node to spread logic virus
	e.Aliens.Add(NodeInfector, int32(centerX-20), int32(centerY-20), 12)

	return e
}

func (e *Engine) CheckFabricationPool() {
	width, height := e.Grid.Width, e.Grid.Height
	centerX, centerY := width/2, height/2

	if e.GlobalSilicates >= FabricationThreshold && e.Registry.Count < MaxSwarmCapacity {
		e.GlobalSilicates -= FabricationThreshold
		e.HistoricalTotal++

		// Trigger slice inflation inside the runtime registry
		e.Registry.Spawn(centerX, centerY, 100*crysmath.Precision)

		fmt.Printf("[REPLICATION SUCCESS] Resources Consumed. New Entity Created. Swarm Count: %d | Global Cache: %d\n",
			e.Registry.Count, e.GlobalSilicates)
	}
}

// BeginTick stages the environment and applies all involuntary systems.
// Architect logic executes after this call and before CommitTick.
func (e *Engine) BeginTick() {
	e.Grid.TickPheromones()
	e.processHazards()
	e.processInfections()
	e.SpreadsInfection()
	e.CheckFabricationPool()

	width, height := e.Grid.Width, e.Grid.Height
	idx := e.Grid.GetIndex(width/2, height/2)
	e.Grid.NextCells[idx].HomePheromone = MaxPheromone
}

func (e *Engine) CommitTick() {
	e.Grid.SwapBuffers()
	e.Tick++
	e.EvaluateMission()
}

// Step advances the built-in fallback AI by one authoritative tick.
func (e *Engine) Step() {
	e.BeginTick()
	e.stepDrones()
	e.CommitTick()
}

func (e *Engine) SenseResource(i int) bool {
	if !e.validDrone(i) {
		return false
	}
	x := int(e.Registry.PositionX[i].V / crysmath.Precision)
	y := int(e.Registry.PositionY[i].V / crysmath.Precision)
	_, _, val := e.Grid.SenseHighestGradient(x, y, true)
	return val > 0
}

func (e *Engine) SenseHome(i int) bool {
	if !e.validDrone(i) {
		return false
	}
	x := int(e.Registry.PositionX[i].V / crysmath.Precision)
	y := int(e.Registry.PositionY[i].V / crysmath.Precision)
	_, _, val := e.Grid.SenseHighestGradient(x, y, false)
	return val > 0
}

func (e *Engine) Harvest(i int) {
	if !e.canAct(i) {
		return
	}
	if e.Registry.Inventory[i] != 0 {
		return
	}
	x := int(e.Registry.PositionX[i].V / crysmath.Precision)
	y := int(e.Registry.PositionY[i].V / crysmath.Precision)
	idx := e.Grid.GetIndex(x, y)
	if e.Grid.CurrentCells[idx].ResourceCount > 0 && e.Grid.NextCells[idx].ResourceCount > 0 {
		e.Grid.NextCells[idx].ResourceCount--
		e.Registry.Inventory[i] = 1
		e.Registry.State[i] = StateReturning
	}
}

func (e *Engine) DropResource(i int) {
	if !e.canAct(i) {
		return
	}
	x := int(e.Registry.PositionX[i].V / crysmath.Precision)
	y := int(e.Registry.PositionY[i].V / crysmath.Precision)
	idx := e.Grid.GetIndex(x, y)
	if e.Grid.CurrentCells[idx].IsBase {
		if e.Registry.Inventory[i] > 0 {
			e.GlobalSilicates += e.Registry.Inventory[i]
			e.TotalDeposited += e.Registry.Inventory[i]
			e.Registry.Inventory[i] = 0
		}
		e.Registry.State[i] = StateSearching
	}
}

func (e *Engine) MoveTowardsResource(i int) {
	if !e.canAct(i) {
		return
	}
	e.stepSearching(i)
}

func (e *Engine) MoveTowardsHome(i int) {
	if !e.canAct(i) {
		return
	}
	e.stepReturning(i)
}

func (e *Engine) MoveRandom(i int) {
	if !e.canAct(i) {
		return
	}
	x := int(e.Registry.PositionX[i].V / crysmath.Precision)
	y := int(e.Registry.PositionY[i].V / crysmath.Precision)
	dx := e.rng.Intn(3) - 1
	dy := e.rng.Intn(3) - 1
	x += dx
	y += dy
	if x < 0 {
		x = 0
	}
	if x >= e.Grid.Width {
		x = e.Grid.Width - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= e.Grid.Height {
		y = e.Grid.Height - 1
	}
	e.Registry.PositionX[i] = crysmath.NewFixedPoint(int64(x))
	e.Registry.PositionY[i] = crysmath.NewFixedPoint(int64(y))
}

func (e *Engine) SenseCargo(i int) bool {
	return e.validDrone(i) && e.Registry.Inventory[i] > 0
}

func (e *Engine) GetState() map[string]interface{} {
	drones := make([]map[string]interface{}, e.Registry.Count)
	for i := 0; i < e.Registry.Count; i++ {
		drones[i] = map[string]interface{}{
			"id":    e.Registry.ID[i],
			"x":     e.Registry.PositionX[i].V,
			"y":     e.Registry.PositionY[i].V,
			"state": e.Registry.State[i],
			"inv":   e.Registry.Inventory[i],
			"bat":   e.Registry.Battery[i],
			"comp":  e.Registry.Compromised[i],
			"trust": e.Registry.TrustScore[i],
			"corr":  e.Registry.CorruptionFactor[i],
		}
	}

	// Sparse collection of active trail cells
	var activeTrails []map[string]interface{}
	for y := 0; y < e.Grid.Height; y++ {
		for x := 0; x < e.Grid.Width; x++ {
			idx := e.Grid.GetIndex(x, y)
			cell := e.Grid.CurrentCells[idx]

			// Only stream cells that have active metrics to save bandwidth
			if cell.HomePheromone > 0 || cell.ResourcePheromone > 0 || cell.ResourceCount > 0 || cell.AlienSignal > 0 {
				activeTrails = append(activeTrails, map[string]interface{}{
					"x":     x,
					"y":     y,
					"home":  cell.HomePheromone,
					"res":   cell.ResourcePheromone,
					"alien": cell.AlienSignal,
					"cnt":   cell.ResourceCount,
				})
			}
		}
	}

	// Collection of active hazards
	var activeHazards []map[string]interface{}
	for i := 0; i < e.Hazards.Capacity; i++ {
		if e.Hazards.Active[i] {
			activeHazards = append(activeHazards, map[string]interface{}{
				"type": e.Hazards.Type[i],
				"x":    e.Hazards.X[i],
				"y":    e.Hazards.Y[i],
				"rad":  e.Hazards.Radius[i],
			})
		}
	}

	// Collection of active alien nodes
	var activeAliens []map[string]interface{}
	for i := 0; i < e.Aliens.Capacity; i++ {
		if e.Aliens.Active[i] {
			activeAliens = append(activeAliens, map[string]interface{}{
				"type": e.Aliens.Type[i],
				"x":    e.Aliens.X[i],
				"y":    e.Aliens.Y[i],
				"rad":  e.Aliens.Radius[i],
			})
		}
	}

	return map[string]interface{}{
		"tick":       e.Tick,
		"drones":     drones,
		"grid":       activeTrails,
		"hazards":    activeHazards,
		"aliens":     activeAliens,
		"colony_res": e.GlobalSilicates,
		"mission":    e.Mission,
		"swarm_size": e.Registry.Count,
	}
}

func (e *Engine) processHazards() {
	for h := 0; h < e.Hazards.Capacity; h++ {
		if !e.Hazards.Active[h] {
			continue
		}

		hx := e.Hazards.X[h]
		hy := e.Hazards.Y[h]
		hr := e.Hazards.Radius[h]
		intensity := e.Hazards.Intensity[h]

		for i := 0; i < e.Registry.Count; i++ {
			dx := int32(e.Registry.PositionX[i].V/crysmath.Precision) - hx
			dy := int32(e.Registry.PositionY[i].V/crysmath.Precision) - hy

			// Simple squared distance check to avoid sqrt
			distSq := dx*dx + dy*dy
			if distSq <= hr*hr {
				// Apply mutation: Drain battery
				e.Registry.Battery[i] -= intensity
				if e.Registry.Battery[i] <= 0 {
					e.Registry.Battery[i] = 0
					e.Registry.State[i] = StateInert
				}
			}
		}
	}
}

func (e *Engine) processInfections() {
	for n := 0; n < e.Aliens.Capacity; n++ {
		if !e.Aliens.Active[n] {
			continue
		}

		nx := e.Aliens.X[n]
		ny := e.Aliens.Y[n]
		nr := e.Aliens.Radius[n]

		for i := 0; i < e.Registry.Count; i++ {
			if e.Registry.Compromised[i] {
				continue
			}

			dx := int32(e.Registry.PositionX[i].V/crysmath.Precision) - nx
			dy := int32(e.Registry.PositionY[i].V/crysmath.Precision) - ny
			distSq := dx*dx + dy*dy

			if distSq <= nr*nr {
				// 5% chance to become compromised per tick while in radius
				if e.rng.Intn(100) < 5 {
					e.Registry.Compromised[i] = true
					e.Registry.TrustScore[i] = 50 // Initial drop in trust
					fmt.Printf("[ALIEN VIRUS] Drone %d Compromised at (%d, %d)\n", i, nx, ny)
				}
			}
		}
	}
}

func (e *Engine) stepDrones() {
	for i := 0; i < e.Registry.Count; i++ {
		if e.Registry.State[i] == StateSearching {
			e.stepSearching(i)
		} else if e.Registry.State[i] == StateReturning {
			e.stepReturning(i)
		}
	}
}

func (e *Engine) SenseAlienSignal(i int) bool {
	if !e.validDrone(i) {
		return false
	}
	x := int(e.Registry.PositionX[i].V / crysmath.Precision)
	y := int(e.Registry.PositionY[i].V / crysmath.Precision)
	_, _, val := e.Grid.SenseHighestAlienGradient(x, y)
	return val > 0
}

// SenseQuorum Consensus Pass: Evaluates nearest 8-neighbor vectors for logic drift
func (e *Engine) SenseQuorum(entityIndex int) bool {
	if !e.validDrone(entityIndex) {
		return false
	}
	ix := int32(e.Registry.PositionX[entityIndex].V / crysmath.Precision)
	iy := int32(e.Registry.PositionY[entityIndex].V / crysmath.Precision)

	votesForTrue := 0
	totalPeers := 0

	for j := 0; j < e.Registry.Count; j++ {
		if entityIndex == j || e.Registry.State[j] == StateInert {
			continue
		}

		jx := int32(e.Registry.PositionX[j].V / crysmath.Precision)
		jy := int32(e.Registry.PositionY[j].V / crysmath.Precision)

		dx, dy := ix-jx, iy-jy
		if dx*dx+dy*dy <= InfectionRadius*InfectionRadius {
			totalPeers++
			// Drones check if their neighbor's logic registry appears sound
			if e.Registry.TrustScore[j] >= 70 && !e.Registry.Compromised[j] {
				votesForTrue++
			}
		}
	}

	// Quorum consensus: If more than 50% of verified local peers are sound, return true
	if totalPeers == 0 {
		return true
	}
	return (votesForTrue * 100 / totalPeers) > 50
}

func (e *Engine) stepSearching(i int) {
	x := int(e.Registry.PositionX[i].V / crysmath.Precision)
	y := int(e.Registry.PositionY[i].V / crysmath.Precision)
	idx := e.Grid.GetIndex(x, y)

	// If standing on resource and not already carrying cargo: harvest
	if e.Registry.Inventory[i] == 0 && e.Grid.CurrentCells[idx].ResourceCount > 0 && e.Grid.NextCells[idx].ResourceCount > 0 {
		e.Grid.NextCells[idx].ResourceCount--
		e.Registry.Inventory[i] = 1 // MaxCargo
		e.Registry.State[i] = StateReturning
		return
	}

	// Drain battery for movement (1 unit per tick)
	e.Registry.Battery[i] -= 1 * crysmath.Precision / 1000
	if e.Registry.Battery[i] <= 0 {
		e.Registry.Battery[i] = 0
		e.Registry.State[i] = StateInert
		return
	}

	// Move: Sense highest resource gradient
	targetX, targetY, val := e.Grid.SenseHighestGradient(x, y, true)

	if val <= 0 {
		// Random walk if no trail found
		dx := e.rng.Intn(3) - 1
		dy := e.rng.Intn(3) - 1
		x += dx
		y += dy
		// Clamp to grid bounds
		if x < 0 {
			x = 0
		}
		if x >= e.Grid.Width {
			x = e.Grid.Width - 1
		}
		if y < 0 {
			y = 0
		}
		if y >= e.Grid.Height {
			y = e.Grid.Height - 1
		}
	} else {
		x, y = targetX, targetY
	}

	e.Registry.PositionX[i] = crysmath.NewFixedPoint(int64(x))
	e.Registry.PositionY[i] = crysmath.NewFixedPoint(int64(y))

	// Deposit signals
	if e.Registry.Compromised[i] {
		e.Grid.NextCells[idx].AlienSignal = saturateAdd(e.Grid.NextCells[idx].AlienSignal, 150_000)
	} else {
		e.Grid.NextCells[idx].HomePheromone = saturateAdd(e.Grid.NextCells[idx].HomePheromone, 100_000)
	}
}

func (e *Engine) stepReturning(i int) {
	x := int(e.Registry.PositionX[i].V / crysmath.Precision)
	y := int(e.Registry.PositionY[i].V / crysmath.Precision)
	idx := e.Grid.GetIndex(x, y)

	// If at base: drop resource
	if e.Grid.CurrentCells[idx].IsBase {
		if e.Registry.Inventory[i] > 0 {
			e.GlobalSilicates += e.Registry.Inventory[i]
			e.TotalDeposited += e.Registry.Inventory[i]
			e.Registry.Inventory[i] = 0
		}
		e.Registry.State[i] = StateSearching
		return
	}

	// Drain battery for movement (1 unit per tick)
	e.Registry.Battery[i] -= 1 * crysmath.Precision / 1000
	if e.Registry.Battery[i] <= 0 {
		e.Registry.Battery[i] = 0
		e.Registry.State[i] = StateInert
		return
	}

	// Move: Sense highest home gradient
	targetX, targetY, val := e.Grid.SenseHighestGradient(x, y, false)
	baseX, baseY := e.Grid.Width/2, e.Grid.Height/2

	if val > 0 && chebyshevDistance(targetX, targetY, baseX, baseY) < chebyshevDistance(x, y, baseX, baseY) {
		x, y = targetX, targetY
	} else {
		x = stepTowardsCoord(x, baseX)
		y = stepTowardsCoord(y, baseY)
	}

	e.Registry.PositionX[i] = crysmath.NewFixedPoint(int64(x))
	e.Registry.PositionY[i] = crysmath.NewFixedPoint(int64(y))

	// Deposit signals
	if e.Registry.Compromised[i] {
		e.Grid.NextCells[idx].AlienSignal = saturateAdd(e.Grid.NextCells[idx].AlienSignal, 150_000)
	} else {
		e.Grid.NextCells[idx].ResourcePheromone = saturateAdd(e.Grid.NextCells[idx].ResourcePheromone, 100_000)
	}
}

func (e *Engine) PrintTelemetry() {
	returning := 0
	for i := 0; i < e.Registry.Count; i++ {
		if e.Registry.State[i] == StateReturning {
			returning++
		}
	}
	fmt.Printf("[Tick %d] Active Drones: %d | Returning with Silicates: %d\n", e.Tick, e.Registry.Count, returning)
}

func saturateAdd(a, b int32) int32 {
	res := int64(a) + int64(b)
	if res > int64(MaxPheromone) {
		return MaxPheromone
	}
	if res < 0 {
		return 0
	}
	return int32(res)
}

func stepTowardsCoord(current, target int) int {
	if current < target {
		return current + 1
	}
	if current > target {
		return current - 1
	}
	return current
}

func chebyshevDistance(ax, ay, bx, by int) int {
	dx := ax - bx
	if dx < 0 {
		dx = -dx
	}
	dy := ay - by
	if dy < 0 {
		dy = -dy
	}
	if dx > dy {
		return dx
	}
	return dy
}

func (e *Engine) validDrone(i int) bool {
	return i >= 0 && i < e.Registry.Count
}

func (e *Engine) canAct(i int) bool {
	return e.validDrone(i) && e.Registry.State[i] != StateInert && e.Registry.Battery[i] > 0
}
