package gameloop

import (
	"time"

	"challenge-to-you/backend/internal/eventbus"
)

type GameState int

const (
	StatePlaying GameState = iota
	StatePaused
	StateWon
	StateLost
	StatePurged
)

func (s GameState) String() string {
	switch s {
	case StatePlaying:
		return "playing"
	case StatePaused:
		return "paused"
	case StateWon:
		return "won"
	case StateLost:
		return "lost"
	case StatePurged:
		return "purged"
	default:
		return "unknown"
	}
}

type EventType int

const (
	EventPlayerInput EventType = iota
	EventTick
	EventSystem
)

type GameEvent struct {
	Type      EventType
	PlayerID  string
	Action    string
	Payload   string
	Timestamp time.Time
}

type PlayerInput struct {
	Action  string `json:"action"`
	Payload string `json:"payload"`
}

type TickResult struct {
	Tick          int
	State         GameState
	Events        []GameEvent
	Snapshot      map[string]interface{}
	MissionStatus *MissionStatus
	Errors        []string
}

type MissionStatus struct {
	ID       string  `json:"id"`
	Status   string  `json:"status"`
	Progress float64 `json:"progress"`
	Won      bool    `json:"won"`
	Lost     bool    `json:"lost"`
}

type GameLoopConfig struct {
	TickInterval time.Duration
	EventBus     *eventbus.EventBus
	Bus          *eventbus.EventBus
	MaxTicks     int
	SessionKey   string       // Unique session identifier for eviction
	EvictionFn   func(string) // Called with SessionKey when the fabric should be evicted
}

func DefaultConfig() GameLoopConfig {
	return GameLoopConfig{
		TickInterval: 500 * time.Millisecond,
		MaxTicks:     0,
	}
}

type ReplayFrame struct {
	Tick      int                    `json:"tick"`
	State     GameState              `json:"state"`
	Events    []GameEvent            `json:"events"`
	StateHash string                 `json:"state_hash"`
	Snapshot  map[string]interface{} `json:"snapshot"`
	ElapsedMS int64                  `json:"elapsed_ms"`
}

type TelemetryPoint struct {
	Tick      int                    `json:"tick"`
	Vigilance float64                `json:"vigilance"`
	Entropy   float64                `json:"entropy"`
	State     GameState              `json:"state"`
	Metrics   map[string]float64     `json:"metrics"`
	Snapshot  map[string]interface{} `json:"snapshot,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type ResourceType string

const (
	ResourceEnergy    ResourceType = "energy"
	ResourceMaterial  ResourceType = "material"
	ResourceData      ResourceType = "data"
	ResourceArtifact  ResourceType = "artifact"
	ResourceRuneShard ResourceType = "rune_shard"
)

type Inventory struct {
	Resources map[ResourceType]int `json:"resources"`
}

func NewInventory() *Inventory {
	return &Inventory{
		Resources: make(map[ResourceType]int),
	}
}

func (inv *Inventory) Add(r ResourceType, amount int) {
	inv.Resources[r] += amount
}

func (inv *Inventory) Remove(r ResourceType, amount int) int {
	current := inv.Resources[r]
	if current < amount {
		amount = current
	}
	inv.Resources[r] -= amount
	if inv.Resources[r] <= 0 {
		delete(inv.Resources, r)
	}
	return amount
}

func (inv *Inventory) Has(r ResourceType, amount int) bool {
	return inv.Resources[r] >= amount
}

func (inv *Inventory) Count(r ResourceType) int {
	return inv.Resources[r]
}
