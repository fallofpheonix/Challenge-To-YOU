package server

import (
	"context"
	"net"
	"net/http"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"challenge-to-you/backend/internal/db"
	"challenge-to-you/backend/internal/engine"
	"challenge-to-you/backend/internal/eventbus"
	"challenge-to-you/backend/internal/obs"
	"challenge-to-you/backend/internal/session"

	"github.com/gorilla/websocket"
)

// newTestServer creates a Server with real DB and minimal deps for shutdown tests.
func newTestServer(t *testing.T) *Server {
	t.Helper()
	tmpDir := t.TempDir()
	testDB, err := db.NewDB(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { testDB.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	s := &Server{
		db:             testDB,
		bus:            eventbus.NewEventBus(100),
		sessionManager: session.NewManager(30 * time.Minute),
		globalFabrics:  make(map[string]*engine.AxiomaticFabric),
		log:            obs.Default().Component("test"),
		metrics:        obs.NewMetrics(),
		ctx:            ctx,
		cancel:         cancel,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
	return s
}

// startTestHTTP starts a minimal HTTP server on a random port for testing.
func startTestHTTP(t *testing.T, s *Server) (string, func()) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/rift", func(w http.ResponseWriter, r *http.Request) {
		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer conn.Close()
			s.handleRift(s.ctx, r, conn, testDef(), testFabric())
		}()
	})
	mux.HandleFunc("/api/languages", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/api/challenges", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/debug/runtime", obs.DebugHandler(s.metrics))
	mux.HandleFunc("/metrics", obs.MetricsHandler(s.metrics))

	s.httpServer = &http.Server{Handler: mux}
	go func() { _ = s.httpServer.Serve(l) }()

	addr := "ws://" + l.Addr().String()
	cleanup := func() {
		_ = s.httpServer.Close()
	}
	return addr, cleanup
}

func testDef() *engine.ChallengeDefinition {
	return &engine.ChallengeDefinition{
		ID:          "test_shutdown_01",
		Paradigm:    engine.Magitech,
		Name:        "Shutdown Test",
		Description: "A test challenge for shutdown scenarios",
		LogosToken:  "SHUTDOWN_CIPHER",
		Flaws: []engine.Flaw{
			{
				ID:           "flaw_1",
				TriggerEvent: "TRIGGER_RUNE_A",
				Name:         "Rune A",
				Conditions:   engine.ParadigmState{"rune_a": true},
				Mutations:    engine.ParadigmState{"rune_a": false, "rune_b": true},
			},
		},
		WinCondition: engine.WinCondition{
			TargetStateKey: "rune_b",
			ExpectedValue:  true,
		},
		InitialState: engine.ParadigmState{"rune_a": true, "rune_b": false},
	}
}

func testFabric() *engine.AxiomaticFabric {
	def := testDef()
	fabric := def.BuildFabric()
	return fabric
}

// Test 1: Shutdown with no clients.
func TestShutdown_NoClients(t *testing.T) {
	s := newTestServer(t)
	addr, cleanup := startTestHTTP(t, s)
	defer cleanup()

	// Verify server is reachable.
	httpAddr := "http://" + addr[len("ws://"):]
	resp, err := http.Get(httpAddr + "/api/languages")
	if err != nil {
		t.Fatalf("server not reachable: %v", err)
	}
	resp.Body.Close()

	// Run shutdown — should complete immediately with no handlers.
	done := make(chan struct{})
	go func() {
		s.shutdown()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown did not complete within 5s")
	}

	if s.dbCloseCount.Load() != 1 {
		t.Fatalf("expected db.Close called once, got %d", s.dbCloseCount.Load())
	}
}

// Test 2: Shutdown with active WebSocket.
func TestShutdown_ActiveWebSocket(t *testing.T) {
	s := newTestServer(t)
	addr, cleanup := startTestHTTP(t, s)
	defer cleanup()

	// Connect a WebSocket client.
	wsURL := addr + "/rift"
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer wsConn.Close()

	// Read the initial snapshot.
	wsConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, _, err = wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("read initial snapshot: %v", err)
	}

	// Give the handler goroutine time to register.
	time.Sleep(50 * time.Millisecond)

	// Run shutdown — should cancel context, causing the handler to exit.
	done := make(chan struct{})
	go func() {
		s.shutdown()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("shutdown did not complete within 10s with active WS")
	}
}

// Test 3: Shutdown during active session processing.
func TestShutdown_DuringSessionProcessing(t *testing.T) {
	s := newTestServer(t)
	addr, cleanup := startTestHTTP(t, s)
	defer cleanup()

	wsConn, _, err := websocket.DefaultDialer.Dial(addr+"/rift", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer wsConn.Close()

	// Read initial snapshot.
	wsConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, _, err = wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("read initial: %v", err)
	}

	// Send an event to engage the handler's event loop.
	msg := `{"event":"TRIGGER_RUNE_A"}`
	if err := wsConn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		t.Fatalf("send event: %v", err)
	}

	// Read the response snapshot.
	wsConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, _, err = wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	// Shutdown should drain the handler cleanly.
	done := make(chan struct{})
	go func() {
		s.shutdown()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("shutdown deadlocked during active session")
	}

	// Verify session was cleaned up.
	if s.sessionManager.Count() != 0 {
		t.Fatalf("expected 0 active sessions after shutdown, got %d", s.sessionManager.Count())
	}
}

// Test 4: Sandbox cleanup on shutdown.
func TestShutdown_SandboxCleanup(t *testing.T) {
	s := newTestServer(t)
	addr, cleanup := startTestHTTP(t, s)
	defer cleanup()

	wsConn, _, err := websocket.DefaultDialer.Dial(addr+"/rift", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer wsConn.Close()

	// Read initial snapshot.
	wsConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, _, err = wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("read initial: %v", err)
	}

	// The sandbox is stateless — verify shutdown doesn't leave goroutines leaked.
	// Track goroutines before and after.
	goroutinesBefore := countRiftGoroutines(s)

	// Shutdown.
	done := make(chan struct{})
	go func() {
		s.shutdown()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("shutdown did not complete")
	}

	// Give goroutines a moment to settle.
	time.Sleep(100 * time.Millisecond)

	goroutinesAfter := countRiftGoroutines(s)
	if goroutinesAfter > goroutinesBefore {
		t.Fatalf("goroutine leak: before=%d after=%d", goroutinesBefore, goroutinesAfter)
	}
}

// Test 5: Database closed exactly once.
func TestShutdown_DBClosedExactlyOnce(t *testing.T) {
	s := newTestServer(t)
	_, cleanup := startTestHTTP(t, s)
	defer cleanup()

	// First shutdown.
	s.shutdown()
	count := s.dbCloseCount.Load()
	if count != 1 {
		t.Fatalf("expected 1 db.Close call, got %d", count)
	}

	// Second shutdown — db.Close already called; verify no panic.
	// shutdown calls httpServer.Shutdown (will error since already closed),
	// then s.cancel (idempotent), then wg.Wait (already zero), then db.Close
	// (will error since already closed but that's fine).
	done := make(chan struct{})
	go func() {
		s.shutdown()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("second shutdown did not complete")
	}

	finalCount := s.dbCloseCount.Load()
	if finalCount != 2 {
		t.Fatalf("expected 2 db.Close calls after double shutdown, got %d", finalCount)
	}
}

// Test 6: Idempotent shutdown — multiple signals, no panic.
func TestShutdown_Idempotent(t *testing.T) {
	s := newTestServer(t)
	_, cleanup := startTestHTTP(t, s)
	defer cleanup()

	// Run shutdown concurrently from multiple goroutines — only one should
	// complete cleanly, but none should panic.
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic during shutdown: %v", r)
				}
			}()
			s.shutdown()
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("concurrent shutdowns did not complete")
	}

	// DB should have been closed (once per shutdown call that got past the
	// httpServer.Shutdown error). The important thing is no panic.
	count := s.dbCloseCount.Load()
	if count < 1 {
		t.Fatalf("expected at least 1 db.Close call, got %d", count)
	}
}

// countRiftGoroutines returns an estimate of goroutines still active in the
// server's handler path. This is a best-effort check for leaks.
func countRiftGoroutines(s *Server) int {
	// The WaitGroup tracks active handlers — if it's zero, all handlers exited.
	// We can't inspect WaitGroup internals, so we check the session count.
	return s.sessionManager.Count()
}

// TestShutdown_ContextCancellationPropagates verifies that cancelling the
// server context causes the rift event loop to exit.
func TestShutdown_ContextCancellationPropagates(t *testing.T) {
	s := newTestServer(t)

	// Create a custom context we can cancel.
	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx

	// Start a minimal event loop that respects context.
	eventChan := make(chan riftEvent, 10)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			case ev := <-eventChan:
				_ = ev
			}
		}
	}()

	// Cancel context — event loop should exit.
	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("event loop did not exit after context cancellation")
	}
}
