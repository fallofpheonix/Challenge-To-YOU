# Challenge To YOU — CONTEXT

## Project Status

**Current Project**: Challenge To YOU
**Status**: Core Engine Complete — Ready for Playtest
**Last Updated**: 2026-07-10
**Documentation**: `docs/ARCHITECTURE-PHASE1.md`

---

## Domain Glossary

| Term | Definition |
|------|-----------|
| **Challenge To YOU** | Roguelike coding puzzle game across multiple fantasy/sci-fi eras |
| **Era** | Thematic world with unique code type and aesthetic (Magitech, Cyberpunk, etc.) |
| **Mode** | Gameplay style — Architect (build), Ghost (stealth), Saboteur (break) |
| **Passcode** | Key to advance; emerges from code interactions, not direct output |
| **Frankenstein Code** | Broken/unrelated scripts combined to create glitches/loopholes |
| **Luck Stat** | Player attribute affecting level complexity and AI monitoring |
| **Procedural Generation** | Seed-based RNG stitching modular code segments |
| **Glitch** | Intentional exploit created by combining broken code |
| **Loophole** | Security flaw discovered through code interaction |
| **Environmental Side-Effect** | Code affecting virtual hardware (overload, reboot, etc.) |
| **Detection Meter** | Ghost mode indicator; spikes if CPU usage too high |
| **Chain Reaction** | Saboteur mode; one break causes cascading failures |
| **AxiomaticFabric** | Universal logic graph manager for all game state |
| **Archon** | AI supervisor; vigilance rises with entropy |
| **Hydrator** | Procedural noise injector based on player Luck |

---

## Era Progression (v1 = 3 Eras)

| Tier | Era | Theme | Code Type | Aesthetic |
|------|-----|-------|-----------|-----------|
| 1 | Medieval Magitech | Dark fantasy | Custom DSL (Runes) | Mystical, ancient |
| 2 | Cyberpunk Neon | Dystopian future | Real scripting (Python/JS) | Neon, terminal |
| 3 | Cosmic Void | Deep space physics | Quantum logic | Abstract, reality-warping |

---

## Gameplay Modes (v1 = 3 Modes)

| Mode | Role | Objective | Skill Tested |
|------|------|-----------|--------------|
| **Architect** | Builder | Write clean new modules | Algorithmic thinking, planning |
| **Ghost** | Stealth Hacker | Modify code under AI detection | Optimization, stealth mindset |
| **Saboteur** | Chaos Agent | Break code for chain reactions | Judgment, critical thinking |

---

## Core Mechanics

### 1. Procedural Challenge Generation
- Seed-based RNG stitches modular "junk code segments"
- Luck stat affects: flaw count, AI monitoring strictness, entropy penalties
- At least one "golden path" solution always exists

### 2. The AxiomaticFabric
- Universal logic graph evaluating `TriggerOntologicalShift(eventID)`
- Conditions → Effects → FallbackEffects (entropy penalties)
- Win conditions via state key/value matching

### 3. The Hydrator
- Takes base `ChallengeDefinition` + `Luck` float (0.0–1.0)
- Injects 1–5 junk flaws per paradigm (Magitech/Cyberpunk/Cosmic)
- Shuffles flaw order to randomize golden path visibility
- Core flaws immutable, junk flaws penalize entropy

### 4. Entropy Economy
- Junk trap hits → +10-25 entropy
- Entropy > 0 → vigilance rises 0.2% per tick (500ms)
- Vigilance ≥ 100% → ONTOLOGICAL_PURGE (game over)
- Tension: ~25 seconds at max entropy before purge

### 5. AI Archon (Ollama)
- Dynamic taunts on junk trap hits (paradigm-specific)
- Mending Protocol: structured JSON repair mutations
- 2-second timeout to prevent game loop blocking
- Graceful fallback if Ollama unavailable

---

## System Boundaries

```
┌─────────────────────────────────────────────────────────────┐
│                    Godot 4 Client (Desktop)                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ Terminal UI  │  │ Command Line │  │ Era-Specific     │  │
│  │ (BBCode)     │  │ Parser       │  │ Visual Themes    │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────────────┘  │
│         │                 │                                  │
│  ┌──────▼─────────────────▼──────────────────────────────┐  │
│  │              WebSocketPeer (ws://localhost:8080)       │  │
│  └──────┬────────────────────────────────────────────────┘  │
└─────────┼───────────────────────────────────────────────────┘
          │
┌─────────▼───────────────────────────────────────────────────┐
│                    Go Backend (WebSocket Server)             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ Axiomatic    │  │ Procedural   │  │ AI Archon        │  │
│  │ Fabric       │  │ Generator    │  │ (Ollama)         │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Hydrator (Luck-Based Noise)              │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## Project Structure

```
challenge-to-you/
├── backend/                    # Go backend
│   ├── cmd/sandbox/            # WebSocket server entry point
│   ├── internal/
│   │   ├── engine/             # AxiomaticFabric, challenges, evaluator
│   │   ├── generator/          # Procedural generation (vocab pools)
│   │   ├── ai/                 # Ollama integration (taunts, repairs)
│   │   └── db/                 # State persistence
│   ├── challenges/             # Static challenge JSON files
│   │   ├── magitech_tier1/     # 7 challenges
│   │   ├── cyberpunk_tier1/    # 6 challenges
│   │   └── cosmic_tier1/       # 8 challenges
│   ├── go.mod
│   └── go.sum
├── client/                     # Godot 4 project
│   ├── scenes/                 # Main.tscn
│   ├── scripts/                # main.gd, network_bridge.gd
│   └── project.godot
├── docs/                       # Documentation
│   ├── CHALLENGE-TO-YOU-PLAN.md
│   ├── ARCHITECTURE.md
│   ├── GAME-DESIGN.md
│   ├── API.md
│   ├── DEPLOYMENT.md
│   └── DECISIONS/              # ADRs
└── README.md
```

---

## Architecture Decisions

### ADR-001: Desktop-First with Godot 4
**Decision**: Use Godot 4 as frontend engine, targeting desktop platforms.
**Rationale**: WASM sandbox needs raw power; GDExtension allows native Go integration.

### ADR-002: Go Backend with WebSocket
**Decision**: Use Go for backend with WebSocket for client communication.
**Rationale**: Concurrency for multi-layer execution; security via isolation.

### ADR-003: Procedural Generation with Luck Mechanic
**Decision**: Seed-based RNG with Luck stat affecting difficulty.
**Rationale**: Infinite replayability; roguelike appeal.

### ADR-004: Three-Era Progression
**Decision**: Launch with Magitech → Cyberpunk → Cosmic.
**Rationale**: Proves multiverse hook; visual contrast.

### ADR-005: Three Core Gameplay Modes
**Decision**: Launch with Architect, Ghost, Saboteur.
**Rationale**: Full spectrum of hacker fantasies; skill variety.

### ADR-006: Local AI (Ollama)
**Decision**: Use local LLM (Ollama) for dynamic taunts and repairs.
**Rationale**: Cost-free; private; offline capable.

---

## Data Coverage

| Era | Challenges | Pack Status |
|-----|-----------|-------------|
| Magitech | 7 (breach, centrifuge, vault, golem, grimoire, astrolabe, loom) | Complete |
| Cyberpunk | 6 (autodoc, elevator, server, barista, drone, traffic) | Complete |
| Cosmic | 8 (airlock, nav, stasis, winch, seed, singularity, relay, valve) | Complete |

**Total**: 21 hand-crafted challenges + infinite procedural permutations

---

*Last updated: 2026-07-10*
*Status: Core Engine Complete — Ready for Playtest*
