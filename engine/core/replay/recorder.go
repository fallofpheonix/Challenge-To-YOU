// Package replay records the simulation event stream and periodic world
// snapshots so that any tick can be reconstructed without replaying from
// the beginning. It is a pure consumer of the EventBus — it reads snapshots
// but never writes to the engine. The engine has no knowledge that recording
// exists.
//
// Architecture:
//
//	Engine → EventBus → Commit() → bus.Events()
//	                                    │
//	                                    ▼
//	                              Recorder.Record()
//	                                    │
//	                          ┌─────────┴──────────┐
//	                          ▼                    ▼
//	                     FrameEvents          Checkpoint
//	                     (every tick)         (every N ticks)
package replay

import (
	"chrysalis-engine/core/simulation"
	"encoding/json"
)

// ReplayMeta captures immutable session metadata recorded at the start of a run.
// Used for validation on load and for display in the replay UI.
type ReplayMeta struct {
	Seed       int64  `json:"seed"`
	Version    string `json:"version"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	DroneCount int    `json:"drone_count"`
	StartedAt  int64  `json:"started_at"` // Unix timestamp
	LevelID    string `json:"level_id,omitempty"`
}

// FrameEvents holds the event log for a single simulation tick.
type FrameEvents struct {
	Tick   int64              `json:"tick"`
	Events []simulation.Event `json:"events"`
}

// Checkpoint is a full world-state snapshot used for efficient seeking.
// WorldHash is the FNV-64a hash of the engine state at this tick; it is used
// during replay validation to detect determinism divergence immediately.
type Checkpoint struct {
	Tick      int64                  `json:"tick"`
	State     map[string]interface{} `json:"state"`
	WorldHash uint64                 `json:"world_hash,omitempty"`
}

// ReplayData is the serializable record of a complete run. It can be written
// to disk and loaded back to replay any tick range.
type ReplayData struct {
	Meta        ReplayMeta    `json:"meta"`
	Initial     Checkpoint    `json:"initial"`
	Checkpoints []Checkpoint  `json:"checkpoints"`
	Frames      []FrameEvents `json:"frames"`
}

// Recorder subscribes to the EventBus snapshot each tick and accumulates the
// replay archive. Checkpoints are created automatically every CheckpointEvery
// ticks when a stateProvider is supplied via Checkpoint().
//
// Usage — one call per tick after CommitTick:
//
//	recorder.Record(engine.Tick, engine.Bus.Events())
//	if engine.Tick%recorder.CheckpointEvery == 0 {
//	    recorder.Checkpoint(engine.Tick, engine.GetState())
//	}
type Recorder struct {
	CheckpointEvery int64
	Meta            ReplayMeta
	frames          []FrameEvents
	checkpoints     []Checkpoint
	initial         Checkpoint
	totalEvents     int
}

// NewRecorder creates a Recorder with an initial world snapshot.
// checkpointEvery controls how often full state snapshots are created.
func NewRecorder(initialState map[string]interface{}, checkpointEvery int64) *Recorder {
	return &Recorder{
		CheckpointEvery: checkpointEvery,
		initial: Checkpoint{
			Tick:  0,
			State: initialState,
		},
	}
}

// Record appends this tick's events to the archive. Events is a reference to
// the EventBus snapshot — Record copies the slice to protect against the next
// Commit rotating the buffer.
func (r *Recorder) Record(tick int64, events []simulation.Event) {
	if len(events) == 0 {
		// Still record the frame so tick numbers are contiguous.
		r.frames = append(r.frames, FrameEvents{Tick: tick})
		return
	}
	copied := make([]simulation.Event, len(events))
	copy(copied, events)
	r.frames = append(r.frames, FrameEvents{Tick: tick, Events: copied})
	r.totalEvents += len(events)
}

// SetMeta records session-level metadata (seed, version, map dimensions).
// Call once after NewRecorder before the first tick.
func (r *Recorder) SetMeta(meta ReplayMeta) {
	r.Meta = meta
}

// CheckpointCount returns the number of intermediate checkpoints recorded.
func (r *Recorder) CheckpointCount() int {
	return len(r.checkpoints)
}

// Checkpoint records a full world snapshot at the given tick. Call this every
// CheckpointEvery ticks alongside Record. State is the output of engine.GetState();
// it is shallow-copied so later mutation by the caller cannot corrupt the archive.
// The optional worldHash argument is stored alongside the state for replay validation.
func (r *Recorder) Checkpoint(tick int64, state map[string]interface{}, worldHash ...uint64) {
	copied := make(map[string]interface{}, len(state))
	for k, v := range state {
		copied[k] = v
	}
	var hash uint64
	if len(worldHash) > 0 {
		hash = worldHash[0]
	}
	r.checkpoints = append(r.checkpoints, Checkpoint{Tick: tick, State: copied, WorldHash: hash})
}

// TotalFrames returns the number of ticks recorded.
func (r *Recorder) TotalFrames() int {
	return len(r.frames)
}

// TotalEvents returns the total number of events recorded across all ticks.
func (r *Recorder) TotalEvents() int {
	return r.totalEvents
}

// NearestCheckpoint returns the checkpoint closest to but not exceeding targetTick.
// Returns the initial snapshot if no intermediate checkpoint precedes targetTick.
func (r *Recorder) NearestCheckpoint(targetTick int64) Checkpoint {
	best := r.initial
	for _, cp := range r.checkpoints {
		if cp.Tick <= targetTick {
			best = cp
		}
	}
	return best
}

// EventsInRange returns all recorded events between startTick and endTick inclusive.
// Used by replay playback to feed events into the inspector or renderer.
func (r *Recorder) EventsInRange(startTick, endTick int64) []simulation.Event {
	var result []simulation.Event
	for _, frame := range r.frames {
		if frame.Tick >= startTick && frame.Tick <= endTick {
			result = append(result, frame.Events...)
		}
	}
	return result
}

// Serialize produces the full replay archive as JSON.
// The result can be written to a save file and loaded back with Deserialize.
func (r *Recorder) Serialize() ([]byte, error) {
	data := ReplayData{
		Meta:        r.Meta,
		Initial:     r.initial,
		Checkpoints: r.checkpoints,
		Frames:      r.frames,
	}
	return json.Marshal(data)
}

// Deserialize reconstructs a Recorder from a previously serialized archive.
func Deserialize(raw []byte, checkpointEvery int64) (*Recorder, error) {
	var data ReplayData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	r := &Recorder{
		CheckpointEvery: checkpointEvery,
		Meta:            data.Meta,
		initial:         data.Initial,
		checkpoints:     data.Checkpoints,
		frames:          data.Frames,
	}
	for _, f := range r.frames {
		r.totalEvents += len(f.Events)
	}
	return r, nil
}
