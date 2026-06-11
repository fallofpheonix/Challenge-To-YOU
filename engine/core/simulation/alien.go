package simulation

type AlienNodeType uint8

const (
	NodeInfector AlienNodeType = iota // Spreads logic virus
	NodeJammer                        // Jams communications
)

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
