package simulation

import "encoding/json"

// EventType is the canonical tag for every simulation event.
type EventType string

const (
	EventDroneSpawned   EventType = "drone_spawned"
	EventDroneDied      EventType = "drone_died"
	EventDroneInfected  EventType = "drone_infected"
	EventHarvested      EventType = "harvested"
	EventDeposited      EventType = "deposited"
	EventTrustChanged   EventType = "trust_changed"
	EventMissionChanged EventType = "mission_changed"
	EventFabricated     EventType = "fabricated"
	EventHazardDamage   EventType = "hazard_damage"
)

// EventPayload is a sealed interface. The unexported marker method prevents
// external packages from accidentally satisfying it, so Event.Data can only
// hold the explicitly defined payload types below — not arbitrary maps or structs.
type EventPayload interface {
	isEventPayload()
}

// --- Typed payload structs ---
// Each event type owns its own struct with named fields. No generic ValueA/ValueB slots.

type SpawnedData struct {
	SwarmSize int `json:"swarm_size"`
}

type DiedData struct {
	Cause string `json:"cause"` // "battery" | "hazard"
}

type InfectedData struct {
	Vector string `json:"vector"` // "alien_node" | "peer_spread"
}

type HarvestData struct {
	ResourcesRemaining int32 `json:"resources_remaining"`
}

type DepositData struct {
	Amount      int32 `json:"amount"`
	ColonyTotal int32 `json:"colony_total"`
}

type TrustData struct {
	OldTrust int32 `json:"old_trust"`
	NewTrust int32 `json:"new_trust"`
}

type HazardData struct {
	Damage           int64 `json:"damage"`
	BatteryRemaining int64 `json:"battery_remaining"`
}

type MissionData struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

// Sealed interface implementations — only these types satisfy EventPayload.
func (SpawnedData) isEventPayload()  {}
func (DiedData) isEventPayload()     {}
func (InfectedData) isEventPayload() {}
func (HarvestData) isEventPayload()  {}
func (DepositData) isEventPayload()  {}
func (TrustData) isEventPayload()    {}
func (HazardData) isEventPayload()   {}
func (MissionData) isEventPayload()  {}

// Event is an immutable record of one thing that happened during a tick.
// DroneID is -1 for events not tied to a specific drone.
// Data is a sealed EventPayload — the compiler rejects any non-payload type at the assignment site.
type Event struct {
	TickNum int64        `json:"tick"`
	Type    EventType    `json:"type"`
	DroneID int32        `json:"drone_id"`
	X       int32        `json:"x"`
	Y       int32        `json:"y"`
	Data    EventPayload `json:"data,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler for Event, dispatching the Data
// field to the correct concrete payload type based on the Type field.
// This is required because the sealed EventPayload interface has no default
// JSON dispatch — without this method the decoder cannot reconstruct the type.
func (e *Event) UnmarshalJSON(b []byte) error {
	var raw struct {
		TickNum int64           `json:"tick"`
		Type    EventType       `json:"type"`
		DroneID int32           `json:"drone_id"`
		X       int32           `json:"x"`
		Y       int32           `json:"y"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	e.TickNum = raw.TickNum
	e.Type = raw.Type
	e.DroneID = raw.DroneID
	e.X = raw.X
	e.Y = raw.Y

	if len(raw.Data) == 0 || string(raw.Data) == "null" {
		return nil
	}

	switch raw.Type {
	case EventDroneSpawned, EventFabricated:
		var d SpawnedData
		if err := json.Unmarshal(raw.Data, &d); err != nil {
			return err
		}
		e.Data = d
	case EventDroneDied:
		var d DiedData
		if err := json.Unmarshal(raw.Data, &d); err != nil {
			return err
		}
		e.Data = d
	case EventDroneInfected:
		var d InfectedData
		if err := json.Unmarshal(raw.Data, &d); err != nil {
			return err
		}
		e.Data = d
	case EventHarvested:
		var d HarvestData
		if err := json.Unmarshal(raw.Data, &d); err != nil {
			return err
		}
		e.Data = d
	case EventDeposited:
		var d DepositData
		if err := json.Unmarshal(raw.Data, &d); err != nil {
			return err
		}
		e.Data = d
	case EventTrustChanged:
		var d TrustData
		if err := json.Unmarshal(raw.Data, &d); err != nil {
			return err
		}
		e.Data = d
	case EventHazardDamage:
		var d HazardData
		if err := json.Unmarshal(raw.Data, &d); err != nil {
			return err
		}
		e.Data = d
	case EventMissionChanged:
		var d MissionData
		if err := json.Unmarshal(raw.Data, &d); err != nil {
			return err
		}
		e.Data = d
	}
	return nil
}

// EventBus uses a double-buffer so the snapshot is never overwritten by
// in-progress Emit calls. buffers[active] collects the current tick's events;
// buffers[active^1] is the stable snapshot from the previous Commit.
//
// Invariant: between Commit calls, only buffers[active] is written to.
// Nothing may call Emit after Commit() and before the next BeginTick.
type EventBus struct {
	buffers [2][]Event
	active  int // index of the buffer currently being emitted to
}

func NewEventBus() *EventBus {
	b := &EventBus{}
	b.buffers[0] = make([]Event, 0, 256)
	b.buffers[1] = make([]Event, 0, 256)
	return b
}

func (b *EventBus) Emit(e Event) {
	b.buffers[b.active] = append(b.buffers[b.active], e)
}

// BeginTick asserts the active buffer is empty at the start of each tick.
// If events are present, it means Emit was called after the last Commit —
// cross-tick contamination that would corrupt the snapshot for other readers.
// This turns a silent data bug into an immediate panic.
func (b *EventBus) BeginTick() {
	if len(b.buffers[b.active]) != 0 {
		panic("EventBus invariant violated: active buffer is not empty at BeginTick — cross-tick event contamination detected")
	}
}

// Commit closes the current tick's event window.
// The active buffer becomes the new snapshot; the previous snapshot buffer
// is reset and becomes the active buffer for the next tick.
// Called once per CommitTick — do not call Emit after this until BeginTick.
func (b *EventBus) Commit() {
	b.active ^= 1
	b.buffers[b.active] = b.buffers[b.active][:0]
}

// Events returns the stable snapshot from the last Commit.
// Safe for multiple consumers; the EventBus never writes to this slice
// after Commit and before the next Commit. Callers must not mutate the
// returned slice or its elements — doing so corrupts the snapshot for all
// other readers. Copy with slices.Clone if independent mutation is needed.
func (b *EventBus) Events() []Event {
	return b.buffers[b.active^1]
}

// PendingLen returns the number of events accumulated since the last Commit.
func (b *EventBus) PendingLen() int {
	return len(b.buffers[b.active])
}
