package replay

import (
	"chrysalis-engine/core/simulation"
	"encoding/json"
	"testing"
)

func makeEvents(tick int64, n int) []simulation.Event {
	evs := make([]simulation.Event, n)
	for i := range evs {
		evs[i] = simulation.Event{
			TickNum: tick,
			Type:    simulation.EventHarvested,
			DroneID: int32(i),
			Data:    simulation.HarvestData{ResourcesRemaining: int32(100 - i)},
		}
	}
	return evs
}

func TestRecorderAccumulatesFrames(t *testing.T) {
	r := NewRecorder(map[string]interface{}{"tick": 0}, 500)

	r.Record(1, makeEvents(1, 3))
	r.Record(2, makeEvents(2, 2))
	r.Record(3, nil)

	if r.TotalFrames() != 3 {
		t.Fatalf("expected 3 frames, got %d", r.TotalFrames())
	}
	if r.TotalEvents() != 5 {
		t.Fatalf("expected 5 events, got %d", r.TotalEvents())
	}
}

func TestRecorderCopiesEventSlice(t *testing.T) {
	r := NewRecorder(nil, 500)
	evs := makeEvents(1, 2)
	r.Record(1, evs)

	// Mutate the original slice — recorder's copy must be unaffected
	evs[0].DroneID = 999

	stored := r.EventsInRange(1, 1)
	if stored[0].DroneID == 999 {
		t.Fatal("recorder did not copy the event slice — mutation propagated")
	}
}

func TestNearestCheckpointFallsBackToInitial(t *testing.T) {
	initialState := map[string]interface{}{"tick": int64(0)}
	r := NewRecorder(initialState, 500)

	// No checkpoints added yet
	cp := r.NearestCheckpoint(250)
	if tick, ok := cp.State["tick"]; !ok || tick != int64(0) {
		t.Fatalf("expected initial snapshot, got %v", cp.State)
	}
}

func TestNearestCheckpointReturnsClosestPrior(t *testing.T) {
	r := NewRecorder(map[string]interface{}{"tick": int64(0)}, 500)
	r.Checkpoint(500, map[string]interface{}{"tick": int64(500)})
	r.Checkpoint(1000, map[string]interface{}{"tick": int64(1000)})
	r.Checkpoint(1500, map[string]interface{}{"tick": int64(1500)})

	cp := r.NearestCheckpoint(1200)
	if tick := cp.State["tick"].(int64); tick != 1000 {
		t.Fatalf("expected checkpoint at tick 1000, got tick %d", tick)
	}
}

func TestNearestCheckpointDoesNotReturnFutureCheckpoint(t *testing.T) {
	r := NewRecorder(map[string]interface{}{"tick": int64(0)}, 500)
	r.Checkpoint(500, map[string]interface{}{"tick": int64(500)})

	// Target is before the checkpoint
	cp := r.NearestCheckpoint(300)
	if tick := cp.State["tick"].(int64); tick != 0 {
		t.Fatalf("expected initial (tick 0), got tick %d", tick)
	}
}

func TestEventsInRange(t *testing.T) {
	r := NewRecorder(nil, 500)
	r.Record(10, makeEvents(10, 2))
	r.Record(11, makeEvents(11, 3))
	r.Record(12, makeEvents(12, 1))

	evs := r.EventsInRange(10, 11)
	if len(evs) != 5 {
		t.Fatalf("expected 5 events in range [10,11], got %d", len(evs))
	}
}

func TestSerializeDeserializeRoundTrip(t *testing.T) {
	r := NewRecorder(map[string]interface{}{"tick": int64(0), "seed": int64(42)}, 500)
	r.Record(1, makeEvents(1, 2))
	r.Record(2, makeEvents(2, 1))
	r.Checkpoint(500, map[string]interface{}{"tick": int64(500)})

	raw, err := r.Serialize()
	if err != nil {
		t.Fatalf("serialize failed: %v", err)
	}

	r2, err := Deserialize(raw, 500)
	if err != nil {
		t.Fatalf("deserialize failed: %v", err)
	}

	if r2.TotalFrames() != r.TotalFrames() {
		t.Fatalf("frame count mismatch: %d vs %d", r2.TotalFrames(), r.TotalFrames())
	}
	if r2.TotalEvents() != r.TotalEvents() {
		t.Fatalf("event count mismatch: %d vs %d", r2.TotalEvents(), r.TotalEvents())
	}
	if len(r2.checkpoints) != len(r.checkpoints) {
		t.Fatalf("checkpoint count mismatch")
	}
}

func TestSerializedJSONIsStable(t *testing.T) {
	// Two recorders with identical inputs must produce identical JSON.
	build := func() []byte {
		r := NewRecorder(map[string]interface{}{"tick": int64(0)}, 500)
		r.Record(1, makeEvents(1, 3))
		r.Record(2, makeEvents(2, 1))
		raw, _ := r.Serialize()
		return raw
	}

	if string(build()) != string(build()) {
		t.Fatal("serialized JSON is not stable across identical runs")
	}
}

func TestDeserializePreservesEventPayloadTypes(t *testing.T) {
	r := NewRecorder(nil, 500)
	r.Record(1, []simulation.Event{
		{
			TickNum: 1,
			Type:    simulation.EventHarvested,
			DroneID: 5,
			Data:    simulation.HarvestData{ResourcesRemaining: 42},
		},
	})

	raw, _ := r.Serialize()

	// After round-trip, Data will be map[string]interface{} (JSON unmarshals interfaces
	// as maps). Verify the field value is preserved even after type erasure.
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	frames := data["frames"].([]interface{})
	firstFrame := frames[0].(map[string]interface{})
	events := firstFrame["events"].([]interface{})
	ev := events[0].(map[string]interface{})
	payload := ev["data"].(map[string]interface{})

	if remaining, ok := payload["resources_remaining"]; !ok {
		t.Fatal("resources_remaining field missing after round-trip")
	} else {
		// JSON numbers deserialize as float64
		if remaining.(float64) != 42 {
			t.Fatalf("expected resources_remaining=42, got %v", remaining)
		}
	}
}
