# Research Workspace

Curated local research workspace for **Challenge To YOU** — a roguelike coding puzzle game across fantasy/sci-fi eras.

## Repository Overview

| Repository | Type | Priority | Keep |
|------------|------|----------|------|
| [godot-go/godot-go](godot-go/) | Go bindings for Godot 4 GDExtension | **Critical** | ✅ |
| [godot-go/godot-go-demo-projects](godot-go-demo-projects/) | Working Go+Godot templates | **Critical** | ✅ |
| [nathanfranke/gdextension](gdextension/) | C++ GDExtension template (reference) | Medium | ✅ |
| [godot-academy/godot-coding-challenge](godot-coding-challenge/) | Coding challenges in Godot | **High** | ✅ |
| [russmatney/dothop](dothop/) | Grid puzzle game with custom puzzle format | **High** | ✅ |
| [vivisuke/GodotSudoku](GodotSudoku/) | Sudoku puzzle game | **High** | ✅ |
| [r1z11/Sudoku-puzzle-generator](Sudoku-puzzle-generator/) | Backtracking Sudoku generator | **High** | ✅ |
| [nathanhoad/godot_puzzle_dependencies](godot_puzzle_dependencies/) | Puzzle dependency graph addon | **High** | ✅ |
| [kidscancode/godot_recipes](godot_recipes/) | Godot patterns & recipes (Hugo site) | **High** | ✅ |
| [godotengine/awesome-godot](awesome-godot/) | Curated resource list | Reference | ❌ (list only) |
| [sagarneeli/coding-challenges](coding-challenges/) | LeetCode solutions by problem | **High** | ✅ |
| [kothariji/competitive-programming](competitive-programming/) | CP solutions organized by topic | **High** | ✅ |
| [lnishan/awesome-competitive-programming](awesome-competitive-programming/) | Curated CP resource list | Reference | ❌ (list only) |
| [pawelborkar/awesome-repos](awesome-repos/) | Curated repo list | Reference | ❌ (list only) |

---

## Quick Start

```bash
# Godot + Go integration (primary reference)
cd godot-go-demo-projects/2d/dodge_the_creeps
make build    # builds GDExtension
# Open project/ in Godot 4

# Puzzle systems
cd godot_puzzle_dependencies
# Open in Godot - examine addons/puzzle_dependencies/

# Dothop puzzle game
cd dothop
# Open in Godot - examine PuzzleWorld.gd, .puzz format

# Algorithm reference
cd coding-challenges/1-two-sum
cat two-sum.py
```

---

## Architecture Decisions

### Godot + Go Integration (godot-go)
- **Pattern**: GDExtension via `godot-go` bindings
- **Entry point**: `main.go` with `//export GodotGo*Init` function
- **Class registration**: `ClassDBRegisterClass` with virtual methods & signals
- **Build**: Makefile → produces `.so`/`.dll`/`.dylib` in `project/addons/`
- **Godot project**: Standard `project.godot`, loads extension via `addons/`

### Puzzle System (dothop + godot_puzzle_dependencies)
- **Custom resource format**: `.puzz` files for level data
- **Dependency graph**: Nodes → edges for puzzle prerequisites
- **State machine**: Parsed game → PuzzleWorld → active puzzle
- **Extensible**: Add new puzzle types via `ParsedGame.gd`

### Sudoku Generation (Sudoku-puzzle-generator)
- **Algorithm**: Backtracking with constraint propagation
- **Phases**: Fill diagonal 3x3 → recursive fill → remove clues
- **Validation**: Row/col/box constraint checking

### Competitive Programming Reference
- **coding-challenges**: One folder per LeetCode problem, Python solutions
- **competitive-programming**: Organized by topic (DP, Graph, Trees, etc.) - C++/Java/Python

---

## Key Files to Study First

```
godot-go-demo-projects/2d/dodge_the_creeps/
├── main.go                 # GDExtension entry point
├── pkg/demo/hud.go         # Godot class in Go
├── Makefile                # Build script
└── project/                # Godot project

godot_puzzle_dependencies/addons/puzzle_dependencies/
├── plugin.gd               # Editor plugin entry
├── components/
│   ├── board.gd            # Puzzle board logic
│   ├── thing.gd            # Puzzle node/entity
│   └── graph_popup_menu.gd # Dependency graph UI
└── views/                  # Editor UI

dothop/
├── PuzzleWorld.gd          # Core puzzle state machine
├── ParsedGame.gd           # .puzz format parser
├── PuzzleSetData.gd        # Puzzle collection
└── *.puzz                  # Level data format

Sudoku-puzzle-generator/
├── Main.gd                 # Backtracking generator
└── Panel.tscn              # Cell scene
```

---

## Reusable Components Identified

| System | Source | Adaptation |
|--------|--------|------------|
| GDExtension build pipeline | godot-go-demo-projects | Direct use |
| Godot class binding in Go | godot-go/pkg/core | Direct use |
| Puzzle dependency graph | godot_puzzle_dependencies | Core architecture |
| Custom resource format (.puzz) | dothop | Adapt for challenge format |
| Backtracking puzzle gen | Sudoku-puzzle-generator | Adapt for coding challenges |
| Topic-organized algorithms | competitive-programming | Reference implementation |
| LeetCode problem solutions | coding-challenges | Challenge content |

---

## Discarded (Reference Only)
- `awesome-godot` - just links, no code
- `awesome-competitive-programming` - just links
- `awesome-repos` - just links

---

## Next Steps

1. **Week 1**: Build minimal GDExtension from `dodge_the_creeps` template
2. **Week 2**: Implement puzzle dependency graph from `godot_puzzle_dependencies`
3. **Week 3**: Design `.challenge` format based on `.puzz` + backtracking generator
4. **Week 4**: Port algorithm implementations from `competitive-programming` to Go VM

See `implementation_plan.md` for detailed phases.