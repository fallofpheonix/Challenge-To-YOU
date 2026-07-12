package missionengine

import (
	"testing"
	"time"

	"challenge-to-you/backend/internal/eventbus"
)

func createTestMission() *Mission {
	return &Mission{
		ID:          "test_mission_1",
		Title:       "Test Mission",
		Description: "A test mission",
		Era:         "magitech",
		Act:         1,
		Chapter:     1,
		Paradigm:    "MAGITECH",
		Difficulty:  0.3,
		EstTimeMins: 10,
		Objectives: []Objective{
			{
				ID:          "obj_1",
				Type:        "complete_challenge",
				Target:      "challenge_01",
				Description: "Complete the challenge",
				IsRequired:  true,
			},
			{
				ID:          "obj_2",
				Type:        "find_item",
				Target:      "item_01",
				Description: "Find the item",
				IsRequired:  true,
			},
		},
		Reward: &MissionReward{
			XP:           100,
			Credits:      50,
			Unlocks:      []string{"mission_2"},
			Achievements: []string{"test_achievement"},
		},
	}
}

func createOptionalObjMission() *Mission {
	return &Mission{
		ID:          "test_mission_optional",
		Title:       "Optional Mission",
		Description: "Has optional objectives",
		Era:         "magitech",
		Act:         1,
		Chapter:     1,
		Difficulty:  0.3,
		EstTimeMins: 10,
		Objectives: []Objective{
			{
				ID:          "obj_1",
				Type:        "complete_challenge",
				Target:      "challenge_01",
				Description: "Complete the challenge",
				IsRequired:  true,
			},
			{
				ID:          "obj_2",
				Type:        "find_item",
				Target:      "item_01",
				Description: "Find the hidden item",
				IsOptional:  true,
			},
		},
		Reward: &MissionReward{
			XP:      100,
			Credits: 50,
		},
	}
}

func TestLoader(t *testing.T) {
	registry := NewMissionRegistry()
	err := registry.LoadMissionsFromDir("../../../data/missions")
	if err != nil {
		t.Fatalf("Failed to load missions: %v", err)
	}

	if registry.Count() == 0 {
		t.Fatal("No missions loaded")
	}

	t.Logf("Loaded %d missions", registry.Count())
	for _, m := range registry.GetAll() {
		t.Logf("  - %s: %s (%s)", m.ID, m.Title, m.Era)
	}
}

func TestRegistry(t *testing.T) {
	registry := NewMissionRegistry()
	mission := createTestMission()
	registry.Register(mission)

	got, ok := registry.Get("test_mission_1")
	if !ok {
		t.Fatal("Failed to get mission")
	}
	if got.ID != "test_mission_1" {
		t.Errorf("Expected test_mission_1, got %s", got.ID)
	}

	byEra := registry.GetByEra("magitech")
	if len(byEra) != 1 {
		t.Errorf("Expected 1 mission for magitech, got %d", len(byEra))
	}

	all := registry.GetAll()
	if len(all) != 1 {
		t.Errorf("Expected 1 mission total, got %d", len(all))
	}
}

func TestStateMachine(t *testing.T) {
	sm := NewStateMachine()
	mission := createTestMission()
	session := &MissionSession{
		Mission:           mission,
		PlayerID:          "player_1",
		Status:            MissionStatusAvailable,
		CurrentStep:       0,
		ObjectiveStatuses: make(map[string]ObjectiveStatus),
		FabricState:       make(map[string]interface{}),
		StartTime:         time.Time{},
	}

	if err := sm.StartMission(session); err != nil {
		t.Fatalf("Failed to start mission: %v", err)
	}
	if session.Status != MissionStatusActive {
		t.Fatalf("Expected Active status, got %v", session.Status)
	}

	if err := sm.CompleteObjective(session, "obj_1"); err != nil {
		t.Fatalf("Failed to complete objective obj_1: %v", err)
	}
	if session.ObjectiveStatuses["obj_1"] != ObjectiveStatusCompleted {
		t.Error("Objective 1 should be completed")
	}
	if session.Status != MissionStatusActive {
		t.Fatalf("Mission should still be active after first objective, got %v", session.Status)
	}

	if err := sm.CompleteObjective(session, "obj_2"); err != nil {
		t.Fatalf("Failed to complete objective obj_2: %v", err)
	}
	if session.Status != MissionStatusCompleted {
		t.Errorf("Expected Completed status, got %v", session.Status)
	}
	if session.EndTime == nil {
		t.Error("End time should be set on completion")
	}
}

func TestStateMachine_OnlyRequiredCompletesMission(t *testing.T) {
	sm := NewStateMachine()
	mission := createOptionalObjMission()
	session := &MissionSession{
		Mission:           mission,
		PlayerID:          "player_1",
		Status:            MissionStatusAvailable,
		ObjectiveStatuses: make(map[string]ObjectiveStatus),
		FabricState:       make(map[string]interface{}),
	}

	sm.StartMission(session)

	// Complete only the required objective - mission should complete
	if err := sm.CompleteObjective(session, "obj_1"); err != nil {
		t.Fatalf("Failed to complete objective: %v", err)
	}
	if session.Status != MissionStatusCompleted {
		t.Errorf("Mission should complete with only required objectives done, got %v", session.Status)
	}
}

func TestStateMachine_Fail(t *testing.T) {
	sm := NewStateMachine()
	mission := createTestMission()
	session := &MissionSession{
		Mission:           mission,
		PlayerID:          "player_1",
		Status:            MissionStatusAvailable,
		ObjectiveStatuses: make(map[string]ObjectiveStatus),
		FabricState:       make(map[string]interface{}),
	}

	sm.StartMission(session)

	if err := sm.FailMission(session, "test failure"); err != nil {
		t.Fatalf("Failed to fail mission: %v", err)
	}
	if session.Status != MissionStatusFailed {
		t.Errorf("Expected Failed status, got %v", session.Status)
	}
	if session.FabricState["failure_reason"] != "test failure" {
		t.Error("Fail reason not set in fabric state")
	}
}

func TestStateMachine_FailRequiredObjective(t *testing.T) {
	sm := NewStateMachine()
	mission := createTestMission()
	session := &MissionSession{
		Mission:           mission,
		PlayerID:          "player_1",
		Status:            MissionStatusAvailable,
		ObjectiveStatuses: make(map[string]ObjectiveStatus),
		FabricState:       make(map[string]interface{}),
	}

	sm.StartMission(session)

	// Failing a required objective should fail the whole mission
	if err := sm.FailObjective(session, "obj_1"); err != nil {
		t.Fatalf("Failed to fail objective: %v", err)
	}
	if session.Status != MissionStatusFailed {
		t.Errorf("Expected Failed mission after failing required objective, got %v", session.Status)
	}
}

func TestObjectiveEngine(t *testing.T) {
	oe := NewObjectiveEngine()
	mission := createTestMission()
	session := &MissionSession{
		Mission:           mission,
		PlayerID:          "player_1",
		Status:            MissionStatusActive,
		ObjectiveStatuses: make(map[string]ObjectiveStatus),
		FabricState:       make(map[string]interface{}),
	}

	event := ObjectiveEvent{
		Type:   "complete_challenge",
		Target: "challenge_01",
		Player: "player_1",
	}

	completed := oe.ProcessEvent(session, event)
	if len(completed) != 1 {
		t.Fatalf("Expected 1 completed objective, got %d", len(completed))
	}
	if completed[0] != "obj_1" {
		t.Errorf("Expected obj_1, got %s", completed[0])
	}

	event2 := ObjectiveEvent{
		Type:   "find_item",
		Target: "item_01",
		Player: "player_1",
	}

	completed2 := oe.ProcessEvent(session, event2)
	if len(completed2) != 1 {
		t.Fatalf("Expected 1 completed objective, got %d", len(completed2))
	}
	if completed2[0] != "obj_2" {
		t.Errorf("Expected obj_2, got %s", completed2[0])
	}
}

func TestObjectiveEngine_IgnoresCompleted(t *testing.T) {
	oe := NewObjectiveEngine()
	mission := createTestMission()
	session := &MissionSession{
		Mission:  mission,
		PlayerID: "player_1",
		Status:   MissionStatusActive,
		ObjectiveStatuses: map[string]ObjectiveStatus{
			"obj_1": ObjectiveStatusCompleted,
		},
		FabricState: make(map[string]interface{}),
	}

	event := ObjectiveEvent{
		Type:   "complete_challenge",
		Target: "challenge_01",
		Player: "player_1",
	}

	completed := oe.ProcessEvent(session, event)
	if len(completed) != 0 {
		t.Errorf("Should not re-complete already completed objective, got %d", len(completed))
	}
}

func TestManager_FullFlow(t *testing.T) {
	bus := eventbus.NewEventBus(100)
	registry := NewMissionRegistry()
	mission := createTestMission()
	registry.Register(mission)

	manager := NewMissionManager(registry, bus, newTestDB(t))

	available := manager.GetAvailableMissions("player_1")
	if len(available) != 1 {
		t.Fatalf("Expected 1 available mission, got %d", len(available))
	}

	session, err := manager.StartMission("player_1", "test_mission_1")
	if err != nil {
		t.Fatalf("Failed to start mission: %v", err)
	}
	if session.Status != MissionStatusActive {
		t.Errorf("Expected Active status, got %v", session.Status)
	}

	available = manager.GetAvailableMissions("player_1")
	if len(available) != 0 {
		t.Errorf("Expected 0 available missions, got %d", len(available))
	}

	err = manager.CompleteChallenge("player_1", "test_mission_1", "challenge_01")
	if err != nil {
		t.Fatalf("Failed to complete challenge: %v", err)
	}

	manager.ProcessEvent("player_1", ObjectiveEvent{
		Type:   "find_item",
		Target: "item_01",
		Player: "player_1",
	})

	if !manager.IsMissionCompleted("player_1", "test_mission_1") {
		t.Error("Mission should be completed")
	}

	_, err = manager.StartMission("player_1", "test_mission_1")
	if err == nil {
		t.Error("Should not be able to start completed mission")
	}
}

func TestManager_CancelMission(t *testing.T) {
	bus := eventbus.NewEventBus(100)
	registry := NewMissionRegistry()
	mission := createTestMission()
	registry.Register(mission)

	manager := NewMissionManager(registry, bus, newTestDB(t))

	_, err := manager.StartMission("player_1", "test_mission_1")
	if err != nil {
		t.Fatalf("Failed to start mission: %v", err)
	}

	err = manager.CancelMission("player_1", "test_mission_1")
	if err != nil {
		t.Fatalf("Failed to cancel mission: %v", err)
	}

	available := manager.GetAvailableMissions("player_1")
	if len(available) != 1 {
		t.Errorf("Expected 1 available mission after cancel, got %d", len(available))
	}
}

func TestManager_GetProgress(t *testing.T) {
	bus := eventbus.NewEventBus(100)
	registry := NewMissionRegistry()
	mission := createTestMission()
	registry.Register(mission)

	manager := NewMissionManager(registry, bus, newTestDB(t))

	_, err := manager.StartMission("player_1", "test_mission_1")
	if err != nil {
		t.Fatalf("Failed to start mission: %v", err)
	}

	progress := manager.GetMissionProgress("player_1", "test_mission_1")
	if progress == nil {
		t.Fatal("Progress should not be nil")
	}

	if progress["status"] != "active" {
		t.Errorf("Expected status 'active', got %v", progress["status"])
	}

	objectManager_Requirements(t, registry, manager)
}

func objectManager_Requirements(t *testing.T, registry *MissionRegistry, manager *MissionManager) {
	t.Helper()

	// Create a mission with requirements
	dep := &Mission{
		ID:          "dep_mission",
		Title:       "Dependency",
		Era:         "magitech",
		Act:         1,
		Chapter:     1,
		Difficulty:  0.1,
		EstTimeMins: 5,
		Objectives: []Objective{
			{ID: "dep_obj", Type: "complete_challenge", Target: "dep_challenge", IsRequired: true},
		},
	}
	chained := &Mission{
		ID:           "chained_mission",
		Title:        "Chained",
		Era:          "magitech",
		Act:          1,
		Chapter:      2,
		Difficulty:   0.2,
		EstTimeMins:  5,
		Requirements: []string{"dep_mission"},
		Objectives: []Objective{
			{ID: "chain_obj", Type: "complete_challenge", Target: "chain_challenge", IsRequired: true},
		},
	}
	registry.Register(dep)
	registry.Register(chained)

	// chained_mission should not be available until dep_mission is completed
	available := manager.GetAvailableMissions("player_2")
	hasChained := false
	for _, m := range available {
		if m.ID == "chained_mission" {
			hasChained = true
		}
	}
	if hasChained {
		t.Error("chained_mission should not be available before dep_mission is completed")
	}

	// Complete dep_mission
	_, err := manager.StartMission("player_2", "dep_mission")
	if err != nil {
		t.Fatalf("Failed to start dep_mission: %v", err)
	}
	err = manager.CompleteChallenge("player_2", "dep_mission", "dep_challenge")
	if err != nil {
		t.Fatalf("Failed to complete dep_mission challenge: %v", err)
	}

	// Now chained_mission should be available
	available = manager.GetAvailableMissions("player_2")
	hasChained = false
	for _, m := range available {
		if m.ID == "chained_mission" {
			hasChained = true
		}
	}
	if !hasChained {
		t.Error("chained_mission should be available after dep_mission is completed")
	}
}
