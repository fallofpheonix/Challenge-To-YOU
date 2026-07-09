# Challenge To YOU — Complete Implementation Plan

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

## 📅 1-Month Production Timeline

### Week 1 (Days 1-7): Core Infrastructure

**Goal**: Basic Go backend + Godot editor that can execute simple code

| Task | Owner | Deliverable |
|------|-------|-------------|
| Set up Go module with GDExtension | Solo Dev | `challenge-to-you/backend/` |
| Implement WASM sandbox (Extism) | Solo Dev | Basic execution pipeline |
| Create Godot project structure | Solo Dev | `challenge-to-you/client/` |
| Build minimal code editor UI | Solo Dev | Text input → execute → output |
| Pay Steam Direct fee ($100) | Solo Dev | Store page approval started |
| Create terminal UI theme | Freelancer | Retro-cyber visual theme |

**Milestone**: Code typed in Godot executes in Go sandbox

### Week 2 (Days 8-14): Procedural Generation

**Goal**: Seed-based challenge generation with luck mechanic

| Task | Owner | Deliverable |
|------|-------|-------------|
| Design 10 modular "junk code blocks" | Solo Dev | `backend/modules/` |
| Build seed-based RNG generator | Solo Dev | `backend/generator.go` |
| Implement Luck stat system | Solo Dev | `backend/luck.go` |
| Create Magitech DSL parser | Solo Dev | `backend/pscript/` |
| Write level flavor text | Collaborator | `backend/narrative/` |
| Build level selection UI | Solo Dev | Era/mode selection screen |

**Milestone**: Randomly generated challenges execute and produce output

### Week 3 (Days 15-21): Gameplay Modes

**Goal**: Implement Architect, Ghost, and Saboteur modes

| Task | Owner | Deliverable |
|------|-------|-------------|
| Implement Architect mode | Solo Dev | Write/fix code objectives |
| Implement Ghost mode | Solo Dev | CPU usage tracking, stealth meter |
| Implement Saboteur mode | Solo Dev | Break code, chain reactions |
| Build tracking meter UI | Solo Dev | Detection/stealth indicators |
| Create passcode generation engine | Solo Dev | `backend/passcode.go` |
| Add CC0 synth music | Sound Designer | Audio atmosphere |

**Milestone**: All 3 modes playable with passcode generation

### Week 4 (Days 22-30): Polish & Launch

**Goal**: Bug fixes, safety clamps, Itch.io launch

| Task | Owner | Deliverable |
|------|-------|-------------|
| Implement infinite loop timeout | Solo Dev | 10s execution cap |
| Add hint archive system | Solo Dev | Save/load hints |
| Package desktop builds | Solo Dev | Windows/Mac/Linux |
| Set up Itch.io page | Solo Dev | Store listing, screenshots |
| Smoke test with players | Community | Bug reports |
| Launch on Itch.io | Solo Dev | Free/PWYW Alpha |
| Direct traffic to Steam wishlist | Solo Dev | Marketing push |

**Milestone**: Itch.io alpha live, Steam page collecting wishlists

---

## 🧠 Universal Logic Engine

**Key Insight**: This is NOT a code compiler game — it's a **Universal Logic Manipulation Engine**.

The engine treats all game elements (magic runes, code scripts, physical gears) as the same data structure: **Events, Conditions, and Effects** in a Directed Acyclic Graph (DAG).

### Core Concept

```
[INPUT EVENT] ───► [CONDITION/STATE CHECK] ───► [OUTPUT EFFECT]
```

### Theme Examples

| Theme | Player Sees | Engine Sees |
|-------|-------------|-------------|
| **Magitech** | Fire rune on spell-book | `Event: fire_rune, Power: 100` |
| **Cyberpunk** | Python code in terminal | `Event: audio_driver, Port: 80` |
| **Industrial** | Metal gear on conveyor | `Event: gear, Voltage: 220` |

### Implementation

See `docs/UNIVERSAL-LOGIC-ENGINE.md` for complete technical specification.

**Key Components**:
1. **GameNode** — Universal element (Event/Condition/Effect)
2. **GameGraph** — Cause-and-effect network
3. **GlitchDetector** — Finds emergent solutions
4. **WinCondition** — State/event/chain-based verification

---

## 📁 Project Structure

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

*Last updated: 2026-07-10*
*Status: Approved — Ready for Implementation*
