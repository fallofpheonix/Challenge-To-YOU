# Repository Summary

Detailed analysis of each cloned repository.

---

## 1. godot-go/godot-go

**Purpose**: Go bindings for Godot 4.3 GDExtension API  
**Tech Stack**: Go 1.21+, GDExtension C API, cgo  
**Structure**:
```
godot-go/
├── pkg/
│   ├── builtin/      # Godot built-in types (Vector2, String, Array, etc.)
│   ├── constant/     # Godot constants/enums
│   ├── core/         # Core API: ClassDB, Object, Variant, Callable
│   ├── ffi/          # FFI bindings to GDExtension C API
│   ├── gdclassimpl/  # Class implementation helpers
│   ├── gdclassinit/  # Class initialization
│   ├── log/          # Logging (zap)
│   └── util/         # Utilities
├── cmd/              # Code generation tools
└── test/             # Integration tests
```

**Build System**: Makefile with `go build -buildmode=c-shared`  
**Architecture**: Generated bindings + hand-written wrappers. Uses `godot_headers` for GDExtension API.  
**Key Pattern**: `ClassDBRegisterClass` with `GDClassImpl` embedding for Godot classes in Go.

**Reusable**:
- `pkg/core.InitObject` - GDExtension initialization
- `pkg/core.ClassDBRegisterClass` - Register Go types as Godot classes
- `pkg/builtin/*` - All Godot built-in types in Go
- Build system (Makefile) for cross-platform `.so`/`.dll`/`.dylib`

**Learning Value**: High - complete GDExtension binding reference

---

## 2. godot-go/godot-go-demo-projects

**Purpose**: Working templates for Go + Godot games  
**Tech Stack**: Go, Godot 4, GDExtension  
**Structure**:
```
godot-go-demo-projects/
├── go.work                 # Go workspace
├── 2d/
│   ├── dodge_the_creeps/   # Official Godot tutorial in Go
│   └── topdown/            # Top-down movement demo
```

**dodge_the_creeps**:
- `main.go` - GDExtension entry point with `GodotGoDemo2DDodgeTheCreepsInit`
- `pkg/demo/hud.go` - HUD class with signals, virtual methods, property binding
- `Makefile` - Builds extension to `project/addons/`
- `project/` - Godot project with scenes

**Reusable**: 
- Complete build pipeline (Makefile → GDExtension → Godot)
- Signal/virtual method binding patterns
- Node path resolution in Go (`GetNode`, `ObjectCastTo`)

---

## 3. nathanfranke/gdextension

**Purpose**: C++ GDExtension template with SConstruct build  
**Tech Stack**: C++17, Godot 4, SCons  
**Structure**:
```
gdextension/
├── SConstruct              # SCons build config
├── src/
│   ├── register_types.cpp  # Class registration
│   ├── my_node.cpp/hpp     # Example Node3D
│   └── my_singleton.cpp/hpp # Autoload singleton
└── project/                # Godot project
```

**Reusable**: 
- SCons build for GDExtension (alternative to Makefile)
- Singleton/autoload pattern in C++
- Clean separation of registration vs implementation

---

## 4. godot-academy/godot-coding-challenge

**Purpose**: Coding challenges implemented as Godot projects  
**Tech Stack**: GDScript, Godot 4  
**Structure**: 7 independent Godot projects (each has own `project.godot`):
- 1-starfield - Particle system
- 2-menger-sponge - 3D fractal
- 3-snake-game - Classic snake
- 4-purple-rain - Particle rain
- 5-space-invaders - Arcade shooter
- 6-mitosis - Cell division sim
- RBG Worm - Worm movement
- tween-visualizer - Animation tweening

**Reusable**:
- Project-per-challenge structure (isolated, testable)
- Scene organization patterns
- Tween/animation usage

---

## 5. russmatney/dothop

**Purpose**: Grid-based puzzle game with seasonal puzzle packs  
**Tech Stack**: GDScript, Godot 4  
**Structure**:
```
dothop/
├── PuzzleWorld.gd          # Core state machine
├── ParsedGame.gd           # .puzz format parser
├── PuzzleSetData.gd        # Puzzle collection metadata
├── components/             # Reusable components
└── *.puzz                  # Custom puzzle format (JSON-like)
```

**.puzz format**:
```json
{
  "meta": {"title": "...", "author": "...", "season": "spring"},
  "puzzles": [
    {"id": "p1", "grid": [[...]], "constraints": {...}}
  ]
}
```

**Key Systems**:
- `ParsedGame` → tokenizes `.puzz` → `PuzzleWorld` state machine
- Seasonal puzzle packs (spring/summer/fall/winter/extra)
- Custom resource import for `.puzz` files
- Grid-based puzzle logic with constraint validation

**Reusable**:
- Custom resource format + importer pattern
- Puzzle state machine (load → validate → play → complete)
- Seasonal/content pack organization
- Grid constraint system

---

## 6. vivisuke/GodotSudoku

**Purpose**: Complete Sudoku game with UI  
**Tech Stack**: GDScript, Godot 4  
**Structure**:
```
GodotSudoku/
├── Main.gd                 # Game controller
├── LevelScene.gd           # Level management
├── Global.gd               # Autoload singleton
├── fallingNumber.gd        # Animation component
└── questButton.gd          # UI component
```

**Reusable**:
- Autoload pattern for global state (`Global.gd`)
- Falling number animation (tween-based)
- Level scene management

---

## 7. r1z11/Sudoku-puzzle-generator

**Purpose**: Pure GDScript Sudoku generator using backtracking  
**Tech Stack**: GDScript  
**Key Algorithm** (`Main.gd`):
```
1. Fill diagonal 3x3 boxes (independent)
2. Recursive backtracking for remaining cells
3. Constraint propagation (row/col/box)
4. Remove clues for difficulty
```

**Reusable**:
- Backtracking with constraint propagation
- `get_candidates(row, col)` - valid numbers for cell
- `clean_numbers()` - remove unsolvable cells
- Efficient bitmask/array operations in GDScript

---

## 8. nathanhoad/godot_puzzle_dependencies

**Purpose**: Godot editor addon for puzzle dependency graphs  
**Tech Stack**: GDScript, Godot 4 EditorPlugin  
**Structure**:
```
addons/puzzle_dependencies/
├── plugin.gd               # EditorPlugin entry
├── components/
│   ├── board.gd            # Puzzle board logic
│   ├── thing.gd            # Puzzle node (prerequisites, unlocks)
│   ├── graph_popup_menu.gd # Dependency visualization
│   └── download_update_panel.gd
├── views/                  # Editor UI
└── utilities/
```

**Key Classes**:
- `Thing` - Puzzle node with `prerequisites[]`, `unlocks[]`, `state`
- `Board` - Manages collection of Things, validates dependencies
- `GraphPopupMenu` - Visual dependency graph editor

**Reusable**:
- **Core architecture**: DAG-based puzzle dependencies
- `Thing` resource - serializable puzzle node
- Editor integration for visual graph editing
- State machine: LOCKED → AVAILABLE → COMPLETED

---

## 9. kidscancode/godot_recipes

**Purpose**: Godot 4 recipes/patterns documentation (Hugo site)  
**Structure**: Markdown content organized by topic:
```
content/
├── 2D/           # Movement, tilemaps, physics
├── 3D/           # Camera, lighting, meshes
├── Math/         # Vectors, transforms, interpolation
├── ai/           # State machines, behavior trees, GOAP
├── animation/    # Tween, AnimationPlayer, AnimationTree
├── basics/       # Nodes, scenes, signals, resources
├── games/        # Complete game tutorials
├── input/        # Actions, mappings, controllers
├── physics/      # RigidBody, CharacterBody, Area
├── recipes/      # Common patterns
├── shaders/      # Visual shaders
└── ui/           # Containers, themes, custom controls
```

**Reusable**:
- State machine pattern (`ai/state_machine.md`)
- Signal bus / event system (`basics/signals.md`)
- Resource-based data (`basics/resources.md`)
- UI patterns (`ui/`)

---

## 10. godotengine/awesome-godot

**Purpose**: Curated list of Godot resources (plugins, templates, tutorials)  
**Type**: Link collection only - **DISCARD** (no code)

---

## 11. sagarneeli/coding-challenges

**Purpose**: LeetCode solutions, one folder per problem  
**Structure**: 150+ folders, each with `README.md` + solution (Python)  
**Example** (`1-two-sum/`):
```
README.md       # Problem description
two-sum.py      # Python solution
```

**Reusable**:
- Problem descriptions for challenge content
- Algorithm implementations to port to Go VM
- Topic coverage: arrays, strings, trees, graphs, DP, etc.

---

## 12. kothariji/competitive-programming

**Purpose**: Competitive programming solutions organized by topic  
**Structure**:
```
competitive-programming/
├── Arrays/
├── Dynamic Programming/
├── Graph/
├── Tree/
├── Bit-Manipulations/
├── Number Theory/
├── Greedy/
├── Backtracking/
├── Searching/
├── Sorting/
├── String/
├── CSES/                    # CSES problem set
├── Codechef solutions/
├── Leetcode - Top Interview Questions/
└── Cryptographic algos/
```

**Reusable**:
- **Best algorithm reference** - organized by CS topic
- C++/Java/Python implementations
- CSES/Codeforces/LeetCode problem coverage
- Ready to port to Go for VM builtins

---

## 13. lnishan/awesome-competitive-programming

**Purpose**: Curated CP resource list - **DISCARD** (links only)

---

## 14. pawelborkar/awesome-repos

**Purpose**: Curated repo list - **DISCARD** (links only)

---

## Summary: Keep vs Discard

| Keep (Code/Architecture) | Discard (Lists Only) |
|-------------------------|---------------------|
| godot-go | awesome-godot |
| godot-go-demo-projects | awesome-competitive-programming |
| gdextension (reference) | awesome-repos |
| godot-coding-challenge | |
| dothop | |
| GodotSudoku | |
| Sudoku-puzzle-generator | |
| godot_puzzle_dependencies | |
| godot_recipes | |
| coding-challenges | |
| competitive-programming | |