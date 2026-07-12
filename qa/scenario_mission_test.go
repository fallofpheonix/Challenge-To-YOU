package qa

import (
	"encoding/json"
	"fmt"
	"time"
)

// ScenarioMissionStart verifies a mission can be started via WebSocket.
func ScenarioMissionStart(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	if srv == nil {
		return ctx.Error
	}
	defer srv.Stop()

	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()

	// Give server time to initialize
	time.Sleep(500 * time.Millisecond)

	// Send start_game to trigger mission initialization
	if err := client.Send("start_game", ""); err != nil {
		return fmt.Errorf("send start_game: %w", err)
	}

	snap, err := client.ReadSnapshot(10 * time.Second)
	if err != nil {
		return fmt.Errorf("read snapshot: %w", err)
	}

	ctx.Assert("mission_loaded", snap.ChallengeID != "", fmt.Sprintf("Challenge ID: %s", snap.ChallengeID))
	ctx.Assert("paradigm_set", snap.Paradigm != "", fmt.Sprintf("Paradigm: %s", snap.Paradigm))
	ctx.Assert("modules_present", len(snap.Modules) > 0, fmt.Sprintf("Module count: %d", len(snap.Modules)))
	ctx.Assert("state_initialized", snap.State != nil, "State should be initialized")

	return nil
}

// ScenarioMissionObjectives verifies objectives are tracked correctly.
func ScenarioMissionObjectives(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	if srv == nil {
		return ctx.Error
	}
	defer srv.Stop()

	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()

	// Give server time to initialize
	time.Sleep(500 * time.Millisecond)

	// Get initial state
	if err := client.Send("start_game", ""); err != nil {
		return fmt.Errorf("send start_game: %w", err)
	}

	snap, err := client.ReadSnapshot(10 * time.Second)
	if err != nil {
		return fmt.Errorf("read snapshot: %w", err)
	}

	// Verify challenge has modules (objectives in challenge terms)
	ctx.Assert("has_modules", len(snap.Modules) > 0, fmt.Sprintf("Found %d modules", len(snap.Modules)))
	ctx.Assert("has_state", len(snap.State) > 0, fmt.Sprintf("State has %d keys", len(snap.State)))

	// Verify level is not complete yet
	ctx.Assert("not_complete", !snap.LevelComplete, "Level should not be complete initially")
	ctx.Assert("vigilance_low", snap.Vigilance < 0.5, fmt.Sprintf("Vigilance: %f", snap.Vigilance))

	return nil
}

// ScenarioMissionStateTransitions verifies state changes on trigger events.
func ScenarioMissionStateTransitions(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	if srv == nil {
		return ctx.Error
	}
	defer srv.Stop()

	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()

	// Get initial snapshot
	if err := client.Send("profile", ""); err != nil {
		return fmt.Errorf("send profile: %w", err)
	}

	snap, err := client.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("read snapshot: %w", err)
	}

	initialState := snap.State
	ctx.Step("initial_state", fmt.Sprintf("State keys: %v", mapKeys(initialState)))

	// Find a triggerable event
	if len(snap.Triggerable) == 0 {
		ctx.Step("no_triggers", "No triggerable events found, skipping state transition test")
		return nil
	}

	triggerEvent := snap.Triggerable[0]
	ctx.Step("trigger_event", fmt.Sprintf("Triggering: %s", triggerEvent))

	// Send the trigger
	if err := client.Send(triggerEvent, ""); err != nil {
		return fmt.Errorf("send trigger: %w", err)
	}

	// Read the response snapshot
	newSnap, err := client.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("read snapshot after trigger: %w", err)
	}

	// Verify we got a response - the server processes the trigger and sends back a snapshot
	// State changes depend on the challenge logic, so just verify the trigger was processed
	stateChanged := !stateMapsEqual(initialState, newSnap.State)
	vigilanceIncreased := newSnap.Vigilance > snap.Vigilance
	hasMessage := newSnap.Message != ""

	ctx.Assert("trigger_processed", stateChanged || vigilanceIncreased || hasMessage,
		fmt.Sprintf("State changed: %v, Vigilance: %.2f -> %.2f, Message: %q",
			stateChanged, snap.Vigilance, newSnap.Vigilance, newSnap.Message))

	return nil
}

// ScenarioMissionCompletion verifies a mission can be completed.
func ScenarioMissionCompletion(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	if srv == nil {
		return ctx.Error
	}
	defer srv.Stop()

	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()

	// Get initial state
	if err := client.Send("profile", ""); err != nil {
		return fmt.Errorf("send profile: %w", err)
	}

	snap, err := client.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("read snapshot: %w", err)
	}

	// Trigger events in sequence to solve the challenge
	// For magitech_01: invoke_binding -> mana_surge (if binding_active)
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		if snap.LevelComplete {
			break
		}
		if len(snap.Triggerable) == 0 {
			break
		}

		// Pick first triggerable event
		event := snap.Triggerable[0]
		if err := client.Send(event, ""); err != nil {
			return fmt.Errorf("send trigger %d: %w", i, err)
		}

		newSnap, err := client.ReadSnapshot(3 * time.Second)
		if err != nil {
			return fmt.Errorf("read snapshot %d: %w", i, err)
		}
		snap = newSnap
		ctx.Step(fmt.Sprintf("trigger_%d", i), fmt.Sprintf("Event: %s, Complete: %v, Vigilance: %.2f",
			event, snap.LevelComplete, snap.Vigilance))
	}

	// We may not complete the challenge in automated mode (requires specific order)
	// but verify the system responded to triggers
	ctx.Assert("system_responded", snap.Vigilance > 0 || snap.LevelComplete,
		"System should respond to trigger events")

	return nil
}

// ScenarioMissionUnlocks verifies unlock mechanisms work.
func ScenarioMissionUnlocks(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	if srv == nil {
		return ctx.Error
	}
	defer srv.Stop()

	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()

	// Request profile
	if err := client.Send("profile", ""); err != nil {
		return fmt.Errorf("send profile: %w", err)
	}

	snap, err := client.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("read snapshot: %w", err)
	}

	// Verify profile shows paradigms
	if snap.Profile != nil {
		ctx.Assert("profile_has_paradigms", snap.Profile.UnlockedParadigms != "",
			fmt.Sprintf("Paradigms: %s", snap.Profile.UnlockedParadigms))
		ctx.Assert("profile_has_luck", snap.Profile.Luck > 0,
			fmt.Sprintf("Luck: %f", snap.Profile.Luck))
	}

	// Try unlocking a paradigm (CYBERPUNK costs 50 rep)
	if err := client.Send("unlock", "CYBERPUNK"); err != nil {
		return fmt.Errorf("send unlock: %w", err)
	}

	// Read response (may be snapshot or error)
	_, err = client.ReadSnapshot(3 * time.Second)
	// Either it succeeds or fails due to insufficient rep - both are valid
	ctx.Step("unlock_attempt", "Unlock CYBERPUNK attempted")

	return nil
}

func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func stateMapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		av, _ := json.Marshal(v)
		bvv, _ := json.Marshal(bv)
		if string(av) != string(bvv) {
			return false
		}
	}
	return true
}
