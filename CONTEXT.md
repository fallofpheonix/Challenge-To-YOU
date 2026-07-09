# Challenge To YOU — CONTEXT

## Project Status

**Current Project**: Challenge To YOU (formerly Project Chrysalis)  
**Status**: Planning Complete — Ready for Implementation  
**Last Updated**: 2026-07-10  
**Documentation**: `docs/CHALLENGE-TO-YOU-PLAN.md`

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

---

## Era Progression (v1 = 2 Eras)

| Tier | Era | Theme | Code Type | Aesthetic |
|------|-----|-------|-----------|-----------|
| 1 | Medieval Magitech | Dark fantasy | Custom DSL (Runes) | Mystical, ancient |
| 2 | Cyberpunk Neon | Dystopian future | Real scripting (Python/JS) | Neon, terminal |

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
- Seed-based RNG stitches 15-20 modular "junk code segments"
- Luck stat affects: code obfuscation, AI monitoring strictness, glitch availability
- At least one "glitch/loophole" solution always exists

### 2. Multi-Layer Code Interaction
- **Frankenstein Code**: Broken scripts combine to create exploits
- **Loophole Exploitation**: Race conditions, timing attacks
- **Environmental Side-Effects**: Code affects virtual hardware

### 3. Dynamic Passcode System
- Passcodes emerge from code interactions (not direct output)
- Hidden in error logs, memory leaks, CPU fluctuations
- Different per player based on approach and luck

### 4. Luck & Volatility Engine
- **High Luck**: Easy flaws, perfect glitch alignment
- **Low Luck**: Obfuscated code, aggressive AI monitoring

---

## Architecture Decisions

### ADR-001: Desktop-First with Godot 4
**Decision**: Use Godot 4 as frontend engine, targeting desktop platforms.  
**Rationale**: WASM sandbox needs raw power; GDExtension allows native Go integration.  
**Consequences**: No web version for MVP; must distribute via Steam/Itch.io.

### ADR-002: Go Backend with WASM Sandbox
**Decision**: Use Go for backend with Extism/Wasmer for WASM execution.  
**Rationale**: Concurrency for multi-layer execution; security via isolation.  
**Consequences**: Must learn WASM integration; long-term scalability excellent.

### ADR-003: Procedural Generation with Luck Mechanic
**Decision**: Seed-based RNG with Luck stat affecting difficulty.  
**Rationale**: Infinite replayability; roguelike appeal.  
**Consequences**: Must ensure at least one solution always exists.

### ADR-004: Multi-Era Progression (2 Eras in v1)
**Decision**: Launch with Magitech → Cyberpunk.  
**Rationale**: Proves multiverse hook; visual contrast.  
**Consequences**: Must build era-specific UI themes.

### ADR-005: Three Core Gameplay Modes
**Decision**: Launch with Architect, Ghost, Saboteur.  
**Rationale**: Full spectrum of hacker fantasies; skill variety.  
**Consequences**: Must implement tracking meters for Ghost mode.

### ADR-006: Local AI (Ollama) vs Cloud API
**Decision**: Use local LLM (Ollama + Llama 3) with AST parsing fallback.  
**Rationale**: Cost-free; private; offline capable.  
**Consequences**: Requires Ollama installation; less powerful than GPT-4.

### ADR-007: Skip WASM for Itch.io Alpha
**Decision**: Use Go text parsing for alpha; add WASM for Steam.  
**Rationale**: 1-month deadline; alpha tests mechanics, not security.  
**Consequences**: Alpha less secure; WASM becomes Phase 2 priority.

### ADR-008: Itch.io First, Steam Later
**Decision**: Free/PWYW alpha on Itch.io (Week 4); Steam Early Access (Month 3).  
**Rationale**: Free alpha attracts testers; drives wishlists.  
**Consequences**: Must set up Steam page in Week 1.

### ADR-009: Project Rename to "Challenge To YOU"
**Decision**: Rename from "Project Chrysalis" to "Challenge To YOU".  
**Rationale**: Direct, personal, challenging, memorable.  
**Consequences**: Must update all documentation.

### ADR-010: Document Everything
**Decision**: Document all conversations, decisions, changes.  
**Rationale**: Solo dev with fluid team needs clear documentation.  
**Consequences**: Must maintain structured docs; essential for team scaling.

---

## System Boundaries

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

---

## Project Structure

```
challenge-to-you/
├── backend/                    # Go backend
│   ├── cmd/                    # Entry points
│   │   └── sandbox/            # WASM sandbox runner
│   ├── internal/               # Private packages
│   │   ├── generator/          # Procedural generation
│   │   ├── sandbox/            # WASM execution
│   │   ├── analyzer/           # Code analysis
│   │   ├── passcode/           # Passcode generation
│   │   └── narrative/          # Level text/flavor
│   ├── pscript/                # Magitech DSL
│   ├── go.mod
│   └── go.sum
├── client/                     # Godot 4 project
│   ├── scenes/                 # Game scenes
│   ├── scripts/                # GDScript files
│   ├── themes/                 # Visual themes
│   ├── addons/                 # Godot plugins
│   └── project.godot
├── docs/                       # Documentation
│   ├── CHALLENGE-TO-YOU-PLAN.md
│   ├── ARCHITECTURE.md
│   ├── GAME-DESIGN.md
│   ├── DECISIONS/              # ADRs
│   └── CONVERSATIONS/          # Conversation logs
├── tools/                      # Build tools
└── README.md
```

---

## 1-Month Timeline

| Week | Focus | Deliverable |
|------|-------|-------------|
| 1 | Core Infrastructure | Go backend + Godot editor |
| 2 | Procedural Generation | Seed-based RNG + luck mechanic |
| 3 | Gameplay Modes | Architect, Ghost, Saboteur |
| 4 | Polish & Launch | Itch.io alpha live |

---

## Documentation

| Document | Purpose |
|----------|---------|
| `docs/CHALLENGE-TO-YOU-PLAN.md` | Complete implementation plan |
| `docs/DECISIONS/ADR-001-to-010.md` | Architecture decision records |
| `docs/CONVERSATIONS/2026-07-10-pivot.md` | Planning session log |
| `docs/ARCHITECTURE.md` | System design (to create) |
| `docs/GAME-DESIGN.md` | Mechanics, modes, eras (to create) |
| `docs/API.md` | Go ↔ Godot interface (to create) |

---

## Archived: Project Chrysalis

The original project (swarm simulation) is archived but preserved in:
- `engine/core/` — Go simulation engine
- `engine/client/` — Godot telemetry client
- `CONTEXT.md` (this file, above)

**Decision**: Project Chrysalis code is preserved but no longer actively developed. Focus shifted entirely to Challenge To YOU.

---

*Last updated: 2026-07-10*
*Status: Planning Complete — Ready for Implementation*
