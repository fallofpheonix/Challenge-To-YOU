# Project Chrysalis — Design Decision Record

Living record of the phased architecture review. Each phase is locked once its
questions are resolved. Later phases may add ADRs to `CONTEXT.md`; this file
holds the product/architecture intent behind them.

---

## ✅ Phase 0 — Vision & Product Definition (LOCKED)

| Item | Decision |
|------|----------|
| Player takeaway | "I wrote the mind of a swarm, then watched it out-think an alien intelligence I couldn't directly control." |
| Genres (identity) | Programming Game · Swarm Strategy · Distributed Systems Simulation |
| Genres (store tags) | `Programming` · `Strategy` · `Simulation` · `Automation` · `Colony Sim` |
| Core mechanic | Programming behaviors (P-Script). Debugging is the second pillar. |
| Act emotions | I: Wonder · II: Tension · III: Mastery |
| Audience | CS students / developers, **intermediate** programming assumed |
| Teaching goal | Teach *Chrysalis concepts* (swarm APIs, debugging, distributed systems), not basic programming |
| Success ranking | Correctness → Survival → Adaptability → Robustness → Efficiency → Output → Low complexity |
| Primary metric | **Composite Mission Score**, robustness-dominant; small code-quality bonus that never outweighs mission success |
| Scope | Single-player at launch; async competition via deterministic replays later. **Server never simulates.** |
| Modding | Mission Editor (validated JSON) + sandboxed P-Script sharing. Workshop/assets = post-launch. |
| USP | You never command a unit — you write the logic every drone runs, debug it live against adversarial AI, replay it to the exact tick. |

**Non-negotiables:** deterministic core · P-Script-only control · emergent swarm ·
adversarial alien AI · live debugging tools.

---

## ✅ Phase 1 — Engine & Technical Architecture (LOCKED)

| Item | Decision |
|------|----------|
| Determinism contract | **Cross-platform, same-version.** Bit-identical `WorldHash` across OS for a given engine build. Cross-version replay compat only via explicit migration. |
| Max swarm (engine capacity) | **5,000 drones @ 10 Hz.** Typical campaign 1,000–2,500; challenge/stress maps up to 5,000. |
| Tick execution | **Single-threaded by default.** Deterministic parallelism considered *only* after profiling proves the 100 ms budget is missed. |
| Backend (v1) | **Local replay/score files only.** No server ships in v1, but replay format is versioned, self-contained, and upload-ready. |
| Telemetry | **Delta + subscription + LOD.** Bandwidth proportional to scene *activity*, not swarm size. Client stays presentation-only. |

### Per-tick budget target (100 ms @ 10 Hz, 5,000 drones)
| System | Budget |
|--------|-------:|
| VM execution | 35 ms |
| Swarm simulation | 25 ms |
| Spatial queries | 15 ms |
| Pheromones | 10 ms |
| Hazards & Alien AI | 5 ms |
| Event generation | 5 ms |
| Misc | 5 ms |

### Derived engineering obligations
- **Independent version fields** on every replay: `engine`, `replay-format`, `level-schema`, `pscript`, plus `pscript-hash`, `seed`, `level-id`, `WorldHash`, tick count.
- **Golden-replay CI suite**: recorded replay → re-run → `WorldHash` must match, on every PR.
- **Profiling gate** before any parallelism: benchmark `Stress_5000` (5,000 drones, 10 min) must hold avg tick < 80 ms, worst < 100 ms on reference desktop.
- **Hardened P-Script sandbox**: no fs / net / process; memory + step limits (step limit exists at 1000/tick — revisit at 5k scale).
- **Upload-ready replay package**: `manifest.json` + `replay.bin` + `metadata.json` (+ optional thumbnail).
- If parallelism ever needed, parallelize read-heavy isolated systems first (spatial, pheromone diffusion, heatmaps); keep VM, state commits, EventBus, replay single-threaded.

### Telemetry channel model (locked)
Split the single stream into channels; the sim decides what changed, the client subscribes to what it needs:
| Channel | Rate | Payload |
|---------|------|---------|
| Global stats | 10 Hz | tick, drone count, resources, threat, mission state (~hundreds of bytes) |
| Visible drones | 10 Hz | delta updates for on-camera drones only |
| Inspector | 10 Hz | one subscribed drone: decision frame, vars, trace, events |
| Heatmaps | 2–5 Hz | pheromone / trust / danger / resource-density aggregates |
| Events | as they occur | EventBus (death, harvest, vote, infection) — naturally sparse |

Bandwidth ∝ activity, not swarm size. Never send unchanged drone state. LOD by camera distance
(full → reduced → density-only → nothing offscreen). Same model serves replay, spectating, and profiling.

---

## ✅ Phase 2 — P-Script Language & Swarm API (LOCKED)

| Item | Decision |
|------|----------|
| Expressiveness | **Fixed-size arrays + local vars + functions + loops + consts.** No structs, maps, dynamic allocation, heap, pointers, or recursion. "Deterministic, resource-bounded language for swarm intelligence" — not a small Python. |
| Drone memory | **Persistent per-drone local memory** (registers + fixed arrays). Same program on every drone; specialization *emerges* from local state, never assigned roles. |
| Coordination | **Stigmergy + deterministic local broadcast messages.** Pheromones stay primary/indirect; bounded messages add explicit consensus/negotiation/trust. No global writable blackboard. |
| Error model | **4-layer: compile-time → deploy-validation → isolated runtime fault → fatal engine invariant.** Faulted drone idles safely; fault streams to inspector + replay; hot-recovery without mission restart. |

### Language philosophy
P-Script is a **deterministic, resource-bounded systems language for autonomous agents**.
Specialization (scout/harvester/relay/defender) is an *emergent projection* of local memory +
decisions + communication, never a first-class role assignment. The colony itself is the debugger.

### Message-passing contract (must stay deterministic)
- API surface: `SEND(id, type, value)` · `BROADCAST(type, value)` · `RECEIVE()` · `HAS_MESSAGE()` · `MESSAGE_COUNT()`
- Fixed-size payloads only (sender, receiver, type, payload, TTL). No strings/objects.
- Sent at tick N, **delivered at N+1**, order = (tick, senderID, receiverID, sequence).
- Per-drone fixed inbox, max sends/tick, TTL, delivery only within comm radius, FIFO per sender.

### Derived engine obligations
- **Persistent drone memory is canonical state** → serialize in checkpoints, restore via `SetState`, include in replay + `WorldHash`, reset only on spawn/destroy. Touches `ecs.go`, `setstate.go`, `simulation.go`, `replay/`.
- **Array support** → compile-time size + bounds checking, new array opcodes (extends the 18-opcode set), stack/frame allocation. Touches `parser.go`, `vm/compiler.go`, `vm/vm.go`. (A `for` loop likely wanted alongside arrays — currently `while`-only.)
- **Message subsystem** → new deterministic sim system + 5 builtins; part of `WorldHash`/replay.
- **Fault model** → `FAULTED` drone state + safe-idle; structured `ScriptFaultEvent` (tick, droneID, line, instruction, fault type, message, call stack, memory snapshot) on the EventBus; inspector fault view; fault markers on the replay timeline; hot-redeploy resumes faulted drones. Fatal invariants (WorldHash divergence, VM corruption) halt immediately.
- **Runtime fault categories** (keep small/fixed): Execution (step budget), Memory (bounds), Arithmetic (div-zero), API (bad builtin), Communication (queue overflow), Trust (bad vote), Internal (VM invariant → fatal).

---

## ✅ Phase 3 — Adversary & Simulation Systems (LOCKED)

| Item | Decision |
|------|----------|
| Alien AI | **Deterministic adaptive intelligence.** Perceives canonical swarm state, keeps persistent internal memory, evaluates a library of strategies (EXPAND/INFILTRATE/ISOLATE/STARVE/DECEIVE/OVERWHELM), adapts *within a mission*. No ML, no RNG-drift, no cross-mission learning. |
| Defense | **Layered: distributed consensus (detect) + quarantine/recovery (respond).** Trust/quorum decides *who to believe*; quarantine (reboot/purify/destroy/observe/release) decides *what to do*. Alien attacks the pipeline via information warfare. |
| Economy | **Scarcity-driven, attrition-aware.** Key metric = net colony growth (fabrication − losses). Every drone is a meaningful investment; robustness beats throughput. |
| Failure model | **Resilience threshold + rapid iteration.** Fail only when recovery is impossible (swarm < minimum, base lost, objective deadline). On fail → straight to replay review + hot-redeploy retry, no punitive loss. Difficulty scales by shrinking recovery margin, not harsher punishment. Hardcore/ironman = post-launch optional. |

### ⚠️ Binding constraints on the adaptive alien (determinism-critical)
- **Utility/strategy scores MUST be `crysmath.FixedPoint`, never floats** (ADR-001). Illustrative `0.84`-style scores are floats and would break cross-platform `WorldHash`.
- **No cheating / no omniscience:** the alien may only read canonical simulation state it could legitimately sense (drone density, trails, trust graph, mission progress). No direct access to player script variables or hidden map data. Victory must come from better strategy, not privileged information.
- Alien perception + memory + active-strategy state are **canonical state** → in `WorldHash`, `SetState`, replay, checkpoints. Structurally mirrors the drone AI (perception → memory → planner → strategy library → executor).
- Campaign scales alien capability: Act I simple expand/infect · Act II observe+counter-strategy · Act III information warfare (false messages, false pheromones, consensus disruption, network partition).

### Layered-defense pipeline
`observe → suspicion/trust↓ → exchange info → quorum vote → consensus → quarantine/recovery`.
Quorum threshold is a player-tuned trade-off (low = fast but false positives; high = safe but slow spread). False positives carry a real economic cost under the scarcity economy — trust thresholds become genuine engineering decisions.

---

## ✅ Phase 4 — Content, Campaign & Progression (LOCKED)

| Item | Decision |
|------|----------|
| Length | **~12 hours, 3 acts × ~8 missions (~24).** Defined by *concepts mastered*, not playtime. Longevity from replay/score optimization + post-launch challenges + community content, not filler. (Resolves Phase 0 Q12.) |
| Progression | **Hybrid tiered.** Campaign gates *technology tiers*; within a tier the research tree lets players choose unlock order. |
| Progression rule | **Research tree unlocks Swarm APIs + drone hardware, NOT language syntax.** P-Script (arrays/loops/functions) is always available and stable. You research `SEND()`/`QUARANTINE()`/sensors, never `if`. |
| Balance rule | **No mission may require an optional unlock.** When a concept is taught, its whole tier is already available; optional unlocks add efficiency/strategy diversity, never gate winnability. |
| Missions | **7 canonical archetypes:** Harvest & Logistics · Exploration & Mapping · Survival & Defense · Detection & Consensus · Recovery & Self-Healing · Optimization · Adaptive Alien Encounter (act climaxes). Each mission = primary + secondary archetype + ≤1 new concept. |
| Onboarding | **Diegetic — Act I 'Wonder' IS the tutorial.** One concept per mission, narrative justifies each mechanic, no separate tutorial mode. Completed missions become replayable training sims; API reference always visible; hints explain *why it failed*, never *how to solve*. |

### Act pacing (target)
- **Act I — Wonder / Discovery (~3–4h):** movement → sensors → memory → arrays → multi-drone → pheromones → debugger → optimization. Goal: "I wrote simple code and a colony emerged."
- **Act II — Tension / Mastery (~4h):** hazards → trust → messaging → consensus → quarantine → scarcity → recovery → mixed. Goal: "My swarm survives because my algorithm is resilient."
- **Act III — Mastery / Conflict (~4h):** adaptive alien → information warfare → fake messages → trust manipulation → large colony → multi-front → final confrontation. Goal: "I beat another distributed intelligence with a better algorithm."

### Tech tiers (unlock order = campaign; intra-tier order = player)
1. Foundations (auto) · 2. Local Intelligence (memory/arrays/sensors) · 3. Coordination (pheromones/messaging/voting/trust) · 4. Resilience (quarantine/recovery/fabrication/defense policy) · 5. Advanced Swarm Intelligence (knowledge graph/distributed planning/optimization tools).

### Derived obligations
- Mission framework must express objectives per archetype + validate them (e.g. a Detection mission must contain compromised drones + a quorum objective) — feeds the future mission editor's templates + validation.
- Difficulty scales *within* archetypes by shrinking recovery margin (Phase 3), not by new mechanics.

---

## Roadmap status
- **Vision complete (Phases 0–4):** product identity → engine architecture → language → adversary/systems → campaign/progression. No new gameplay fundamentals remain to define.
- **Remaining phases** are execution, not invention:
  - **Phase 5 — Implementation & Tooling:** build order, mission-editor + debugger tooling, content pipeline, telemetry-channel + message-subsystem + persistent-memory implementation, fixed-point alien mind.
  - **Phase 6 — Testing, QA & Determinism Verification:** golden-replay CI, `Stress_5000` profiling gate, cross-platform WorldHash matrix, sandbox/fuzz tests for P-Script.
  - **Phase 7 — Release Engineering:** packaging, versioning/migration policy, store presence, upload-ready replay backend, post-launch (Workshop, challenges).
