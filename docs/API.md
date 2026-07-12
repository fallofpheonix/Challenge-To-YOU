# API Document: Challenge To YOU

## Overview

This document specifies the interface between the **Go backend** and **Godot frontend** for Challenge To YOU.

> ⚠️ **Status correction (2026-07-12):** the primary/fallback below are **inverted
> relative to the implementation**. The live transport is **WebSocket** (the GDExtension
> native path was never built). The authoritative wire protocol is
> [`ARCHITECTURE-PHASE1.md`](ARCHITECTURE-PHASE1.md) §10; the GDExtension function
> signatures in this document describe the *original* design and are retained for context.

---

## Communication Protocol

### Live transport: WebSocket
- **Type**: JSON messages over WebSocket
- **Port**: 127.0.0.1:8080
- **Usage**: the actual client↔backend link (`internal/server`, `client/scripts/network_bridge.gd`)

### Original design (not built): GDExtension native
- **Type**: Direct function calls
- **Language**: Go shared library → Godot GDExtension
- **Latency**: ~1ms (native calls)
- **Status**: superseded by WebSocket; signatures below are historical

---

## GDExtension Functions

### Initialize

```go
// Exported to Godot
func Initialize() error
```

**Description**: Initialize the Go backend, load modules, set up AI.

**Returns**: Error if initialization fails

---

### ExecuteCode

```go
// Exported to Godot
func ExecuteCode(code string, era string, mode string) ExecutionResult
```

**Description**: Execute player's code in WASM sandbox.

**Parameters**:
| Name | Type | Description |
|------|------|-------------|
| `code` | string | Player's source code |
| `era` | string | Current era ("magitech" or "cyberpunk") |
| `mode` | string | Current mode ("architect", "ghost", or "saboteur") |

**Returns**:
```go
type ExecutionResult struct {
    Output    string      `json:"output"`    // stdout/stderr output
    Passcode  string      `json:"passcode"`  // Generated passcode
    Error     string      `json:"error"`     // Error message (if any)
    Glitches  []string    `json:"glitches"`  // Detected glitches
    Style     CodeProfile `json:"style"`     // Code analysis
    Duration  int64       `json:"duration"`  // Execution time (ms)
    Steps     int         `json:"steps"`     // Execution steps used
}
```

**Example**:
```go
result := ExecuteCode(`
RUNE fire = IGNITE(power: 100)
RUNE water = FLOW(direction: NORTH)
EFFECT { fire.COMBINE(water) }
`, "magitech", "architect")

// result.Output = "Rune combination successful"
// result.Passcode = "a1b2c3d4e5f6g7h8"
// result.Glitches = ["rune_interference"]
```

---

### GenerateChallenge

```go
// Exported to Godot
func GenerateChallenge(seed int64, luck float64, era string) Challenge
```

**Description**: Generate a procedural challenge.

**Parameters**:
| Name | Type | Description |
|------|------|-------------|
| `seed` | int64 | RNG seed for reproducibility |
| `luck` | float64 | Luck value (0.0-1.0) |
| `era` | string | Target era |

**Returns**:
```go
type Challenge struct {
    ID        string       `json:"id"`        // Unique challenge ID
    Seed      int64        `json:"seed"`      // Seed used
    Code      string       `json:"code"`      // Generated junk code
    Modules   []CodeModule `json:"modules"`   // Individual modules
    Era       string       `json:"era"`       // Era name
    Mode      string       `json:"mode"`      // Suggested mode
    Luck      float64      `json:"luck"`      // Luck value used
    Passcode  string       `json:"passcode"`  // Expected passcode (hint)
    Hints     []string     `json:"hints"`     // Help text
}

type CodeModule struct {
    ID           string `json:"id"`
    Type         string `json:"type"`         // INPUT, CORE, OUTPUT, DECOY, EXPLOIT
    Complexity   int    `json:"complexity"`   // 1-10
    GlitchWeight float64 `json:"glitch_weight"` // 0.0-1.0
    Code         string `json:"code"`         // Module source code
    Description  string `json:"description"`  // Human-readable description
}
```

---

### AnalyzeCode

```go
// Exported to Godot
func AnalyzeCode(code string) CodeProfile
```

**Description**: Analyze code style and detect patterns.

**Parameters**:
| Name | Type | Description |
|------|------|-------------|
| `code` | string | Source code to analyze |

**Returns**:
```go
type CodeProfile struct {
    Complexity    int            `json:"complexity"`    // 1-10
    Patterns      []string       `json:"patterns"`     // Detected patterns
    Glitches      []string       `json:"glitches"`     // Potential glitches
    Style         string         `json:"style"`        // Coding style
    Readability   float64        `json:"readability"`  // 0.0-1.0
    Efficiency    float64        `json:"efficiency"`   // 0.0-1.0
    StealthScore  float64        `json:"stealth_score"` // For Ghost mode
    ChaosScore    float64        `json:"chaos_score"`  // For Saboteur mode
}
```

---

### GetEras

```go
// Exported to Godot
func GetEras() []Era
```

**Description**: Get all available eras.

**Returns**:
```go
type Era struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Theme       string `json:"theme"`
    CodeType    string `json:"code_type"`    // DSL, PYTHON, JS
    Unlocked    bool   `json:"unlocked"`
    Description string `json:"description"`
}
```

---

### GetModes

```go
// Exported to Godot
func GetModes() []Mode
```

**Description**: Get all available gameplay modes.

**Returns**:
```go
type Mode struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Role        string `json:"role"`        // BUILDER, HACKER, AGENT
    Unlocked    bool   `json:"unlocked"`
    Description string `json:"description"`
    Objectives  []string `json:"objectives"` // Mode-specific goals
}
```

---

### VerifyPasscode

```go
// Exported to Godot
func VerifyPasscode(passcode string, challengeID string) bool
```

**Description**: Verify if a passcode is valid for a challenge.

**Parameters**:
| Name | Type | Description |
|------|------|-------------|
| `passcode` | string | Passcode to verify |
| `challengeID` | string | Challenge ID |

**Returns**: `true` if valid, `false` otherwise

---

### GetPlayerProgress

```go
// Exported to Godot
func GetPlayerProgress() PlayerProgress
```

**Description**: Get player's current progress.

**Returns**:
```go
type PlayerProgress struct {
    Level           int              `json:"level"`
    XP              int              `json:"xp"`
    ErasUnlocked    []string         `json:"eras_unlocked"`
    ModesUnlocked   []string         `json:"modes_unlocked"`
    ChallengesCompleted int          `json:"challenges_completed"`
    GlitchesDiscovered []string      `json:"glitches_discovered"`
    PasscodeCollection []string      `json:"passcode_collection"`
    CurrentLuck     float64          `json:"current_luck"`
}
```

---

### SaveProgress

```go
// Exported to Godot
func SaveProgress(progress PlayerProgress) error
```

**Description**: Save player progress to disk.

**Parameters**:
| Name | Type | Description |
|------|------|-------------|
| `progress` | PlayerProgress | Progress to save |

**Returns**: Error if save fails

---

## Data Structures

### Core Types

```go
// Challenge represents a procedural challenge
type Challenge struct {
    ID        string       `json:"id"`
    Seed      int64        `json:"seed"`
    Code      string       `json:"code"`
    Modules   []CodeModule `json:"modules"`
    Era       string       `json:"era"`
    Mode      string       `json:"mode"`
    Luck      float64      `json:"luck"`
    Passcode  string       `json:"passcode"`
    Hints     []string     `json:"hints"`
}

// CodeModule represents a single code segment
type CodeModule struct {
    ID           string  `json:"id"`
    Type         string  `json:"type"`
    Complexity   int     `json:"complexity"`
    GlitchWeight float64 `json:"glitch_weight"`
    Code         string  `json:"code"`
    Description  string  `json:"description"`
}

// ExecutionResult represents code execution output
type ExecutionResult struct {
    Output    string      `json:"output"`
    Passcode  string      `json:"passcode"`
    Error     string      `json:"error"`
    Glitches  []string    `json:"glitches"`
    Style     CodeProfile `json:"style"`
    Duration  int64       `json:"duration"`
    Steps     int         `json:"steps"`
}

// CodeProfile represents code analysis results
type CodeProfile struct {
    Complexity   int      `json:"complexity"`
    Patterns     []string `json:"patterns"`
    Glitches     []string `json:"glitches"`
    Style        string   `json:"style"`
    Readability  float64  `json:"readability"`
    Efficiency   float64  `json:"efficiency"`
    StealthScore float64  `json:"stealth_score"`
    ChaosScore   float64  `json:"chaos_score"`
}

// Era represents a game era
type Era struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Theme       string `json:"theme"`
    CodeType    string `json:"code_type"`
    Unlocked    bool   `json:"unlocked"`
    Description string `json:"description"`
}

// Mode represents a gameplay mode
type Mode struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Role        string   `json:"role"`
    Unlocked    bool     `json:"unlocked"`
    Description string   `json:"description"`
    Objectives  []string `json:"objectives"`
}

// PlayerProgress represents player state
type PlayerProgress struct {
    Level                int        `json:"level"`
    XP                   int        `json:"xp"`
    ErasUnlocked         []string   `json:"eras_unlocked"`
    ModesUnlocked        []string   `json:"modes_unlocked"`
    ChallengesCompleted  int        `json:"challenges_completed"`
    GlitchesDiscovered   []string   `json:"glitches_discovered"`
    PasscodeCollection   []string   `json:"passcode_collection"`
    CurrentLuck          float64    `json:"current_luck"`
}
```

---

## Error Handling

### Error Types

```go
type ErrorCode int

const (
    ErrNone ErrorCode = iota
    ErrSyntax
    ErrRuntime
    ErrTimeout
    ErrMemoryLimit
    ErrStepLimit
    ErrNetworkAccess
    ErrFileAccess
    ErrInvalidEra
    ErrInvalidMode
    ErrChallengeNotFound
    ErrPasscodeInvalid
    ErrSaveFailed
    ErrLoadFailed
)

type Error struct {
    Code    ErrorCode `json:"code"`
    Message string    `json:"message"`
    Line    int       `json:"line,omitempty"`
    Column  int       `json:"column,omitempty"`
}
```

### Error Messages

| Code | Message | Example |
|------|---------|---------|
| `ErrSyntax` | Syntax error in code | `Unexpected token at line 5` |
| `ErrRuntime` | Runtime error | `Division by zero` |
| `ErrTimeout` | Execution exceeded time limit | `Execution timed out after 5s` |
| `ErrMemoryLimit` | Memory usage exceeded | `Memory limit exceeded (64MB)` |
| `ErrStepLimit` | Execution steps exceeded | `Step limit exceeded (10000)` |
| `ErrNetworkAccess` | Network access attempted | `Network access not allowed` |
| `ErrFileAccess` | File access attempted | `File access not allowed` |

---

## WebSocket Protocol (live transport)

### Connection

```
ws://127.0.0.1:8080/telemetry
```

### Message Format

```json
{
    "packet_type": "EXECUTE_CODE",
    "tick": 0,
    "payload": {
        "code": "...",
        "era": "magitech",
        "mode": "architect"
    }
}
```

### Packet Types

| Type | Direction | Description |
|------|-----------|-------------|
| `EXECUTE_CODE` | Client → Server | Execute player code |
| `EXECUTION_RESULT` | Server → Client | Code execution result |
| `GENERATE_CHALLENGE` | Client → Server | Generate new challenge |
| `CHALLENGE_DATA` | Server → Client | Generated challenge |
| `ANALYZE_CODE` | Client → Server | Analyze code style |
| `ANALYSIS_RESULT` | Server → Client | Code analysis |
| `PLAYER_PROGRESS` | Server → Client | Progress update |
| `ERROR` | Server → Client | Error message |

---

## Usage Examples

### Godot (GDScript)

```gdscript
# Initialize backend
var backend = ChallengeBackend.new()
backend.initialize()

# Generate challenge
var challenge = backend.generate_challenge(12345, 0.7, "magitech")
print(challenge.code)  # Shows generated junk code

# Execute player code
var result = backend.execute_code(player_code, "magitech", "architect")
if result.error == "":
    print("Passcode: ", result.passcode)
    for glitch in result.glitches:
        print("Glitch found: ", glitch)
else:
    print("Error: ", result.error)

# Verify passcode
var valid = backend.verify_passcode("a1b2c3d4e5f6g7h8", challenge.id)
if valid:
    print("Challenge complete!")
```

### Go (Internal)

```go
// Generate challenge
challenge := GenerateChallenge(12345, 0.7, "magitech")

// Execute code
result := ExecuteCode(challenge.Code, "magitech", "architect")

// Check for glitches
if len(result.Glitches) > 0 {
    fmt.Printf("Found glitches: %v\n", result.Glitches)
}

// Verify passcode
if VerifyPasscode(result.Passcode, challenge.ID) {
    fmt.Println("Passcode valid!")
}
```

---

*Last updated: 2026-07-10*
