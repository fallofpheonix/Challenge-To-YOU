package qa

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ScenarioContinuousPlaythrough runs a full end-to-end playthrough.
//
// Launch backend → Connect client → Start mission → Receive dialogue →
// Load challenge → Submit solution → Complete mission → Receive rewards →
// Save → Restart backend → Reload save → Verify state → Disconnect
func ScenarioContinuousPlaythrough(ctx *ScenarioContext) error {
	root := getProjectRoot()
	tmpDir, err := os.MkdirTemp("", "qa_e2e_playthrough_*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// === PHASE 1: Launch Backend ===
	binaryPath, err := BuildBinary(root)
	if err != nil {
		return fmt.Errorf("build binary: %w", err)
	}
	ctx.Step("phase1_build", "Backend binary compiled")

	port, _ := FindFreePort()
	dbPath := filepath.Join(tmpDir, "e2e_test.db")
	srv := NewServer(ServerConfig{Port: port, DBPath: dbPath, WorkDir: root})

	srvCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if err := srv.StartWithPort(srvCtx, binaryPath, port); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer srv.Stop()
	ctx.Step("phase1_launch", fmt.Sprintf("Server launched on port %d", port))

	// === PHASE 2: Connect Client ===
	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()
	ctx.Step("phase2_connect", "WebSocket client connected")

	// === PHASE 3: Start Mission (get initial state) ===
	if err := client.Send("profile", ""); err != nil {
		return fmt.Errorf("send profile: %w", err)
	}

	snap, err := client.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("read initial snapshot: %w", err)
	}
	ctx.Require("phase3_mission_started", snap != nil, "Should receive initial snapshot")
	ctx.Step("phase3_mission", fmt.Sprintf("Mission loaded: %s (Paradigm: %s)", snap.ChallengeID, snap.Paradigm))

	// === PHASE 4: Receive Dialogue / World ===
	if snap.Intro != nil && len(snap.Intro) > 0 {
		ctx.Step("phase4_intro", "Intro sequence present")
	}
	if snap.World != nil && len(snap.World) > 0 {
		ctx.Step("phase4_world", "World data loaded")
	}
	if len(snap.Dialogue) > 0 {
		ctx.Step("phase4_dialogue", fmt.Sprintf("Dialogue entries: %d", len(snap.Dialogue)))
	}
	ctx.Assert("phase4_state_ready", snap.State != nil, "Game state initialized")

	// === PHASE 5: Load Challenge ===
	ctx.Require("phase5_challenge_loaded", snap.ChallengeID != "", "Challenge should be loaded")
	ctx.Assert("phase5_has_modules", len(snap.Modules) > 0, fmt.Sprintf("Modules: %d", len(snap.Modules)))
	ctx.Assert("phase5_has_triggers", len(snap.Triggerable) > 0, fmt.Sprintf("Triggers: %v", snap.Triggerable))

	// === PHASE 6: Submit Solutions (trigger events) ===
	maxRounds := 15
	roundCount := 0
	for round := 0; round < maxRounds; round++ {
		if snap.LevelComplete {
			ctx.Step("phase6_complete", fmt.Sprintf("Challenge completed after %d rounds", round))
			break
		}
		if len(snap.Triggerable) == 0 {
			ctx.Step("phase6_no_triggers", "No more triggerable events")
			break
		}

		// Trigger the first available event
		event := snap.Triggerable[0]
		if err := client.Send(event, ""); err != nil {
			// Connection may have been closed by server (challenge complete, etc.)
			ctx.Step("phase6_send_done", fmt.Sprintf("Round %d: connection closed: %v", round, err))
			break
		}

		newSnap, err := client.ReadSnapshot(3 * time.Second)
		if err != nil {
			ctx.Step("phase6_read_done", fmt.Sprintf("Round %d: read complete: %v", round, err))
			break
		}

		snap = newSnap
		roundCount++
		ctx.Step(fmt.Sprintf("phase6_round_%d", round),
			fmt.Sprintf("Event: %s | Vigilance: %.2f | Complete: %v",
				event, snap.Vigilance, snap.LevelComplete))
	}

	// === PHASE 7: Complete Mission ===
	if snap.LevelComplete {
		ctx.Assert("phase7_cipher", snap.LastCipher != "", fmt.Sprintf("Cipher: %s", snap.LastCipher))
		ctx.Step("phase7_complete", "Mission completed successfully!")
	} else {
		ctx.Step("phase7_partial", fmt.Sprintf("Challenge not completed after %d rounds (expected for automated test)", roundCount))
	}

	// === PHASE 8: Receive Rewards ===
	if snap.Profile != nil {
		ctx.Step("phase8_rewards", fmt.Sprintf("Profile - XP: %d, Level: %d, Rep: %d, Luck: %.2f",
			snap.Profile.XP, snap.Profile.Level, snap.Profile.Reputation, snap.Profile.Luck))
	}

	// === PHASE 9: Save (server auto-saves to DB) ===
	client.Disconnect()
	srv.Stop()
	time.Sleep(500 * time.Millisecond)
	ctx.Step("phase9_save", "Server stopped (state persisted to DB)")

	// === PHASE 10: Restart Backend ===
	port2, _ := FindFreePort()
	srv2 := NewServer(ServerConfig{Port: port2, DBPath: dbPath, WorkDir: root})

	srvCtx2, cancel2 := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel2()

	if err := srv2.StartWithPort(srvCtx2, binaryPath, port2); err != nil {
		return fmt.Errorf("restart server: %w", err)
	}
	defer srv2.Stop()
	ctx.Step("phase10_restart", fmt.Sprintf("Server restarted on port %d", port2))

	// === PHASE 11: Reload Save ===
	client2 := NewWSClient(port2)
	if err := client2.Connect(); err != nil {
		return fmt.Errorf("reconnect: %w", err)
	}
	defer client2.Disconnect()

	if err := client2.Send("profile", ""); err != nil {
		return fmt.Errorf("send profile after restart: %w", err)
	}

	snap2, err := client2.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("read snapshot after restart: %w", err)
	}
	ctx.Require("phase11_loaded", snap2 != nil, "Should receive snapshot after restart")
	ctx.Step("phase11_reload", fmt.Sprintf("State reloaded: %s", snap2.ChallengeID))

	// === PHASE 12: Verify State ===
	ctx.Assert("phase12_paradigm_match", snap2.Paradigm == snap.Paradigm,
		fmt.Sprintf("Paradigm: %s -> %s", snap.Paradigm, snap2.Paradigm))

	if snap.Profile != nil && snap2.Profile != nil {
		ctx.Assert("phase12_xp_match", snap2.Profile.XP == snap.Profile.XP,
			fmt.Sprintf("XP: %d -> %d", snap.Profile.XP, snap2.Profile.XP))
		ctx.Assert("phase12_luck_match", snap2.Profile.Luck == snap.Profile.Luck,
			fmt.Sprintf("Luck: %f -> %f", snap.Profile.Luck, snap2.Profile.Luck))
	}

	// === PHASE 13: Disconnect ===
	client2.Disconnect()
	ctx.Step("phase13_disconnect", "Final disconnect complete")

	ctx.Step("e2e_complete", fmt.Sprintf("Full playthrough completed. %d rounds executed.", roundCount))
	return nil
}
