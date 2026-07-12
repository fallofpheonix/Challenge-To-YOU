# Architecture Notes

Key architectural patterns extracted from research repositories.

---

## 1. Godot + Go (GDExtension) Architecture

### Entry Point Pattern
```go
// main.go
import "C"
//export GodotGoMyGameInit
func GodotGoMyGameInit(p_get_proc_address, p_library, r_initialization unsafe.Pointer) bool {
    initObj := core.NewInitObject(...)
    initObj.RegisterSceneInitializer(func() {
        demo.RegisterClassPlayer()
        demo.RegisterClassEnemy()
    })
    return initObj.Init()
}
```

### Class Registration
```go
func RegisterClassPlayer() {
    ClassDBRegisterClass[*Player](&Player{}, []GDExtensionPropertyInfo{}, nil, func(t GDClass) {
        // Virtual methods
        ClassDBBindMethodVirtual(t, "V_OnHit", "_on_hit", nil, nil)
        // Properties
        ClassDBBindMethod(t, "GetHealth", "get_health", nil, nil)
        ClassDBBindMethod(t, "TakeDamage", "take_damage", []string{"amount"}, nil)
        // Signals
        ClassDBAddSignal(t, "died")
    })
}
```

### Build Pipeline (Makefile)
```makefile
build:
	go build -buildmode=c-shared -o project/addons/godot-go/libgodotgo.so .
	
# Cross-platform:
# Linux:   libgodotgo.so
# macOS:   libgodotgo.dylib  
# Windows: godotgo.dll
```

### Godot Project Integration
- `project.godot` loads extension via `[extensions]` section
- Extension `.gdextension` manifest in `addons/godot-go/`
- Scenes reference Go classes by registered name

---

## 2. Puzzle Dependency Graph (godot_puzzle_dependencies + dothop)

### Core Data Model
```gdscript
# Thing.gd - Puzzle node
class_name Thing
var prerequisites: Array[Thing] = []  # Must complete first
var unlocks: Array[Thing] = []        # Unlocked on completion
var state: State = State.LOCKED       # LOCKED, AVAILABLE, COMPLETED
var puzzle_data: Dictionary           # Puzzle-specific config

enum State { LOCKED, AVAILABLE, COMPLETED }
```

### Board Manager
```gdscript
# Board.gd
var things: Array[Thing] = []

func validate_dependencies() -> bool:
    # Check for cycles in prerequisite graph
    # Topological sort for unlock order
    
func get_available_things() -> Array[Thing]:
    return things.filter(func(t): return t.state == State.AVAILABLE)
    
func on_thing_completed(thing: Thing):
    thing.state = State.COMPLETED
    for unlocked in thing.unlocks:
        if unlocked.all_prereqs_met():
            unlocked.state = State.AVAILABLE
```

### Custom Resource Format (.puzz)
```json
{
  "meta": { "title": "Spring Pack", "version": 1 },
  "things": [
    { "id": "t1", "type": "grid_puzzle", "prereqs": [], "data": {...} },
    { "id": "t2", "type": "code_puzzle", "prereqs": ["t1"], "data": {...} }
  ]
}
```

### ParsedGame → PuzzleWorld Pipeline
```
.puzz file → ParsedGame.parse() → PuzzleSetData → PuzzleWorld (runtime state)
```

---

## 3. Procedural Puzzle Generation (Sudoku-puzzle-generator)

### Backtracking with Constraint Propagation
```gdscript
func generate():
    # 1. Fill diagonal 3x3 boxes (no conflicts possible)
    fill_diagonal_boxes()
    
    # 2. Recursive backtracking for remaining
    fill_remaining(0, 3)
    
    # 3. Remove clues for difficulty
    remove_clues(difficulty)

func fill_remaining(row, col) -> bool:
    if row >= 9: return true
    if col >= 9: return fill_remaining(row + 1, 0)
    if grid[row][col] != 0: return fill_remaining(row, col + 1)
    
    for num in shuffled(1..9):
        if is_valid(row, col, num):
            grid[row][col] = num
            if fill_remaining(row, col + 1): return true
            grid[row][col] = 0
    return false

func is_valid(row, col, num) -> bool:
    return num not in row_set[row] 
        and num not in col_set[col] 
        and num not in box_set[box_index(row, col)]
```

### Adaptation for Coding Challenges
```
Sudoku grid (9x9)           →  Challenge graph (DAG)
Cell value (1-9)            →  Solution approach (algorithm choice)
Row/col/box constraint      →  Prerequisite/dependency constraint
Backtracking fill           →  Topological generation
Clue removal                →  Difficulty tiering
```

---

## 4. Competitive Programming Algorithm Patterns

### Topic → Implementation Mapping
| Topic | Key Algorithms | Use in Game |
|-------|---------------|-------------|
| Arrays/Strings | Two pointers, sliding window, KMP | String manipulation challenges |
| Trees | DFS/BFS, LCA, segment tree | Hierarchical puzzles, parsing |
| Graphs | Dijkstra, A*, topological sort, SCC | Dependency resolution, pathfinding |
| DP | Knapsack, LIS, edit distance, interval DP | Optimization challenges |
| Bit Manipulation | Subset enumeration, bit DP | Flag/permission puzzles |
| Number Theory | GCD, sieve, modular arithmetic | Crypto/cyberpunk era |
| Greedy | Interval scheduling, Huffman | Resource allocation |
| Backtracking | N-Queens, Sudoku, permutations | Puzzle generation |

### Implementation Strategy
- Port core algorithms to **Go VM** as built-in functions
- Expose as `builtin` in pscript DSL: `sort()`, `dijkstra()`, `bit_count()`
- Use competitive-programming repo as reference implementation

---

## 5. State Management Patterns

### Godot Autoload (Global.gd / Store.gd)
```gdscript
# Store.gd (dothop pattern)
class_name Store
extends Node

signal data_changed(key)

var _data: Dictionary = {}

func set(key: String, value: Variant) -> void:
    _data[key] = value
    emit_signal("data_changed", key)

func get(key: String, default: Variant = null) -> Variant:
    return _data.get(key, default)
```

### Event Bus (Events.gd)
```gdscript
# Events.gd - Global signal bus
class_name Events
extends Node

signal puzzle_started(puzzle_id)
signal puzzle_completed(puzzle_id, solution)
signal era_unlocked(era_id)
signal vigilance_changed(level)
```

---

## 6. Resource-Based Data Architecture

### Custom Resources (ScriptableObjects equivalent)
```gdscript
# PuzzleData.gd
class_name PuzzleData
extends Resource

@export var id: String
@export var title: String
@export var description: String
@export var era: String
@export var difficulty: int
@export var prerequisites: Array[String]
@export var test_cases: Array[TestCase]
@export var starter_code: String
@export var solution_template: String
```

### Loading Pipeline
```gdscript
# PuzzleLoader.gd
static func load_puzzle(path: String) -> PuzzleData:
    var resource = ResourceLoader.load(path)
    if resource == null:
        push_error("Failed to load puzzle: ", path)
    return resource

static func load_era_pack(era: String) -> Array[PuzzleData]:
    var dir = DirAccess.open("res://puzzles/" + era)
    var puzzles = []
    for file in dir.get_files():
        if file.ends_with(".tres"):
            puzzles.append(load_puzzle("res://puzzles/" + era + "/" + file))
    return puzzles
```

---

## 7. Testing Architecture

### Unit Tests (Go VM)
```go
// vm_test.go
func TestBytecodeExecution(t *testing.T) {
    vm := NewVM(bytecode)
    vm.SetEmitCallback(func(s string) { emitted = append(emitted, s) })
    err := vm.Run()
    assert.NoError(t, err)
    assert.Equal(t, []string{"expected"}, emitted)
}
```

### Integration Tests (Godot)
```gdscript
# test_puzzle_runner.gd
extends GUTTestClass

func test_puzzle_dependency_unlock():
    var board = Board.new()
    board.add_thing(Thing.new("a"))
    board.add_thing(Thing.new("b", ["a"]))
    
    board.get_thing("a").complete()
    assert_true(board.get_thing("b").state == State.AVAILABLE)
```

---

## 8. Recommended Project Structure

```
challenge-to-you/
├── backend/                    # Go backend
│   ├── vm/                     # Bytecode VM (standalone module)
│   │   ├── lexer/
│   │   ├── parser/
│   │   ├── compiler/
│   │   ├── bytecode/
│   │   ├── scheduler/
│   │   └── cmd/vm/             # CLI harness
│   ├── generator/              # Procedural generation
│   ├── analyzer/               # AST analysis
│   ├── passcode/               # Passcode generation
│   └── cmd/sandbox/            # GDExtension entry
├── client/                     # Godot 4 project
│   ├── scenes/
│   │   ├── editor/             # Code editor UI
│   │   ├── terminal/           # Output terminal
│   │   ├── puzzle/             # Puzzle scenes
│   │   └── era/                # Era-specific themes
│   ├── scripts/
│   │   ├── core/               # Autoloads (Store, Events)
│   │   ├── puzzle/             # Puzzle system
│   │   ├── era/                # Era managers
│   │   └── bridge/             # GDExtension bridge
│   ├── addons/
│   │   └── godot-go/           # Built extension
│   └── project.godot
├── tools/                      # Build scripts
└── docs/                       # Architecture docs
```

---

## 9. Cross-Cutting Concerns

### Error Handling
- Go: Return `(Result, error)` everywhere
- Godot: `OK`/`ERR_*` constants, `assert()` for invariants

### Logging
- Go: `zap` structured logging (godot-go uses this)
- Godot: `push_error`, `push_warning`, `print_debug`

### Configuration
- Go: YAML/JSON config files, env vars
- Godot: `ProjectSettings`, `.cfg` files

### Serialization
- Puzzle data: JSON (`.puzz`/`.challenge`) or Godot `.tres`
- Save games: JSON or binary (Go handles both)