package simulation

import (
	"chrysalis-engine/core/crysmath"
	"encoding/binary"
	"hash/fnv"
)

// SetState restores the engine to the world described by state, which must be the
// output of a prior GetState() call (either in-memory or after a JSON round-trip).
// All Go-typed values (int32, int64, bool) and JSON-decoded values (float64, bool)
// are handled transparently via the toXxx helpers.
//
// The EventBus is reset to an empty state; the caller must call Bus.BeginTick()
// before the next simulation tick.
func (e *Engine) SetState(state map[string]interface{}) {
	// 1. Scalar fields
	e.Tick = toI64(state["tick"])
	e.GlobalSilicates = int32(toI64(state["colony_res"]))
	e.TotalDeposited = int32(toI64(state["total_deposited"]))
	e.HistoricalTotal = int32(toI64(state["historical_total"]))

	// 2. RNG — restore exact sequence position
	rngSeed := toI64(state["rng_seed"])
	rngCalls := toI64(state["rng_calls"])
	if rngSeed == 0 {
		rngSeed = toI64(state["seed"]) // fallback for older checkpoints
	}
	e.Seed = rngSeed
	e.rng = restoreDetRNG(rngSeed, rngCalls)

	// 3. Registry
	drones := toSlice(state["drones"])
	n := len(drones)
	if n > len(e.Registry.ID) {
		e.Registry = NewSwarmRegistry(n)
	}
	e.Registry.Count = n
	for i, d := range drones {
		dm := toMap(d)
		if dm == nil {
			continue
		}
		e.Registry.ID[i] = uint32(toI64(dm["id"]))
		e.Registry.PositionX[i] = crysmath.NewFixedPointRaw(toI64(dm["x"]))
		e.Registry.PositionY[i] = crysmath.NewFixedPointRaw(toI64(dm["y"]))
		e.Registry.Battery[i] = toI64(dm["bat"])
		e.Registry.State[i] = DroneState(toI64(dm["state"]))
		e.Registry.Inventory[i] = int32(toI64(dm["inv"]))
		e.Registry.Compromised[i] = toBool(dm["comp"])
		e.Registry.TrustScore[i] = int32(toI64(dm["trust"]))
		e.Registry.CorruptionFactor[i] = uint8(toI64(dm["corr"]))
		e.Registry.InertTTL[i] = int32(toI64(dm["inert_ttl"]))
	}

	// Restore the monotonic ID counter so fabrication after a restore mints the
	// same IDs the original run would have. Older checkpoints predate next_id;
	// migrate by deriving max(ID)+1 so restored+fabricated worlds stay collision-free.
	if raw, ok := state["next_id"]; ok {
		e.Registry.NextID = uint32(toI64(raw))
	} else {
		var maxNext uint32
		for i := 0; i < n; i++ {
			if e.Registry.ID[i]+1 > maxNext {
				maxNext = e.Registry.ID[i] + 1
			}
		}
		e.Registry.NextID = maxNext
	}

	// 4. Grid — zero all cells then restore active ones
	for i := range e.Grid.CurrentCells {
		e.Grid.CurrentCells[i] = Cell{}
		e.Grid.NextCells[i] = Cell{}
	}
	for _, c := range toSlice(state["grid"]) {
		cm := toMap(c)
		if cm == nil {
			continue
		}
		x := int(toI64(cm["x"]))
		y := int(toI64(cm["y"]))
		idx := e.Grid.GetIndex(x, y)
		cell := Cell{
			HomePheromone:     int32(toI64(cm["home"])),
			ResourcePheromone: int32(toI64(cm["res"])),
			AlienSignal:       int32(toI64(cm["alien"])),
			ResourceCount:     int32(toI64(cm["cnt"])),
			IsBase:            toBool(cm["base"]),
		}
		e.Grid.CurrentCells[idx] = cell
		e.Grid.NextCells[idx] = cell
	}

	// 5. Hazards
	for i := range e.Hazards.Active {
		e.Hazards.Active[i] = false
	}
	for _, h := range toSlice(state["hazards"]) {
		hm := toMap(h)
		if hm == nil {
			continue
		}
		e.Hazards.Add(
			HazardType(toI64(hm["type"])),
			int32(toI64(hm["x"])),
			int32(toI64(hm["y"])),
			int32(toI64(hm["rad"])),
			toI64(hm["intensity"]),
		)
	}

	// 6. Alien nodes
	for i := range e.Aliens.Active {
		e.Aliens.Active[i] = false
	}
	for _, a := range toSlice(state["aliens"]) {
		am := toMap(a)
		if am == nil {
			continue
		}
		e.Aliens.Add(
			AlienNodeType(toI64(am["type"])),
			int32(toI64(am["x"])),
			int32(toI64(am["y"])),
			int32(toI64(am["rad"])),
		)
	}

	// 7. Mission — try direct struct first (in-memory), then map (post-JSON)
	switch m := state["mission"].(type) {
	case MissionState:
		e.Mission = m
	case map[string]interface{}:
		e.Mission = MissionState{
			Status:                 MissionStatus(toString(m["status"])),
			Reason:                 toString(m["reason"]),
			TargetResources:        int(toI64(m["target_resources"])),
			MaxTicks:               toI64(m["max_ticks"]),
			InfectionLossThreshold: toF64(m["infection_loss_threshold"]),
			ResourcesDeposited:     int(toI64(m["resources_deposited"])),
			InfectedRatio:          toF64(m["infected_ratio"]),
			Tick:                   toI64(m["tick"]),
		}
	}

	// 8. Fresh EventBus — the caller must call BeginTick before the next tick
	e.Bus = NewEventBus()
}

// WorldHash returns a stable FNV-64a hash over every piece of canonical simulation
// state — any field that can influence future simulation output. Two engines with
// identical canonical state always produce the same hash.
//
// Canonical state = everything that affects future tick output:
//   engine scalars, drone registry, grid cells, hazards, aliens, mission, RNG position.
//
// Non-canonical state (EventBus, DecisionFrames, inspector flag, recorder) is
// intentionally excluded: it is ephemeral or reconstructed by consumers.
func (e *Engine) WorldHash() uint64 {
	h := fnv.New64a()

	var buf [8]byte
	put64 := func(v int64) {
		binary.LittleEndian.PutUint64(buf[:], uint64(v))
		h.Write(buf[:])
	}
	put32 := func(v int32) { put64(int64(v)) }
	putBool := func(b bool) {
		if b {
			h.Write([]byte{1})
		} else {
			h.Write([]byte{0})
		}
	}

	// Engine scalars
	put64(e.Tick)
	put32(e.GlobalSilicates)
	put32(e.TotalDeposited)
	put32(e.HistoricalTotal)
	put64(int64(e.Registry.Count))
	put64(int64(e.Registry.NextID))

	// Drone registry — every component column
	for i := 0; i < e.Registry.Count; i++ {
		put64(int64(e.Registry.ID[i]))
		put64(e.Registry.PositionX[i].V)
		put64(e.Registry.PositionY[i].V)
		put64(e.Registry.Battery[i])
		h.Write([]byte{byte(e.Registry.State[i])})
		put32(e.Registry.Inventory[i])
		putBool(e.Registry.Compromised[i])
		put32(e.Registry.TrustScore[i])
		h.Write([]byte{e.Registry.CorruptionFactor[i]})
		put32(e.Registry.InertTTL[i])
	}

	// Grid — sparse: only cells with any non-zero metric, keyed by cell index.
	// The index encodes position; zero-value cells are implicitly absent.
	for i, cell := range e.Grid.CurrentCells {
		if cell.HomePheromone == 0 && cell.ResourcePheromone == 0 &&
			cell.ResourceCount == 0 && cell.AlienSignal == 0 && !cell.IsBase {
			continue
		}
		put64(int64(i))
		put32(cell.HomePheromone)
		put32(cell.ResourcePheromone)
		put32(cell.ResourceCount)
		put32(cell.AlienSignal)
		putBool(cell.IsBase)
	}

	// Hazards — all capacity slots (Active flag included so slot identity is preserved)
	for i := 0; i < e.Hazards.Capacity; i++ {
		putBool(e.Hazards.Active[i])
		if e.Hazards.Active[i] {
			h.Write([]byte{byte(e.Hazards.Type[i])})
			put32(e.Hazards.X[i])
			put32(e.Hazards.Y[i])
			put32(e.Hazards.Radius[i])
			put64(e.Hazards.Intensity[i])
		}
	}

	// Alien network — all capacity slots
	for i := 0; i < e.Aliens.Capacity; i++ {
		putBool(e.Aliens.Active[i])
		if e.Aliens.Active[i] {
			h.Write([]byte{byte(e.Aliens.Type[i])})
			put32(e.Aliens.X[i])
			put32(e.Aliens.Y[i])
			put32(e.Aliens.Radius[i])
		}
	}

	// Mission — canonical fields only (ResourcesDeposited/InfectedRatio are derived)
	switch e.Mission.Status {
	case MissionRunning:
		h.Write([]byte{0})
	case MissionVictory:
		h.Write([]byte{1})
	case MissionDefeat:
		h.Write([]byte{2})
	default:
		h.Write([]byte{255})
	}
	put64(int64(e.Mission.TargetResources))
	put64(e.Mission.MaxTicks)

	// RNG position — a different (seed, callCount) means all future randomness diverges
	rngSeed, rngCalls := e.rng.snapshot()
	put64(rngSeed)
	put64(rngCalls)

	return h.Sum64()
}

// --- type-coercion helpers ---
// These handle both direct Go types (stored in in-memory checkpoints) and
// float64/bool values produced by encoding/json unmarshaling.

func toI64(v interface{}) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case float64:
		return int64(x)
	case int32:
		return int64(x)
	case int:
		return int64(x)
	case uint32:
		return int64(x)
	case uint8:
		return int64(x)
	case uint64:
		return int64(x)
	}
	return 0
}

func toBool(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func toF64(v interface{}) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int64:
		return float64(x)
	}
	return 0
}

func toSlice(v interface{}) []interface{} {
	switch x := v.(type) {
	case []interface{}:
		return x
	case []map[string]interface{}:
		out := make([]interface{}, len(x))
		for i, m := range x {
			out[i] = m
		}
		return out
	}
	return nil
}

func toMap(v interface{}) map[string]interface{} {
	m, _ := v.(map[string]interface{})
	return m
}
