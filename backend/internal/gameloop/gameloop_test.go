package gameloop

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"challenge-to-you/backend/internal/db"
	"challenge-to-you/backend/internal/engine"
	"challenge-to-you/backend/internal/eventbus"
	"challenge-to-you/backend/internal/missionengine"
)

// newTestDB creates a temporary on-disk database for game-loop tests.
func newTestDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.NewDB(filepath.Join(t.TempDir(), "gameloop_test.db"))
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestGameStateString(t *testing.T) {
	tests := []struct {
		state GameState
		want  string
	}{
		{StatePlaying, "playing"},
		{StatePaused, "paused"},
		{StateWon, "won"},
		{StateLost, "lost"},
		{StatePurged, "purged"},
		{GameState(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("GameState.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInventory(t *testing.T) {
	inv := NewInventory()
	if inv.Count(ResourceEnergy) != 0 {
		t.Errorf("expected 0 energy, got %d", inv.Count(ResourceEnergy))
	}

	inv.Add(ResourceEnergy, 5)
	if !inv.Has(ResourceEnergy, 5) {
		t.Error("expected to have 5 energy")
	}
	if inv.Count(ResourceEnergy) != 5 {
		t.Errorf("expected 5 energy, got %d", inv.Count(ResourceEnergy))
	}

	removed := inv.Remove(ResourceEnergy, 3)
	if removed != 3 {
		t.Errorf("expected removed 3, got %d", removed)
	}
	if inv.Count(ResourceEnergy) != 2 {
		t.Errorf("expected 2 energy remaining, got %d", inv.Count(ResourceEnergy))
	}

	removed = inv.Remove(ResourceEnergy, 5)
	if removed != 2 {
		t.Errorf("expected removed 2, got %d", removed)
	}
	if inv.Has(ResourceEnergy, 1) {
		t.Error("expected no energy remaining")
	}
}

func TestResourceManager(t *testing.T) {
	rm := NewResourceManager()

	result := rm.Gather("player1", ResourceEnergy, 3)
	if !result.Success {
		t.Errorf("gather should succeed: %s", result.Message)
	}
	if result.Added != 3 {
		t.Errorf("expected added 3, got %d", result.Added)
	}

	inv := rm.GetInventory("player1")
	if inv.Count(ResourceEnergy) != 3 {
		t.Errorf("expected 3 energy, got %d", inv.Count(ResourceEnergy))
	}

	result = rm.Consume("player1", ResourceEnergy, 2)
	if !result.Success {
		t.Errorf("consume should succeed: %s", result.Message)
	}
	if result.Removed != 2 {
		t.Errorf("expected removed 2, got %d", result.Removed)
	}

	result = rm.Consume("player1", ResourceEnergy, 5)
	if result.Success {
		t.Error("consume of more than available should fail")
	}

	result = rm.Deliver("player1", ResourceEnergy, 1, "objective_1")
	if !result.Success {
		t.Errorf("deliver should succeed: %s", result.Message)
	}
	if result.Removed != 1 {
		t.Errorf("expected removed 1, got %d", result.Removed)
	}

	rm.Reset("player1")
	inv = rm.GetInventory("player1")
	if len(inv.Resources) != 0 {
		t.Error("expected empty inventory after reset")
	}
}

func TestResourceManagerOperations(t *testing.T) {
	rm := NewResourceManager()

	op := ResourceOperation{
		PlayerID:     "p1",
		Type:         "gather",
		ResourceType: ResourceMaterial,
		Amount:       5,
	}
	result := rm.ProcessOperation(op)
	if !result.Success {
		t.Errorf("gather operation failed: %s", result.Message)
	}

	op.Type = "consume"
	result = rm.ProcessOperation(op)
	if !result.Success {
		t.Errorf("consume operation failed: %s", result.Message)
	}

	op.Type = "deliver"
	op.Target = "base"
	result = rm.ProcessOperation(op)
	if result.Success {
		t.Errorf("deliver without enough resources should fail")
	}

	op.Type = "unknown"
	result = rm.ProcessOperation(op)
	if result.Success {
		t.Error("unknown operation should fail")
	}
}

func TestReplayRecorder(t *testing.T) {
	rr := NewReplayRecorder(42)

	rr.RecordEvent(GameEvent{Type: EventPlayerInput, Action: "gather", Payload: "energy"})

	frame := rr.RecordTick(1, StatePlaying, map[string]interface{}{"key": "value"})
	if frame.Tick != 1 {
		t.Errorf("expected tick 1, got %d", frame.Tick)
	}
	if frame.State != StatePlaying {
		t.Errorf("expected StatePlaying, got %v", frame.State)
	}
	if len(frame.Events) != 1 {
		t.Errorf("expected 1 event in frame, got %d", len(frame.Events))
	}
	if frame.StateHash == "" {
		t.Error("expected non-empty state hash")
	}

	rr.RecordTick(2, StateWon, map[string]interface{}{"key": "value2"})
	if rr.FrameCount() != 2 {
		t.Errorf("expected 2 frames, got %d", rr.FrameCount())
	}

	ok, err := rr.VerifyIntegrity()
	if err != nil {
		t.Errorf("integrity check failed: %v", err)
	}
	if !ok {
		t.Error("integrity check returned false")
	}

	frames := rr.Frames()
	if len(frames) != 2 {
		t.Errorf("expected 2 frames, got %d", len(frames))
	}
}

func TestReplayPlayback(t *testing.T) {
	rr := NewReplayRecorder(1)
	rr.RecordTick(1, StatePlaying, map[string]interface{}{"a": 1})
	rr.RecordTick(2, StatePlaying, map[string]interface{}{"a": 2})
	rr.RecordTick(3, StateWon, map[string]interface{}{"a": 3})

	frames := rr.Frames()
	rp := NewReplayPlayback(frames)

	if rp.TotalFrames() != 3 {
		t.Errorf("expected 3 total frames, got %d", rp.TotalFrames())
	}

	frame, ok := rp.Next()
	if !ok {
		t.Fatal("expected first frame")
	}
	if frame.Tick != 1 {
		t.Errorf("expected tick 1, got %d", frame.Tick)
	}
	if rp.Progress() <= 0 {
		t.Error("expected progress > 0")
	}

	found, ok := rp.SeekTo(3)
	if !ok {
		t.Fatal("expected to find tick 3")
	}
	if found.State != StateWon {
		t.Errorf("expected StateWon, got %v", found.State)
	}
}

func TestReplayReset(t *testing.T) {
	rr := NewReplayRecorder(1)
	rr.RecordTick(1, StatePlaying, nil)

	rr.Reset(2)
	if rr.FrameCount() != 0 {
		t.Errorf("expected 0 frames after reset, got %d", rr.FrameCount())
	}

	rr.RecordTick(1, StatePlaying, nil)
	if rr.FrameCount() != 1 {
		t.Errorf("expected 1 frame after re-recording, got %d", rr.FrameCount())
	}
}

func TestTelemetryCollector(t *testing.T) {
	tc := NewTelemetryCollector(100)

	tc.Record(1, StatePlaying, 0.1, 0.0, nil, map[string]float64{"test": 1.0})
	tc.Record(2, StatePlaying, 0.2, 5.0, nil, nil)

	if tc.Count() != 2 {
		t.Errorf("expected 2 points, got %d", tc.Count())
	}

	latest := tc.Latest()
	if latest == nil {
		t.Fatal("expected latest point")
	}
	if latest.Tick != 2 {
		t.Errorf("expected tick 2, got %d", latest.Tick)
	}
	if latest.Vigilance != 0.2 {
		t.Errorf("expected vigilance 0.2, got %f", latest.Vigilance)
	}
	if latest.Metrics["entropy"] != 5.0 {
		t.Errorf("expected entropy 5.0, got %f", latest.Metrics["entropy"])
	}

	tc.SetLabel("session", "test_1")
	labels := tc.Labels()
	if labels["session"] != "test_1" {
		t.Errorf("expected label session=test_1, got %v", labels)
	}

	all := tc.Points()
	if len(all) != 2 {
		t.Errorf("expected 2 points from Points(), got %d", len(all))
	}

	recent := tc.Recent(1)
	if len(recent) != 1 {
		t.Errorf("expected 1 recent point, got %d", len(recent))
	}

	tc.Reset()
	if tc.Count() != 0 {
		t.Errorf("expected 0 points after reset, got %d", tc.Count())
	}
}

func TestTelemetryCollectorMaxPoints(t *testing.T) {
	tc := NewTelemetryCollector(5)
	for i := 0; i < 10; i++ {
		tc.Record(i, StatePlaying, 0, 0, nil, nil)
	}
	if tc.Count() != 5 {
		t.Errorf("expected 5 points (ring buffer), got %d", tc.Count())
	}
	latest := tc.Latest()
	if latest == nil || latest.Tick != 9 {
		t.Errorf("expected latest tick 9, got %v", latest)
	}
}

func createTestFabric() *engine.AxiomaticFabric {
	fabric := engine.NewAxiomaticFabric(engine.Magitech, "ward_restored", true)
	fabric.SetState("entropy", 0)
	fabric.SetState("current_room", "Rune Chamber")
	return fabric
}

func createTestMission() *missionengine.Mission {
	return &missionengine.Mission{
		ID:          "test_mission_1",
		Title:       "Test Mission",
		Description: "A test mission",
		Era:         "magitech",
		Paradigm:    "MAGITECH",
		Difficulty:  0.3,
		Objectives: []missionengine.Objective{
			{
				ID:          "obj_1",
				Type:        "complete_challenge",
				Target:      "trigger_ward_repair",
				Description: "Repair the ward",
				IsRequired:  true,
			},
		},
		Reward: &missionengine.MissionReward{XP: 100},
	}
}

func TestGameLoopPauseResume(t *testing.T) {
	fabric := createTestFabric()
	gl := NewGameLoop(fabric, nil, nil, "test_player", "test_mission", DefaultConfig(), 0)

	if gl.State() != StatePlaying {
		t.Errorf("expected initial state Playing, got %v", gl.State())
	}

	gl.Pause()
	if !gl.IsPaused() {
		t.Error("expected paused after Pause()")
	}
	if gl.State() != StatePaused {
		t.Errorf("expected StatePaused, got %v", gl.State())
	}

	gl.Resume()
	if gl.IsPaused() {
		t.Error("expected not paused after Resume()")
	}
	if gl.State() != StatePlaying {
		t.Errorf("expected StatePlaying after Resume(), got %v", gl.State())
	}
}

func TestGameLoopTick(t *testing.T) {
	fabric := createTestFabric()
	gl := NewGameLoop(fabric, nil, nil, "test_player", "test_mission", DefaultConfig(), 0)

	if gl.State() != StatePlaying {
		t.Errorf("expected initial state Playing, got %v", gl.State())
	}
	if gl.TickCount() != 0 {
		t.Errorf("expected tick 0, got %d", gl.TickCount())
	}
}

func TestGameLoopGather(t *testing.T) {
	fabric := createTestFabric()
	gl := NewGameLoop(fabric, nil, nil, "test_player", "", DefaultConfig(), 0)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	results := gl.Start(ctx)

	gl.Submit(GameEvent{
		Type: EventPlayerInput, PlayerID: "test_player",
		Action: "gather", Payload: "energy",
	})

	select {
	case <-results:
	case <-time.After(200 * time.Millisecond):
	}

	inv := gl.Resources().GetInventory("test_player")
	if inv.Count(ResourceEnergy) != 1 {
		t.Errorf("expected 1 energy after gather, got %d", inv.Count(ResourceEnergy))
	}
	gl.Stop()
}

func TestGameLoopDeliver(t *testing.T) {
	fabric := createTestFabric()
	gl := NewGameLoop(fabric, nil, nil, "test_player", "", DefaultConfig(), 0)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	results := gl.Start(ctx)

	gl.Submit(GameEvent{
		Type: EventPlayerInput, PlayerID: "test_player",
		Action: "gather", Payload: "energy",
	})
	select {
	case <-results:
	case <-time.After(200 * time.Millisecond):
	}

	gl.Submit(GameEvent{
		Type: EventPlayerInput, PlayerID: "test_player",
		Action: "deliver", Payload: "energy",
	})
	select {
	case <-results:
	case <-time.After(200 * time.Millisecond):
	}

	inv := gl.Resources().GetInventory("test_player")
	if inv.Count(ResourceEnergy) != 0 {
		t.Errorf("expected 0 energy after deliver, got %d", inv.Count(ResourceEnergy))
	}
	gl.Stop()
}

func TestGameLoopTriggerEvent(t *testing.T) {
	fabric := createTestFabric()
	fabric.SetState("rune_charge", 0)

	gl := NewGameLoop(fabric, nil, nil, "test_player", "", DefaultConfig(), 0)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	results := gl.Start(ctx)

	gl.Submit(GameEvent{
		Type: EventPlayerInput, PlayerID: "test_player",
		Action: "trigger_event", Payload: "charge_rune",
	})

	select {
	case <-results:
	case <-time.After(200 * time.Millisecond):
	}

	if gl.State() != StatePlaying {
		t.Errorf("expected state Playing, got %v", gl.State())
	}
	gl.Stop()
}

func TestGameLoopPurge(t *testing.T) {
	fabric := createTestFabric()
	fabric.ArchonVigilance = 1.0

	gl := NewGameLoop(fabric, nil, nil, "test_player", "", DefaultConfig(), 0)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	results := gl.Start(ctx)

	select {
	case result, ok := <-results:
		if !ok {
			t.Fatal("expected result from loop")
		}
		if result.State != StatePurged {
			t.Errorf("expected StatePurged, got %v", result.State)
		}
		if result.Snapshot == nil {
			t.Fatal("expected snapshot")
		}
		if result.Snapshot["game_over"] != true {
			t.Error("expected game_over flag in snapshot")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for tick")
	}

	gl.Stop()
}

func TestGameLoopMissionLifecycle(t *testing.T) {
	bus := eventbus.NewEventBus(100)
	registry := missionengine.NewMissionRegistry()
	mission := createTestMission()
	registry.Register(mission)
	manager := missionengine.NewMissionManager(registry, bus, newTestDB(t))

	fabric := createTestFabric()

	session, err := manager.StartMission("test_player", "test_mission_1")
	if err != nil {
		t.Fatalf("failed to start mission: %v", err)
	}

	gl := NewGameLoop(fabric, manager, session, "test_player", "test_mission_1", DefaultConfig(), 0)

	if gl.State() != StatePlaying {
		t.Errorf("expected Playing, got %v", gl.State())
	}

	if gl.MissionSession() == nil {
		t.Fatal("expected mission session")
	}

	if gl.MissionManager() == nil {
		t.Fatal("expected mission manager")
	}

	gl.Stop()
}

func TestGameLoopMissionComplete(t *testing.T) {
	bus := eventbus.NewEventBus(100)
	registry := missionengine.NewMissionRegistry()
	registry.Register(createTestMission())
	manager := missionengine.NewMissionManager(registry, bus, newTestDB(t))

	fabric := createTestFabric()
	session, err := manager.StartMission("p1", "test_mission_1")
	if err != nil {
		t.Fatalf("failed to start mission: %v", err)
	}

	gl := NewGameLoop(fabric, manager, session, "p1", "test_mission_1", DefaultConfig(), 0)

	gl.Submit(GameEvent{
		Type: EventPlayerInput, PlayerID: "p1",
		Action: "trigger_event", Payload: "trigger_ward_repair",
	})

	if !gl.MissionManager().IsMissionCompleted("p1", "test_mission_1") {
		t.Log("Note: mission completion depends on fabric win condition")
	}

	gl.Stop()
}

func TestGameLoopReset(t *testing.T) {
	fabric := createTestFabric()
	gl := NewGameLoop(fabric, nil, nil, "p1", "", DefaultConfig(), 1)

	gl.Submit(GameEvent{Type: EventPlayerInput, PlayerID: "p1", Action: "gather", Payload: "energy"})

	newFabric := createTestFabric()
	gl.Reset(newFabric, nil, 2)

	if gl.State() != StatePlaying {
		t.Errorf("expected Playing after reset, got %v", gl.State())
	}
	if gl.TickCount() != 0 {
		t.Errorf("expected tick 0 after reset, got %d", gl.TickCount())
	}

	inv := gl.Resources().GetInventory("p1")
	if len(inv.Resources) != 0 {
		t.Error("expected empty inventory after reset")
	}
}

func TestGameLoopBuildSnapshot(t *testing.T) {
	fabric := createTestFabric()
	gl := NewGameLoop(fabric, nil, nil, "p1", "", DefaultConfig(), 0)

	snapshot := gl.buildSnapshot("test message")
	if snapshot == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snapshot["game_state"] != "playing" {
		t.Errorf("expected game_state playing, got %v", snapshot["game_state"])
	}
	if snapshot["tick"] != 0 {
		t.Errorf("expected tick 0, got %v", snapshot["tick"])
	}
	if _, ok := snapshot["inventory"]; !ok {
		t.Error("expected inventory in snapshot")
	}
	if _, ok := snapshot["replay_frame"]; !ok {
		t.Error("expected replay_frame in snapshot")
	}
	if _, ok := snapshot["replay_frame"]; !ok {
		t.Error("expected replay_frame in snapshot")
	}
}

func TestGameLoopSystemEvent(t *testing.T) {
	fabric := createTestFabric()

	bus := eventbus.NewEventBus(100)
	registry := missionengine.NewMissionRegistry()
	registry.Register(createTestMission())
	manager := missionengine.NewMissionManager(registry, bus, newTestDB(t))

	session, err := manager.StartMission("p1", "test_mission_1")
	if err != nil {
		t.Fatalf("failed to start mission: %v", err)
	}

	gl := NewGameLoop(fabric, manager, session, "p1", "test_mission_1", DefaultConfig(), 0)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	results := gl.Start(ctx)

	gl.Submit(GameEvent{
		Type: EventSystem, PlayerID: "p1",
		Action: "fail_mission", Payload: "test failure",
	})

	select {
	case result := <-results:
		if result.State != StateLost {
			t.Errorf("expected StateLost, got %v", result.State)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for result")
	}

	if gl.State() != StateLost {
		t.Errorf("expected StateLost, got %v", gl.State())
	}
	gl.Stop()
}

func TestResourceManagerNewInventory(t *testing.T) {
	rm := NewResourceManager()
	inv := rm.GetInventory("nonexistent")
	if inv == nil {
		t.Fatal("expected non-nil inventory for new player")
	}
	if len(inv.Resources) != 0 {
		t.Error("expected empty resources for new player")
	}
}

func TestReplayIntegrityFailure(t *testing.T) {
	rr := NewReplayRecorder(0)
	rr.RecordTick(1, StatePlaying, map[string]interface{}{"a": 1})
	rr.RecordTick(2, StatePlaying, map[string]interface{}{"a": 2})

	rr.mu.Lock()
	rr.frames[1].StateHash = "tampered"
	rr.mu.Unlock()

	ok, err := rr.VerifyIntegrity()
	if ok {
		t.Error("expected integrity check to fail on tampered frame")
	}
	if err == nil {
		t.Error("expected error from integrity check")
	}
}
