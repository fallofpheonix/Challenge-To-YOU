package simulation

import (
	"chrysalis-engine/core/crysmath"
)

// DroneState defines the high-level behavior state of an agent
type DroneState uint8

const (
	StateSearching DroneState = iota
	StateReturning
	StateInert
)

// SwarmRegistry holds the contiguous memory arrays for all drones in the simulation.
// This data-oriented layout ensures maximum CPU cache efficiency.
type SwarmRegistry struct {
	Count int

	// Component Slices
	ID               []uint32
	PositionX        []crysmath.FixedPoint
	PositionY        []crysmath.FixedPoint
	Battery          []int64 // Scaled by crysmath.Precision
	State            []DroneState
	Inventory        []int32
	Compromised      []bool  // Logic virus infection status
	TrustScore       []int32 // Peer-to-peer validation score
	CorruptionFactor []uint8 // Percentage tracking local logic manipulation [0 to 100]
}

// NewSwarmRegistry initializes a registry with a fixed capacity.
func NewSwarmRegistry(capacity int) *SwarmRegistry {
	return &SwarmRegistry{
		Count:            0,
		ID:               make([]uint32, capacity),
		PositionX:        make([]crysmath.FixedPoint, capacity),
		PositionY:        make([]crysmath.FixedPoint, capacity),
		Battery:          make([]int64, capacity),
		State:            make([]DroneState, capacity),
		Inventory:        make([]int32, capacity),
		Compromised:      make([]bool, capacity),
		TrustScore:       make([]int32, capacity),
		CorruptionFactor: make([]uint8, capacity),
	}
}

// Spawn adds a new drone to the registry at the given position.
// If capacity is reached, slices are automatically expanded.
func (r *SwarmRegistry) Spawn(x, y int, battery int64) {
	if r.Count >= len(r.ID) {
		// Expand capacity (double the current size)
		newCap := len(r.ID) * 2
		if newCap == 0 { newCap = 1 }

		newID := make([]uint32, newCap)
		newPX := make([]crysmath.FixedPoint, newCap)
		newPY := make([]crysmath.FixedPoint, newCap)
		newBat := make([]int64, newCap)
		newState := make([]DroneState, newCap)
		newInv := make([]int32, newCap)
		newComp := make([]bool, newCap)
		newTrust := make([]int32, newCap)
		newCorr := make([]uint8, newCap)

		copy(newID, r.ID)
		copy(newPX, r.PositionX)
		copy(newPY, r.PositionY)
		copy(newBat, r.Battery)
		copy(newState, r.State)
		copy(newInv, r.Inventory)
		copy(newComp, r.Compromised)
		copy(newTrust, r.TrustScore)
		copy(newCorr, r.CorruptionFactor)

		r.ID = newID
		r.PositionX = newPX
		r.PositionY = newPY
		r.Battery = newBat
		r.State = newState
		r.Inventory = newInv
		r.Compromised = newComp
		r.TrustScore = newTrust
		r.CorruptionFactor = newCorr
	}

	i := r.Count
	r.ID[i] = uint32(i)
	r.PositionX[i] = crysmath.NewFixedPoint(int64(x))
	r.PositionY[i] = crysmath.NewFixedPoint(int64(y))
	r.Battery[i] = battery
	r.State[i] = StateSearching
	r.Inventory[i] = 0
	r.Compromised[i] = false
	r.TrustScore[i] = 100 // Default full trust
	r.CorruptionFactor[i] = 0

	r.Count++
}
