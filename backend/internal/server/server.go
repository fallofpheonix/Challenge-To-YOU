package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"challenge-to-you/backend/internal/ai"
	"challenge-to-you/backend/internal/compiler"
	"challenge-to-you/backend/internal/db"
	"challenge-to-you/backend/internal/engine"
	"challenge-to-you/backend/internal/eventbus"
	"challenge-to-you/backend/internal/executionengine"
	pythonexecutor "challenge-to-you/backend/internal/executor/python"
	"challenge-to-you/backend/internal/missionengine"
	"challenge-to-you/backend/internal/obs"
	"challenge-to-you/backend/internal/sandbox"
	"challenge-to-you/backend/internal/session"

	"github.com/gorilla/websocket"
)

type Server struct {
	oracle          *ai.OracleClient
	compilerManager *compiler.Manager
	execEngine      *executionengine.Engine
	sb              *sandbox.ProcessSandbox
	bus             *eventbus.EventBus
	db              *db.DB
	challengePath   string
	missionRegistry *missionengine.MissionRegistry
	missionManager  *missionengine.MissionManager
	sessionManager  *session.Manager
	upgrader        websocket.Upgrader

	globalFabrics map[string]*engine.AxiomaticFabric

	// Observability
	log     *obs.Logger
	metrics *obs.Metrics

	// Shutdown infrastructure
	httpServer   *http.Server
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	dbCloseCount atomic.Int32
}

func NewServer() *Server {
	s := &Server{
		globalFabrics: make(map[string]*engine.AxiomaticFabric),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		log:     obs.Default().Component("server"),
		metrics: obs.NewMetrics(),
	}
	s.initDeps()
	return s
}

func (s *Server) initDeps() {
	s.oracle = ai.NewOracleClient(os.Getenv("OLLAMA_URL"), os.Getenv("OLLAMA_MODEL"))
	s.bus = eventbus.NewEventBus(1000)

	s.compilerManager = compiler.NewManager()
	s.compilerManager.RegisterLanguage(&compiler.Language{
		ID:          "python",
		Name:        "Python 3",
		Extensions:  []string{".py"},
		RunCmd:      "python3 {file}",
		TimeoutMs:   5000,
		MemoryBytes: 256 * 1024 * 1024,
	}, pythonexecutor.NewExecutor())

	s.sb = sandbox.NewProcessSandbox()
	s.execEngine = executionengine.NewEngine(s.compilerManager, s.sb)

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "challenge.db"
	}
	var err error
	s.db, err = db.NewDB(dbPath)
	if err != nil {
		s.log.Error("database init failed", "error", err, "class", obs.ClassPersistence)
		log.Fatalf("Failed to initialize database: %v", err)
	}

	s.challengePath = os.Getenv("CHALLENGE_PATH")
	if s.challengePath == "" {
		s.challengePath = filepath.Join("challenges", "magitech_tier1", "magitech_01.json")
	}

	s.missionRegistry = missionengine.NewMissionRegistry()
	// Resolve the missions dir relative to the working directory. The server may
	// be launched from the repo root or from backend/, so try both before warning.
	missionsDir := os.Getenv("MISSIONS_PATH")
	candidates := []string{missionsDir, filepath.Join("data", "missions"), filepath.Join("..", "data", "missions")}
	loaded := false
	for _, dir := range candidates {
		if dir == "" {
			continue
		}
		if _, statErr := os.Stat(dir); statErr != nil {
			continue
		}
		if err := s.missionRegistry.LoadMissionsFromDir(dir); err != nil {
			s.log.Warn("mission load failed", "dir", dir, "error", err)
		} else {
			loaded = true
		}
		break
	}
	if !loaded {
		s.log.Warn("no missions directory found", "tried", candidates)
	}
	s.missionManager = missionengine.NewMissionManager(s.missionRegistry, s.bus, s.db)
	s.sessionManager = session.NewManager(30 * time.Minute)
}

// Start loads the challenge and runs the server until a signal is received.
// It blocks until shutdown completes.
func (s *Server) Start() {
	s.Run()
}

// Run loads the challenge, starts the HTTP server, and blocks until SIGINT or
// SIGTERM is received. On signal it performs a deterministic shutdown:
//
//	Stop accepting connections → close listeners → notify handlers →
//	persist state → close database → exit
func (s *Server) Run() {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	def, err := engine.LoadChallenge(s.challengePath)
	if err != nil {
		s.log.Error("challenge load failed", "path", s.challengePath, "error", err)
		log.Fatalf("Failed to load challenge %s: %v", s.challengePath, err)
	}

	fabric := def.BuildFabric()

	mux := http.NewServeMux()
	mux.HandleFunc("/rift", func(w http.ResponseWriter, r *http.Request) {
		requestID := obs.NewRequestID()
		s.metrics.TotalRequests.Add(1)
		s.metrics.ActiveWebSockets.Add(1)
		defer s.metrics.ActiveWebSockets.Add(-1)

		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			s.log.Error("rift upgrade failed", "request_id", requestID, "error", err)
			return
		}

		ctx := obs.WithRequestID(s.ctx, requestID)
		s.log.Lifecycle("ClientConnected", "request_id", requestID, "remote", r.RemoteAddr)

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer conn.Close()
			s.handleRift(ctx, r, conn, def, fabric)
		}()
	})
	mux.HandleFunc("/api/languages", s.handleListLanguages)
	mux.HandleFunc("/api/challenges", s.handleListChallenges)
	mux.HandleFunc("/debug/runtime", obs.DebugHandler(s.metrics))
	mux.HandleFunc("/metrics", obs.MetricsHandler(s.metrics))

	s.log.Lifecycle("ServerStarted", "challenge_id", def.ID, "challenge_name", def.Name)
	s.log.Info("languages registered", "count", len(s.compilerManager.ListLanguages()))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ReadHeaderTimeout bounds slow-header (Slowloris) attacks. WriteTimeout is
	// left unset because WebSocket connections are intentionally long-lived.
	s.httpServer = &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Start server in a goroutine so we can listen for signals.
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("listening", "port", port)
		errCh <- s.httpServer.ListenAndServe()
	}()

	// Wait for shutdown signal or server error.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		s.log.Lifecycle("ShutdownStarted", "signal", sig.String())
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			s.log.Error("server error", "error", err)
			log.Fatalf("Server error: %v", err)
		}
	}

	s.shutdown()
}

// shutdown performs the deterministic shutdown sequence.
func (s *Server) shutdown() {
	shutdownStart := time.Now()

	// 1. Stop accepting new connections (in-flight requests get a deadline).
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.log.Error("http shutdown error", "error", err)
	}

	// 2. Cancel server context — signals all active handlers to exit.
	s.cancel()

	// 3. Wait for all handlers to finish (with timeout).
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.log.Info("all handlers drained")
	case <-time.After(15 * time.Second):
		s.log.Warn("handler drain timeout — forcing shutdown")
	}

	// 4. Close database (checkpoints WAL, releases file descriptors).
	s.dbCloseCount.Add(1)
	if err := s.db.Close(); err != nil {
		s.log.Error("database close error", "error", err)
	}

	duration := time.Since(shutdownStart)
	s.metrics.ShutdownDuration.Store(int64(duration))
	s.log.Lifecycle("ShutdownCompleted", "duration_ms", duration.Milliseconds())
}
