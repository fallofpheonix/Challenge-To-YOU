# Project Chrysalis — Architecture ↔ Implementation Gap Analysis

_Triage of the frozen design contracts against the current engine. Evidence is `file:line`.
Deliverable is a traceability matrix, not code. Status = rough % of the design contract met._

## Traceability matrix

| Decision / Contract | Status | Evidence | Required work | Priority |
|---|---|---|---|---|
| **Canonical state → WorldHash** | ✅ 90% | `simulation/setstate.go:145` FNV-64a over engine scalars, full drone registry, grid, hazards, aliens, mission, **RNG position** (`:235`). Clean canonical/non-canonical doc (`:143`). Replay stores WorldHash per checkpoint (`replay/recorder.go:44,122`). | No code gap today. **Discipline risk**: 3 not-yet-built systems (persistent memory, message queues, alien memory) must be added to `WorldHash()` when built. Add a registry checklist + a test that fails when an ECS column is unhashed. | P0 |
| **Determinism (no float / no map-iter / RNG)** | ✅ 85% | Fixed-point `crysmath`; `rng.go` wraps math/rand with call-count tracking for serialization; sim uses SoA slices, no map iteration in hot path; spatial hash deterministic (`spatial.go`). | **One float smell**: win/lose uses `float64` — `InfectedRatio() = float64(infected)/float64(count)` vs `float64` threshold (`mission.go:66,77`). Replay-safe (excluded from hash as derived) but convert to integer compare (`infected*100 >= pct*count`) to remove the last float from a gameplay-affecting condition. | P0→P1 |
| **P-Script arrays** | ❌ 0% | VM opcodes are Const/Load/Store/Call/arith/cmp/jump only (`vm/vm.go:12-31`). No `OpIndex`/`OpNewArray`. Vars are flat `[128]interface{}` (`vm.go:61`). | Array opcodes + indexed storage. | P0 |
| **Persistent drone memory** | ❌ ~5% | **`VM.Run()` resets `Variables` to `nil` every tick** (`vm.go:76-78`). Variables are per-tick scratch; nothing survives to next tick. | Per-drone persistent memory block in ECS, serialized, hashed into WorldHash, with an **explicit budget** (e.g. 32 registers + 256 array slots). | P0 |
| **Typed messaging (SEND/RECEIVE/MessageType)** | ❌ ~10% | Only `BROADCAST_VOTE` builtin (`main.go:378`). No SEND/RECEIVE, no inbox/queue, no `MessageType` enum, no delivery ordering. | Deterministic queue + typed `MessageType` enum + per-drone inbox in ECS, hashed. (Review §3: type messages, don't use raw ints.) | P0 |
| **Adaptive alien AI** | ❌ ~20% | `alien.go` is pure reactive proximity infection spread; `NodeJammer` declared but unused (`alien.go:11`). No FSM, perception, memory, utility, or planner. | Below even Gen-1. Build **Gen-1 deterministic FSM** (Expand/Harvest/Attack) first, then utility, then memory-driven. | P1 |
| **Telemetry (delta / subscription / LOD / schema ver)** | ❌ ~10% | Full-state `GetState()` builds `map[string]interface{}` every broadcast (`simulation.go:234`); `hub.go` broadcasts to all clients (`hub.go:74`). No deltas/subscriptions/LOD/version. | Delta channels + subscriptions + `Telemetry Schema v1` tag independent of replay/engine version. | P1 |
| **Mission archetypes** | ⚠️ 25% | Single win/lose: target resources + infection ratio + max ticks (`mission.go`). Levels are JSON (`levels/chrysalis_campaign.json`, `_2`, `_3`). | Generalize toward the 7 archetypes + progression rules. | P2 |
| **Scoring vector** | ❌ 0% | No mission score exists. `TrustScore` (`ecs.go:29`) is per-drone peer trust, unrelated. | Explicit vector: Completion, Robustness, Efficiency, Energy, DroneLosses, CodeSize, ExecTime → composite. | P2 |
| **Research → hardware (not syntax)** | ⚠️ presentation only | `client/ui/screens/research_tree.gd` renders tiers/cards/RESEARCH buttons from a data Dictionary; no engine-side unlock logic exists. | Define engine gating: Core APIs always on (MOVE/SENSE_*/HARVEST — already registered `main.go:351-386`); advanced APIs gated by researched **hardware**. | P2 |

## Notes / incidental findings

- **Duplicate builtin map** in `main.go` (~`:351-386` and `:393-428`) — two near-identical registrations. Confirm intent (interpreter vs VM split) or dedupe.
- Both an **interpreter** (`pscript/interpreter`) and a **VM** (`pscript/vm`) exist. Decide which is canonical for shipping; the VM is the allocation-free, budget-friendly path.

## Recommended sequence (dependency order)

1. **Foundational VM/memory** (P0): persistent memory block + array opcodes + budget → wire into WorldHash + SetState + replay.
2. **Deterministic message subsystem** (P0): typed queue/inbox → wire into WorldHash.
3. **Determinism cleanup** (P1): remove float from win condition.
4. **Adaptive alien Gen-1 FSM** (P1).
5. **Telemetry deltas + schema v1** (P1).
6. **Mission archetypes + scoring vector** (P2).
7. **Research hardware-gating** (P2).

Rationale: every higher-level feature depends on the canonical core staying bit-stable, so all new canonical state (memory, messages, alien memory) lands and gets hashed **before** the systems that consume it.
