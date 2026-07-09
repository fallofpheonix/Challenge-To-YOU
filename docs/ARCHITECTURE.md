# Architecture Document: Challenge To YOU

## System Overview

**Challenge To YOU** is a desktop-first roguelike hacking game built with:
- **Godot 4** (Frontend) — UI, code editor, visual themes
- **Go 1.26+** (Backend) — Code execution, procedural generation, AI analysis
- **WASM** (Sandbox) — Secure player code execution
- **Ollama + Llama 3** (AI) — Code style analysis, passcode generation

---

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Godot 4 Client (Desktop)                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ Code Editor  │  │ Terminal UI  │  │ Era-Specific     │  │
│  │ (Syntax HL)  │  │ (Output)     │  │ Visual Themes    │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────────────┘  │
│         │                 │                                  │
│  ┌──────▼─────────────────▼──────────────────────────────┐  │
│  │              GDExtension Bridge (Native)               │  │
│  └──────┬────────────────────────────────────────────────┘  │
└─────────┼───────────────────────────────────────────────────┘
          │
┌─────────▼───────────────────────────────────────────────────┐
│                    Go Backend (Shared Library)               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ WASM Sandbox │  │ Procedural   │  │ AI/AST Analyzer  │  │
│  │ (Execution)  │  │ Generation   │  │ (Code Style)     │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Passcode Generation Engine               │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## Component Details

### 1. Godot 4 Client

#### 1.1 Code Editor
- **Purpose**: Player writes code here
- **Features**:
  - Syntax highlighting (era-specific)
  - Auto-completion (optional)
  - Error indicators
  - Line numbers
- **File**: `client/scripts/editor.gd`

#### 1.2 Terminal UI
- **Purpose**: Shows code output, errors, system messages
- **Features**:
  - Scrolling output buffer
  - Color-coded messages (error, warning, info)
  - Real-time updates during execution
- **File**: `client/scripts/terminal.gd`

#### 1.3 Era Themes
- **Purpose**: Visual theming per era
- **Eras**:
  - **Magitech**: Dark backgrounds, mystical fonts, rune symbols
  - **Cyberpunk**: Neon colors, terminal fonts, glitch effects
- **File**: `client/themes/`

#### 1.4 GDExtension Bridge
- **Purpose**: Native communication between Godot and Go
- **Protocol**: Direct function calls (no WebSocket for desktop)
- **File**: `client/scripts/bridge.gd`

---

### 2. Go Backend

#### 2.1 WASM Sandbox
- **Purpose**: Secure execution of player code
- **Technology**: Extism (WASM framework)
- **Features**:
  - Isolated VM instances per execution
  - Resource limits (CPU, memory, steps)
  - Timeout enforcement (10s max)
  - No network access
- **File**: `backend/internal/sandbox/`

```go
// Example: Sandbox execution
type Sandbox struct {
    instance wasmer.Instance
    limits   ResourceLimits
}

func (s *Sandbox) Execute(code string) (output string, err error) {
    // 1. Compile code to WASM module
    // 2. Instantiate VM with resource limits
    // 3. Execute with timeout
    // 4. Capture output/errors
    // 5. Tear down instance
}
```

#### 2.2 Procedural Generation Engine
- **Purpose**: Create random challenges from modular components
- **Algorithm**: Seed-based RNG + modular stitching
- **File**: `backend/internal/generator/`

```go
// Example: Challenge generation
type Generator struct {
    rng    *rand.Rand
    modules []CodeModule
}

func (g *Generator) Generate(seed int64, luck float64) Challenge {
    // 1. Seed RNG
    // 2. Select 15-20 modules based on luck
    // 3. Stitch modules together
    // 4. Ensure at least one glitch exists
    // 5. Return challenge with passcode
}
```

#### 2.3 AI/AST Analyzer
- **Purpose**: Analyze code style, detect patterns
- **Technology**: 
  - Primary: Ollama + Llama 3 (local)
  - Fallback: go/ast parsing (deterministic)
- **File**: `backend/internal/analyzer/`

```go
// Example: Code analysis
type Analyzer struct {
    ollama *ollama.Client
    ast    *ast.Parser
}

func (a *Analyzer) Analyze(code string) CodeProfile {
    // 1. Parse AST
    // 2. Extract features (complexity, patterns)
    // 3. If Ollama available, get style analysis
    // 4. Return profile with passcode seed
}
```

#### 2.4 Passcode Generation Engine
- **Purpose**: Generate passcodes from code interactions
- **Algorithm**: Multi-layer glitch detection + hash generation
- **File**: `backend/internal/passcode/`

```go
// Example: Passcode generation
type PasscodeEngine struct {
    analyzer *Analyzer
}

func (p *PasscodeEngine) Generate(challenge Challenge, solution string) string {
    // 1. Execute solution in sandbox
    // 2. Detect glitches/loopholes
    // 3. Analyze code style
    // 4. Combine factors into seed
    // 5. Generate deterministic passcode
}
```

---

### 3. WASM Sandbox Details

#### 3.1 Security Model
- **Isolation**: Each execution runs in separate VM instance
- **Resource Limits**:
  - CPU time: 5 seconds max
  - Memory: 64 MB max
  - Execution steps: 10,000 max
  - Network: Disabled
  - File system: Read-only (code only)

#### 3.2 Execution Flow
```
1. Player writes code in Godot editor
2. Code sent to Go backend via GDExtension
3. Backend compiles code to WASM module
4. Backend instantiates sandbox VM
5. VM executes with resource limits
6. Output/errors captured
7. Results sent back to Godot
8. VM torn down
```

#### 3.3 Multi-Language Support
- **Era 1 (Magitech)**: Custom DSL (runes/incantations)
  - Custom parser in Go
  - Compiles to WASM or interprets directly
- **Era 2 (Cyberpunk)**: Real languages (Python/JS)
  - Use existing WASM compilers (Python WASM, JS WASM)
  - Or interpret via Go bindings

---

### 4. Procedural Generation System

#### 4.1 Modular Code Segments
Each "junk code block" is a module with:
- **Type**: Input processing, core logic, output generation
- **Complexity**: Low, medium, high
- **Glitch Potential**: Can create exploit when combined
- **Era**: Magitech, Cyberpunk, or universal

Example modules:
```go
type CodeModule struct {
    ID            string
    Type          ModuleType  // INPUT, CORE, OUTPUT
    Complexity    int         // 1-10
    GlitchWeight  float64     // 0.0-1.0
    Era           Era         // MAGICTECH, CYBERPUNK
    Code          string      // The actual code
    Dependencies  []string    // Other module IDs
}
```

#### 4.2 Seed-Based Generation
```go
func GenerateChallenge(seed int64, luck float64) Challenge {
    rng := rand.New(rand.NewSource(seed))
    
    // Select modules based on luck
    // High luck = easier modules, more glitches
    // Low luck = harder modules, fewer glitches
    
    modules := selectModules(rng, luck, 15, 20)
    
    // Stitch modules together
    code := stitchModules(modules)
    
    // Ensure at least one glitch exists
    ensureGlitchExists(code, modules)
    
    return Challenge{
        Seed:    seed,
        Code:    code,
        Modules: modules,
        Luck:    luck,
    }
}
```

#### 4.3 Luck Mechanic
| Luck Value | Effect |
|------------|--------|
| 0.0-0.3 | Hard: Obfuscated code, aggressive AI monitoring |
| 0.3-0.7 | Medium: Balanced challenge |
| 0.7-1.0 | Easy: Obvious flaws, perfect glitch alignment |

---

### 5. Passcode System

#### 5.1 Passcode Sources
Passcodes emerge from:
1. **Error Logs**: Hidden in stack traces
2. **Memory Leaks**: Patterns in allocation failures
3. **CPU Fluctuations**: Timing-based exploits
4. **Glitch Interactions**: Frankenstein code combinations

#### 5.2 Generation Algorithm
```go
func GeneratePasscode(solution string, style CodeProfile) string {
    // Combine multiple factors
    factors := []interface{}{
        solution,           // The code itself
        style.Complexity,   // How complex
        style.Patterns,     // What patterns used
        style.Glitches,     // Glitches detected
        time.Now().UnixNano(), // Timestamp (for uniqueness)
    }
    
    // Hash factors into passcode
    hash := sha256.New()
    for _, f := range factors {
        fmt.Fprintf(hash, "%v", f)
    }
    
    return hex.EncodeToString(hash.Sum(nil))[:16] // 16-char passcode
}
```

#### 5.3 Passcode Verification
- Passcodes are deterministic given same inputs
- Different approaches produce different passcodes
- No "correct" answer — any valid passcode works
- Passcodes can be saved and shared

---

### 6. AI Integration

#### 6.1 Ollama + Llama 3
- **Purpose**: Analyze code style, generate passcodes
- **Setup**: Local model, no API costs
- **Usage**:
  - Analyze player's coding approach
  - Detect if player is "thinking like a hacker"
  - Generate personalized passcodes

#### 6.2 AST Parsing Fallback
- **Purpose**: Deterministic analysis when AI unavailable
- **Technology**: go/ast package
- **Features**:
  - Detect code patterns
  - Measure complexity
  - Identify potential glitches

#### 6.3 Hybrid Approach
```go
func AnalyzeCode(code string) CodeProfile {
    // 1. Always run AST analysis (fast, deterministic)
    profile := analyzeAST(code)
    
    // 2. If Ollama available, enhance with AI analysis
    if ollamaAvailable {
        aiProfile := analyzeWithAI(code)
        profile = mergeProfiles(profile, aiProfile)
    }
    
    return profile
}
```

---

## Data Flow

### 1. Challenge Generation Flow
```
Seed + Luck
    ↓
Generator (Go)
    ↓
Select Modules
    ↓
Stitch Code
    ↓
Ensure Glitch
    ↓
Challenge Object
    ↓
Send to Godot
    ↓
Display in Editor
```

### 2. Solution Execution Flow
```
Player Code (Godot)
    ↓
GDExtension Bridge
    ↓
Go Backend
    ↓
WASM Sandbox
    ↓
Execute with Limits
    ↓
Capture Output/Errors
    ↓
Analyze Code (AI/AST)
    ↓
Detect Glitches
    ↓
Generate Passcode
    ↓
Return to Godot
    ↓
Display in Terminal
```

### 3. Era Transition Flow
```
Complete Era 1 Challenges
    ↓
Earn Era 2 Unlock
    ↓
Switch Theme (Godot)
    ↓
Load New Modules (Go)
    ↓
Change Code Type (DSL → Real)
    ↓
New Challenges Available
```

---

## File Structure

```
challenge-to-you/
├── backend/
│   ├── cmd/
│   │   └── sandbox/
│   │       └── main.go              # Entry point
│   ├── internal/
│   │   ├── generator/
│   │   │   ├── generator.go         # Main generator
│   │   │   ├── modules.go           # Code modules
│   │   │   ├── luck.go              # Luck mechanics
│   │   │   └── stitch.go            # Module stitching
│   │   ├── sandbox/
│   │   │   ├── sandbox.go           # WASM wrapper
│   │   │   ├── limits.go            # Resource limits
│   │   │   └── executor.go          # Execution logic
│   │   ├── analyzer/
│   │   │   ├── analyzer.go          # Main analyzer
│   │   │   ├── ast.go               # AST parsing
│   │   │   ├── style.go             # Style detection
│   │   │   └── ollama.go            # AI integration
│   │   ├── passcode/
│   │   │   ├── engine.go            # Passcode generator
│   │   │   ├── glitch.go            # Glitch detection
│   │   │   └── hash.go              # Hashing utilities
│   │   └── narrative/
│   │       ├── magitech.go          # Era 1 text
│   │       └── cyberpunk.go         # Era 2 text
│   ├── pscript/                     # Magitech DSL
│   │   ├── lexer.go
│   │   ├── parser.go
│   │   └── ast.go
│   ├── go.mod
│   └── go.sum
├── client/
│   ├── scenes/
│   │   ├── main.tscn                # Main menu
│   │   ├── editor.tscn              # Code editor
│   │   ├── terminal.tscn            # Output terminal
│   │   └── eras/
│   │       ├── magitech.tscn
│   │       └── cyberpunk.tscn
│   ├── scripts/
│   │   ├── main.gd                  # Main controller
│   │   ├── editor.gd                # Code editor logic
│   │   ├── terminal.gd              # Terminal output
│   │   └── bridge.gd                # GDExtension bridge
│   ├── themes/
│   │   ├── magitech/
│   │   │   ├── theme.tres
│   │   │   └── colors.gd
│   │   └── cyberpunk/
│   │       ├── theme.tres
│   │       └── colors.gd
│   ├── addons/
│   │   └── gdextension/
│   │       └── libchallenge.so      # Go shared library
│   └── project.godot
├── docs/
│   ├── ARCHITECTURE.md              # This file
│   ├── GAME-DESIGN.md               # Game mechanics
│   ├── API.md                       # Interface specs
│   └── CHALLENGE-TO-YOU-PLAN.md     # Implementation plan
├── tools/
│   ├── build.sh                     # Build script
│   └── package.sh                   # Packaging script
└── README.md
```

---

## API Interface

### Go ↔ Godot Communication

#### GDExtension Functions
```go
// Exported to Godot
func ExecuteCode(code string, era string) (output string, passcode string, err string)
func GenerateChallenge(seed int64, luck float64) (challengeJSON string)
func AnalyzeCode(code string) (profileJSON string)
func GetEras() (erasJSON string)
func GetModes() (modesJSON string)
```

#### Data Structures
```go
type Challenge struct {
    ID        string       `json:"id"`
    Seed      int64        `json:"seed"`
    Code      string       `json:"code"`
    Modules   []CodeModule `json:"modules"`
    Era       string       `json:"era"`
    Mode      string       `json:"mode"`
    Luck      float64      `json:"luck"`
    Passcode  string       `json:"passcode"`
}

type ExecutionResult struct {
    Output    string `json:"output"`
    Passcode  string `json:"passcode"`
    Error     string `json:"error"`
    Glitches  []string `json:"glitches"`
    Style     CodeProfile `json:"style"`
}
```

---

## Security Considerations

### 1. WASM Isolation
- Player code cannot access host system
- No network, file system, or process access
- Resource limits prevent DoS attacks

### 2. Timeout Enforcement
- Maximum 5 seconds execution time
- Automatic termination if exceeded
- Graceful error handling

### 3. Input Validation
- Code size limits (10 MB max)
- Syntax validation before execution
- Rate limiting on submissions

### 4. Output Sanitization
- Error messages sanitized (no stack traces)
- Output truncated if too large
- No sensitive data exposure

---

## Performance Considerations

### 1. WASM Instantiation
- Cold start: ~100ms (acceptable for desktop)
- Warm start: ~10ms (reuse VM instances)
- Memory: ~64MB per instance

### 2. Procedural Generation
- Seed-based: O(1) lookup
- Module selection: O(n) where n = module count
- Stitching: O(m) where m = modules per challenge

### 3. AI Analysis
- AST parsing: ~10ms
- Ollama analysis: ~500ms (local)
- Combined: ~500ms max

### 4. Passcode Generation
- Hash computation: ~1ms
- Deterministic: Same inputs → same output
- Unique: Different approaches → different passcodes

---

## Testing Strategy

### Unit Tests
| Component | Test Focus |
|-----------|------------|
| Generator | Seed reproducibility, module validity |
| Sandbox | Execution isolation, timeout enforcement |
| Passcode | Hash consistency, uniqueness |
| Analyzer | AST parsing, style detection |
| Luck System | Distribution fairness |

### Integration Tests
| Test | Description |
|------|-------------|
| End-to-End | Code → Execute → Passcode |
| Mode Switching | Architect ↔ Ghost ↔ Saboteur |
| Era Transition | Magitech → Cyberpunk |
| Procedural | 1000 seed runs, all valid |

### Playtesting
| Phase | Focus |
|-------|-------|
| Week 3 | Internal testing, fix bugs |
| Week 4 | Community alpha, gather feedback |
| Post-Launch | Iterate on player data |

---

*Last updated: 2026-07-10*
