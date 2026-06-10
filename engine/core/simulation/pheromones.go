package simulation

// TickPheromones handles global evaporation and prepares the NextCells layer
func (g *Grid) TickPheromones() {
	g.mu.Lock()
	defer g.mu.Unlock()

	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			idx := g.GetIndex(x, y)
			currentCell := g.CurrentCells[idx]

			// Evaporate Home Pheromone
			nextHome := currentCell.HomePheromone - PheromoneDecay
			if nextHome < 0 { nextHome = 0 }

			// Evaporate Resource Pheromone
			nextResource := currentCell.ResourcePheromone - PheromoneDecay
			if nextResource < 0 { nextResource = 0 }

			// Stage into the double buffer without leaking into active execution
			g.NextCells[idx].HomePheromone = nextHome
			g.NextCells[idx].ResourcePheromone = nextResource
			g.NextCells[idx].ResourceCount = currentCell.ResourceCount
			g.NextCells[idx].IsBase = currentCell.IsBase
		}
	}
}

// SenseHighestGradient looks at the immediate 8-cell neighborhood.
// Returns the coordinates of the best cell and the pheromone value found.
func (g *Grid) SenseHighestGradient(startX, startY int, lookForResource bool) (targetX, targetY int, maxVal int32) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	bestVal := int32(-1)
	bestX, bestY := startX, startY

	// Scan 3x3 neighborhood bounds
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := startX+dx, startY+dy
			
			// Rigid simulation walls check
			if nx < 0 || nx >= g.Width || ny < 0 || ny >= g.Height {
				continue
			}

			idx := g.GetIndex(nx, ny)
			var val int32
			if lookForResource {
				val = g.CurrentCells[idx].ResourcePheromone
			} else {
				val = g.CurrentCells[idx].HomePheromone
			}

			if val > bestVal {
				bestVal = val
				bestX, bestY = nx, ny
			}
		}
	}
	return bestX, bestY, bestVal
}
