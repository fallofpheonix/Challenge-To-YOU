# Challenge To YOU — Complete Implementation Plan

> **Implementation status (2026-07-12):** The core is built and verified — the
> Go backend builds, the full test suite passes, and all CI workflows are green.
> Two design details in this plan have evolved: the **sandbox** runs player code
> in a hardened host **subprocess** (WASM is intentionally deferred, as noted in
> this plan's own Post-MVP section), and the client↔backend link uses a
> **WebSocket** server. See [`TRACEABILITY-AND-CONFLICTS.md`](TRACEABILITY-AND-CONFLICTS.md)
> for the requirement→code coverage matrix and
> [`ARCHITECTURE-PHASE1.md`](ARCHITECTURE-PHASE1.md) for the authoritative architecture.

## Project Overview

**Challenge To YOU** is a desktop-first, roguelike hacking game where players solve procedurally generated coding challenges across multiple fantasy/sci-fi eras. The core mechanic is **Emergent Multi-Layer Systems** — combining broken/unrelated code to create glitches, loopholes, and side-effects that produce passcodes.

### Core Concept
- **Genre**: Roguelike Coding Puzzle Game
- **Platform**: Desktop (Godot 4) → Steam + Itch.io
- **Backend**: Go (Golang) for code execution and procedural generation
- **Sandbox**: WASM (WebAssembly) for secure code execution
- **AI**: Local model (Ollama/Llama) + AST parsing for code analysis

---

## 🎮 Game Design Summary

### Era Progression (v1 = 2 Eras)

| Era | Theme | Code Type | Aesthetic |
|-----|-------|-----------|-----------|
| **Tier 1** | Medieval Magitech | Custom DSL (Runes/Incantations) | Dark, mystical, ancient |
| **Tier 2** | Cyberpunk Neon | Real scripting (Python/JS) | Neon, terminal, corporate |

### Gameplay Modes (v1 = 3 Modes)

| Mode | Role | Objective | Skill Tested |
|------|------|-----------|--------------|
| **The Architect** | Builder | Write clean new modules to automate systems | Algorithmic thinking, planning |
| **The Ghost** | Stealth Hacker | Modify code under AI detection threshold | Optimization, stealth mindset |
| **The Saboteur** | Chaos Agent | Break code to cause environmental chain reactions | Judgment, critical thinking |

### Core Mechanics

1. **Procedural Challenge Generation**
   - Seed-based RNG stitches together modular "junk code segments"
   - Luck stat affects level complexity and AI monitoring strictness
   - At least one "glitch/loophole" solution always exists

2. **Multi-Layer Code Interaction**
   - Frankenstein Code: Broken scripts combine to create exploits
   - Loophole Exploitation: Race conditions, timing attacks
   - Environmental Side-Effects: Code affects virtual hardware

3. **Dynamic Passcode System**
   - Passcodes emerge from code interactions (not direct output)
   - Hidden in error logs, memory leaks, CPU fluctuations
   - Different per player based on approach and luck

4. **Luck & Volatility Engine**
   - High Luck: Easy flaws, perfect glitch alignment
   - Low Luck: Obfuscated code, aggressive AI monitoring

---

## 🏗️ Technical Architecture

### System Components

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

### Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| **Frontend** | Godot 4 (GDScript/C#) | UI, visual themes, editor |
| **Backend** | Go 1.26+ | Code execution, generation, analysis |
| **Sandbox** | WASM (Extism/Wasmer) | Secure player code execution |
| **AI** | Ollama + Llama 3 | Code style analysis, passcode generation |
| **Parsing** | go/ast | AST analysis for code patterns |
| **Bridge** | GDExtension | Native Go ↔ Godot communication |

### Security Model

1. **WASM Isolation**: Player code runs in isolated VM instances
2. **Resource Limits**: CPU time, memory, execution steps capped
3. **No Network Access**: Sandbox cannot make external calls
4. **Timeout Enforcement**: Infinite loops killed after threshold

---

## Milestone Roadmap

This is the authoritative long-range implementation sequence for architecture-driven work. Each milestone is gated by explicit dependencies and exit criteria so future implementation slices, ADRs, and issues can derive from it without drifting into feature work too early.

### Milestone 3: Infrastructure Decoupling

**Prerequisites**

- Single runtime `ChallengeDefinition`
- Thin server entry point
- Green verification pipeline

**Objectives**

- Eliminate remaining package-level globals.
- Establish explicit ownership for infrastructure services.
- Complete dependency injection.

**Deliverables**

- DB service ownership.
- SessionManager.
- Service-oriented server composition.
- Clear lifecycle boundaries.

**Exit criteria**

- No package-level mutable runtime state.
- Session lifecycle owned entirely by `SessionManager`.
- `make verify` and QA suite remain green.

### Milestone 3 Vertical Slices

Each slice is independently deliverable and must leave the repository releasable with `make verify` and the QA suite passing.

#### Slice 1: Database Service

**Objective**

Convert the database into an owned service.

**Tasks**

- Create `type DB struct`.
- Move initialization into `server.New()`.
- Inject `*db.DB` into dependent services.
- Remove singleton access.

**Deliverables**

- Constructor-based ownership.
- Unit tests unchanged.
- QA unchanged.

**Risk**

- Low.

#### Slice 2: SessionManager

**Objective**

Extract session lifecycle from transport code.

**Tasks**

- Reconnect.
- Session creation.
- Persistence.
- Lookup.
- Eviction.

**API**

```go
type SessionManager interface {
   Create(...)
   Attach(...)
   Save(...)
   Remove(...)
}
```

**Exit criteria**

- WebSocket no longer manages session lifecycle.
- Integration tests unchanged.

#### Slice 3: Service Wiring

**Objective**

Move construction into one place.

**Current**

```
handler
creates DB

creates sandbox

creates oracle
```

**Target**

```
server.New()

↓

constructs everything

↓

passes dependencies
```

#### Slice 4: Constructor Migration

**Objective**

Convert remaining constructors to explicit dependency injection.

**Deliverables**

- Explicit dependencies.
- Easier testing.
- Deterministic initialization.

#### Slice 5: Shutdown Lifecycle

**Objective**

Implement graceful shutdown.

**Sequence**

```
SIGINT
     ↓
Stop accepting clients
     ↓
Persist active sessions
     ↓
Stop game loops
     ↓
Close DB
     ↓
Exit
```

**Tests**

- Shutdown during mission.
- Reconnect after restart.
- Persistence verification.

#### Slice 6: Observability

**Objective**

Standardize logging across request paths.

Every request carries:

```
SessionID

MissionID

PlayerID

Tick

RequestID
```

This makes reconnect and replay debugging significantly easier.

### Milestone 4: Event Model and Transport Abstraction

This milestone should begin by stabilizing the event model before introducing new transports.

**Objectives**

Define a canonical simulation event schema.

Example events:

```
MissionStarted
MissionLoaded
TickAdvanced
CodeExecuted
ResourceCollected
ObjectiveCompleted
PasscodeGenerated
MissionWon
MissionLost
```

Then introduce a transport interface:

```
GameLoop
   ↓
Simulation Events
   ↓
Transport
   ├── WebSocket
   ├── Replay
   ├── QA Runner
   └── Future CLI
```

**Exit criteria**

- GameLoop contains no WebSocket-specific logic.
- QA runner consumes the same event stream.
- Replay is event-driven rather than transport-driven.

### Milestone 5: Constraint Solver

This becomes a foundational subsystem rather than an isolated feature.

Dependencies:

- Stable event model.
- Stable deterministic simulation.

Subsystems consuming it include:

- procedural generation,
- challenge validation,
- difficulty estimation,
- hint generation,
- dependency analysis,
- glitch propagation.

Core algorithms:

- AC-3,
- forward checking,
- MRV,
- degree heuristic,
- IDA*,
- memoization,
- cycle detection.

**Exit criteria**

- Every generated challenge can be validated automatically for solvability and structural correctness.

### Milestone 6: P-Script and Swarm Evolution

These should progress together because communication primitives directly affect emergent behavior.

P-Script additions:

- deterministic collections,
- persistent state,
- message passing,
- timers,
- richer VM instructions.

Swarm capabilities:

- local communication,
- broadcast,
- quorum voting,
- pheromone systems,
- coordinated task allocation.

**Exit criteria**

- Player scripts can express meaningful multi-agent coordination without sacrificing deterministic execution.

### Milestone 7: Emergent Gameplay

Build gameplay on top of the mature engine.

Features include:

- coordinated drone strategies,
- sabotage mechanics,
- dynamic passcodes,
- multi-step dependency chains,
- emergent puzzle solutions.

The emphasis shifts from adding systems to creating interactions between existing systems.

### Milestone 8: Content Pipeline

Designer tooling depends on the earlier milestones.

Pipeline:

```
Author
   ↓
Validator
   ↓
Constraint Solver
   ↓
Replay Verification
   ↓
Difficulty Estimator
   ↓
Publish
```

Every challenge should satisfy:

- deterministic execution,
- solvability,
- replay reproducibility,
- dependency correctness,
- expected difficulty.

### Milestone 9: Polish and Production

Focus areas:

- visual polish,
- audio,
- balancing,
- accessibility,
- performance,
- telemetry,
- progression,
- achievements,
- release preparation.

## Development Contract

Each milestone follows the same engineering discipline:

1. Implement a focused vertical slice.
2. Add or update automated tests.
3. Run the QA scenario suite.
4. Run `make verify`.
5. Merge only when all verification gates remain green.

This keeps the repository continuously releasable while allowing increasingly sophisticated gameplay systems to be layered onto a stable deterministic engine.

```
challenge-to-you/
├── backend/                    # Go backend
│   ├── cmd/                    # Entry points
│   │   └── sandbox/            # WASM sandbox runner
│   ├── internal/               # Private packages
│   │   ├── generator/          # Procedural generation
│   │   │   ├── generator.go    # Seed-based RNG
│   │   │   ├── modules.go      # Junk code segments
│   │   │   └── luck.go         # Luck mechanics
│   │   ├── sandbox/            # WASM execution
│   │   │   ├── sandbox.go      # Extism/Wasmer wrapper
│   │   │   └── limits.go       # Resource limits
│   │   ├── analyzer/           # Code analysis
│   │   │   ├── ast.go          # AST parsing
│   │   │   └── style.go        # Code style detection
│   │   ├── passcode/           # Passcode generation
│   │   │   ├── engine.go       # Main generator
│   │   │   └── glitch.go       # Glitch detection
│   │   └── narrative/          # Level text/flavor
│   │       ├── magitech.go     # Era 1 text
│   │       └── cyberpunk.go    # Era 2 text
│   ├── pscript/                # Magitech DSL
│   │   ├── lexer.go
│   │   ├── parser.go
│   │   └── ast.go
│   ├── go.mod
│   └── go.sum
├── client/                     # Godot 4 project
│   ├── scenes/                 # Game scenes
│   │   ├── main.tscn           # Main menu
│   │   ├── editor.tscn         # Code editor
│   │   ├── terminal.tscn       # Output terminal
│   │   └── eras/               # Era-specific scenes
│   │       ├── magitech.tscn
│   │       └── cyberpunk.tscn
│   ├── scripts/                # GDScript files
│   │   ├── main.gd
│   │   ├── editor.gd
│   │   ├── terminal.gd
│   │   └── bridge.gd           # GDExtension bridge
│   ├── themes/                 # Visual themes
│   │   ├── magitech/
│   │   └── cyberpunk/
│   ├── addons/                 # Godot plugins
│   │   └── gdextension/        # Go bridge
│   └── project.godot
├── docs/                       # Documentation
│   ├── CHALLENGE-TO-YOU-PLAN.md
│   ├── ARCHITECTURE.md
│   ├── GAME-DESIGN.md
│   ├── API.md
│   ├── UNIVERSAL-LOGIC-ENGINE.md
│   └── LOCAL-LLM-INTEGRATION.md
├── tools/                      # Build tools
│   ├── build.sh                # Build script
│   └── package.sh              # Packaging script
└── README.md
```

---

## 🧪 Testing Strategy

### Unit Tests

| Component | Test Focus |
|-----------|------------|
| **Generator** | Seed reproducibility, module combination validity |
| **Sandbox** | Execution isolation, timeout enforcement, memory limits |
| **Passcode** | Glitch detection, hash consistency |
| **AST Analyzer** | Code pattern recognition, style scoring |
| **Luck System** | Distribution fairness, boundary conditions |

### Integration Tests

| Test | Description |
|------|-------------|
| **End-to-End** | Code input → Execution → Passcode output |
| **Mode Switching** | Architect ↔ Ghost ↔ Saboteur transitions |
| **Era Transition** | Magitech → Cyberpunk unlock flow |
| **Procedural Generation** | 1000 seed runs, all produce valid challenges |

### Playtesting

| Phase | Focus |
|-------|-------|
| **Week 3** | Internal testing, fix game-breaking bugs |
| **Week 4** | Community alpha on Itch.io, gather feedback |
| **Post-Launch** | Iterate based on player data |

---

## 📚 Documentation Plan

### What to Document

| Document | When | Purpose |
|----------|------|---------|
| **ARCHITECTURE.md** | Week 1 | System design, component boundaries |
| **GAME-DESIGN.md** | Week 1 | Mechanics, modes, era progression |
| **API.md** | Week 2 | Go ↔ Godot interface specification |
| **CONTRIBUTING.md** | Week 3 | Team collaboration guidelines |
| **CHANGELOG.md** | Week 4 | Version history, fixes |

### Conversation Documentation

All design decisions, conversations, and changes will be documented in:
- `docs/CONVERSATIONS/` — Timestamped conversation logs
- `docs/DECISIONS/` — Architecture decision records (ADRs)
- `docs/CHANGES/` — Change log with rationale

---

## 🚀 Post-MVP Roadmap

### Phase 2 (Month 2-3): Steam Early Access

| Task | Priority |
|------|----------|
| Implement true WASM sandbox | High |
| Add Era 3 (Dieselpunk) | Medium |
| Add Leaderboards | Medium |
| Add Tutorial/Onboarding | High |
| Steam Early Access launch ($4.99-$9.99) | High |

### Phase 3 (Month 4-6): Full Release

| Task | Priority |
|------|----------|
| Add Era 4 (Cosmic Space) | Medium |
| Add 3 more gameplay modes | Low |
| AI challenge generation | High |
| Multiplayer (cooperative hacking) | Low |
| Full Steam release | High |

---

## 💰 Budget Estimate

| Item | Cost |
|------|------|
| Steam Direct fee | $100 |
| Freelancer (UI theme) | $200-$500 |
| Sound designer (CC0 music) | $100-$300 |
| Ollama/Llama hosting | Free (local) |
| Itch.io hosting | Free |
| **Total MVP** | **$400-$900** |

---

## 📊 Success Metrics

### Itch.io Alpha (Week 4)

| Metric | Target |
|--------|--------|
| Downloads | 100+ |
| Play time average | 15+ minutes |
| Bug reports | 10+ |
| Steam wishlist clicks | 50+ |

### Steam Early Access (Month 3)

| Metric | Target |
|--------|--------|
| Wishlists | 500+ |
| Reviews | 10+ positive |
| Revenue | $1,000+ |

---

## 🎯 Key Decisions Log

| Decision | Rationale | Date |
|----------|-----------|------|
| Desktop first (Godot) | WASM sandbox needs raw power | 2026-07-10 |
| Go backend | Concurrency for multi-layer execution | 2026-07-10 |
| 2 eras in v1 | Prove multiverse hook | 2026-07-10 |
| 3 modes in v1 | Cover full spectrum (build/steal/break) | 2026-07-10 |
| Procedural generation | Infinite replayability | 2026-07-10 |
| Luck mechanic | Roguelike elements, replay value | 2026-07-10 |
| Itch.io first | Free alpha for feedback | 2026-07-10 |
| Steam later | Early Access for revenue | 2026-07-10 |
| WASM for MVP | Skip for alpha, add for Steam | 2026-07-10 |

---

*Last updated: 2026-07-12*
*Status: Core implemented and verified (build/tests/CI green); roadmap milestones 5–9 and Phase 2/3 remain future work as documented above.*
