package simulation

import "chrysalis-engine/core/crysmath"

// SpatialHash partitions drones into grid cells for O(n) neighbor queries.
// Cell size is InfectionRadius so each drone only checks 9 cells (3x3 neighborhood).
type SpatialHash struct {
	cellSize int32
	cols     int
	rows     int
	cells    [][]int // cells[row*cols + col] = list of drone indices
}

// NewSpatialHash creates a spatial hash for the given world dimensions.
func NewSpatialHash(width, height int, cellSize int32) *SpatialHash {
	cols := int(int32(width)/cellSize) + 1
	rows := int(int32(height)/cellSize) + 1
	cells := make([][]int, cols*rows)
	return &SpatialHash{
		cellSize: cellSize,
		cols:     cols,
		rows:     rows,
		cells:    cells,
	}
}

// Clear resets all cells for reuse (avoids allocation per tick).
func (sh *SpatialHash) Clear() {
	for i := range sh.cells {
		sh.cells[i] = sh.cells[i][:0]
	}
}

// Insert adds a drone index to the appropriate cell.
func (sh *SpatialHash) Insert(index int, x, y int32) {
	cx := int(x / sh.cellSize)
	cy := int(y / sh.cellSize)
	if cx < 0 {
		cx = 0
	}
	if cx >= sh.cols {
		cx = sh.cols - 1
	}
	if cy < 0 {
		cy = 0
	}
	if cy >= sh.rows {
		cy = sh.rows - 1
	}
	idx := cy*sh.cols + cx
	sh.cells[idx] = append(sh.cells[idx], index)
}

// QueryNearby returns all drone indices in the 3x3 neighborhood around (x, y).
func (sh *SpatialHash) QueryNearby(x, y int32) []int {
	cx := int(x / sh.cellSize)
	cy := int(y / sh.cellSize)
	if cx < 0 {
		cx = 0
	}
	if cx >= sh.cols {
		cx = sh.cols - 1
	}
	if cy < 0 {
		cy = 0
	}
	if cy >= sh.rows {
		cy = sh.rows - 1
	}

	var result []int
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := cx+dx, cy+dy
			if nx < 0 || nx >= sh.cols || ny < 0 || ny >= sh.rows {
				continue
			}
			cell := sh.cells[ny*sh.cols+nx]
			result = append(result, cell...)
		}
	}
	return result
}

// BuildFromRegistry populates the spatial hash from the current drone positions.
func (sh *SpatialHash) BuildFromRegistry(reg *SwarmRegistry) {
	sh.Clear()
	for i := 0; i < reg.Count; i++ {
		x := int32(reg.PositionX[i].V / crysmath.Precision)
		y := int32(reg.PositionY[i].V / crysmath.Precision)
		sh.Insert(i, x, y)
	}
}
