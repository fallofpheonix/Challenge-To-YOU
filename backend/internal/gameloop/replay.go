package gameloop

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

type ReplayRecorder struct {
	mu          sync.Mutex
	frames      []ReplayFrame
	seed        int64
	startTime   time.Time
	eventBuffer []GameEvent
}

func NewReplayRecorder(seed int64) *ReplayRecorder {
	return &ReplayRecorder{
		frames:    make([]ReplayFrame, 0, 1024),
		seed:      seed,
		startTime: time.Now(),
	}
}

func (rr *ReplayRecorder) RecordEvent(ev GameEvent) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	rr.eventBuffer = append(rr.eventBuffer, ev)
}

func hashSnapshot(tick int, state GameState, snapshot map[string]interface{}) string {
	data := struct {
		Tick     int                    `json:"tick"`
		State    string                 `json:"state"`
		Snapshot map[string]interface{} `json:"snapshot"`
	}{
		Tick:     tick,
		State:    state.String(),
		Snapshot: snapshot,
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return fmt.Sprintf("err:%v", err)
	}
	h := sha256.Sum256(encoded)
	return fmt.Sprintf("%x", h)
}

func sortMapKeys(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		result[k] = m[k]
	}
	return result
}

func (rr *ReplayRecorder) RecordTick(tick int, state GameState, snapshot map[string]interface{}) ReplayFrame {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	orderedSnapshot := sortMapKeys(snapshot)
	stateHash := hashSnapshot(tick, state, orderedSnapshot)

	frame := ReplayFrame{
		Tick:      tick,
		State:     state,
		Events:    make([]GameEvent, len(rr.eventBuffer)),
		StateHash: stateHash,
		Snapshot:  orderedSnapshot,
		ElapsedMS: time.Since(rr.startTime).Milliseconds(),
	}
	copy(frame.Events, rr.eventBuffer)
	rr.eventBuffer = rr.eventBuffer[:0]

	rr.frames = append(rr.frames, frame)
	return frame
}

func (rr *ReplayRecorder) Frames() []ReplayFrame {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	result := make([]ReplayFrame, len(rr.frames))
	copy(result, rr.frames)
	return result
}

func (rr *ReplayRecorder) FrameCount() int {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	return len(rr.frames)
}

func (rr *ReplayRecorder) VerifyIntegrity() (bool, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	for i, frame := range rr.frames {
		expected := hashSnapshot(frame.Tick, frame.State, frame.Snapshot)
		if frame.StateHash != expected {
			return false, fmt.Errorf("replay integrity check failed at frame %d: hash mismatch (expected %s, got %s)", i, expected, frame.StateHash)
		}
	}
	return true, nil
}

func (rr *ReplayRecorder) Reset(seed int64) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	rr.frames = make([]ReplayFrame, 0, 1024)
	rr.seed = seed
	rr.startTime = time.Now()
	rr.eventBuffer = nil
}

type ReplayPlayback struct {
	frames []ReplayFrame
	cursor int
}

func NewReplayPlayback(frames []ReplayFrame) *ReplayPlayback {
	return &ReplayPlayback{
		frames: frames,
		cursor: 0,
	}
}

func (rp *ReplayPlayback) Next() (*ReplayFrame, bool) {
	if rp.cursor >= len(rp.frames) {
		return nil, false
	}
	frame := rp.frames[rp.cursor]
	rp.cursor++
	return &frame, true
}

func (rp *ReplayPlayback) SeekTo(tick int) (*ReplayFrame, bool) {
	for i, frame := range rp.frames {
		if frame.Tick == tick {
			rp.cursor = i
			return &frame, true
		}
	}
	return nil, false
}

func (rp *ReplayPlayback) TotalFrames() int {
	return len(rp.frames)
}

func (rp *ReplayPlayback) Progress() float64 {
	if len(rp.frames) == 0 {
		return 0
	}
	return float64(rp.cursor) / float64(len(rp.frames))
}
