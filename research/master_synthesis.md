# Master Research Synthesis — Challenge-To-YOU

## Cross-Repository Scoring Matrix

| Dimension | godot-go | gdextension | dothop | puzzle_deps | Sudoku-gen | GodotSudoku | godot_recipes | comp-prog | coding-challenges | godot-coding |
|-----------|----------|-------------|--------|-------------|------------|-------------|---------------|-----------|-------------------|--------------|
| Architecture | 5 | 8 | 7 | 6 | 3 | 2 | 4 | 4 | 4 | 4 |
| Code Quality | 4 | 7 | 7 | 5 | 4 | 3 | 3 | 5 | 6 | 5 |
| Scalability | 6 | 6 | 5 | 4 | 2 | 1 | 3 | 5 | 4 | 3 |
| Maintainability | 3 | 8 | 6 | 5 | 3 | 2 | 4 | 3 | 5 | 4 |
| Performance | 3 | 7 | 4 | 6 | 5 | 3 | 4 | 6 | 6 | 4 |
| Documentation | 5 | 8 | 5 | 3 | 1 | 1 | 6 | 2 | 2 | 2 |
| Testing | 2 | 1 | 7 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| Reusability | 7 | 5 | 7 | 8 | 5 | 2 | 6 | 8 | 7 | 4 |
| Suitability C2U | 5 | 4 | 6 | 7 | 6 | 1 | 5 | 9 | 7 | 3 |
| **Avg** | **4.4** | **6.0** | **6.0** | **4.9** | **3.2** | **1.7** | **3.9** | **4.7** | **4.6** | **3.7** |

## Critical Findings Per Repo

### godot-go (Score: 4.4)
- **Don't use as library**: `log.Panic` error handling, reflection-based dispatch, debug stacktraces on every cgo call, dot-imports
- **Key pattern to adapt**: Codegen pipeline for type bindings, C-heap allocation for cgo-crossed structs
- **Our WebSocket architecture already avoids all cgo pain** — this validates the decision

### gdextension C++ (Score: 6.0)
- **Best init pattern**: 4-level `gdextension_initialize` with proper terminate callbacks
- **Best singleton guards**: `ERR_FAIL_COND` on double-init/double-free
- **Adopt**: Self-contained addon structure, platform support table, SCons auto-discovery

### dothop (Score: 6.0)
- **Best puzzle architecture**: Pure-logic `PuzzleState` separated from scene nodes → maps to Go backend
- **Best test coverage**: Parameterized solver tests + full content sweep for solvability
- **Critical bugs**: `remove_empty_columns()` corrupts indices; solver is exponential O(b^d) with no transposition table; `dot_count()` hardcodes legend chars; save format has no checksum/atomicity
- **Adopt**: `.puzz`-style custom DSL with legend system, event-driven mode architecture, GdUnit4 test patterns

### godot_puzzle_dependencies (Score: 4.9)
- **Best editor addon pattern**: GraphEdit + GraphNode with undo/redo integration
- **Best data model**: DAG-based puzzle dependency graph with LOCKED/AVAILABLE/COMPLETED states
- **Critical flaws**: No cycle detection (!!!), data stored in ProjectSettings not Resource files, no runtime consumption path
- **Adopt**: GraphEdit visual editor for AxiomaticFabric, Graphviz DOT export, undo/redo in editor tooling
- **Adapt**: Must add cycle detection, topological sort validation, export to JSON consumed by Go backend

### Sudoku-puzzle-generator (Score: 3.2)
- **Best algorithm**: Correct backtracking with candidate pruning and diagonal-box prefill
- **Critical flaw**: `clean_numbers()` does NOT verify unique solution — generated puzzles are unvalidated
- **Adopt**: Backtracking + constraint propagation for challenge/glitch generation
- **Avoid**: Missing seed-based RNG, no solvability verification, flat-array index math

### competitive-programming (Score: 4.7)
- **Best algorithm reference**: 600+ solutions across 35 topics, correct textbook implementations
- **Critical quality issues**: Zero tests, global RNG (no seed), fixed-size arrays, C++ patterns (VLA, template, `#include <bits/stdc++.h>`) that need translation
- **Port**: Kadane, KMP, DSU, BFS/DFS, Dijkstra, sieve, balanced brackets, topological sort, trie
- **Skip**: union_find without path compression, stub topological_sort.py, Floyd-Warshall with V=4

### coding-challenges (Score: 4.6)
- **Best problem descriptions**: 150+ LeetCode solutions with READMEs → ready-made challenge content
- **Port**: Sliding window, two-pointer, LRU cache, merge intervals, bucket-sort top-k
- **Skip**: Buggy two-sum (`[-1. -1]`), overly-specific platform problems

### godot_recipes (Score: 3.9)
- **Best patterns**: "Call down, signal up" architecture, grid-based movement, steering behaviors, screen shake
- **Critical gaps**: No WebSocket patterns (missing for our architecture), no Resource-based data modeling, no signal bus / EventBus pattern, old Tween API, Godot 3 holdovers
- **Adapt**: Grid movement → deterministic puzzle-piece placement; steering seek → smooth interpolation; heart containers → entropy/vigilance display

---

## Adopt / Adapt / Reference / Reject Matrix

### ADOPT (use as-is or with minimal wrapper)

| Pattern | Source | Location |
|---------|--------|----------|
| Pure-logic puzzle state machine | dothop | `PuzzleState.gd` → Go struct |
| Backtracking + constraint propagation | Sudoku-gen | `fill_remaining()` + `get_candidates()` |
| DAG dependency graph (Thing/Board) | puzzle_deps | `Thing.gd` + `Board.gd` data model |
| Custom DSL with legend system | dothop | `.puzz` PRELUDE/LEGEND/PUZZLES format |
| Event-driven mode subscription | dothop | `ClassicMode`/`RandomMode` subscribing to win events |
| Parameterized solver + full-content-sweep tests | dothop | `puzzle_analysis_test.gd` |
| "Call down, signal up" node communication | godot_recipes | `basics/node_communication.md` |
| Self-contained addon structure | gdextension | `project/addons/example/` |
| Graphviz DOT export for dependency graphs | puzzle_deps | `export.gd` |
| TileMap-based grid rendering (not 81 nodes) | GodotSudoku | TileMap approach (fix the code, keep the idea) |
| Steering/seek behavior for entity movement | godot_recipes | `ai/homing_missile.md` seek() |
| Screen shake with noise for feedback | godot_recipes | `2D/screen_shake.md` |

### ADAPT (redesign for our architecture, don't copy)

| Pattern | Source | How to adapt |
|---------|--------|-------------|
| godot-go GDExtension binding pattern | godot-go | Use only codegen ideas for type binding; our WebSocket separation avoids cgo crashes entirely |
| 4-level GDExtension init pattern | gdextension | Adapt to Go initialization sequence if we use GDExtension later; skip for WebSocket architecture |
| godot-go class registration with generics | godot-go | Use for bridge types if we add GDExtension; not needed for current architecture |
| .puzz custom resource format | dothop | Port to Go-side parser + embed challenge pool at build time |
| GraphEdit visual dependency editor | puzzle_deps | Build as Godot addon that outputs JSON for Go backend. Add cycle detection (required). |
| dothop solver with transposition table | dothop | Replace O(b^d) DFS with BFS/IDA* + memoization. Add benchmark gate. |
| AudioManager pool with preloaded resources | godot_recipes | Fix the `load()`-per-play bug (preload into Dictionary) |
| Grid movement with tween interpolation | godot_recipes | Extract as reusable `snap_and_tween()` for puzzle-piece placement |
| Heart containers (partial mode) | godot_recipes | Replace textures with era-themed icons; use for entropy/vigilance display |
| Debug overlay with property registration | godot_recipes | Adapt to display WebSocket latency, entropy, vigilance, tick count |
| Cooldown button (radial TextureProgressBar) | godot_recipes | Use for action-ability cooldowns in Architect/Ghost/Saboteur modes |

### REFERENCE ONLY (study pattern, don't port)

| Pattern | Source | Why reference |
|---------|--------|--------------|
| SConstruct build auto-discovery | gdextension | SCons pattern not applicable to Go project, but auto-discovery idea is worth adopting for our build system |
| Competitive programming solutions | comp-prog / coding-challenges | Reference implementations for porting algorithms to Go VM builtins; don't copy code directly |
| C-heap allocation for cgo-crossed structs | godot-go | Only relevant if we ever use cgo; bookmark for future |
| Singletion lifecycle guards (ERR_FAIL_COND) | gdextension | Good pattern but expressed in C++; adapt the idea to Go's error handling |
| 4-level module initialization | gdextension | Standard Godot init pattern to understand if we ever write a GDExtension |
| Pandora entity/PuzzleWorld composition | dothop | Interesting approach to composing puzzles/themes/icons but over-coupled for our needs |

### REJECT (do not use, anti-pattern for our architecture)

| Pattern | Source | Reason for rejection |
|---------|--------|---------------------|
| `log.Panic` as error handling | godot-go | Kills entire Godot process on any type mismatch. Return errors instead. |
| dot-imports | godot-go | Hides symbol origin; makes code unreadable |
| `printStacktrace()` on every C bridge call | godot-go | Unconditional debug output = 100x slowdown in production |
| Reflection-based method dispatch on every call | godot-go | Allocations per call; no inlining. Use generated switch statements instead. |
| Godot 3 Tween API (`interpolate_property` + `yield`) | godot_recipes | Will break in Godot 4.3+. Use `create_tween().tween_property()` + `await`. |
| `rect_position` / `rect_size` | godot_recipes | Removed in Godot 4. Use `position` / `size` on Control nodes. |
| `is_OK()` with `&&` instead of `||` in box validation | GodotSudoku | Bug: skips 4 additional cells. Box constraint check must use `||`. |
| `clean_numbers()` without unique-solution verification | Sudoku-gen | Produces invalid/unsolvable puzzles. Generator must verify uniqueness via a separate solver pass. |
| 81 Panel instances for a 9×9 grid | GodotSudoku | 81 Godot nodes = 81× overhead. Use TileMap (1 node) or `_draw()`. |
| Save format without checksum/atomicity | dothop | `SaveGame.gd` writes JSON without atomic swap or CRC. Partial write = data loss. |
| Events global mutable state (`Events.puzzle_node.current`) | dothop | Late-bound listeners get stale references. Use immutable event structs. |
| Puzzle data stored in ProjectSettings | puzzle_deps | Blows up `project.godot` with large data. Use `Resource` files instead. |
| Package-level StringName globals with manual lifecycle | godot-go-demo-projects | Easy to double-destroy/miss-destroy. Use lazy initialization or store on struct. |
| Global RNG without seed control | Sudoku-gen, godot-coding-challenge | Non-deterministic output breaks replay system. Always seed RNG. |

---

## Prioritized Implementation Backlog

### Tie-break legend: (Impact / Difficulty / Risk)

### HIGH priority, LOW difficulty — immediate wins

| Priority | Task | Source | Value |
|----------|------|--------|-------|
| H/0.1 | Fix DSU without path compression in port (use comp-prog dsu.cpp, not union_find.cpp) | comp-prog | Determinism + correctness |
| H/0.1 | Make all RNG seed-based in builtins | (cross-cutting) | Replay determinism |
| H/0.2 | Port Kadane's algorithm as ARRAY_MAX_SUB builtin | comp-prog | Architect mode array challenges |
| H/0.2 | Port KMP string search as STRING_FIND builtin | comp-prog | Magitech rune string matching |
| H/0.2 | Port DSU with path compression + union by rank | comp-prog | Faction detection, connectivity |
| H/0.2 | Port balanced brackets parser | comp-prog | Magic circle validation |
| H/0.3 | Port BFS/DFS as GRAPH_FLOOD_FILL builtins | comp-prog | Grid-based dungeon puzzles |
| H/0.3 | Port sieve/primality as builtin | comp-prog | Magitech rune crafting |
| H/0.3 | Implement stable sort (merge sort) as default | comp-prog | Deterministic sorting |
| H/0.3 | Add `ALLOC` opcode for 2D/3D arrays in VM | comp-prog | Grid/challenge data structures |
| H/0.3 | Add `INDEX2D` opcode for flattened grid access | comp-prog | Grid addressing without manual index math |
| H/0.3 | Port sliding window (longest substring) builtin | coding-challenges | Cyberpunk packet analysis puzzles |

### HIGH priority, MEDIUM difficulty — core systems

| Priority | Task | Source | Value |
|----------|------|--------|-------|
| H/0.5 | **Implement challenge generator with unique-solution verification** | Sudoku-gen (redesign) | Guarantees golden path exists; critical for procedural generation |
| H/0.5 | **Build AxiomaticFabric event editor as Godot addon** | puzzle_deps GraphEdit | Visual editing of condition→effect→fallback chains |
| H/0.5 | **Port dothop .puzz parser to Go** | dothop | Custom puzzle format with legend system for Rune DSL |
| H/0.5 | **Add cycle detection to dependency/event graphs** | puzzle_deps (missing feature) | Must reject cyclic condition chains in AxiomaticFabric |
| H/0.5 | Implement topological sort on event graphs | puzzle_deps (missing feature) | Correct event chain execution order |
| M/0.5 | Port Dijkstra (integer weights, priority queue) | comp-prog | Cyberpunk network shortest path |
| M/0.5 | Port 0/1 Knapsack as DP builtin | comp-prog | Resource optimization challenges |
| M/0.5 | Port LRU cache (doubly-linked-list + hashmap) | coding-challenges 146 | Cyberpunk memory management puzzles |
| M/0.5 | Port merge intervals (greedy sort) | coding-challenges 56 | Scheduling/optimization puzzles |

### MEDIUM priority, MEDIUM difficulty — important but not blocking

| Priority | Task | Source | Value |
|----------|------|--------|-------|
| M/0.4 | Port topological sort (Kahn's algorithm) | comp-prog | Dependency resolution puzzles |
| M/0.4 | Port coin change (unbounded + 0/1) | comp-prog | Currency/score optimization |
| M/0.4 | Port edit distance | comp-prog | String mutation (Cyberpunk DNA hacking) |
| M/0.4 | Port TRIE with prefix search | comp-prog | Lexicon validation, autocomplete |
| M/0.4 | Port GCD/LCM, fast exponentiation | comp-prog | Rune harmony math |
| M/0.4 | Port BST operations | comp-prog | Grimoire indexing for Magitech |
| M/0.5 | Port ACM-3 / forward checking for constraint propagation | dothop (redesign) | Chain reaction / glitch propagation with pruning |
| M/0.5 | Add BFS/IDA* solver with transposition table | dothop (redesign) | Replace exponential DFS; benchmark gate required |
| M/0.5 | Implement WebSocket client in Godot with exponential backoff | godot_recipes (missing) | Critical missing pattern; needed for client-server communication |
| M/0.5 | Implement EventBus autoload for Godot side | godot_recipes (missing) | Decouples WebSocket handler from UI; needed for maintainable client |
| M/0.5 | Implement Resource-based data contracts matching Go structs | godot_recipes (missing) | Type-safe data sharing between Godot and Go |

### LOW priority — nice-to-haves

| Priority | Task | Source | Value |
|----------|------|--------|-------|
| L/0.3 | Screen shake with noise for entropy feedback | godot_recipes | Polish; useful for immersion |
| L/0.3 | Debug overlay for WebSocket latency + tick count | godot_recipes | Developer tooling |
| L/0.3 | Heart containers (partial) for vigilance meter | godot_recipes | UI polish |
| L/0.4 | Steering/seek behavior for entity movement | godot_recipes | Smooth interpolation of puzzle entities |
| L/0.5 | Grid movement with tween interpolation | godot_recipes | Puzzle-piece placement animation |
| L/0.5 | AudioManager pool with preloaded resources | godot_recipes | Audio feedback on actions |
| L/0.5 | Cooldown buttons for action abilities | godot_recipes | Ghost/Saboteur mode UI |
| L/0.5 | Count sort, bucket sort builtins | comp-prog | Cosmic era bounded-integer puzzles |
| L/0.6 | Floyd-Warshall with reasonable size limits | comp-prog | All-pairs routing (Cosmic spatial mapping) |
| L/0.6 | SCC Kosaraju | comp-prog | Cosmic fabric analysis |
| L/0.7 | Kruskal MST | comp-prog | Network optimization puzzles |
| L/0.7 | Menger sponge fractal subdivide | godot-coding | Procedural dungeon generation (long-term) |

---

## Key Architecture Insights

### 1. The WebSocket separation is the right call
godot-go's `log.Panic` + reflection dispatch + debug stacktraces confirm that in-process GDExtension in Go is fragile. Your WebSocket architecture:
- Avoids cgo crashes killing the Godot process
- Allows independent debugging (Go tests run without Godot)
- Enables replay testing of the entire backend
- Supports hot-reload of the Go backend

### 2. The biggest missing system: a proper constraint solver
dothop's exponential DFS, Sudoku-gen's missing unique-solution verification, and GodotSudoku's buggy `is_OK()` all point to the same gap: **none of these repos have a production-grade constraint solver**. Your AxiomaticFabric needs:
- AC-3 or forward checking for glitch/chain-reaction propagation
- BFS/IDA* with transposition table for challenge validation
- Cycle detection in event graphs (mandatory — puzzle_deps lacks this)
- Unique-solution verification in the challenge generator

### 3. The best puzzle architecture model is dothop's PuzzleState
- `PuzzleDef` (data-only Resource) + `PuzzleState` (pure-logic state machine) + scene nodes (visualization) = clean three-layer separation
- Maps perfectly to your architecture: Go backend owns PuzzleState simulation, Godot client owns visualization
- dothop's solver-test + content-sweep pattern should be mandatory for every era's challenge pool

### 4. Editor tooling is underinvested across all repos
Only puzzle_deps has a real editor addon, and it lacks cycle detection, topological sort, and runtime consumption. An AxiomaticFabric event editor built on GraphEdit + GraphNode — with proper validation — would be uniquely valuable for puzzle authoring.

### 5. Algorithm porting guide
| VM Use Case | Algorithm | Priority |
|------------|-----------|----------|
| Array analysis | Kadane, two-pointer, sliding window | H |
| String search | KMP | H |
| Graph traversal | BFS, DFS, Dijkstra (int weights) | H |
| Discrete math | DSU, GCD, modpow, sieve | H |
| Data structures | Trie, BST, LRU, balanced stack | H |
| Dependency resolution | Topological sort, DSU | M |
| DP puzzles | Knapsack, coin change, edit distance | M |
| Spatial partitioning | QuadTree, spatial hash | L (future) |
| Pathfinding | A* with JPS | L (future) |

---

## Things to Avoid (Summary of Rejects)

1. **`log.Panic` in production** — return errors to the caller. A Godot process crash due to a GDScript type mismatch is unacceptable.
2. **Exponential DFS without memoization** — dothop's `collect_move_tree` proves O(b^d) is not viable. Use transposition tables or IDA*.
3. **Unvalidated puzzle generation** — Sudoku-gen's `clean_numbers()` without unique-solution verification will ship unsolvable content. Always run a solver pass.
4. **`&&` instead of `||` in box validation** — GodotSudoku's bug is easy to replicate. Use a correct constraint check matrix.
5. **81 Panel nodes for a grid** — Use TileMap (1 node) or `_draw()` (1 node). 81 separate scene instances is unacceptable.
6. **Save format without atomic write / checksum** — dothop's `SaveGame.gd` will corrupt on crash. Use temp file + rename + CRC.
7. **Godot 3 API in Godot 4 code** — `rect_position`, old Tween, `yield()` will break. Lint for these.
8. **Package-level global mutable state** — `Events.puzzle_node.current` (dothop), package-level StringNames (godot-go-demos). Encapsulate state in structs/objects.
9. **Global RNG without seed** — Every algorithm using randomness must accept a seed parameter for deterministic replay.
10. **Data in ProjectSettings** — puzzle_deps storing board data in ProjectSettings bloats `project.godot`. Use `Resource` files.

---

## Final Verdict

### Most Valuable Repos (for our specific architecture)
1. **competitive-programming** (9/10 suitability) — best algorithm reference for porting to Go VM
2. **godot_puzzle_dependencies** (7/10) — best editor addon + data model (with fixes)
3. **dothop** (6/10) — best puzzle architecture + test patterns (with fixes)
4. **coding-challenges** (7/10) — ready-made challenge content + problem descriptions
5. **godot_recipes** (5/10) — useful Godot patterns (with heavy fixes for version correctness)

### Most Important Missing Capability
**Constraint propagation with cycle detection** — none of the repos implement this correctly. Your AxiomaticFabric requires it for deterministic chain reactions, glitch propagation, and dependency validation. This is the highest-impact engineering investment.

### Things This Analysis Revealed
- The C2U architecture (Go simulation + Godot visualization via WebSocket) is structurally superior to any GDExtension approach for determinism and debuggability
- Procedural generation without unique-solution verification is dangerous — always validate generated content
- The competitive programming repo is more valuable as an algorithm spec library than as importable code
- dothop's test suite is the gold standard for puzzle game testing
- All repos lack proper error boundaries, cycle detection, and deterministic replay considerations
