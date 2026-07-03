---
Status: Active
Implementation: 50% (Phases 0–8 complete except REPLAY_LOAD, Phase 9 foundation complete)
Confidence: Authoritative
Last Updated: 2026-07-03
---

# Project Chrysalis — Development Roadmap

Status key: ✅ Complete · 🔶 Partial / Foundation Only · ⬜ Not Started

---

## Phase 0 — Vision & Research ✅

**Objective:** Define exactly what game is being built.

Deliverables: VISION.md, GDD, TDD, architecture diagrams, P-Script spec, coding standards, performance targets, determinism spec, milestone roadmap.

Exit criteria: entire project architecture agreed upon.

---

## Phase 1 — Core Engine ✅

**Objective:** Deterministic simulation engine that produces identical state from the same seed.

Systems built:
- Fixed-point math library (`crysmath`, precision 10⁶)
- Seeded deterministic RNG (per-engine instance)
- Tick scheduler (10 Hz, `BeginTick` / `CommitTick`)
- ECS/Registry (data-oriented column arrays, dynamic capacity)
- Double-buffered spatial grid
- Entity lifecycle (spawn, inert, fabrication)

Exit criteria: ✅ Same seed → identical simulation every run (verified by test suite with race detector).

---

## Phase 2 — World Simulation ✅

**Objective:** Living autonomous colony without player input.

Systems built:
- Resource generation (seeded placement, `ResourceCount` per cell)
- Base (center tile, `IsBase`, home pheromone source)
- Drone spawning at base
- Pheromone trails (Home, Resource, Alien — decay per tick)
- Harvesting and cargo
- Depositing at base
- Battery drain and inert state
- Replication via fabrication pool

Exit criteria: ✅ Autonomous colony harvests, deposits, and fabricates without P-Script.

---

## Phase 3 — P-Script ✅

**Objective:** Replace hardcoded AI entirely with player-authored code.

Components built:
- Lexer (tokens, identifiers, operators)
- Pratt parser (infix precedence, if/while/fn/let)
- AST node types
- Bounded AST interpreter (while loop cap, per-drone variables)
- Builtin registry

Builtins: `SENSE_RESOURCE`, `SENSE_HOME`, `SENSE_BATTERY`, `SENSE_TRUST`, `SENSE_CORRUPTION`, `SENSE_COMPROMISED`, `SENSE_ALIEN_SIGNAL`, `SENSE_SWARM_SIZE`, `SENSE_COLONY_RESOURCES`, `SENSE_CARGO`, `HARVEST`, `DROP_RESOURCE`, `MOVE_RANDOM`, `MOVE_TOWARDS_RESOURCE`, `MOVE_TOWARDS_HOME`, `BROADCAST_VOTE`

Exit criteria: ✅ Entire swarm behavior driven by `scripts/agent.ps`.

---

## Phase 4 — Runtime Programming ✅

**Objective:** Player edits code while simulation continues; no pause required.

Features built:
- File watch hot-reload (mod-time polling per tick)
- WebSocket `COMMAND_INJECTION` — remote AST replacement
- Parse validation before swap (rollback on error)
- Error reporting to stderr

Exit criteria: ✅ Simulation never pauses during code replacement; invalid code is rejected without interrupting the run.

---

## Phase 5 — Visualization ✅

**Objective:** Entire simulation observable through Godot client.

Godot screens built:
- Telemetry dashboard (swarm metrics)
- Swarm view (drone positions, states)
- Resource / logistics overlay
- Pheromone view
- Hazard monitor
- Alien detector
- Uplink terminal (P-Script editor + injection)
- Research tree (stub)
- Structure manager (stub)
- Replay controls (UI shell — awaiting Phase 8 backend)
- WebSocket bridge with reconnect timer

Exit criteria: ✅ Full simulation visible; terminal banner renders VICTORY/DEFEAT.

---

## Phase 6 — Event Architecture ✅

**Objective:** Simulation emits a canonical, immutable event stream that all downstream systems can consume independently.

Systems built:
- `EventBus` — double-buffer, `BeginTick` invariant guard, `Commit` / `Events`
- 9 typed event types with sealed `EventPayload` interface
- `UnmarshalJSON` dispatch for JSON round-trips
- Centralized emit helpers (`emitHarvest`, `emitDeposit`, `emitDeath`, `emitInfection`, `emitHazardDamage`, `emitMissionChanged`, `emitFabricated`, `emitSpawn`)
- `DecisionFrame` — immutable per-drone per-tick decision log; private `traceBuilder` inside interpreter
- `EvalTraced` — returns completed frame without exposing mutable state
- `replay.Recorder` — external consumer; `Record` / `Checkpoint` / `Serialize` / `Deserialize` / `NearestCheckpoint` / `EventsInRange`

Consumers receiving event stream: WebSocket network layer (`"events"` in every `EMISSION_SNAPSHOT` packet), `replay.Recorder`.

Exit criteria: ✅ Simulation emits canonical event stream; snapshots are immutable after `Commit`; same seed produces byte-identical JSON event stream across runs.

---

## Phase 7 — Debugging Infrastructure ✅

**Objective:** Every drone decision is explainable through in-game tooling.

**Go side (complete):**
- `DecisionFrame` emitted per inspected drone each tick via `EvalTraced`
- `INSPECT_DRONE` / `INSPECT_CLEAR` WebSocket commands handled in `main.go`
- `"trace"` field in `EMISSION_SNAPSHOT` payload when a drone is being inspected

**Godot side (complete):**
- `drone_inspector.gd` sends `INSPECT_DRONE` on drone selection, `INSPECT_CLEAR` on deselect or screen exit
- `load_trace(trace)` renders the `DecisionStep` list: `✓/✗ [IF] condition → result` for branches, `[•] action → result` for calls
- `load_events(events)` accumulates per-drone event timeline (filtered by `drone_id`, rolling 50-event buffer, most-recent-first)
- Live variable watch: battery, trust, state, position, cargo update every tick via `load_swarm`
- `game_hub.gd` routes `"trace"` and `"events"` fields from every `EMISSION_SNAPSHOT` to the active drone inspector screen
- Conditional breakpoint (battery threshold) already in `main.gd` via `_trigger_breakpoint_halt`

Exit criteria: ✅ Player can select any drone, see its last tick's decision trace, and filter its full event history.

---

## Phase 8 — Replay System 🔶

**Objective:** Any tick in any run is reconstructable and seekable.

**Complete:**
- `replay.Recorder` — `Record`, `Checkpoint` (copies state), `NearestCheckpoint`, `EventsInRange`, `Serialize`, `Deserialize`
- `ReplayMeta` struct (seed, version, width, height, drone_count, started_at) — embedded in `.chrysalis_replay` JSON
- Always-on recording in `main.go`: events recorded every tick, full world-state checkpoint every 500 ticks
- `REPLAY_SEEK {"tick":N}` WebSocket command → engine finds nearest checkpoint ≤ N, broadcasts synthetic `EMISSION_SNAPSHOT` with checkpoint world state + events from checkpoint to N
- `REPLAY_SAVE {}` command → serializes archive to `replay_<timestamp>.chrysalis_replay`
- `"replay"` field in every `EMISSION_SNAPSHOT` (`{recording, total_ticks, current_tick}`)
- `replay_controls.gd` signals wired to WebSocket commands: seek, step forward/back, reset, save
- `replay_controls.gd` event log rendered from seek response events, newest-first
- `game_hub.gd` routes `data["replay"]` directly to `load_replay()` (fixed wrapping bug)
- `Checkpoint()` performs shallow copy — callers cannot corrupt stored state by mutating the map after the call

**Phase 8A — Engine.SetState() (complete):**
- `DetRNG` wrapper in `simulation/rng.go`: tracks `(seed, callCount)`, restores any historical RNG position by fast-forwarding from seed in O(callCount)
- `Engine.Seed` field stored alongside `Engine.rng`
- `GetState()` expanded: `rng_seed`, `rng_calls`, `total_deposited`, `historical_total`, `hazard.intensity`, `grid.base` flag — all fields needed for bit-perfect restoration
- `Engine.SetState(map[string]interface{})` in `simulation/setstate.go`: handles both in-memory types and JSON-decoded `float64` values via coercion helpers
- `Engine.WorldHash() uint64`: FNV-64a over tick + economy + all drone positions/batteries/states — determinism oracle
- `Checkpoint{WorldHash uint64}` — stored per checkpoint for replay validation
- Test suite: `TestSetStateRoundTrip`, `TestSetStateResumesSimulation`, `TestWorldHashDeterminism`, `TestRNGRestoreProducesIdenticalSequence`, `TestSetStatePreservesRNGPosition` — all pass

**Phase 8B — True replay reconstruction (complete):**
- `REPLAY_SEEK` command now: saves live state → `SetState(checkpoint)` → forward-simulates to target tick → broadcasts reconstructed state → restores live state
- Replay is fully faithful; world state at any sought tick is reconstructed, not approximated from the nearest checkpoint
- WorldHash logged on every seek for determinism auditing

**Phase 8C — Replay validation (complete):**
- `WorldHash()` expanded to cover ALL canonical state: engine scalars, drone registry (including `Compromised`, `ID`), grid (sparse by cell index), hazards (all slots + intensity), alien network (all slots), mission status/targets, RNG position (seed + callCount)
- Non-canonical state explicitly excluded by contract: EventBus, DecisionFrames, inspector ID, recorder
- 7 coverage tests (one per canonical subsystem): each mutates exactly one field and asserts hash changes — if WorldHash has a gap, a test fails
- `stepEngine(engine, program, interp, inspectID)` extracted as shared function; live simulation and replay forward-pass both call it — identical code path, no divergence from branching
- `REPLAY_SEEK` validates `WorldHash()` immediately after `SetState(checkpoint)` vs stored `cp.WorldHash`; logs `[REPLAY VALIDATION FAIL]` to stderr if they differ

**Remaining (Phase 8D):**
- `REPLAY_LOAD {filename}` command: deserialize `.chrysalis_replay` from disk, enter replay-only streaming mode (simulation paused)

Architecture:
```
Initial Snapshot → Events → Events → Checkpoint → Events → Replay
```

Exit criteria: Player can scrub to any tick in a completed run; replay reconstructs identical world state to original.

---

## Phase 9 — Mission Framework 🔶

**Objective:** Content creators build missions without touching engine code.

**Complete:**
- `MissionState` struct with victory/defeat/tick-limit conditions
- `EvaluateMission` with `EventMissionChanged` emission
- `level.json` schema defined: world dimensions, seed, resource placement, hazard placement, alien nodes, mission targets (target_resources, max_ticks, infection_loss_threshold), unlocked builtins, narrative strings
- `levels/level.go` — `Level` type, `LoadLevel(path)`, `Level.CreateEngine()`: produces a fully wired `*simulation.Engine` from JSON alone
- `simulation.NewBaseEngineWithSeed` — engine constructor with no hardcoded hazards/aliens; used by `CreateEngine()` so JSON fully controls world content
- `main.go` reads `PHX_LEVEL_PATH` env var; falls back to legacy defaults if unset
- `ReplayMeta.LevelID` — replay archives record which level was played
- `chrysalis_1.json` — canonical first level definition (10 drones, 100×100, magnetic hazard + alien infector, target 10 silicates, 3000 tick limit)

**Remaining:**
- Campaign progression (unlock next level on victory, persist between sessions)
- Mission scripting hooks (narrative event triggers tied to EventBus events)
- Reward/unlock system tied to mission completion
- `REPLAY_LOAD {filename}` command (Phase 8D — deserialize `.chrysalis_replay` from disk)

Exit criteria: Adding a new mission requires only a JSON file; no Go or GDScript changes needed. ✅ **Exit criteria met** for content authoring; campaign progression and scripting hooks are stretch goals.

---

## Phase 10 — Environment ⬜

Dynamic world that actively challenges player algorithms.

Systems: Radiation zones, thermal damage, magnetic storms, EMP bursts, cave collapse, resource depletion curves, day/night cycle.

Depends on: Phase 9 (environment defined in mission JSON).

Exit criteria: World state changes during a run in ways that require algorithmic adaptation.

---

## Phase 11 — Communication Layer ⬜

Distributed coordination mechanics.

Systems: Explicit broadcast range, signal strength and decay, relay drones, directional messages, voting protocols, pheromone amplification.

Concepts introduced: message passing, leader election, distributed consensus.

Depends on: Phase 9.

Exit criteria: Player can write protocols that coordinate swarm decisions across spatial distance.

---

## Phase 12 — Trust System ⬜

Information reliability becomes a strategic resource.

Systems: Per-drone reputation scores, trust decay, quorum voting with Byzantine detection, identity spoofing, trust recovery protocols.

Gameplay question: *Who should be believed?*

Depends on: Phase 11 (trust flows through communication channels).

Exit criteria: Compromised drones can mislead the swarm; player can write detection and quarantine algorithms.

---

## Phase 13 — Knowledge Graph ⬜

Shared colony memory with confidence and decay.

Stores: resource locations, hazard positions, alien activity, exploration coverage, confidence scores, source drone ID, timestamp.

Supports: memory decay, contradiction resolution, confidence-weighted consensus.

Depends on: Phase 12 (trust determines confidence weight of incoming facts).

Exit criteria: Colony maintains a queryable shared world model; stale facts decay; conflicting reports require resolution.

---

## Phase 14 — Alien AI ⬜

Transform the alien from a passive hazard into an algorithmic adversary.

Abilities: counterfeit pheromone trails, trust poisoning, vote manipulation, drone hijacking, network partitioning, fake broadcasts, adaptive policy learning.

Goal: algorithm vs. algorithm.

Depends on: Phase 13 (alien attacks the knowledge graph and trust layer).

Exit criteria: Alien adapts to player strategies; no single P-Script policy defeats it permanently.

---

## Phase 15 — Colony Evolution ⬜

Unlockable hardware, sensors, protocols, and drone specialization.

Roles: Scout, Miner, Builder, Relay, Defender, Researcher.

New P-Script builtins unlocked per mission tier.

Depends on: Phase 14 (specialization driven by adversarial pressure).

---

## Phase 16 — Campaign ⬜

Three-act narrative structured around the emergent systems.

**Act I — Wonder:** Learn emergence. 3-line policy builds a colony.

**Act II — Mastery:** Optimization, multi-colony logistics, hazard routing, distributed consensus.

**Act III — War:** Alien AI. Information warfare. Adaptive algorithms. Final distributed battle.

Depends on: Phases 9, 10, 11, 12, 13, 14, 15.

---

## Phase 17 — Content Expansion ⬜

40–60 campaign missions, sandbox maps, challenge missions, tutorials, endless mode, procedural map generator.

Depends on: Phase 9 (mission JSON schema complete).

---

## Phase 18 — Polish ⬜

**Technical:** profiling, memory budgets, save migration, crash reporting, accessibility.

**Presentation:** music, SFX, particle effects, animation, UI polish, GPU pheromone shader.

---

## Phase 19 — Modding ⬜

Custom P-Script libraries, mission editor, campaign editor, custom hazard types, custom drone hardware, plugin API, Steam Workshop integration.

---

## Phase 20 — Multiplayer (Optional) ⬜

Shared swarm co-op, competitive colonies, spectator mode, replay sharing, leaderboards.

*Must come after single-player campaign is complete and stable.*

---

## Phase 21 — Release ⬜

Closed alpha → open alpha → closed beta → open beta → Steam launch. Documentation, marketing, community Discord, Workshop integration.

---

## Phase 22 — Live Operations ⬜

Balance patches, new campaigns, seasonal challenges, community maps, DLC, performance improvements.

---

## Dependency Graph

```
Phase 0  Vision
     │
     ▼
Phase 1  Core Engine
     │
     ▼
Phase 2  World Simulation
     │
     ▼
Phase 3  P-Script
     │
     ▼
Phase 4  Runtime Programming
     │
     ▼
Phase 5  Visualization
     │
     ▼
Phase 6  Event Architecture
     │
     ▼
Phase 7  Debugging Infrastructure
     │
     ▼
Phase 8  Replay System
     │
     ▼
Phase 9  Mission Framework             ← YOU ARE HERE
     │
     ├───────────────────┐
     ▼                   ▼
Phase 10 Environment  Phase 11 Communication
     │                   │
     └────────┬──────────┘
              ▼
        Phase 12 Trust
              │
              ▼
     Phase 13 Knowledge Graph
              │
              ▼
       Phase 14 Alien AI
              │
              ▼
    Phase 15 Colony Evolution
              │
              ▼
       Phase 16 Campaign
              │
              ▼
   Phase 17 Content Expansion
              │
              ▼
       Phase 18 Polish
              │
              ▼
       Phase 19 Modding
              │
              ▼
  Phase 20 Multiplayer (Optional)
              │
              ▼
       Phase 21 Release
              │
              ▼
   Phase 22 Live Operations
```
