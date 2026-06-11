package simulation

import (
	"chrysalis-engine/core/crysmath"
	"math/rand"
)

type AlienNodeType uint8

const (
	NodeInfector    AlienNodeType = iota // Spreads logic virus
	NodeJammer                           // Jams communications
	InfectionRadius int32         = 3    // Proximity bounds for wireless viral spread (in grid cells)
)

// SpreadsInfection scans the registry for compromised entities and bleeds corruption
func (e *Engine) SpreadsInfection() {
	for i := 0; i < e.Registry.Count; i++ {
		if !e.Registry.Compromised[i] {
			continue
		}

		ix := int32(e.Registry.PositionX[i].V / crysmath.Precision)
		iy := int32(e.Registry.PositionY[i].V / crysmath.Precision)

		// Search for nearby healthy drones to infect
		for j := 0; j < e.Registry.Count; j++ {
			if i == j || e.Registry.Compromised[j] {
				continue
			}

			jx := int32(e.Registry.PositionX[j].V / crysmath.Precision)
			jy := int32(e.Registry.PositionY[j].V / crysmath.Precision)

			dx, dy := ix-jx, iy-jy
			distSq := dx*dx + dy*dy

			if distSq <= InfectionRadius*InfectionRadius {
				// Bleed corruption: increase factor by 1-5% per tick
				factor := uint8(rand.Intn(5) + 1)
				current := e.Registry.CorruptionFactor[j]
				
				if uint32(current) + uint32(factor) >= 100 {
					e.Registry.CorruptionFactor[j] = 100
					e.Registry.Compromised[j] = true
					e.Registry.TrustScore[j] = 50
				} else {
					e.Registry.CorruptionFactor[j] += factor
				}
			}
		}
	}
}

type AlienNetwork struct {
	Capacity  int
	Active    []bool
	Type      []AlienNodeType
	X         []int32
	Y         []int32
	Radius    []int32
}

func NewAlienNetwork(capacity int) *AlienNetwork {
	return &AlienNetwork{
		Capacity: capacity,
		Active:   make([]bool, capacity),
		Type:     make([]AlienNodeType, capacity),
		X:        make([]int32, capacity),
		Y:        make([]int32, capacity),
		Radius:   make([]int32, capacity),
	}
}

func (an *AlienNetwork) Add(nType AlienNodeType, x, y, radius int32) {
	for i := 0; i < an.Capacity; i++ {
		if !an.Active[i] {
			an.Active[i] = true
			an.Type[i] = nType
			an.X[i] = x
			an.Y[i] = y
			an.Radius[i] = radius
			return
		}
	}
}
