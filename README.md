# Challenge To YOU

A roguelike coding puzzle game across multiple fantasy/sci-fi eras.

## Overview

**Challenge To YOU** is a desktop-first game where players solve procedurally generated coding challenges across different worlds. The core mechanic is **Emergent Multi-Layer Systems** — combining broken/unrelated code to create glitches, loopholes, and side-effects that produce passcodes.

### Features

- **Multi-Era Progression**: From Medieval Magitech to Cyberpunk Neon to Cosmic Void
- **Three Gameplay Modes**: Architect (build), Ghost (stealth), Saboteur (break)
- **Procedural Generation**: Seed-based RNG creates infinite challenges
- **Luck Mechanic**: Roguelike volatility affects difficulty (0.0–1.0)
- **Dynamic Passcodes**: Different approaches produce different passcodes
- **AI Archon**: Local Ollama-powered taunts and mending protocols

## Tech Stack

| Layer | Technology |
|-------|------------|
| Frontend | Godot 4 (GDScript) |
| Backend | Go 1.26+ |
| WebSocket | gorilla/websocket |
| AI | Ollama (local LLM) |

## Quick Start

### Prerequisites

- Go 1.26+
- Godot 4.x
- Ollama (optional, for AI features)

### Installation

```bash
# Clone repository
git clone https://github.com/yourusername/challenge-to-you.git
cd challenge-to-you

# Build Go backend
cd backend
go build -o sandbox ./cmd/sandbox/
```

### Run

```bash
# Terminal 1: Start backend
cd backend
CHALLENGE_PATH=challenges/magitech_01.json ./sandbox

# Terminal 2: Open Godot client
cd client
open project.godot
```

### Procedural Mode

```bash
# Run with procedural generation (luck=0.5, Cyberpunk era)
./sandbox?seed=42&luck=0.5&paradigm=CYBERPUNK
```

## Project Structure

```
challenge-to-you/
├── backend/                    # Go backend
│   ├── cmd/sandbox/            # WebSocket server
│   ├── internal/
│   │   ├── engine/             # AxiomaticFabric core
│   │   ├── generator/          # Procedural generation
│   │   ├── ai/                 # Ollama integration
│   │   └── db/                 # State persistence
│   ├── challenges/             # Static challenge JSON
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
│   └── API.md
└── README.md
```

## Architecture

```
Godot Client ←→ WebSocket ←→ Go Backend
                                  ├─ AxiomaticFabric (state machine)
                                  ├─ Procedural Generator (vocab pools)
                                  ├─ Hydrator (Luck-based noise)
                                  └─ AI Archon (Ollama taunts/repairs)
```

## Documentation

| Document | Purpose |
|----------|---------|
| [CHALLENGE-TO-YOU-PLAN.md](docs/CHALLENGE-TO-YOU-PLAN.md) | Complete implementation plan |
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | System design |
| [GAME-DESIGN.md](docs/GAME-DESIGN.md) | Game mechanics |
| [API.md](docs/API.md) | Go ↔ Godot interface |
| [DEPLOYMENT.md](docs/DEPLOYMENT.md) | Deployment guide |

## Data Coverage

| Era | Challenges | Pack Status |
|-----|-----------|-------------|
| Magitech | 7 | Complete |
| Cyberpunk | 6 | Complete |
| Cosmic | 8 | Complete |

**Total**: 21 hand-crafted challenges + infinite procedural permutations

## Development Timeline

| Week | Focus | Deliverable |
|------|-------|-------------|
| 1 | Core Infrastructure | Go backend + Godot editor |
| 2 | Procedural Generation | Seed-based RNG + luck mechanic |
| 3 | Gameplay Modes | Architect, Ghost, Saboteur |
| 4 | Polish & Launch | Itch.io alpha live |

## License

MIT License - see [LICENSE](LICENSE) for details.

---

*Last updated: 2026-07-10*
