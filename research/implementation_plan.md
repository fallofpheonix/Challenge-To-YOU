# Implementation Plan

Phased roadmap for building **Challenge To YOU** based on research findings.

---

## Phase 0: Foundation (Week 0-1) ✅ **IN PROGRESS**

### VM Module (Standalone Go Package)
```
backend/vm/
├── internal/
│   ├── lexer/       ✅ Lexer for Pscript DSL
│   ├── parser/      ✅ Pratt parser with AST
│   ├── compiler/    ✅ Bytecode compiler
│   ├── bytecode/    ✅ Instruction set
│   ├── scheduler/   ✅ Stack VM with limits
│   └── limits/      ✅ Instruction/time/memory limits
├── cmd/vm/          ✅ CLI harness
└── go.mod           ✅ Module definition
```

**Completed**:
- Lexer with Pscript keywords (rune, bind, channel, if/then/else, fn, while)
- Parser with precedence climbing
- Compiler with symbol table, closures, jumps
- Bytecode VM with stack, frames, globals
- Limits: max instructions, time, stack depth
- CLI: `vm -timeout 5s -max-steps 1M -ast -bytecode file.psi`

**Next**:
- Builtin function registry (log, emit, rune, sleep, len, etc.)
- Rune handler callback for game integration
- Comprehensive test suite
- Benchmark suite

---

## Phase 1: Godot + Go Integration (Week 1-2)

### GDExtension Setup
```
backend/
├── cmd/sandbox/           # GDExtension entry point
│   └── main.go
├── internal/bridge/       # Godot ↔ Go bridge
│   ├── bridge.go          # Class registration
│   ├── vm_wrapper.go      # Godot VM class
│   └── event_bus.go       # Signal bridge
└── Makefile               # Cross-platform build
```

### Build Pipeline
```makefile
# Makefile
build:
	go build -buildmode=c-shared -o ../client/addons/godot-go/libchallenge.so ./cmd/sandbox

build-mac:
	GOOS=darwin GOARCH=arm64 go build -buildmode=c-shared -o ../client/addons/godot-go/libchallenge.dylib ./cmd/sandbox

build-win:
	GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -o ../client/addons/godot-go/challenge.dll ./cmd/sandbox
```

### Godot Project Structure
```
client/
├── project.godot
├── addons/
│   └── godot-go/
│       ├── challenge.gdextension
│       └── libchallenge.so/dylib/dll
├── scenes/
│   ├── main.tscn           # Main menu
│   ├── editor.tscn         # Code editor
│   ├── terminal.tscn       # Output terminal
│   ├── puzzle_map.tscn     # Dependency graph
│   └── era/                # Era-specific themes
├── scripts/
│   ├── core/
│   │   ├── Store.gd        # Global state
│   │   ├── Events.gd       # Signal bus
│   │   ├── ThemeManager.gd # Era themes
│   │   └── SaveManager.gd  # Persistence
│   ├── editor/
│   │   ├── CodeEditor.gd   # Syntax highlighting
│   │   ├── Terminal.gd     # Output with ANSI
│   │   └── VigilanceMeter.gd
│   ├── puzzle/
│   │   ├── PuzzleNode.gd   # Base puzzle scene
│   │   ├── CodePuzzle.gd   # Architect/Ghost
│   │   ├── GridPuzzle.gd   # Saboteur/logic
│   │   └── PuzzleMap.gd    # Dependency graph
│   └── era/
│       ├── MagitechTheme.gd
│       └── CyberpunkTheme.gd
└── themes/
    ├── magitech.tres
    └── cyberpunk.tres
```

### Bridge Classes
```go
// bridge.go
func RegisterClasses(initObj *core.InitObject) {
    initObj.RegisterSceneInitializer(func() {
        // VM wrapper
        ClassDBRegisterClass[*VMWrapper](&VMWrapper{}, ...)
        
        // Puzzle system
        ClassDBRegisterClass[*PuzzleBoard](&PuzzleBoard{}, ...)
        ClassDBRegisterClass[*PuzzleNode](&PuzzleNode{}, ...)
        
        // Theme system
        ClassDBRegisterClass[*ThemeManager](&ThemeManager{}, ...)
    })
}

// vm_wrapper.go
type VMWrapper struct {
    CanvasItemImpl
    vm *scheduler.VM
}

func (v *VMWrapper) ExecuteCode(code String) {
    // Compile → Run → Emit signals
}

func (v *VMWrapper) SetEmitCallback(callable Callable) {
    v.vm.SetEmitCallback(func(s string) {
        callable.Call(builtin.NewVariantString(s))
    })
}
```

---

## Phase 2: Puzzle System (Week 2-3)

### Dependency Graph (from godot_puzzle_dependencies)
```gdscript
# PuzzleBoard.gd
class_name PuzzleBoard
extends Resource

var nodes: Dictionary = {}  # id -> PuzzleNode

func add_node(node: PuzzleNode):
    nodes[node.id] = node

func get_available() -> Array[PuzzleNode]:
    var result = []
    for n in nodes.values():
        if n.state == PuzzleNode.State.AVAILABLE:
            result.append(n)
    return result

func complete_node(id: String):
    var n = nodes[id]
    n.state = PuzzleNode.State.COMPLETED
    for unlock_id in n.unlocks:
        var u = nodes[unlock_id]
        if u and u.state == PuzzleNode.State.LOCKED and u.all_prereqs_met(self):
            u.state = PuzzleNode.State.AVAILABLE
    Events.instance.puzzle_completed.emit(id)
```

### Custom Resource Format (.challenge)
```json
{
  "version": 1,
  "era": "magitech",
  "pack": "tier1_basics",
  "nodes": [
    {
      "id": "rune_binding_01",
      "type": "code",
      "title": "Cracked Rune of Binding",
      "prereqs": [],
      "unlocks": ["rune_release_01"],
      "difficulty": 1,
      "data": {
        "starter": "fn bind(a, b) => /* TODO */",
        "tests": [
          {"input": "bind(1, 2)", "expect": "3"}
        ],
        "hints": [
          {"tier": 1, "text": "Mana flows between connected runes", "cost": 0},
          {"tier": 2, "text": "The bind rune adds two values", "cost": 5}
        ]
      }
    }
  ]
}
```

### Loader
```gdscript
# ChallengeLoader.gd
static func load_pack(era: String, pack: String) -> PuzzleBoard:
    var path = "res://data/challenges/%s/%s.challenge" % [era, pack]
    var file = FileAccess.open(path, FileAccess.READ)
    var json = JSON.parse_string(file.get_as_text())
    var board = PuzzleBoard.new()
    for node_data in json.nodes:
        var node = _create_node(node_data)
        board.add_node(node)
    return board
```

---

## Phase 3: Gameplay Modes (Week 3-4)

### Architect Mode
```gdscript
# CodePuzzle.gd (Architect)
extends PuzzleNode

func _ready():
    $CodeEditor.code_changed.connect(_on_code_changed)
    $RunButton.pressed.connect(_on_run)
    
func _on_run():
    var code = $CodeEditor.get_code()
    var result = VMWrapper.execute(code)
    _check_tests(result)
    
func _check_tests(result: Dictionary):
    var passed = 0
    for test in puzzle_data.tests:
        if _run_test(test, result):
            passed += 1
    if passed == test_count:
        board.complete_node(id)
        _generate_passcode(result)
```

### Ghost Mode
```gdscript
# GhostPuzzle.gd
extends PuzzleNode

@export var vigilance_budget: float = 0.7

var current_vigilance: float = 0.0

func _on_code_changed(new_code: String):
    var diff = _compute_diff(original_code, new_code)
    current_vigilance += _calculate_vigilance_cost(diff)
    $VigilanceMeter.set_vigilance(current_vigilance)
    
    if current_vigilance >= vigilance_budget:
        _trigger_purge()

func _calculate_vigilance_cost(diff: Array) -> float:
    var cost = 0.0
    for change in diff:
        match change.type:
            "insert": cost += 0.02
            "delete": cost += 0.03
            "modify_loop": cost += 0.08
            "add_import": cost += 0.10
            "restructure": cost += 0.15
    return cost
```

### Saboteur Mode
```gdscript
# SaboteurPuzzle.gd
extends PuzzleNode

var fault_injection_points: Array[FaultPoint] = []

func _ready():
    _analyze_system_code()
    
func _analyze_system_code():
    # Parse system code, find injection points
    for node in system_ast:
        if node.type == "function_call" and node.name == "cache_get":
            fault_injection_points.append(FaultPoint.new(
                type: "race_condition",
                location: node.position,
                effect: "cache_poison"
            ))

func inject_fault(fault: FaultPoint):
    var result = VMWrapper.inject_fault(system_code, fault)
    if result.chain_reaction:
        _trigger_chain_reaction(result.mutations)
    if result.target_mutation == "cache_poisoned":
        board.complete_node(id)
```

---

## Phase 4: Procedural Generation (Week 4-5)

### Challenge Generator
```go
// generator/challenge.go
type ChallengeGenerator struct {
    era      string
    templates []ChallengeTemplate
    rng      *rand.Rand
    luck     *LuckEngine
}

type ChallengeTemplate struct {
    ID           string
    Type         string      // "architect", "ghost", "saboteur"
    Category     string      // "sorting", "graph", "crypto", etc.
    Prereqs      []string
    Difficulty   float64     // 0.0 - 1.0
    Params       map[string]any
    ModeWeights  map[string]float64
}

func (g *ChallengeGenerator) GeneratePack(seed int64, count int) []Challenge {
    g.rng = rand.New(rand.NewSource(seed))
    
    // 1. Topological sort of templates by prerequisites
    available := g.getRootTemplates()
    var pack []Challenge
    
    for len(pack) < count && len(available) > 0 {
        // Weighted random selection based on difficulty & luck
        idx := g.weightedChoice(available)
        tmpl := available[idx]
        available = append(available[:idx], available[idx+1:]...)
        
        // 2. Instantiate with procedural parameters
        challenge := g.instantiate(tmpl)
        pack = append(pack, challenge)
        
        // 3. Unlock new templates
        available = append(available, g.getUnlocked(tmpl.ID, pack)...)
    }
    return pack
}

func (g *ChallengeGenerator) instantiate(tmpl ChallengeTemplate) Challenge {
    // Apply luck-based variations
    luck := g.luck.Current()
    
    var challenge Challenge
    switch tmpl.Type {
    case "architect":
        challenge = g.genArchitect(tmpl, luck)
    case "ghost":
        challenge = g.genGhost(tmpl, luck)
    case "saboteur":
        challenge = g.genSaboteur(tmpl, luck)
    }
    return challenge
}
```

### Algorithm Templates (from competitive-programming)
```go
// generator/templates.go
var AlgorithmTemplates = map[string]ChallengeTemplate{
    "sorting_basic": {
        ID: "sort_bubble",
        Type: "architect",
        Category: "sorting",
        Difficulty: 0.2,
        Params: map[string]any{
            "algorithm": "bubble",
            "array_size": 10,
        },
    },
    "sorting_advanced": {
        ID: "sort_quick",
        Type: "architect",
        Category: "sorting",
        Prereqs: []string{"sort_bubble"},
        Difficulty: 0.5,
        Params: map[string]any{
            "algorithm": "quicksort",
            "array_size": 100,
            "constraints": "stable, in-place",
        },
    },
    "graph_dijkstra": {
        ID: "netrunner_route",
        Type: "ghost",
        Category: "graph",
        Difficulty: 0.6,
        Params: map[string]any{
            "nodes": 50,
            "edges": 200,
            "vigilance_budget": 0.6,
        },
    },
    "crypto_rsa": {
        ID: "key_crack",
        Type: "saboteur",
        Category: "crypto",
        Difficulty: 0.8,
        Params: map[string]any{
            "key_size": 1024,
            "fault_type": "timing_attack",
        },
    },
}
```

---

## Phase 5: Passcode & Vigilance (Week 5-6)

### Passcode Engine
```go
// passcode/engine.go
type PasscodeEngine struct {
    detector *GlitchDetector
}

func (pe *PasscodeEngine) Generate(ctx *ExecutionContext) *PasscodeResult {
    // 1. Direct output
    if emit := ctx.LastEmit(); emit != "" {
        return &PasscodeResult{
            Code:   hashPasscode(emit),
            Source: SourceDirectOutput,
        }
    }
    
    // 2. Error logs
    if errLog := ctx.ErrorLog(); errLog != "" {
        if isInterestingError(errLog) {
            return &PasscodeResult{
                Code:   hashPasscode(errLog),
                Source: SourceErrorLog,
            }
        }
    }
    
    // 3. Glitch detection
    if glitch := pe.detector.Analyze(ctx); glitch != nil {
        return &PasscodeResult{
            Code:       glitch.Passcode,
            Source:     SourceGlitchExploit,
            Mutations:  glitch.Mutations,
        }
    }
    
    // 4. Timing side-channel
    if timing := ctx.TimingProfile(); timing.Variance > threshold {
        return &PasscodeResult{
            Code:   hashPasscode(fmt.Sprintf("%d", timing.MeanMicros)),
            Source: SourceTimingVariance,
        }
    }
    
    return nil
}
```

### Glitch Patterns
```go
// passcode/glitch.go
var GlitchPatterns = []GlitchPattern{
    {
        Name: "Integer Overflow",
        Detect: func(ctx *ExecutionContext) bool {
            return ctx.StackOverflow() || ctx.IntOverflow()
        },
        Generate: func(ctx *ExecutionContext) string {
            return fmt.Sprintf("OVERFLOW-%X", ctx.OverflowAddress())
        },
        Mutations: map[string]any{"memory_corrupted": true},
        VigilanceCost: 0.3,
    },
    {
        Name: "Race Condition",
        Detect: func(ctx *ExecutionContext) bool {
            return ctx.DetectedRace()
        },
        Generate: func(ctx *ExecutionContext) string {
            return fmt.Sprintf("RACE-%d", ctx.RaceTimestamp())
        },
        Mutations: map[string]any{"cache_poisoned": true},
        VigilanceCost: 0.2,
    },
    {
        Name: "Use-After-Free",
        Detect: func(ctx *ExecutionContext) bool {
            return ctx.UseAfterFree()
        },
        Generate: func(ctx *ExecutionContext) string {
            return fmt.Sprintf("UAF-%X", ctx.FreeAddress())
        },
        Mutations: map[string]any{"heap_corrupted": true},
        VigilanceCost: 0.4,
    },
}
```

---

## Phase 6: Polish & Launch (Week 6-8)

### Save System
```gdscript
# SaveManager.gd
func save_game(slot: int):
    var data = {
        "version": 1,
        "era": Store.get("current_era"),
        "luck": Store.get("luck"),
        "completed": Store.get("completed_puzzles"),
        "hints": Store.get("hint_archive"),
        "stats": Store.get("statistics"),
        "world_state": WorldState.get_all(),
        "rng_state": RNG.get_state(),
        "timestamp": Time.get_unix_time_from_system()
    }
    var file = FileAccess.open("user://save_%d.save" % slot, FileAccess.WRITE)
    file.store_string(JSON.stringify(data))

func load_game(slot: int) -> bool:
    var path = "user://save_%d.save" % slot
    if not FileAccess.file_exists(path): return false
    var file = FileAccess.open(path, FileAccess.READ)
    var data = JSON.parse_string(file.get_as_text())
    _restore_state(data)
    return true
```

### Theme System
```gdscript
# ThemeManager.gd
func apply_era(era: String):
    var theme = themes[era]
    for canvas in get_tree().get_nodes_in_group("themed"):
        if canvas is Control:
            canvas.theme = theme
    # Custom colors
    RenderingServer.set_default_clear_color(theme.bg_color)
    Events.instance.era_changed.emit(era)
```

### Build & Deploy
```bash
# tools/build.sh
#!/bin/bash
set -e

VERSION=$(cat VERSION)

# Build Go extension
cd backend
make build
make build-mac
make build-win

# Export Godot project
cd ../client
godot --headless --export-release "Windows Desktop" ../build/ChallengeToYOU-$VERSION-windows.exe
godot --headless --export-release "macOS" ../build/ChallengeToYOU-$VERSION-macos.zip
godot --headless --export-release "Linux/X11" ../build/ChallengeToYOU-$VERSION-linux.x86_64

# Package for Itch.io
cd ../build
zip -r ChallengeToYOU-$VERSION-windows.zip ChallengeToYOU-$VERSION-windows.exe
zip -r ChallengeToYOU-$VERSION-macos.zip ChallengeToYOU-$VERSION-macos
tar -czf ChallengeToYOU-$VERSION-linux.tar.gz ChallengeToYOU-$VERSION-linux
```

---

## Milestones & Deliverables

| Week | Milestone | Deliverable |
|------|-----------|-------------|
| 0-1 | VM Complete | `vm` CLI runs Pscript, passes all tests |
| 1-2 | Godot+Go | GDExtension loads, VM executes from Godot |
| 2-3 | Puzzle System | Dependency graph loads .challenge packs |
| 3-4 | 3 Modes | Architect/Ghost/Saboteur playable |
| 4-5 | Proc Gen | Infinite challenges from templates |
| 5-6 | Passcodes | Emergent passcodes, vigilance, luck |
| 6-7 | Polish | Themes, saves, hints, audio, UI |
| 7-8 | Launch | Itch.io alpha, Steam page |

---

## Technical Debt Tracker

| Item | Priority | Effort | Notes |
|------|----------|--------|-------|
| TreeSitter for Pscript | Medium | 2 days | Better highlighting |
| WASM fallback | Low | 1 week | Web demo |
| Multiplayer (co-op) | Low | 2 weeks | Phase 2+ |
| Level editor | Medium | 1 week | Content creation |
| Achievements | Low | 3 days | Steam integration |
| Localization | Low | 1 week | i18n support |

---

## Definition of Done

- [ ] VM: All bytecode instructions tested, limits enforced
- [ ] Bridge: Zero-copy Godot↔Go for primitives
- [ ] Puzzles: 50+ challenges across 3 modes, 2 eras
- [ ] Generation: Seed produces identical pack, difficulty scales
- [ ] Passcodes: 5+ sources, mutations persist world state
- [ ] Save: Load/resume works mid-puzzle
- [ ] Themes: Instant era switch, no restart
- [ ] Build: One command produces 3 platform binaries
- [ ] Tests: >80% coverage on VM, generator, passcode
- [ ] Performance: 60fps with 1000 VM steps/frame