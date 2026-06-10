package simulation

import "sync"

const (
	Scale          int32 = 1_000_000
	PheromoneDecay int32 = 5_000     // 0.005 decay per tick
	MaxPheromone   int32 = 1_000_000 // 1.0 maximum density
)

type Cell struct {
	ResourceCount     int32 // Total silicates available
	HomePheromone     int32 // Fixed-point gradient back to base
	ResourcePheromone int32 // Fixed-point gradient to resources
	IsBase            bool
}

type Grid struct {
	Width  int
	Height int
	// Double buffering to guarantee decoupled state updates
	CurrentCells []Cell
	NextCells    []Cell
	mu           sync.RWMutex
}

func NewGrid(width, height int) *Grid {
	cells := make([]Cell, width*height)
	nextCells := make([]Cell, width*height)
	return &Grid{
		Width:        width,
		Height:       height,
		CurrentCells: cells,
		NextCells:    nextCells,
	}
}

// GetIndex flattens 2D coordinates into a 1D slice layout
func (g *Grid) GetIndex(x, y int) int {
	// Rigid simulation boundary safety
	if x < 0 { x = 0 }
	if x >= g.Width { x = g.Width - 1 }
	if y < 0 { y = 0 }
	if y >= g.Height { y = g.Height - 1 }
	return y*g.Width + x
}

// SwapBuffers commits mutations simultaneously at the end of the simulation tick
func (g *Grid) SwapBuffers() {
	g.mu.Lock()
	defer g.mu.Unlock()
	copy(g.CurrentCells, g.NextCells)
}
