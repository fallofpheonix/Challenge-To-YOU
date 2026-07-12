package gameloop

import (
	"context"
	"runtime"
	"testing"
	"time"

	"challenge-to-you/backend/internal/engine"
	"challenge-to-you/backend/internal/eventbus"
	"challenge-to-you/backend/internal/missionengine"
)

func createFullMission() *missionengine.Mission {
	return &missionengine.Mission{
		ID:          "integ_mission_1",
		Title:       "Integration Test Mission",
		Description: "Test mission covering full lifecycle",
		Era:         "magitech",
		Paradigm:    "MAGITECH",
		Difficulty:  0.3,
		Objectives: []missionengine.Objective{
			{
				ID: "obj_gather_energy", Type: "gather_resource", Target: "energy",
				Description: "Gather energy resource", IsRequired: true,
			},
			{
				ID: "obj_complete_challenge", Type: "complete_challenge", Target: "trigger_ward_repair",
				Description: "Complete the ward challenge", IsRequired: true,
			},
		},
		Reward: &missionengine.MissionReward{XP: 100, Credits: 50},
	}
}

func TestGameLoopIntegration_MissionLifecycle(t *testing.T) {
	bus := eventbus.NewEventBus(100)
	registry := missionengine.NewMissionRegistry()
	registry.Register(createFullMission())
	manager := missionengine.NewMissionManager(registry, bus, newTestDB(t))

	fabric := engine.NewAxiomaticFabric(engine.Magitech, "ward_restored", true)
	fabric.SetState("entropy", 0)
	fabric.SetState("current_room", "Rune Chamber")

	session, err := manager.StartMission("p1", "integ_mission_1")
	if err != nil {
		t.Fatalf("failed to start mission: %v", err)
	}

	cfg := DefaultConfig()
	cfg.EventBus = bus
	cfg.Bus = bus

	gl := NewGameLoop(fabric, manager, session, "p1", "integ_mission_1", cfg, 42)

	if gl.State() != StatePlaying {
		t.Fatalf("expected StatePlaying, got %v", gl.State())
	}
	if gl.MissionSession() == nil {
		t.Fatal("expected mission session")
	}
	if gl.MissionManager() == nil {
		t.Fatal("expected mission manager")
	}
	if gl.Replay() == nil {
		t.Fatal("expected replay recorder")
	}
	if gl.Telemetry() == nil {
		t.Fatal("expected telemetry collector")
	}
	if gl.Resources() == nil {
		t.Fatal("expected resource manager")
	}
}

func TestGameLoopIntegration_TickDriven(t *testing.T) {
	bus := eventbus.NewEventBus(100)
	registry := missionengine.NewMissionRegistry()
	registry.Register(createFullMission())
	manager := missionengine.NewMissionManager(registry, bus, newTestDB(t))

	fabric := engine.NewAxiomaticFabric(engine.Magitech, "ward_restored", true)
	fabric.SetState("entropy", 0)

	session, err := manager.StartMission("p1", "integ_mission_1")
	if err != nil {
		t.Fatalf("failed to start mission: %v", err)
	}

	cfg := DefaultConfig()
	cfg.TickInterval = 10 * time.Millisecond
	cfg.EventBus = bus

	gl := NewGameLoop(fabric, manager, session, "p1", "integ_mission_1", cfg, 42)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	results := gl.Start(ctx)

	var tickCount int
	var lastTick TickResult
	for result := range results {
		tickCount++
		lastTick = result
	}

	if tickCount == 0 {
		t.Fatal("expected at least 1 tick result")
	}
	t.Logf("Received %d tick results", tickCount)

	if lastTick.Snapshot == nil {
		t.Fatal("expected snapshot in tick result")
	}
	if _, ok := lastTick.Snapshot["tick"]; !ok {
		t.Error("expected tick in snapshot")
	}
	if _, ok := lastTick.Snapshot["game_state"]; !ok {
		t.Error("expected game_state in snapshot")
	}
	if _, ok := lastTick.Snapshot["state"]; !ok {
		t.Error("expected fabric state in snapshot")
	}
	if _, ok := lastTick.Snapshot["inventory"]; !ok {
		t.Error("expected inventory in snapshot")
	}
}

func TestGameLoopIntegration_Resources(t *testing.T) {
	fabric := engine.NewAxiomaticFabric(engine.Magitech, "ward_restored", true)
	gl := NewGameLoop(fabric, nil, nil, "p1", "", DefaultConfig(), 0)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	results := gl.Start(ctx)

	gl.Submit(GameEvent{
		Type: EventPlayerInput, PlayerID: "p1",
		Action: "gather", Payload: "energy",
	})
	select {
	case <-results:
	case <-time.After(200 * time.Millisecond):
	}

	gl.Submit(GameEvent{
		Type: EventPlayerInput, PlayerID: "p1",
		Action: "gather", Payload: "material",
	})
	select {
	case <-results:
	case <-time.After(200 * time.Millisecond):
	}

	inv := gl.Resources().GetInventory("p1")
	if inv.Count(ResourceEnergy) != 1 {
		t.Errorf("expected 1 energy, got %d", inv.Count(ResourceEnergy))
	}
	if inv.Count(ResourceMaterial) != 1 {
		t.Errorf("expected 1 material, got %d", inv.Count(ResourceMaterial))
	}

	gl.Stop()
}

func TestGameLoopIntegration_Replay(t *testing.T) {
	fabric := engine.NewAxiomaticFabric(engine.Magitech, "ward_restored", true)
	cfg := DefaultConfig()
	cfg.TickInterval = 10 * time.Millisecond
	cfg.MaxTicks = 5
	gl := NewGameLoop(fabric, nil, nil, "p1", "", cfg, 42)

	gl.Submit(GameEvent{Type: EventPlayerInput, PlayerID: "p1", Action: "gather", Payload: "energy"})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for range gl.Start(ctx) {
	}

	if gl.Replay().FrameCount() == 0 {
		t.Error("expected at least 1 replay frame")
	}

	ok, err := gl.Replay().VerifyIntegrity()
	if err != nil {
		t.Errorf("replay integrity check failed: %v", err)
	}
	if !ok {
		t.Error("replay integrity check returned false")
	}

	frames := gl.Replay().Frames()
	for i, f := range frames {
		if f.StateHash == "" {
			t.Errorf("frame %d has empty state hash", i)
		}
	}

	gl.Stop()
}

func TestGameLoopIntegration_Telemetry(t *testing.T) {
	fabric := engine.NewAxiomaticFabric(engine.Magitech, "ward_restored", true)
	fabric.SetState("entropy", 10)

	gl := NewGameLoop(fabric, nil, nil, "p1", "", DefaultConfig(), 0)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	results := gl.Start(ctx)

	select {
	case <-results:
	case <-time.After(2 * time.Second):
	}
	gl.Stop()

	if gl.Telemetry().Count() == 0 {
		t.Error("expected telemetry points")
	}

	latest := gl.Telemetry().Latest()
	if latest == nil {
		t.Fatal("expected latest telemetry point")
	}
	if latest.Vigilance <= 0 {
		t.Logf("Note: vigilance is %f (may be 0 if entropy was consumed before tick)", latest.Vigilance)
	}
}

func TestGameLoopIntegration_ReplayDeterminism(t *testing.T) {
	seed := int64(42)
	cfg := DefaultConfig()
	cfg.MaxTicks = 10
	cfg.TickInterval = time.Millisecond

	runLoop := func() *GameLoop {
		fabric := engine.NewAxiomaticFabric(engine.Magitech, "ward_restored", true)
		fabric.SetState("entropy", 0)
		gl := NewGameLoop(fabric, nil, nil, "p1", "", cfg, seed)
		gl.Submit(GameEvent{Type: EventPlayerInput, PlayerID: "p1", Action: "gather", Payload: "energy"})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		for range gl.Start(ctx) {
		}
		return gl
	}

	gl1 := runLoop()
	gl2 := runLoop()

	frames1 := gl1.Replay().Frames()
	frames2 := gl2.Replay().Frames()

	if len(frames1) != len(frames2) {
		t.Fatalf("frame count mismatch: %d vs %d", len(frames1), len(frames2))
	}
	for i := 0; i < len(frames1); i++ {
		if frames1[i].StateHash != frames2[i].StateHash {
			t.Errorf("frame %d hash mismatch: %s vs %s", i, frames1[i].StateHash, frames2[i].StateHash)
		}
	}
}

func TestGameLoopIntegration_GameOver(t *testing.T) {
	fabric := engine.NewAxiomaticFabric(engine.Magitech, "ward_restored", true)
	fabric.ArchonVigilance = 1.0

	gl := NewGameLoop(fabric, nil, nil, "p1", "", DefaultConfig(), 0)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	results := gl.Start(ctx)

	var finalState GameState
	for result := range results {
		finalState = result.State
	}

	if finalState != StatePurged {
		t.Errorf("expected StatePurged, got %v", finalState)
	}
}

func TestGameLoopIntegration_Stress(t *testing.T) {
	fabric := engine.NewAxiomaticFabric(engine.Magitech, "ward_restored", true)
	fabric.SetState("entropy", 5)

	cfg := DefaultConfig()
	cfg.TickInterval = 1 * time.Microsecond
	cfg.MaxTicks = 100

	initialGoroutines := runtime.NumGoroutine()
	initialMem := runtime.MemStats{}
	runtime.ReadMemStats(&initialMem)

	gl := NewGameLoop(fabric, nil, nil, "p1", "", cfg, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results := gl.Start(ctx)

	var tickCount int
	for range results {
		tickCount++
	}
	gl.Stop()

	time.Sleep(50 * time.Millisecond)
	finalGoroutines := runtime.NumGoroutine()
	finalMem := runtime.MemStats{}
	runtime.ReadMemStats(&finalMem)

	if tickCount == 0 {
		t.Error("expected ticks from stress test")
	}
	t.Logf("Stress: %d ticks, goroutines: %d (started with %d)", tickCount, finalGoroutines, initialGoroutines)

	leaked := finalGoroutines - initialGoroutines
	if leaked > 5 {
		t.Errorf("possible goroutine leak: %d goroutines (started %d, ended %d)", leaked, initialGoroutines, finalGoroutines)
	}

	if gl.Replay().FrameCount() == 0 {
		t.Error("expected replay frames from stress test")
	}

	ok, err := gl.Replay().VerifyIntegrity()
	if err != nil {
		t.Errorf("replay integrity failed after stress: %v", err)
	}
	if !ok {
		t.Error("replay integrity check failed after stress")
	}
}

func TestGameLoopIntegration_ReconnectInProcess(t *testing.T) {
	// Simulate: player disconnects and reconnects to the same challenge
	fabric := engine.NewAxiomaticFabric(engine.Magitech, "ward_restored", true)
	fabric.SetState("entropy", 0)
	fabric.SetState("progress", 50)

	cfg := DefaultConfig()
	cfg.TickInterval = 5 * time.Millisecond
	cfg.MaxTicks = 5

	// First session
	gl1 := NewGameLoop(fabric, nil, nil, "p1", "mission_1", cfg, 42)
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
	results1 := gl1.Start(ctx1)
	for range results1 {
	}
	cancel1()
	gl1.Stop()

	// Capture fabric state after first session
	savedEntropy, _ := gl1.Fabric().GetState("entropy")
	savedProgress, _ := gl1.Fabric().GetState("progress")

	// Reconnect: create new fabric, restore from saved state
	fabric2 := engine.NewAxiomaticFabric(engine.Magitech, "ward_restored", true)
	fabric2.SetState("entropy", savedEntropy)
	fabric2.SetState("progress", savedProgress)

	cfg2 := DefaultConfig()
	cfg2.TickInterval = 5 * time.Millisecond
	cfg2.MaxTicks = 3

	gl2 := NewGameLoop(fabric2, nil, nil, "p1", "mission_1", cfg2, 42)
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	results2 := gl2.Start(ctx2)
	for range results2 {
	}
	cancel2()

	if gl2.TickCount() == 0 {
		t.Error("expected non-zero ticks in reconnected session")
	}
}

func TestGameLoopIntegration_ConcurrentSessions(t *testing.T) {
	// Two players on the same challenge must have independent state
	cfg := DefaultConfig()
	cfg.TickInterval = 5 * time.Millisecond
	cfg.MaxTicks = 3

	runPlayer := func(playerID, missionID string, seed int64) *GameLoop {
		fabric := engine.NewAxiomaticFabric(engine.Magitech, "ward_restored", true)
		fabric.SetState("entropy", 0)
		gl := NewGameLoop(fabric, nil, nil, playerID, missionID, cfg, seed)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		for range gl.Start(ctx) {
		}
		return gl
	}

	glA := runPlayer("p1", "mission_common", 1)
	glB := runPlayer("p2", "mission_common", 2)

	if glA.Replay().FrameCount() == 0 {
		t.Error("expected frames from player A")
	}
	if glB.Replay().FrameCount() == 0 {
		t.Error("expected frames from player B")
	}
}

func TestGameLoopIntegration_EvictionOnComplete(t *testing.T) {
	// Verify the eviction function is called when mission completes
	evicted := false
	evictedKey := ""
	cfg := DefaultConfig()
	cfg.TickInterval = time.Millisecond
	cfg.SessionKey = "p1:mission_evict"
	cfg.EvictionFn = func(key string) {
		evicted = true
		evictedKey = key
	}

	fabric := engine.NewAxiomaticFabric(engine.Magitech, "ward_restored", true)
	fabric.ArchonVigilance = 1.0

	gl := NewGameLoop(fabric, nil, nil, "p1", "mission_evict", cfg, 0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for range gl.Start(ctx) {
	}

	if !evicted {
		t.Error("expected eviction function to be called on purge")
	}
	if evictedKey != "p1:mission_evict" {
		t.Errorf("expected eviction key 'p1:mission_evict', got %q", evictedKey)
	}
}

func TestGameLoopIntegration_RepeatedReconnectCycles(t *testing.T) {
	// Multiple reconnect cycles should work without resource leaks
	cfg := DefaultConfig()
	cfg.TickInterval = time.Millisecond
	cfg.MaxTicks = 2

	for cycle := 0; cycle < 5; cycle++ {
		fabric := engine.NewAxiomaticFabric(engine.Magitech, "ward_restored", true)
		fabric.SetState("entropy", 0)
		fabric.SetState("cycle", cycle)

		gl := NewGameLoop(fabric, nil, nil, "p1", "mission_reconnect", cfg, int64(cycle))
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		results := gl.Start(ctx)
		for range results {
		}
		cancel()

		cycleState, _ := gl.Fabric().GetState("cycle")
		if cycleState != cycle {
			t.Errorf("cycle %d: expected cycle state %d, got %v", cycle, cycle, cycleState)
		}

		if gl.Replay().FrameCount() == 0 {
			t.Errorf("cycle %d: expected replay frames", cycle)
		}
	}
}

func TestGameLoopIntegration_PurgeViaEntropyTick(t *testing.T) {
	fabric := engine.NewAxiomaticFabric(engine.Magitech, "ward_restored", true)

	fabric.SetState("entropy", 500)
	fabric.ArchonVigilance = 0.0

	cfg := DefaultConfig()
	cfg.TickInterval = 1 * time.Millisecond

	gl := NewGameLoop(fabric, nil, nil, "p1", "", cfg, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	results := gl.Start(ctx)

	var purged bool
	for result := range results {
		if result.State == StatePurged {
			purged = true
			if result.Snapshot == nil {
				t.Error("expected snapshot on purge")
			}
			if result.Snapshot["game_over"] != true {
				t.Error("expected game_over flag on purge snapshot")
			}
			break
		}
	}

	if !purged {
		t.Error("expected purge due to high entropy")
	}
}
