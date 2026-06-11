package simulation

import (
	"chrysalis-engine/core/crysmath"
	"fmt"
	"math/rand"
)

const (
	FabricationThreshold int32 = 5   // 5 silicates required to construct a new unit
	MaxSwarmCapacity     int   = 500 // Safety cap for the MVP engine pass
)

type Engine struct {
	Grid            *Grid
	Registry        *SwarmRegistry
	Hazards         *HazardSystem
	Aliens          *AlienNetwork
	Tick            int64
	GlobalSilicates int32
	HistoricalTotal int32
}

func NewEngine(width, height int, droneCount int) *Engine {
	e := &Engine{
		Grid:     NewGrid(width, height),
		Registry: NewSwarmRegistry(droneCount),
		Hazards:  NewHazardSystem(10), // Support up to 10 active hazards
		Aliens:   NewAlienNetwork(5),  // Support up to 5 alien nodes
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

func (e *Engine) Step() {
	e.Tick++

	// 1. Environment & Pheromone Pass (Decay and stage Next layer)
	e.Grid.TickPheromones()

	// 1.5 Process Hazards
	e.processHazards()

	// 1.6 Process Alien Infections
	e.processInfections()
	e.SpreadsInfection()

	// 1.7 Check for Fabrication
	e.CheckFabricationPool()

	// 2. Reinforce Base Pheromone (Base is a constant source)
	width, height := e.Grid.Width, e.Grid.Height
	idx := e.Grid.GetIndex(width/2, height/2)
	e.Grid.NextCells[idx].HomePheromone = MaxPheromone

	// 3. Drone Logic Pass (Read Current, Write Next)
	// This will be replaced by P-Script execution in the main loop
	e.stepDrones()

	// 4. Swap Buffers (Commit mutations simultaneously)
	e.Grid.SwapBuffers()
}

func (e *Engine) SenseResource(i int) bool {
	x := int(e.Registry.PositionX[i].V / crysmath.Precision)
	y := int(e.Registry.PositionY[i].V / crysmath.Precision)
	_, _, val := e.Grid.SenseHighestGradient(x, y, true)
	return val > 0
}

func (e *Engine) SenseHome(i int) bool {
	x := int(e.Registry.PositionX[i].V / crysmath.Precision)
	y := int(e.Registry.PositionY[i].V / crysmath.Precision)
	_, _, val := e.Grid.SenseHighestGradient(x, y, false)
	return val > 0
}

func (e *Engine) Harvest(i int) {
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
	x := int(e.Registry.PositionX[i].V / crysmath.Precision)
	y := int(e.Registry.PositionY[i].V / crysmath.Precision)
	idx := e.Grid.GetIndex(x, y)
	if e.Grid.CurrentCells[idx].IsBase {
		if e.Registry.Inventory[i] > 0 {
			e.GlobalSilicates += e.Registry.Inventory[i]
			e.Registry.Inventory[i] = 0
		}
		e.Registry.State[i] = StateSearching
	}
}

func (e *Engine) MoveTowardsResource(i int) {
	e.stepSearching(i)
}

func (e *Engine) MoveTowardsHome(i int) {
	e.stepReturning(i)
}

func (e *Engine) MoveRandom(i int) {
	x := int(e.Registry.PositionX[i].V / crysmath.Precision)
	y := int(e.Registry.PositionY[i].V / crysmath.Precision)
	dx := rand.Intn(3) - 1
	dy := rand.Intn(3) - 1
	x += dx
	y += dy
	if x < 0 { x = 0 }
	if x >= e.Grid.Width { x = e.Grid.Width - 1 }
	if y < 0 { y = 0 }
	if y >= e.Grid.Height { y = e.Grid.Height - 1 }
	e.Registry.PositionX[i] = crysmath.NewFixedPoint(int64(x))
	e.Registry.PositionY[i] = crysmath.NewFixedPoint(int64(y))
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
				if e.Registry.Battery[i] < 0 {
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
				if rand.Float32() < 0.05 {
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
	x := int(e.Registry.PositionX[i].V / crysmath.Precision)
	y := int(e.Registry.PositionY[i].V / crysmath.Precision)
	_, _, val := e.Grid.SenseHighestAlienGradient(x, y)
	return val > 0
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

	// Move: Sense highest resource gradient
	targetX, targetY, val := e.Grid.SenseHighestGradient(x, y, true)

	if val <= 0 {
		// Random walk if no trail found
		dx := rand.Intn(3) - 1
		dy := rand.Intn(3) - 1
		x += dx
		y += dy
		// Clamp to grid bounds
		if x < 0 { x = 0 }
		if x >= e.Grid.Width { x = e.Grid.Width - 1 }
		if y < 0 { y = 0 }
		if y >= e.Grid.Height { y = e.Grid.Height - 1 }
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
			e.Registry.Inventory[i] = 0
		}
		e.Registry.State[i] = StateSearching
		return
	}

	// Move: Sense highest home gradient
	targetX, targetY, val := e.Grid.SenseHighestGradient(x, y, false)

	if val <= 0 {
		dx := rand.Intn(3) - 1
		dy := rand.Intn(3) - 1
		x += dx
		y += dy
		if x < 0 { x = 0 }
		if x >= e.Grid.Width { x = e.Grid.Width - 1 }
		if y < 0 { y = 0 }
		if y >= e.Grid.Height { y = e.Grid.Height - 1 }
	} else {
		x, y = targetX, targetY
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
	res := a + b
	if res > MaxPheromone {
		return MaxPheromone
	}
	return res
}
