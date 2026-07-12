# Challenge To YOU — Phase 1 Architecture

## 1. Design Principles

| Principle | Description |
|-----------|-------------|
| **Data-Driven** | Challenges, missions, dialogue, progression loaded from JSON files. Zero engine changes to add content. |
| **Modular** | Each subsystem is a Go package with a clean interface. No circular dependencies. |
| **Event-Driven** | Modules communicate via a central EventBus. No direct cross-module calls. |
| **Language-Agnostic** | Compiler/sandbox abstraction supports any language via executor plugins. |
| **Scalable** | Architecture supports 10,000+ challenges, multiple campaigns, future multiplayer. |
| **Offline-First** | Full game works without network. Multiplayer is a future extension. |

---

## 2. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        GODOT 4 CLIENT                               │
│                                                                     │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │   IDE    │ │  Story   │ │  Audio   │ │  Save    │ │  UI      │ │
│  │  System  │ │  System  │ │  System  │ │  System  │ │  System  │ │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ │
│       └─────────────┴────────────┴─────────────┴─────────────┘      │
│                              │                                       │
│                    ┌─────────▼─────────┐                             │
│                    │  Network Bridge   │                             │
│                    │  (WebSocket)      │                             │
│                    └─────────┬─────────┘                             │
└──────────────────────────────┼──────────────────────────────────────┘
                               │ ws://localhost:8080/rift
┌──────────────────────────────┼──────────────────────────────────────┐
│                        GO BACKEND                                   │
│                              │                                       │
│                    ┌─────────▼─────────┐                             │
│                    │   Event Bus      │◄──── All modules publish     │
│                    │   (Central Hub)  │      and subscribe here      │
│                    └─────────┬─────────┘                             │
│                              │                                       │
│  ┌──────────┐ ┌──────────┐ ┌┴─────────┐ ┌──────────┐ ┌──────────┐  │
│  │ Mission  │ │Challenge │ │Compiler  │ │Progression│ │  Save    │  │
│  │ Engine   │ │ Engine   │ │ Manager  │ │  System   │ │  System  │  │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘  │
│       │            │            │             │            │         │
│  ┌────▼─────┐ ┌────▼─────┐ ┌───▼──────┐ ┌───▼──────┐ ┌───▼──────┐  │
│  │ Dialogue │ │Sandbox   │ │Execution │ │Analytics │ │  AI      │  │
│  │ System   │ │ Manager  │ │ Engine   │ │ Tracker  │ │ Integration│ │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    Language Executors                         │   │
│  │  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐  │   │
│  │  │ C/  │ │Java │ │ Py  │ │ JS  │ │Rust │ │ Go  │ │Kotlin│  │   │
│  │  │ C++ │ │     │ │     │ │     │ │     │ │     │ │  C#  │  │   │
│  │  └─────┘ └─────┘ └─────┘ └─────┘ └─────┘ └─────┘ └─────┘  │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    Data Layer                                 │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │   │
│  │  │Challenge │ │ Mission  │ │ Dialogue │ │  Save    │        │   │
│  │  │  Files   │ │  Files   │ │  Files   │ │  Files   │        │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘        │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 3. Module Breakdown

### 3.1 Backend Modules (Go Packages)

#### Module: `eventbus` — Central Event System

```go
// Event types flowing through the system
type EventType string

const (
    // Client-facing events
    EventChallengeLoaded    EventType = "challenge.loaded"
    EventChallengeCompleted EventType = "challenge.completed"
    EventSnapshot           EventType = "snapshot"
    EventArchonTransmission EventType = "archon.transmission"
    EventPurge              EventType = "purge"
    
    // Internal events
    EventCodeSubmitted      EventType = "code.submitted"
    EventCodeCompiled       EventType = "code.compiled"
    EventCodeExecuted       EventType = "code.executed"
    EventTestPassed         EventType = "test.passed"
    EventTestFailed         EventType = "test.failed"
    EventXPEarned           EventType = "xp.earned"
    EventLevelUp            EventType = "level.up"
    EventAchievementUnlocked EventType = "achievement.unlocked"
    EventMissionStarted     EventType = "mission.started"
    EventMissionCompleted   EventType = "mission.completed"
    EventDialogueStarted    EventType = "dialogue.started"
    EventDialogueChoice     EventType = "dialogue.choice"
    EventSaveLoaded         EventType = "save.loaded"
    EventSaveCreated        EventType = "save.created"
)

type Event struct {
    Type      EventType
    Payload   interface{}
    Source    string  // module that published
    Timestamp time.Time
}

type EventHandler func(Event)

type EventBus struct {
    mu          sync.RWMutex
    handlers    map[EventType][]EventHandler
    history     []Event           // for replay
    maxHistory  int
}

func NewEventBus() *EventBus
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler)
func (eb *EventBus) Unsubscribe(eventType EventType, handler EventHandler)
func (eb *EventBus) Publish(event Event)
func (eb *EventBus) GetHistory() []Event
```

#### Module: `challengeengine` — Challenge Loading & Management

```go
// ChallengeDefinition represents a single programming challenge
// Extended from existing engine.ChallengeDefinition
type ChallengeDefinition struct {
    // === Identity ===
    ID          string   `json:"id"`
    Version     string   `json:"version"`
    
    // === Content ===
    Title       string   `json:"title"`
    StoryContext string  `json:"story_context"`
    Briefing    string   `json:"mission_briefing"`
    
    // === Gameplay ===
    Paradigm    string   `json:"paradigm"`
    Category    string   `json:"category"`
    Tags        []string `json:"tags"`
    Difficulty  float64  `json:"difficulty"`    // 0.0-1.0
    EstTimeMins int      `json:"est_time_mins"`
    
    // === Programming ===
    Language    string   `json:"language"`       // "python", "go", "java", etc.
    StarterCode string   `json:"starter_code"`
    InputSpec   string   `json:"input_spec"`
    OutputSpec  string   `json:"output_spec"`
    Constraints []string `json:"constraints"`
    
    // === Test Cases ===
    VisibleTests  []TestCase `json:"visible_tests"`
    HiddenTests   []TestCase `json:"hidden_tests"`
    EdgeCases     []TestCase `json:"edge_cases"`
    Validators    []string   `json:"validators"`    // custom validator scripts
    
    // === Hints ===
    Hints         []Hint     `json:"hints"`
    
    // === Solution ===
    ReferenceSolution string `json:"reference_solution"`
    AltSolutions      []string `json:"alt_solutions"`
    SolutionExplanation string `json:"solution_explanation"`
    TimeComplexity    string `json:"time_complexity"`
    SpaceComplexity   string `json:"space_complexity"`
    
    // === Legacy (from current system) ===
    InitialState     ParadigmState    `json:"initial_state"`
    Flaws            []Flaw           `json:"flaws"`
    WinCondition     WinCondition     `json:"win_condition"`
    TemplateCode     string           `json:"template_code,omitempty"`
    ValidationScript string           `json:"validation_script,omitempty"`
    SkillType        string           `json:"skill_type,omitempty"`
    ExpectedAnswer   string           `json:"expected_answer,omitempty"`
    
    // === Progression ===
    XPReward         int              `json:"xp_reward"`
    UnlockRequirements []string       `json:"unlock_requirements"`
    Achievements     []string         `json:"achievements"`
    
    // === Story Integration ===
    DialogueID       string           `json:"dialogue_id,omitempty"`
    NPCInteractions  []NPCInteraction `json:"npc_interactions,omitempty"`
    
    // === Analytics ===
    LearningObjectives []string       `json:"learning_objectives"`
    FailureMessages    []string       `json:"failure_messages"`
    SuccessMessages    []string       `json:"success_messages"`
    
    // === Metadata ===
    ReplayMetadata    ReplayMetadata  `json:"replay_metadata,omitempty"`
    AnalyticsMetadata AnalyticsMetadata `json:"analytics_metadata,omitempty"`
}

type TestCase struct {
    ID          string      `json:"id"`
    Input       string      `json:"input"`
    ExpectedOutput string   `json:"expected_output"`
    Description string      `json:"description"`
    IsHidden    bool        `json:"is_hidden"`
    Timeout     int         `json:"timeout_ms"`     // override default
    MemoryLimit int         `json:"memory_bytes"`   // override default
}

type Hint struct {
    Level       int    `json:"level"`
    Cost        int    `json:"cost"`           // XP cost to unlock
    Content     string `json:"content"`
    Type        string `json:"type"`           // "text", "code_snippet", "visual"
}

type NPCInteraction struct {
    NPCID       string `json:"npc_id"`
    DialogueLine string `json:"dialogue_line"`
    Trigger     string `json:"trigger"`        // "on_start", "on_hint", "on_complete"
}

type ReplayMetadata struct {
    SolutionHash string `json:"solution_hash"`
    Paradigm     string `json:"paradigm"`
    Difficulty   float64 `json:"difficulty"`
}

type AnalyticsMetadata struct {
    AvgCompletionTime int     `json:"avg_completion_time"`
    SuccessRate       float64 `json:"success_rate"`
    CommonErrors      []string `json:"common_errors"`
}

// ChallengePack represents a collection of challenges (e.g., era tier)
type ChallengePack struct {
    Version      int            `json:"version"`
    Era          string         `json:"era"`
    Tier         int            `json:"tier"`
    Name         string         `json:"name"`
    Description  string         `json:"description"`
    Prerequisites []string      `json:"prerequisites"`
    Unlocks      []string       `json:"unlocks"`
    LuckModifier float64        `json:"luck_modifier"`
    Challenges   []ChallengeRef `json:"challenges"`
}

type ChallengeRef struct {
    ID         string  `json:"id"`
    File       string  `json:"file"`
    Difficulty float64 `json:"difficulty"`
    Mode       string  `json:"mode"`
}

// ChallengeEngine interface
type ChallengeEngine interface {
    LoadChallenge(id string) (*ChallengeDefinition, error)
    LoadPack(packID string) (*ChallengePack, error)
    ListPacks() []string
    ListChallenges(packID string) []string
    GetAvailableChallenges(playerProgress *PlayerProgress) []string
    ValidateChallenge(def *ChallengeDefinition) error
}
```

#### Module: `compiler` — Language Abstraction

```go
// Language defines a supported programming language
type Language struct {
    ID            string   `json:"id"`            // "python", "java", "go", etc.
    Name          string   `json:"name"`          // "Python 3.11"
    Extensions    []string `json:"extensions"`    // [".py"]
    CompileCmd    string   `json:"compile_cmd"`   // "python3 -m py_compile {file}"
    RunCmd        string   `json:"run_cmd"`       // "python3 {file}"
    Timeout       int      `json:"timeout_ms"`    // default 5000
    MemoryLimit   int      `json:"memory_bytes"`  // default 256MB
    DockerImage   string   `json:"docker_image"`  // "python:3.11-slim"
    EnvVars       []string `json:"env_vars"`      // ["PYTHONUNBUFFERED=1"]
}

// CompilationResult represents the outcome of compilation
type CompilationResult struct {
    Success    bool     `json:"success"`
    Output     string   `json:"output"`
    Errors     []CompileError `json:"errors"`
    Warnings   []string `json:"warnings"`
    DurationMs int      `json:"duration_ms"`
}

type CompileError struct {
    Line    int    `json:"line"`
    Column  int    `json:"column"`
    Message string `json:"message"`
    File    string `json:"file"`
}

// Executor defines how to execute code in a language
type Executor interface {
    Compile(ctx context.Context, code string, lang *Language) (*CompilationResult, error)
    Execute(ctx context.Context, code string, input string, lang *Language) (*ExecutionResult, error)
    Validate(ctx context.Context, code string, testCases []TestCase, lang *Language) (*ValidationResult, error)
}

// CompilationResult from execution
type ExecutionResult struct {
    Success    bool          `json:"success"`
    Output     string        `json:"output"`
    Error      string        `json:"error,omitempty"`
    ExitCode   int           `json:"exit_code"`
    DurationMs int           `json:"duration_ms"`
    MemoryUsed int           `json:"memory_used_bytes"`
    TimedOut   bool          `json:"timed_out"`
}

type ValidationResult struct {
    AllPassed   bool           `json:"all_passed"`
    Results     []TestResult   `json:"results"`
    Score       float64        `json:"score"`          // 0.0-1.0
    TotalTests  int            `json:"total_tests"`
    PassedTests int            `json:"passed_tests"`
}

type TestResult struct {
    TestCaseID  string `json:"test_case_id"`
    Passed      bool   `json:"passed"`
    Input       string `json:"input"`
    Expected    string `json:"expected_output"`
    Actual      string `json:"actual_output"`
    DurationMs  int    `json:"duration_ms"`
    Error       string `json:"error,omitempty"`
}

// CompilerManager manages all language executors
type CompilerManager interface {
    RegisterLanguage(lang *Language, executor Executor)
    GetLanguage(id string) (*Language, error)
    GetExecutor(langID string) (Executor, error)
    ListLanguages() []*Language
    Compile(ctx context.Context, code string, langID string) (*CompilationResult, error)
    Execute(ctx context.Context, code string, input string, langID string) (*ExecutionResult, error)
    Validate(ctx context.Context, code string, testCases []TestCase, langID string) (*ValidationResult, error)
}
```

#### Module: `sandbox` — Isolated Execution

```go
// SandboxConfig defines resource limits for execution
type SandboxConfig struct {
    TimeoutMs     int    `json:"timeout_ms"`      // default 5000
    MemoryBytes   int    `json:"memory_bytes"`     // default 256MB
    MaxOutputSize int    `json:"max_output_bytes"` // default 1MB
    NetworkAccess bool   `json:"network_access"`   // default false
    FileSystem    bool   `json:"filesystem_access"` // default false
    PID1          bool   `json:"pid1"`             // run as PID 1
}

// Sandbox provides isolated code execution
type Sandbox interface {
    Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error)
    Cleanup() error
}

type ExecutionRequest struct {
    Code       string        `json:"code"`
    Language   string        `json:"language"`
    Input      string        `json:"input"`
    Config     SandboxConfig `json:"config"`
    WorkingDir string        `json:"working_dir"`
}

type ExecutionResponse struct {
    Success    bool   `json:"success"`
    Output     string `json:"output"`
    Error      string `json:"error,omitempty"`
    ExitCode   int    `json:"exit_code"`
    DurationMs int    `json:"duration_ms"`
    MemoryUsed int    `json:"memory_used_bytes"`
}

// DockerSandbox implements Sandbox using Docker containers
type DockerSandbox struct {
    dockerClient *client.Client
    imageCache   map[string]string
}

// ProcessSandbox implements Sandbox using OS processes (simpler, less isolation)
type ProcessSandbox struct {
    tempDir string
}
```

#### Module: `executionengine` — Test Case Runner

```go
// ExecutionEngine runs player code against test cases
type ExecutionEngine interface {
    RunSingleTest(ctx context.Context, code string, test TestCase, lang string) (*TestResult, error)
    RunAllTests(ctx context.Context, code string, tests []TestCase, lang string) (*ValidationResult, error)
    RunWithValidator(ctx context.Context, code string, validator string, tests []TestCase, lang string) (*ValidationResult, error)
}

// DefaultExecutionEngine implements ExecutionEngine
type DefaultExecutionEngine struct {
    compiler CompilerManager
    sandbox  Sandbox
}
```

#### Module: `missionengine` — Mission Flow & Objectives

```go
// Mission represents a story-driven sequence of challenges
type Mission struct {
    ID          string            `json:"id"`
    Title       string            `json:"title"`
    Description string            `json:"description"`
    Era         string            `json:"era"`
    Act         int               `json:"act"`           // story act
    Chapter     int               `json:"chapter"`       // chapter within act
    
    // Flow
    Intro       *Cutscene         `json:"intro,omitempty"`
    Dialogue    []DialogueNode    `json:"dialogue"`
    Objectives  []Objective       `json:"objectives"`
    Challenges  []ChallengeRef    `json:"challenges"`
    Reward      *MissionReward    `json:"reward"`
    Outro       *Cutscene         `json:"outro,omitempty"`
    
    // Unlock conditions
    Requirements []string         `json:"requirements"`
    
    // Metadata
    EstTimeMins int               `json:"est_time_mins"`
    Difficulty  float64           `json:"difficulty"`
}

type Objective struct {
    ID          string `json:"id"`
    Type        string `json:"type"`     // "complete_challenge", "find_item", "talk_to_npc", "explore_area"
    Target      string `json:"target"`   // challenge_id, item_id, npc_id, area_id
    Description string `json:"description"`
    Optional    bool   `json:"optional"`
}

type MissionReward struct {
    XP          int      `json:"xp"`
    Credits     int      `json:"credits"`
    Unlocks     []string `json:"unlocks"`     // mission_ids, item_ids, area_ids
    Achievements []string `json:"achievements"`
}

type Cutscene struct {
    ID       string         `json:"id"`
    Type     string         `json:"type"`       // "cinematic", "dialogue", "tutorial"
    Scenes   []CutsceneScene `json:"scenes"`
}

type CutsceneScene struct {
    Background string        `json:"background"`
    Characters []string      `json:"characters"`
    Dialogue   []DialogueLine `json:"dialogue"`
    Duration   int           `json:"duration_ms"`
    Effect     string        `json:"effect,omitempty"`
}

// MissionEngine manages mission flow
type MissionEngine interface {
    LoadMission(id string) (*Mission, error)
    StartMission(id string, playerID string) (*MissionSession, error)
    GetCurrentMission(playerID string) (*Mission, error)
    CompleteObjective(playerID string, objectiveID string) error
    CompleteMission(playerID string) (*MissionReward, error)
    GetAvailableMissions(playerProgress *PlayerProgress) []*Mission
}

type MissionSession struct {
    Mission       *Mission
    PlayerID      string
    CurrentStep   int
    CompletedObjectives []string
    StartTime     time.Time
    State         map[string]interface{}
}
```

#### Module: `dialogue` — NPC Dialogue System

```go
// DialogueTree represents a conversation with an NPC
type DialogueTree struct {
    ID        string                `json:"id"`
    NPCID     string                `json:"npc_id"`
    NPCName   string                `json:"npc_name"`
    Nodes     map[string]*DialogueNode `json:"nodes"`
    StartNode string                `json:"start_node"`
}

type DialogueNode struct {
    ID          string            `json:"id"`
    Speaker     string            `json:"speaker"`      // "npc" or "player"
    Text        string            `json:"text"`
    Choices     []DialogueChoice  `json:"choices,omitempty"`
    NextNode    string            `json:"next_node,omitempty"`
    Conditions  []string          `json:"conditions,omitempty"`  // state checks
    Effects     []string          `json:"effects,omitempty"`     // state changes
    Emotion     string            `json:"emotion,omitempty"`     // "neutral", "happy", "angry"
}

type DialogueChoice struct {
    ID          string   `json:"id"`
    Text        string   `json:"text"`
    NextNode    string   `json:"next_node"`
    Requirements []string `json:"requirements,omitempty"`
    XPBonus     int      `json:"xp_bonus,omitempty"`
}

// DialogueEngine manages dialogue trees
type DialogueEngine interface {
    LoadDialogue(id string) (*DialogueTree, error)
    StartDialogue(id string, playerState map[string]interface{}) (*DialogueSession, error)
    MakeChoice(sessionID string, choiceID string) (*DialogueResponse, error)
    AdvanceDialogue(sessionID string) (*DialogueResponse, error)
    EndDialogue(sessionID string)
}

type DialogueSession struct {
    Tree        *DialogueTree
    CurrentNode string
    History     []string          // node IDs visited
    PlayerState map[string]interface{}
}

type DialogueResponse struct {
    Speaker   string
    Text      string
    Choices   []DialogueChoice
    Effects   []string
    Complete  bool
}
```

#### Module: `progression` — XP, Levels, Ranks

```go
// PlayerProgress tracks all player progression
type PlayerProgress struct {
    // Identity
    PlayerID    string `json:"player_id"`
    Name        string `json:"name"`
    Title       string `json:"title"`
    Level       int    `json:"level"`
    Rank        string `json:"rank"`
    
    // Resources
    XP          int    `json:"xp"`
    Credits     int    `json:"credits"`
    Luck        float64 `json:"luck"`
    
    // Progress
    CompletedChallenges []string `json:"completed_challenges"`
    CompletedMissions   []string `json:"completed_missions"`
    UnlockedAreas       []string `json:"unlocked_areas"`
    UnlockedLanguages   []string `json:"unlocked_languages"`
    
    // Skills
    SkillTree   map[string]int `json:"skill_tree"`    // skill_id -> level
    
    // Cosmetics
    Titles      []string `json:"titles"`
    Badges      []string `json:"badges"`
    Themes      []string `json:"themes"`
    
    // Statistics
    Stats       PlayerStats `json:"stats"`
    
    // Timestamps
    CreatedAt   time.Time `json:"created_at"`
    LastPlayed  time.Time `json:"last_played"`
    TotalPlayTime int     `json:"total_play_time_secs"`
}

type PlayerStats struct {
    ChallengesAttempted  int            `json:"challenges_attempted"`
    ChallengesCompleted  int            `json:"challenges_completed"`
    TotalCompiles        int            `json:"total_compiles"`
    TotalRuntime         int            `json:"total_runtime_ms"`
    ErrorsEncountered    int            `json:"errors_encountered"`
    HintsUsed            int            `json:"hints_used"`
    AchievementsEarned   int            `json:"achievements_earned"`
    CategoryStats        map[string]CategoryStats `json:"category_stats"`
}

type CategoryStats struct {
    Attempted int `json:"attempted"`
    Completed int `json:"completed"`
    AvgTime   int `json:"avg_time_ms"`
}

// Skill represents a skill tree node
type Skill struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Category    string   `json:"category"`
    MaxLevel    int      `json:"max_level"`
    CostPerLevel int     `json:"cost_per_level"`  // XP cost
    Prerequisites []string `json:"prerequisites"`
    Effects     []SkillEffect `json:"effects"`
}

type SkillEffect struct {
    Type   string `json:"type"`     // "unlock_language", "reduce_timeout", "bonus_xp"
    Value  string `json:"value"`
    Amount int    `json:"amount"`
}

// ProgressionSystem manages player progression
type ProgressionSystem interface {
    GetProgress(playerID string) (*PlayerProgress, error)
    AddXP(playerID string, amount int) (int, bool, error) // newXP, levelUp, error
    AddCredits(playerID string, amount int) error
    UnlockTitle(playerID string, title string) error
    UnlockBadge(playerID string, badge string) error
    UnlockTheme(playerID string, theme string) error
    UpgradeSkill(playerID string, skillID string) error
    GetLevel(level int) (*LevelDef, error)
    GetRank(xp int) string
}

type LevelDef struct {
    Level    int   `json:"level"`
    XPRequired int `json:"xp_required"`
    Title    string `json:"title"`
    Rewards  []LevelReward `json:"rewards"`
}

type LevelReward struct {
    Type  string `json:"type"`    // "skill_point", "title", "badge", "theme"
    Value string `json:"value"`
}
```

#### Module: `save` — Persistence

```go
// SaveData represents a complete game save
type SaveData struct {
    Version     string          `json:"version"`
    SlotID      int             `json:"slot_id"`
    PlayerName  string          `json:"player_name"`
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
    PlayTime    int             `json:"play_time_secs"`
    
    // Game state
    Progress    *PlayerProgress `json:"progress"`
    MissionState map[string]*MissionSession `json:"mission_state"`
    
    // Settings
    Settings    *GameSettings   `json:"settings"`
}

type GameSettings struct {
    MusicVolume   float64 `json:"music_volume"`
    SFXVolume     float64 `json:"sfx_volume"`
    TextSpeed     int     `json:"text_speed"`
    AutoSave      bool    `json:"auto_save"`
    Theme         string  `json:"theme"`
    FontSize      int     `json:"font_size"`
}

// SaveManager handles persistence
type SaveManager interface {
    CreateSave(slotID int, name string) (*SaveData, error)
    LoadSave(slotID int) (*SaveData, error)
    SaveGame(slotID int, data *SaveData) error
    DeleteSave(slotID int) error
    ListSaves() []*SaveInfo
    AutoSave(playerID string) error
    ExportSave(slotID int) ([]byte, error)
    ImportSave(data []byte) (*SaveData, error)
}

type SaveInfo struct {
    SlotID    int       `json:"slot_id"`
    Name      string    `json:"name"`
    Level     int       `json:"level"`
    PlayTime  int       `json:"play_time_secs"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

#### Module: `analytics` — Tracking

```go
// AnalyticsEvent represents a tracked event
type AnalyticsEvent struct {
    Type      string                 `json:"type"`
    PlayerID  string                 `json:"player_id"`
    Timestamp time.Time              `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
}

// AnalyticsTracker records player behavior
type AnalyticsTracker interface {
    Track(event AnalyticsEvent)
    GetChallengeStats(challengeID string) (*ChallengeAnalytics, error)
    GetPlayerStats(playerID string) (*PlayerAnalytics, error)
    GetCommonErrors(challengeID string) []string
    GetSuccessRate(challengeID string) float64
    GetAvgCompletionTime(challengeID string) int
}

type ChallengeAnalytics struct {
    ChallengeID     string  `json:"challenge_id"`
    Attempts        int     `json:"attempts"`
    Completions     int     `json:"completions"`
    SuccessRate     float64 `json:"success_rate"`
    AvgTime         int     `json:"avg_time_ms"`
    CommonErrors    []string `json:"common_errors"`
    DifficultyRating float64 `json:"difficulty_rating"`
}

type PlayerAnalytics struct {
    PlayerID        string            `json:"player_id"`
    TotalPlayTime   int               `json:"total_play_time_secs"`
    ChallengesPerCategory map[string]int `json:"challenges_per_category"`
    WeakTopics      []string          `json:"weak_topics"`
    StrongTopics    []string          `json:"strong_topics"`
    SkillMastery    map[string]float64 `json:"skill_mastery"`
}
```

#### Module: `ai` — Ollama Integration

```go
// AIIntegration provides AI-powered features
type AIIntegration interface {
    GenerateTaunt(ctx context.Context, context TauntContext) (string, error)
    GenerateRepair(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error)
    AnalyzeSolution(ctx context.Context, code string, lang string) (*SolutionAnalysis, error)
    GenerateHint(ctx context.Context, challengeID string, playerState map[string]interface{}) (string, error)
}

type TauntContext struct {
    Paradigm    string
    Action      string
    Entropy     int
    Vigilance   float64
}

type SolutionAnalysis struct {
    Quality     float64  `json:"quality"`      // 0.0-1.0
    Suggestions []string `json:"suggestions"`
    Complexity  string   `json:"complexity"`    // "O(n)", "O(n^2)", etc.
    Style       string   `json:"style"`         // "clean", "verbose", "optimized"
}
```

---

### 3.2 Client Modules (Godot 4)

#### Module: `IDE` — Code Editor

```
IDE System (res://scripts/ide/)
├── editor/
│   ├── code_editor.gd          # Main editor container
│   ├── syntax_highlighter.gd   # Language-specific highlighting
│   ├── autocomplete.gd         # Intellisense-like completion
│   ├── bracket_matcher.gd      # Highlight matching brackets
│   └── minimap.gd              # Code minimap
├── tabs/
│   ├── tab_bar.gd              # Tab management
│   └── tab.gd                  # Individual tab
├── tools/
│   ├── search_replace.gd       # Search & replace
│   ├── debugger.gd             # Step, breakpoints, variables
│   └── formatter.gd            # Code formatting
└── themes/
    ├── dracula.tres
    ├── monokai.tres
    └── solarized.tres
```

#### Module: `Story` — World & NPCs

```
Story System (res://scripts/story/)
├── dialogue/
│   ├── dialogue_manager.gd     # Dialogue tree walker
│   ├── dialogue_ui.gd          # Text box, choices
│   └── npc_portraits/          # Character art
├── world/
│   ├── world_map.gd            # Area navigation
│   ├── area.gd                 # Individual location
│   └── interactable.gd         # Interactive objects
├── missions/
│   ├── mission_tracker.gd      # Active mission display
│   └── objective_marker.gd     # Waypoint markers
└── cutscenes/
    ├── cutscene_player.gd      # Cutscene playback
    └── effects/                # Visual effects
```

#### Module: `Audio` — Sound System

```
Audio System (res://scripts/audio/)
├── music/
│   ├── music_manager.gd        # Track management, crossfade
│   └── tracks/                 # Music files
├── sfx/
│   ├── sfx_manager.gd          # Sound effect playback
│   └── effects/                # SFX files
└── ambient/
    ├── ambient_manager.gd      # Environmental sounds
    └── environments/           # Ambient tracks
```

#### Module: `Save` — Client Persistence

```
Save System (res://scripts/save/)
├── save_manager.gd             # Save/load logic
├── autosave.gd                 # Autosave timer
└── save_slots/                 # Save file storage
```

---

## 4. Data Formats

### 4.1 Challenge Format (Extended)

```json
{
  "schema_version": "2.0",
  "id": "cyberpunk_01_autodoc",
  "version": "1.0.0",
  
  "title": "The Autodoc Rebellion",
  "story_context": "A rogue medical AI has locked down the clinic. You must reprogram its healing protocols to open the pod.",
  "mission_briefing": "The Autodoc's neural net is fractured. Inject the correct healing sequence to stabilize its core and release the patient.",
  
  "paradigm": "CYBERPUNK",
  "category": "algorithms",
  "tags": ["state_machine", "conditional_logic", "medical"],
  "difficulty": 0.3,
  "est_time_mins": 15,
  
  "language": "python",
  "starter_code": "def heal_patient(patient_data):\n    # Your code here\n    pass",
  "input_spec": "A dictionary with 'heart_rate' (int), 'blood_pressure' (str)",
  "output_spec": "True if patient stabilized, False otherwise",
  "constraints": [
    "Must handle heart rates from 40-200",
    "Blood pressure formats: 'low', 'normal', 'high', 'critical'"
  ],
  
  "visible_tests": [
    {
      "id": "test_1",
      "input": "{'heart_rate': 72, 'blood_pressure': 'normal'}",
      "expected_output": "True",
      "description": "Healthy patient should stabilize"
    }
  ],
  "hidden_tests": [
    {
      "id": "test_hidden_1",
      "input": "{'heart_rate': 200, 'blood_pressure': 'critical'}",
      "expected_output": "False",
      "description": "Extreme case"
    }
  ],
  "edge_cases": [
    {
      "id": "edge_1",
      "input": "{'heart_rate': 0, 'blood_pressure': 'low'}",
      "expected_output": "False",
      "description": "Zero heart rate"
    }
  ],
  
  "hints": [
    {
      "level": 1,
      "cost": 10,
      "content": "Check both heart rate AND blood pressure",
      "type": "text"
    },
    {
      "level": 2,
      "cost": 25,
      "content": "Use logical AND: if heart_rate < 100 and bp == 'normal'",
      "type": "text"
    },
    {
      "level": 3,
      "cost": 50,
      "content": "def heal_patient(d): return d['heart_rate'] < 100 and d['blood_pressure'] in ('normal', 'low')",
      "type": "code_snippet"
    }
  ],
  
  "reference_solution": "def heal_patient(d): return d['heart_rate'] < 100 and d['blood_pressure'] in ('normal', 'low')",
  "alt_solutions": [
    "def heal_patient(d):\n    hr = d['heart_rate']\n    bp = d['blood_pressure']\n    return hr < 100 and bp in ('normal', 'low')"
  ],
  "solution_explanation": "Check if heart rate is below 100 and blood pressure is not high/critical.",
  "time_complexity": "O(1)",
  "space_complexity": "O(1)",
  
  "initial_state": { "pod_opened": false, "heart_rate_bpm": 80, "entropy": 0 },
  "flaws": [],
  "win_condition": { "target_state_key": "pod_opened", "expected_value": true },
  "skill_type": "write_from_spec",
  
  "xp_reward": 100,
  "unlock_requirements": [],
  "achievements": ["first_hack"],
  
  "dialogue_id": "npc_doctor_intro",
  "npc_interactions": [
    { "npc_id": "doctor", "dialogue_line": "Please help me! The AI won't respond!", "trigger": "on_start" }
  ],
  
  "learning_objectives": ["conditional logic", "dictionary access", "boolean expressions"],
  "failure_messages": [
    "The patient's condition worsens...",
    "Error: Invalid healing protocol"
  ],
  "success_messages": [
    "The pod hisses open. The patient is safe.",
    "LOGOS token acquired."
  ],
  
  "analytics_metadata": {
    "avg_completion_time": 900,
    "success_rate": 0.72,
    "common_errors": ["forgot blood_pressure check", "used > instead of <"]
  }
}
```

### 4.2 Mission Format

```json
{
  "id": "mission_cyberpunk_act1_ch1",
  "title": "Neon Awakening",
  "description": "You awaken in a dimly lit apartment. A flickering terminal beckons.",
  "era": "cyberpunk",
  "act": 1,
  "chapter": 1,
  
  "intro": {
    "id": "cutscene_intro",
    "type": "cinematic",
    "scenes": [
      {
        "background": "res://assets/backgrounds/apartment.png",
        "characters": ["player"],
        "dialogue": [
          { "speaker": "narrator", "text": "The year is 2087. You are a Data Ghost.", "emotion": "neutral" },
          { "speaker": "narrator", "text": "Your terminal flickers to life...", "emotion": "neutral" }
        ],
        "duration": 5000,
        "effect": "fade_in"
      }
    ]
  },
  
  "dialogue": ["npc_terminal_intro", "npc_mentor_first_meeting"],
  
  "objectives": [
    {
      "id": "obj_1",
      "type": "complete_challenge",
      "target": "cyberpunk_01_autodoc",
      "description": "Hack the Autodoc to save the patient"
    },
    {
      "id": "obj_2",
      "type": "find_item",
      "target": "item_data_chip",
      "description": "Find the hidden data chip",
      "optional": true
    }
  ],
  
  "challenges": [
    { "id": "cyberpunk_01_autodoc", "file": "cyberpunk_01_autodoc.json", "difficulty": 0.3, "mode": "architect" }
  ],
  
  "reward": {
    "xp": 150,
    "credits": 50,
    "unlocks": ["mission_cyberpunk_act1_ch2"],
    "achievements": ["first_mission"]
  },
  
  "requirements": ["mission_magitech_finale"],
  "est_time_mins": 20,
  "difficulty": 0.3
}
```

### 4.3 Dialogue Format

```json
{
  "id": "npc_mentor_first_meeting",
  "npc_id": "mentor_alex",
  "npc_name": "Alex Chen",
  "nodes": {
    "start": {
      "id": "start",
      "speaker": "npc",
      "text": "You must be the new Ghost. I've been watching your progress.",
      "emotion": "neutral",
      "next_node": "greeting"
    },
    "greeting": {
      "id": "greeting",
      "speaker": "player",
      "text": "Who are you?",
      "choices": [
        {
          "id": "choice_curious",
          "text": "How do you know my name?",
          "next_node": "reveal"
        },
        {
          "id": "choice_hostile",
          "text": "None of your business.",
          "next_node": "hostile_response"
        }
      ]
    },
    "reveal": {
      "id": "reveal",
      "speaker": "npc",
      "text": "I'm Alex. I used to work for MegaCorp. Now I help people like you.",
      "effects": ["trust_alex +1"],
      "next_node": "offer"
    },
    "offer": {
      "id": "offer",
      "speaker": "npc",
      "text": "I can teach you the basics. Interested?",
      "choices": [
        {
          "id": "choice_accept",
          "text": "Teach me.",
          "next_node": "tutorial_intro",
          "requirements": [],
          "xp_bonus": 50
        },
        {
          "id": "choice_decline",
          "text": "I'll figure it out myself.",
          "next_node": "decline_response"
        }
      ]
    },
    "tutorial_intro": {
      "id": "tutorial_intro",
      "speaker": "npc",
      "text": "Good. Let's start with the basics of code injection...",
      "next_node": "end"
    },
    "decline_response": {
      "id": "decline_response",
      "speaker": "npc",
      "text": "Your choice. But you'll need help eventually.",
      "next_node": "end"
    },
    "hostile_response": {
      "id": "hostile_response",
      "speaker": "npc",
      "text": "Fair enough. But remember who helped you when the Archon comes.",
      "effects": ["trust_alex -1"],
      "next_node": "end"
    },
    "end": {
      "id": "end",
      "speaker": "narrator",
      "text": "Alex walks away into the shadows.",
      "next_node": null
    }
  },
  "start_node": "start"
}
```

### 4.4 Save Format

```json
{
  "version": "1.0.0",
  "slot_id": 1,
  "player_name": "DataGhost_42",
  "created_at": "2026-07-11T10:30:00Z",
  "updated_at": "2026-07-11T12:45:00Z",
  "play_time_secs": 8100,
  
  "progress": {
    "player_id": "uuid-1234",
    "name": "DataGhost_42",
    "title": "Code Apprentice",
    "level": 5,
    "rank": "Apprentice",
    "xp": 2450,
    "credits": 320,
    "luck": 1.15,
    
    "completed_challenges": ["cyberpunk_01", "cyberpunk_02", "magitech_01"],
    "completed_missions": ["mission_cyberpunk_act1_ch1"],
    "unlocked_areas": ["neon_district", "underground_lab"],
    "unlocked_languages": ["python", "javascript"],
    
    "skill_tree": {
      "code_optimization": 2,
      "stealth_hacking": 1,
      "data_analysis": 3
    },
    
    "titles": ["Code Apprentice", "First Blood"],
    "badges": ["first_hack", "speedrunner"],
    "themes": ["dracula", "matrix"],
    
    "stats": {
      "challenges_attempted": 12,
      "challenges_completed": 8,
      "total_compiles": 45,
      "total_runtime_ms": 120000,
      "errors_encountered": 23,
      "hints_used": 5,
      "achievements_earned": 6,
      "category_stats": {
        "algorithms": { "attempted": 5, "completed": 3, "avg_time_ms": 900 },
        "data_structures": { "attempted": 4, "completed": 3, "avg_time_ms": 1200 }
      }
    },
    
    "created_at": "2026-07-11T10:30:00Z",
    "last_played": "2026-07-11T12:45:00Z",
    "total_play_time_secs": 8100
  },
  
  "mission_state": {
    "mission_cyberpunk_act1_ch2": {
      "mission_id": "mission_cyberpunk_act1_ch2",
      "player_id": "uuid-1234",
      "current_step": 2,
      "completed_objectives": ["obj_1"],
      "start_time": "2026-07-11T12:30:00Z",
      "state": { "has_data_chip": true }
    }
  },
  
  "settings": {
    "music_volume": 0.7,
    "sfx_volume": 0.8,
    "text_speed": 30,
    "auto_save": true,
    "theme": "dracula",
    "font_size": 14
  }
}
```

### 4.5 Language Plugin Format

```json
{
  "id": "python",
  "name": "Python 3.11",
  "extensions": [".py"],
  "compile_cmd": "python3 -m py_compile {file}",
  "run_cmd": "python3 {file}",
  "timeout_ms": 5000,
  "memory_bytes": 268435456,
  "docker_image": "python:3.11-slim",
  "env_vars": ["PYTHONUNBUFFERED=1"],
  "syntax_patterns": {
    "keywords": ["def", "class", "if", "else", "elif", "for", "while", "return", "import", "from", "as", "try", "except", "finally", "with", "yield", "lambda", "pass", "break", "continue", "and", "or", "not", "in", "is", "None", "True", "False"],
    "builtins": ["print", "range", "len", "int", "str", "float", "list", "dict", "set", "tuple", "bool", "type", "input", "open", "map", "filter", "zip", "enumerate", "sorted", "reversed", "any", "all", "min", "max", "sum", "abs", "round"]
  },
  "autocompletions": [
    { "trigger": "def ", "snippet": "def {name}({params}):\n    {body}" },
    { "trigger": "class ", "snippet": "class {name}:\n    def __init__(self{params}):\n        {body}" },
    { "trigger": "for ", "snippet": "for {item} in {iterable}:\n    {body}" },
    { "trigger": "if ", "snippet": "if {condition}:\n    {body}" },
    { "trigger": "while ", "snippet": "while {condition}:\n    {body}" },
    { "trigger": "try:", "snippet": "try:\n    {body}\nexcept {error}:\n    {handler}" }
  ]
}
```

---

## 5. Event System Design

### Event Flow

```
Player Action (click, keypress)
        │
        ▼
┌───────────────┐
│  Godot Client │
│  (Input)      │
└───────┬───────┘
        │ WebSocket JSON
        ▼
┌───────────────┐
│  WS Server    │
│  (Router)     │
└───────┬───────┘
        │
        ▼
┌───────────────┐     ┌───────────────┐
│  Event Bus    │────▶│  Subscribers  │
│  (Central)    │     │  (Modules)    │
└───────────────┘     └───────────────┘
        │
        ├──▶ MissionEngine    (mission events)
        ├──▶ ChallengeEngine  (challenge events)
        ├──▶ CompilerManager  (compile events)
        ├──▶ ExecutionEngine  (execute events)
        ├──▶ ProgressionSys   (xp/level events)
        ├──▶ DialogueEngine   (dialogue events)
        ├──▶ AnalyticsTracker (analytics events)
        └──▶ SaveManager      (save events)
```

### Event Routing Example

```go
// Server main.go
func main() {
    bus := eventbus.NewEventBus()
    
    // Subscribe modules
    missionEngine := missionengine.New(bus)
    challengeEngine := challengeengine.New(bus)
    compiler := compiler.New(bus)
    execution := executionengine.New(bus, compiler, sandbox)
    progression := progression.New(bus)
    dialogue := dialogue.New(bus)
    analytics := analytics.New(bus)
    saveManager := save.New(bus)
    
    // WebSocket handler
    handler := func(conn *websocket.Conn) {
        for {
            _, msg, _ := conn.ReadMessage()
            var evt struct{ Event string `json:"event"`; Payload string `json:"payload"` }
            json.Unmarshal(msg, &evt)
            
            // Route to event bus
            switch evt.Event {
            case "execute_code":
                bus.Publish(eventbus.Event{
                    Type:    eventbus.EventCodeSubmitted,
                    Payload: CodeSubmittedPayload{Code: evt.Payload, Language: "python"},
                })
            case "submit_answer":
                bus.Publish(eventbus.Event{
                    Type:    eventbus.EventCodeSubmitted,
                    Payload: AnswerSubmittedPayload{Answer: evt.Payload},
                })
            // ... more routing
            }
        }
    }
}
```

---

## 6. Folder Structure

```
challenge-to-you/
├── backend/                          # Go backend
│   ├── cmd/
│   │   └── server/                   # Main server entry point
│   │       └── main.go
│   ├── internal/
│   │   ├── eventbus/                 # Central event system
│   │   │   └── eventbus.go
│   │   ├── challengeengine/          # Challenge loading & management
│   │   │   ├── engine.go
│   │   │   ├── loader.go
│   │   │   └── validator.go
│   │   ├── compiler/                 # Language abstraction
│   │   │   ├── manager.go
│   │   │   ├── language.go
│   │   │   └── compilation.go
│   │   ├── sandbox/                  # Isolated execution
│   │   │   ├── sandbox.go
│   │   │   ├── docker.go
│   │   │   └── process.go
│   │   ├── executionengine/          # Test runner
│   │   │   ├── engine.go
│   │   │   └── validator.go
│   │   ├── missionengine/            # Mission flow
│   │   │   ├── engine.go
│   │   │   ├── loader.go
│   │   │   └── session.go
│   │   ├── dialogue/                 # NPC dialogue
│   │   │   ├── engine.go
│   │   │   ├── tree.go
│   │   │   └── session.go
│   │   ├── progression/              # XP, levels, ranks
│   │   │   ├── system.go
│   │   │   ├── levels.go
│   │   │   └── skills.go
│   │   ├── save/                     # Persistence
│   │   │   ├── manager.go
│   │   │   └── schema.go
│   │   ├── analytics/                # Tracking
│   │   │   ├── tracker.go
│   │   │   └── stats.go
│   │   ├── ai/                       # Ollama integration
│   │   │   ├── oracle.go
│   │   │   └── prompts.go
│   │   └── executor/                 # Language executors
│   │       ├── python/
│   │       │   └── executor.go
│   │       ├── go/
│   │       │   └── executor.go
│   │       ├── java/
│   │       │   └── executor.go
│   │       ├── javascript/
│   │       │   └── executor.go
│   │       ├── rust/
│   │       │   └── executor.go
│   │       ├── cpp/
│   │       │   └── executor.go
│   │       └── csharp/
│   │           └── executor.go
│   ├── data/                         # All game data
│   │   ├── challenges/               # Challenge JSON files
│   │   │   ├── magitech_tier1/
│   │   │   ├── magitech_tier2/
│   │   │   ├── cyberpunk_tier1/
│   │   │   ├── cyberpunk_tier2/
│   │   │   ├── cosmic_tier1/
│   │   │   └── cosmic_tier2/
│   │   ├── missions/                 # Mission JSON files
│   │   │   ├── magitech/
│   │   │   ├── cyberpunk/
│   │   │   └── cosmic/
│   │   ├── dialogue/                 # Dialogue JSON files
│   │   │   ├── npcs/
│   │   │   └── terminals/
│   │   ├── languages/                # Language plugin configs
│   │   │   ├── python.json
│   │   │   ├── go.json
│   │   │   ├── java.json
│   │   │   ├── javascript.json
│   │   │   ├── rust.json
│   │   │   ├── cpp.json
│   │   │   └── csharp.json
│   │   ├── skills/                   # Skill tree definitions
│   │   │   └── skills.json
│   │   ├── achievements/             # Achievement definitions
│   │   │   └── achievements.json
│   │   └── areas/                    # World area definitions
│   │       └── areas.json
│   ├── go.mod
│   └── go.sum
│
├── client/                           # Godot 4 project
│   ├── project.godot
│   ├── scenes/
│   │   ├── main.tscn                 # Main menu
│   │   ├── game.tscn                 # Game screen
│   │   ├── ide.tscn                  # IDE overlay
│   │   ├── world.tscn                # World exploration
│   │   └── dialogue.tscn             # Dialogue UI
│   ├── scripts/
│   │   ├── autoload/
│   │   │   ├── network_bridge.gd     # WebSocket
│   │   │   ├── audio_manager.gd      # Audio
│   │   │   └── save_manager.gd       # Persistence
│   │   ├── ide/
│   │   │   ├── code_editor.gd
│   │   │   ├── syntax_highlighter.gd
│   │   │   ├── autocomplete.gd
│   │   │   └── tabs.gd
│   │   ├── story/
│   │   │   ├── dialogue_manager.gd
│   │   │   ├── world_map.gd
│   │   │   └── mission_tracker.gd
│   │   └── ui/
│   │       ├── hud.gd
│   │       ├── menu.gd
│   │       └── settings.gd
│   ├── shaders/
│   │   └── crt_glitch.gdshader
│   ├── themes/
│   │   ├── dracula.tres
│   │   └── matrix.tres
│   └── assets/
│       ├── fonts/
│       ├── audio/
│       ├── backgrounds/
│       └── sprites/
│
├── docs/                             # Documentation
│   ├── ARCHITECTURE-PHASE1.md        # This file
│   ├── CHALLENGE-FORMAT.md           # Challenge schema guide
│   ├── MISSION-FORMAT.md             # Mission schema guide
│   ├── DIALOGUE-FORMAT.md            # Dialogue schema guide
│   ├── API.md                        # WebSocket protocol
│   └── CONTRIBUTING.md               # How to add content
│
└── README.md
```

---

## 7. Compiler/Sandbox Abstraction

### Language Executor Pipeline

```
Player Code
    │
    ▼
┌───────────────┐
│  Validate     │ ← Language-specific syntax check
│  (Optional)   │
└───────┬───────┘
        │
        ▼
┌───────────────┐
│  Compile      │ ← Language-specific compilation
│  (If needed)  │    C/C++: gcc, Java: javac, Rust: rustc
└───────┬───────┘    Python/JS/Go: skip (interpreted)
        │
        ▼
┌───────────────┐
│  Execute      │ ← Isolated in Docker/process
│  (Sandboxed)  │    Timeout, memory limits enforced
└───────┬───────┘
        │
        ▼
┌───────────────┐
│  Compare      │ ← Output vs expected
│  (Test Cases) │    Custom validators supported
└───────┬───────┘
        │
        ▼
   Result (pass/fail + details)
```

### Docker Sandbox Configuration

```yaml
# docker-compose.yml for sandbox
version: '3.8'
services:
  python-sandbox:
    image: python:3.11-slim
    mem_limit: 256m
    cpus: 0.5
    read_only: true
    tmpfs:
      - /tmp:size=64m
    security_opt:
      - no-new-privileges:true
    
  java-sandbox:
    image: eclipse-temurin:17-jre
    mem_limit: 512m
    cpus: 0.5
    
  go-sandbox:
    image: golang:1.21-alpine
    mem_limit: 256m
    cpus: 0.5
    
  rust-sandbox:
    image: rust:1.72-slim
    mem_limit: 512m
    cpus: 0.5
```

---

## 8. Progression & Save Architecture

### XP Curve

```go
// XP required per level
var LevelXP = map[int]int{
    1: 0,
    2: 100,
    3: 250,
    4: 500,
    5: 850,
    6: 1300,
    7: 1900,
    8: 2700,
    9: 3700,
    10: 5000,
    // ... continues with formula: xp[n] = xp[n-1] + 100 * n + 50 * (n-1)
}

// Rank titles
var RankTitles = []struct{
    MinXP int
    Title string
}{
    {0, "Newcomer"},
    {100, "Script Kiddie"},
    {500, "Code Apprentice"},
    {1500, "Debug Warrior"},
    {3000, "Algorithm Adept"},
    {5000, "System Architect"},
    {8000, "Security Specialist"},
    {12000, "Code Phantom"},
    {18000, "Master Hacker"},
    {25000, "Legendary Ghost"},
}
```

### Save Slot Architecture

```
saves/
├── slot_1/
│   ├── save.json          # Main save data
│   ├── analytics.json     # Player analytics
│   └── settings.json      # Player settings
├── slot_2/
│   └── ...
└── slot_3/
    └── ...
```

### Autosave Triggers

- After completing a challenge
- After completing a mission objective
- After making dialogue choices
- After upgrading skills
- Every 5 minutes of active play
- On game pause/exit

---

## 9. Dependency Graph

```
eventbus (no dependencies)
    │
    ├──▶ challengeengine (depends on: eventbus)
    │        │
    │        └──▶ executionengine (depends on: challengeengine, compiler, sandbox)
    │
    ├──▶ compiler (depends on: eventbus)
    │        │
    │        └──▶ executor/* (depends on: compiler)
    │
    ├──▶ sandbox (depends on: eventbus)
    │
    ├──▶ missionengine (depends on: eventbus, challengeengine)
    │
    ├──▶ dialogue (depends on: eventbus)
    │
    ├──▶ progression (depends on: eventbus)
    │
    ├──▶ save (depends on: eventbus, progression)
    │
    ├──▶ analytics (depends on: eventbus)
    │
    └──▶ ai (depends on: eventbus)
```

---

## 10. WebSocket Protocol (Extended)

### Client → Server Messages

```json
// Execute code in a language
{
  "event": "execute_code",
  "payload": {
    "language": "python",
    "code": "def solve(arr):\n    return sorted(arr)",
    "challenge_id": "cyberpunk_01_autodoc"
  }
}

// Submit answer for recognize challenges
{
  "event": "submit_answer",
  "payload": {
    "answer": "42",
    "challenge_id": "cyberpunk_01_autodoc"
  }
}

// Trigger state machine event
{
  "event": "trigger",
  "payload": {
    "event_id": "inject_adrenaline",
    "challenge_id": "cyberpunk_01_autodoc"
  }
}

// Request hint
{
  "event": "request_hint",
  "payload": {
    "challenge_id": "cyberpunk_01_autodoc",
    "hint_level": 1
  }
}

// Start mission
{
  "event": "start_mission",
  "payload": {
    "mission_id": "mission_cyberpunk_act1_ch1"
  }
}

// Dialogue choice
{
  "event": "dialogue_choice",
  "payload": {
    "session_id": "uuid",
    "choice_id": "choice_accept"
  }
}

// Save game
{
  "event": "save_game",
  "payload": {
    "slot_id": 1
  }
}

// Load game
{
  "event": "load_game",
  "payload": {
    "slot_id": 1
  }
}
```

### Server → Client Messages

```json
// Challenge snapshot (existing)
{
  "type": "snapshot",
  "challenge_id": "...",
  "paradigm": "...",
  "title": "...",
  "description": "...",
  "skill_type": "...",
  "modules": [...],
  "state": {...},
  "vigilance": 0.5,
  "triggerable": [...],
  "level_complete": false,
  "message": "..."
}

// Code execution result
{
  "type": "execution_result",
  "success": true,
  "output": "All tests passed",
  "test_results": [
    { "id": "test_1", "passed": true },
    { "id": "test_2", "passed": true }
  ],
  "score": 1.0,
  "xp_earned": 100
}

// Compilation result
{
  "type": "compilation_result",
  "success": false,
  "errors": [
    { "line": 3, "column": 5, "message": "SyntaxError: unexpected indent" }
  ]
}

// Mission update
{
  "type": "mission_update",
  "mission_id": "...",
  "completed_objectives": ["obj_1"],
  "current_step": 2,
  "message": "Objective completed!"
}

// Dialogue
{
  "type": "dialogue",
  "session_id": "...",
  "speaker": "Alex Chen",
  "text": "Welcome, Ghost.",
  "choices": [
    { "id": "choice_1", "text": "Who are you?" },
    { "id": "choice_2", "text": "I need help." }
  ]
}

// Progression update
{
  "type": "progression_update",
  "level": 5,
  "xp": 2450,
  "xp_to_next": 500,
  "rank": "Apprentice",
  "level_up": false
}

// Achievement unlocked
{
  "type": "achievement_unlocked",
  "achievement_id": "first_hack",
  "title": "First Blood",
  "description": "Complete your first challenge",
  "xp_bonus": 50
}

// Archon transmission (existing)
{
  "type": "archon_transmission",
  "message": "..."
}
```

---

## 11. Implementation Priorities

### Phase 1A: Core Engine (Week 1-2)
1. Event bus implementation
2. Challenge engine (load, validate, manage)
3. Compiler manager + Python executor
4. Sandbox (Docker or process-based)
5. Execution engine (run tests, compare output)
6. Extended WebSocket protocol

### Phase 1B: Content Pipeline (Week 3-4)
1. Challenge format standardization
2. 100+ challenges across all categories
3. Mission format + 10 missions
4. Dialogue system + 20 dialogue trees
5. Skill tree definitions

### Phase 1C: Client Enhancement (Week 5-6)
1. In-game IDE (syntax highlighting, autocomplete)
2. Mission tracker UI
3. Dialogue UI
4. World map/exploration
5. Save/load UI

### Phase 1D: Polish (Week 7-8)
1. Audio system
2. Animations & transitions
3. Tutorial flow
4. Balance testing
5. Performance optimization

---

*Last updated: 2026-07-11*
*Status: Architecture Design Complete*
