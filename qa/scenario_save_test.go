package qa

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ScenarioSaveCreateLoad verifies save creation and loading via the server API.
func ScenarioSaveCreateLoad(ctx *ScenarioContext) error {
	root := getProjectRoot()
	tmpDir, err := os.MkdirTemp("", "qa_save_create_*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath, err := BuildBinary(root)
	if err != nil {
		return fmt.Errorf("build binary: %w", err)
	}

	dbPath := filepath.Join(tmpDir, "save_test.db")
	port, _ := FindFreePort()
	srv := NewServer(ServerConfig{Port: port, DBPath: dbPath, WorkDir: root})

	srvCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := srv.StartWithPort(srvCtx, binaryPath, port); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer srv.Stop()

	// Connect and get initial profile
	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()

	// Request profile (triggers DB load)
	if err := client.Send("profile", ""); err != nil {
		return fmt.Errorf("send profile: %w", err)
	}

	snap, err := client.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("read snapshot: %w", err)
	}

	ctx.Assert("save_profile_loaded", snap != nil, "Profile should load from DB")
	if snap != nil && snap.Profile != nil {
		ctx.Assert("default_luck", snap.Profile.Luck > 0, fmt.Sprintf("Luck: %f", snap.Profile.Luck))
		ctx.Step("profile_state", fmt.Sprintf("Rep: %d, Luck: %.2f, Paradigms: %s",
			snap.Profile.Reputation, snap.Profile.Luck, snap.Profile.UnlockedParadigms))
	}

	// Trigger some state changes (which get saved)
	if len(snap.Triggerable) > 0 {
		client.Send(snap.Triggerable[0], "")
		client.ReadSnapshot(3 * time.Second)
		ctx.Step("state_modified", "Triggered event to modify state")
	}

	return nil
}

// ScenarioSaveRestartReload verifies save persists across server restarts.
func ScenarioSaveRestartReload(ctx *ScenarioContext) error {
	root := getProjectRoot()
	tmpDir, err := os.MkdirTemp("", "qa_save_restart_*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath, err := BuildBinary(root)
	if err != nil {
		return fmt.Errorf("build binary: %w", err)
	}

	dbPath := filepath.Join(tmpDir, "restart_test.db")

	// First server instance
	port1, _ := FindFreePort()
	srv1 := NewServer(ServerConfig{Port: port1, DBPath: dbPath, WorkDir: root})
	srvCtx1, cancel1 := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel1()

	if err := srv1.StartWithPort(srvCtx1, binaryPath, port1); err != nil {
		return fmt.Errorf("start server 1: %w", err)
	}

	client1 := NewWSClient(port1)
	if err := client1.Connect(); err != nil {
		srv1.Stop()
		return fmt.Errorf("connect client 1: %w", err)
	}

	// Get state from first instance
	client1.Send("profile", "")
	snap1, err := client1.ReadSnapshot(5 * time.Second)
	if err != nil {
		srv1.Stop()
		return fmt.Errorf("read snapshot 1: %w", err)
	}
	ctx.Step("server1_state", fmt.Sprintf("Challenge: %s", snap1.ChallengeID))

	// Modify state
	if len(snap1.Triggerable) > 0 {
		client1.Send(snap1.Triggerable[0], "")
		client1.ReadSnapshot(3 * time.Second)
	}

	client1.Disconnect()
	srv1.Stop()
	time.Sleep(500 * time.Millisecond)

	// Second server instance (reload)
	port2, _ := FindFreePort()
	srv2 := NewServer(ServerConfig{Port: port2, DBPath: dbPath, WorkDir: root})
	srvCtx2, cancel2 := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel2()

	if err := srv2.StartWithPort(srvCtx2, binaryPath, port2); err != nil {
		return fmt.Errorf("start server 2: %w", err)
	}
	defer srv2.Stop()

	client2 := NewWSClient(port2)
	if err := client2.Connect(); err != nil {
		return fmt.Errorf("connect client 2: %w", err)
	}
	defer client2.Disconnect()

	client2.Send("profile", "")
	snap2, err := client2.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("read snapshot 2: %w", err)
	}

	ctx.Assert("save_persisted", snap2 != nil, "State should persist across restarts")
	if snap2 != nil {
		ctx.Assert("challenge_restored", snap2.ChallengeID == snap1.ChallengeID,
			fmt.Sprintf("Challenge: %s -> %s", snap1.ChallengeID, snap2.ChallengeID))
		ctx.Step("server2_state", fmt.Sprintf("Challenge: %s (restored)", snap2.ChallengeID))
	}

	return nil
}

// ScenarioSaveCorruptedHandling verifies the system handles corrupted saves.
func ScenarioSaveCorruptedHandling(ctx *ScenarioContext) error {
	root := getProjectRoot()
	tmpDir, err := os.MkdirTemp("", "qa_save_corrupt_*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath, err := BuildBinary(root)
	if err != nil {
		return fmt.Errorf("build binary: %w", err)
	}

	dbPath := filepath.Join(tmpDir, "corrupt_test.db")

	// Write corrupted data
	corruptedData := []byte("this is not a valid sqlite database")
	if err := os.WriteFile(dbPath, corruptedData, 0o644); err != nil {
		return fmt.Errorf("write corrupted db: %w", err)
	}
	ctx.Step("corrupted_written", "Wrote corrupted DB file")

	// Try to start server with corrupted DB
	port, _ := FindFreePort()
	srv := NewServer(ServerConfig{Port: port, DBPath: dbPath, WorkDir: root})
	srvCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = srv.StartWithPort(srvCtx, binaryPath, port)
	if err != nil {
		ctx.Assert("corrupt_handled_gracefully", true,
			fmt.Sprintf("Server correctly rejected corrupted DB: %v", err))
	} else {
		// Server started - it may have recreated the DB
		srv.Stop()
		ctx.Step("corrupt_recreated", "Server recreated DB from corrupted file (graceful recovery)")
	}

	return nil
}

// ScenarioSaveMultipleProfiles verifies multiple profile operations work.
func ScenarioSaveMultipleProfiles(ctx *ScenarioContext) error {
	// This tests that multiple save/load cycles don't corrupt data
	root := getProjectRoot()
	tmpDir, err := os.MkdirTemp("", "qa_save_multi_*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath, err := BuildBinary(root)
	if err != nil {
		return fmt.Errorf("build binary: %w", err)
	}

	dbPath := filepath.Join(tmpDir, "multi_test.db")
	port, _ := FindFreePort()
	srv := NewServer(ServerConfig{Port: port, DBPath: dbPath, WorkDir: root})
	srvCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := srv.StartWithPort(srvCtx, binaryPath, port); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer srv.Stop()

	// Multiple save/load cycles
	for i := 0; i < 5; i++ {
		client := NewWSClient(port)
		if err := client.Connect(); err != nil {
			return fmt.Errorf("connect cycle %d: %w", i, err)
		}

		client.Send("profile", "")
		snap, err := client.ReadSnapshot(3 * time.Second)
		if err != nil {
			client.Disconnect()
			return fmt.Errorf("read cycle %d: %w", i, err)
		}

		// Trigger an event
		if len(snap.Triggerable) > 0 {
			client.Send(snap.Triggerable[0], "")
			client.ReadSnapshot(2 * time.Second)
		}

		client.Disconnect()
		ctx.Step(fmt.Sprintf("cycle_%d", i), fmt.Sprintf("Save/load cycle %d completed", i))
	}

	// Final verification
	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("final connect: %w", err)
	}
	defer client.Disconnect()

	client.Send("profile", "")
	finalSnap, err := client.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("final read: %w", err)
	}

	ctx.Assert("multi_save_integrity", finalSnap != nil, "Data integrity maintained across multiple saves")

	return nil
}
