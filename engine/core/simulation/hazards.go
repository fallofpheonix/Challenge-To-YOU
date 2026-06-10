package simulation

type HazardType uint8

const (
	HazardMagnetic HazardType = iota // Drains battery
	HazardThermal                    // Future: Physical damage
)

type HazardSystem struct {
	Capacity  int
	Active    []bool
	Type      []HazardType
	X         []int32
	Y         []int32
	Radius    []int32
	Intensity []int64 // crysmath.FixedPoint scaled values (e.g., drain per tick)
}

func NewHazardSystem(capacity int) *HazardSystem {
	return &HazardSystem{
		Capacity:  capacity,
		Active:    make([]bool, capacity),
		Type:      make([]HazardType, capacity),
		X:         make([]int32, capacity),
		Y:         make([]int32, capacity),
		Radius:    make([]int32, capacity),
		Intensity: make([]int64, capacity),
	}
}

func (h *HazardSystem) Add(hType HazardType, x, y, radius int32, intensity int64) {
	for i := 0; i < h.Capacity; i++ {
		if !h.Active[i] {
			h.Active[i] = true
			h.Type[i] = hType
			h.X[i] = x
			h.Y[i] = y
			h.Radius[i] = radius
			h.Intensity[i] = intensity
			return
		}
	}
}
