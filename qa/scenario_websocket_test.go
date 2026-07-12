package qa

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ScenarioWSConnect verifies WebSocket connection to the server.
func ScenarioWSConnect(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	defer srv.Stop()

	client := NewWSClient(port)
	err := client.Connect()
	ctx.Require("ws_connect", err == nil, fmt.Sprintf("Connect error: %v", err))
	defer client.Disconnect()

	// Give server time to initialize
	time.Sleep(500 * time.Millisecond)

	ctx.Assert("ws_connected", client.IsConnected(), "Client should be connected")
	return nil
}

// ScenarioWSDisconnect verifies clean WebSocket disconnection.
func ScenarioWSDisconnect(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	defer srv.Stop()

	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	ctx.Assert("ws_connected_before", client.IsConnected(), "Should be connected")

	if err := client.Disconnect(); err != nil {
		return fmt.Errorf("disconnect: %w", err)
	}

	ctx.Assert("ws_disconnected", !client.IsConnected(), "Should be disconnected after Disconnect()")
	return nil
}

// ScenarioWSReconnect verifies a client can disconnect and reconnect.
func ScenarioWSReconnect(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	if srv == nil {
		return ctx.Error
	}
	defer srv.Stop()

	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("first connect: %w", err)
	}
	ctx.Assert("first_connect", client.IsConnected(), "Should be connected")

	if err := client.Disconnect(); err != nil {
		return fmt.Errorf("disconnect: %w", err)
	}
	time.Sleep(200 * time.Millisecond)

	if err := client.Connect(); err != nil {
		return fmt.Errorf("reconnect: %w", err)
	}
	ctx.Assert("reconnect", client.IsConnected(), "Should be reconnected")

	client.Disconnect()
	return nil
}

// ScenarioWSRoomStateResilience verifies that room navigation and inventory state
// is preserved across WebSocket reconnections (global shared fabric).
func ScenarioWSRoomStateResilience(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	if srv == nil {
		return ctx.Error
	}
	defer srv.Stop()

	// === Connection A: navigate and pick up item ===
	clientA := NewWSClient(port)
	if err := clientA.Connect(); err != nil {
		return fmt.Errorf("connect A: %w", err)
	}

	// Consume initial snapshot
	if _, err := clientA.ReadSnapshot(5 * time.Second); err != nil {
		return fmt.Errorf("initial snapshot A: %w", err)
	}
	ctx.Step("conn_a_connected", "Connection A: received initial snapshot")

	// Move to Forbidden Library
	if err := clientA.Send("move_room:Forbidden Library", ""); err != nil {
		return fmt.Errorf("move_room A: %w", err)
	}
	moveSnap, err := clientA.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("move snapshot A: %w", err)
	}
	ctx.Assert("conn_a_moved", moveSnap.State["current_room"] == "Forbidden Library",
		fmt.Sprintf("Room after move: %v", moveSnap.State["current_room"]))

	// Pick up Rune Shard
	if err := clientA.Send("inspect_object:scroll_rack", ""); err != nil {
		return fmt.Errorf("inspect A: %w", err)
	}
	pickupSnap, err := clientA.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("pickup snapshot A: %w", err)
	}
	ctx.Assert("conn_a_has_shard", pickupSnap.State["has_rune_shard"] == true,
		fmt.Sprintf("has_rune_shard after inspect: %v", pickupSnap.State["has_rune_shard"]))
	ctx.Step("conn_a_complete", "Connection A: navigated to Library, picked up Rune Shard")

	// Disconnect Connection A
	clientA.Disconnect()
	time.Sleep(100 * time.Millisecond)

	// === Connection B: verify restored state ===
	clientB := NewWSClient(port)
	if err := clientB.Connect(); err != nil {
		return fmt.Errorf("connect B: %w", err)
	}
	defer clientB.Disconnect()

	reconnSnap, err := clientB.ReadSnapshot(5 * time.Second)
	if err != nil {
		return fmt.Errorf("reconnect snapshot B: %w", err)
	}

	ctx.Assert("reconnect_room_restored",
		reconnSnap.State["current_room"] == "Forbidden Library",
		fmt.Sprintf("Expected 'Forbidden Library', got: %v", reconnSnap.State["current_room"]))

	ctx.Assert("reconnect_shard_restored",
		reconnSnap.State["has_rune_shard"] == true,
		fmt.Sprintf("Expected has_rune_shard=true, got: %v", reconnSnap.State["has_rune_shard"]))

	ctx.Step("resilience_passed", "Connection B: correctly restored room and inventory state")
	return nil
}

// ScenarioWSInvalidMessage verifies server handles invalid messages gracefully.
func ScenarioWSInvalidMessage(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	defer srv.Stop()

	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()

	// Send invalid JSON
	err := client.SendRaw([]byte("not valid json {{{"))
	if err != nil {
		return fmt.Errorf("send invalid: %w", err)
	}
	ctx.Step("invalid_sent", "Sent invalid JSON")

	// Server should still be alive - try sending a valid message
	time.Sleep(500 * time.Millisecond)
	err = client.Send("profile", "")
	ctx.Assert("server_alive_after_invalid", err == nil, "Server should handle invalid messages and stay alive")

	return nil
}

// ScenarioWSTimeout verifies the server handles idle connections.
func ScenarioWSTimeout(ctx *ScenarioContext) error {
	port, srv := startTestServer(ctx)
	defer srv.Stop()

	client := NewWSClient(port)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()

	// Send a valid message to get initial state
	err := client.Send("profile", "")
	if err != nil {
		return fmt.Errorf("send profile: %w", err)
	}
	ctx.Step("initial_message", "Sent profile request")

	// Wait and verify connection is still alive
	time.Sleep(1 * time.Second)
	ctx.Assert("connection_alive", client.IsConnected(), "Connection should remain alive after idle period")

	return nil
}

// startTestServer is a helper that builds and starts a server, returning port and server.
func startTestServer(ctx *ScenarioContext) (int, *Server) {
	root := getProjectRoot()
	tmpDir, _ := os.MkdirTemp("", "qa_ws_*")

	binaryPath, err := BuildBinary(root)
	if err != nil {
		ctx.Error = fmt.Errorf("build binary: %w", err)
		return 0, nil
	}
	_ = binaryPath

	port, _ := FindFreePort()
	srv := NewServer(ServerConfig{
		Port:    port,
		DBPath:  filepath.Join(tmpDir, "test.db"),
		WorkDir: root,
	})

	srvCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	_ = cancel

	if err := srv.StartWithPort(srvCtx, binaryPath, port); err != nil {
		ctx.Error = fmt.Errorf("start server: %w", err)
		return 0, nil
	}

	return port, srv
}


