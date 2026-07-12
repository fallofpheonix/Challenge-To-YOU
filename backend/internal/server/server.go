package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"challenge-to-you/backend/internal/ai"
	"challenge-to-you/backend/internal/compiler"
	"challenge-to-you/backend/internal/db"
	"challenge-to-you/backend/internal/engine"
	"challenge-to-you/backend/internal/eventbus"
	"challenge-to-you/backend/internal/executionengine"
	pythonexecutor "challenge-to-you/backend/internal/executor/python"
	"challenge-to-you/backend/internal/missionengine"
	"challenge-to-you/backend/internal/sandbox"

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
	upgrader        websocket.Upgrader

	globalFabrics map[string]*engine.AxiomaticFabric
}

func NewServer() *Server {
	s := &Server{
		globalFabrics: make(map[string]*engine.AxiomaticFabric),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
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
		log.Fatalf("Failed to initialize database: %v", err)
	}

	s.challengePath = os.Getenv("CHALLENGE_PATH")
	if s.challengePath == "" {
		s.challengePath = filepath.Join("challenges", "magitech_tier1", "magitech_01.json")
	}

	s.missionRegistry = missionengine.NewMissionRegistry()
	if err := s.missionRegistry.LoadMissionsFromDir("data/missions"); err != nil {
		log.Printf("Warning: failed to load missions: %v", err)
	}
	s.missionManager = missionengine.NewMissionManager(s.missionRegistry, s.bus, s.db)
}

func (s *Server) Start() {
	def, err := engine.LoadChallenge(s.challengePath)
	if err != nil {
		log.Fatalf("Failed to load challenge %s: %v", s.challengePath, err)
	}

	fabric := def.BuildFabric()

	http.HandleFunc("/rift", func(w http.ResponseWriter, r *http.Request) {
		s.handleRift(w, r, def, fabric)
	})
	http.HandleFunc("/api/languages", s.handleListLanguages)
	http.HandleFunc("/api/challenges", s.handleListChallenges)

	log.Printf("Challenge Engine online — challenge: %s (%s)", def.ID, def.Name)
	log.Printf("Registered languages: %v", s.compilerManager.ListLanguages())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
