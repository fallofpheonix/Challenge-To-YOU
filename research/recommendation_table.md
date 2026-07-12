# Final Recommendation Table

| Repository | Keep | Discard | Why | Priority | Reusable Parts |
|------------|------|---------|-----|----------|----------------|
| **godot-go/godot-go** | ✅ | | **Primary Go binding for Godot 4.3 GDExtension**. Complete class registration, variant handling, build system. | **Critical** | `pkg/core.InitObject`, `ClassDBRegisterClass`, `pkg/builtin/*`, `pkg/ffi/*`, Makefile |
| **godot-go/godot-go-demo-projects** | ✅ | | **Working templates**. dodge_the_creeps shows full pipeline: Go class → Godot scene → signal binding → build. | **Critical** | `main.go` entry point, `pkg/demo/hud.go` class pattern, `Makefile` cross-platform build |
| **nathanfranke/gdextension** | ✅ | | **C++ GDExtension template** (reference only). Useful for understanding extension manifest, SCons build, singleton pattern. | Medium | `project.gdextension` format, `register_types.cpp`, SConstruct |
| **godot-academy/godot-coding-challenge** | ✅ | | **7 complete Godot coding challenge projects**. Each isolated with own `project.godot`. Pattern: per-challenge scene structure. | **High** | Project-per-challenge structure, `Star.gd` particle pattern, `Tween` visualizer |
| **russmatney/dothop** | ✅ | | **Full puzzle game with custom .puzz format, dependency graph, seasonal packs**. Best architecture reference. | **Critical** | `PuzzleWorld.gd` state machine, `ParsedGame.gd` parser, `.puzz` format, `Store.gd` autoload, `Themes/`, `Events.gd` signal bus |
| **vivisuke/GodotSudoku** | ✅ | | **Complete Sudoku with animations, levels, UI**. Good puzzle game reference. | **High** | `Global.gd` autoload pattern, `fallingNumber.gd` tween animation, level scene management |
| **r1z11/Sudoku-puzzle-generator** | ✅ | | **Pure GDScript backtracking generator with constraint propagation**. Algorithm directly adaptable. | **High** | `Main.gd` `fill_remaining()`, `get_candidates()`, `is_valid()`, `clean_numbers()` |
| **nathanhoad/godot_puzzle_dependencies** | ✅ | | **Editor addon for puzzle dependency graphs**. `Thing`/`Board` classes, visual graph editor, custom resource. | **Critical** | `Thing.gd` (prerequisites/unlocks), `Board.gd` validation, `GraphPopupMenu.gd`, `.tres` resource format |
| **kidscancode/godot_recipes** | ✅ | | **Hugo site with Godot 4 patterns**. Covers state machines, signals, UI, animation, shaders. | **High** | `ai/state_machine.md`, `basics/signals.md`, `ui/` patterns, `animation/tween.md` |
| **godotengine/awesome-godot** | | ❌ | **Link list only**. No code. Discard after extracting useful links. | Discard | None (metadata only) |
| **sagarneeli/coding-challenges** | ✅ | | **150+ LeetCode solutions, one folder per problem**. Great for challenge content. | **High** | Problem descriptions, Python reference implementations, topic coverage (arrays, trees, graphs, DP) |
| **kothariji/competitive-programming** | ✅ | | **CP solutions organized by topic** (DP, Graphs, Trees, Bit Manipulation, Number Theory, Crypto). Best algorithm reference. | **Critical** | Topic-organized C++/Java/Python implementations for porting to Go VM builtins |
| **lnishan/awesome-competitive-programming** | | ❌ | **Curated list only**. No implementations. | Discard | None (metadata only) |
| **pawelborkar/awesome-repos** | | ❌ | **Curated list only**. No implementations. | Discard | None (metadata only) |

---

## Summary

### Keep (11 repos)
| Priority | Count | Repositories |
|----------|-------|--------------|
| **Critical** | 5 | godot-go, godot-go-demo-projects, dothop, godot_puzzle_dependencies, competitive-programming |
| **High** | 5 | godot-coding-challenge, GodotSudoku, Sudoku-puzzle-generator, godot_recipes, coding-challenges |
| **Medium** | 1 | gdextension |

### Discard (3 repos)
- awesome-godot
- awesome-competitive-programming  
- awesome-repos

### Primary References by Subsystem

| Subsystem | Primary Reference | Secondary |
|-----------|------------------|-----------|
| Go ↔ Godot | godot-go + demo-projects | gdextension (manifest) |
| Puzzle Graph | godot_puzzle_dependencies | dothop (runtime) |
| Puzzle Format | dothop (.puzz) | godot_puzzle_dependencies (.tres) |
| Puzzle Generation | Sudoku-puzzle-generator | dothop (seasonal packs) |
| Algorithms | competitive-programming | coding-challenges |
| UI Patterns | godot_recipes | dothop, GodotSudoku |
| State Management | dothop (Store.gd, Events.gd) | GodotSudoku (Global.gd) |
| Theming | dothop (Themes/) | godot_recipes (ui/themes) |
| Challenge Content | godot-coding-challenge | coding-challenges |

---

## Action Items

1. **Immediate**: Clean up discarded repos (`rm -rf awesome-*`)
2. **Week 1**: Use `godot-go-demo-projects/2d/dodge_the_creeps` as GDExtension template
3. **Week 2**: Port `dothop/src/core/Store.gd` + `Events.gd` + `godot_puzzle_dependencies/addons/puzzle_dependencies/components/Thing.gd` → GDScript core
4. **Week 2**: Adapt `Sudoku-puzzle-generator/Main.gd` backtracking → Go VM → `backend/generator/puzzle.go`
5. **Week 3**: Mine `competitive-programming/Graph/` + `DP/` → Go VM builtins
6. **Week 3**: Use `godot-coding-challenge/1-starfield` structure for per-challenge scenes