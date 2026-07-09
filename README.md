# Challenge To YOU

A roguelike coding puzzle game across multiple fantasy/sci-fi eras.

## Overview

**Challenge To YOU** is a desktop-first game where players solve procedurally generated coding challenges across different worlds. The core mechanic is **Emergent Multi-Layer Systems** — combining broken/unrelated code to create glitches, loopholes, and side-effects that produce passcodes.

### Features

- **Multi-Era Progression**: From Medieval Magitech to Cyberpunk Neon
- **Three Gameplay Modes**: Architect (build), Ghost (stealth), Saboteur (break)
- **Procedural Generation**: Seed-based RNG creates infinite challenges
- **Luck Mechanic**: Roguelike volatility affects difficulty
- **Dynamic Passcodes**: Different approaches produce different passcodes
- **AI Integration**: Local AI analyzes your coding style

## Tech Stack

| Layer | Technology |
|-------|------------|
| Frontend | Godot 4 (GDScript/C#) |
| Backend | Go 1.26+ |
| Sandbox | WASM (Extism/Wasmer) |
| AI | Ollama + Llama 3 |

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
go build -o libchallenge.so -buildmode=c-shared ./cmd/sandbox

# Open Godot project
cd ../client
open project.godot
```

### Run

```bash
# Build and run
cd tools
./build.sh
./run.sh
```

## Project Structure

```
challenge-to-you/
├── backend/                    # Go backend
│   ├── cmd/                    # Entry points
│   ├── internal/               # Private packages
│   ├── pscript/                # Magitech DSL
│   └── go.mod
├── client/                     # Godot 4 project
│   ├── scenes/                 # Game scenes
│   ├── scripts/                # GDScript files
│   ├── themes/                 # Visual themes
│   └── project.godot
├── docs/                       # Documentation
│   ├── CHALLENGE-TO-YOU-PLAN.md
│   ├── ARCHITECTURE.md
│   ├── GAME-DESIGN.md
│   └── API.md
├── tools/                      # Build tools
└── README.md
```

## Documentation

| Document | Purpose |
|----------|---------|
| [CHALLENGE-TO-YOU-PLAN.md](docs/CHALLENGE-TO-YOU-PLAN.md) | Complete implementation plan |
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | System design |
| [GAME-DESIGN.md](docs/GAME-DESIGN.md) | Game mechanics |
| [API.md](docs/API.md) | Go ↔ Godot interface |

## Development Timeline

| Week | Focus | Deliverable |
|------|-------|-------------|
| 1 | Core Infrastructure | Go backend + Godot editor |
| 2 | Procedural Generation | Seed-based RNG + luck mechanic |
| 3 | Gameplay Modes | Architect, Ghost, Saboteur |
| 4 | Polish & Launch | Itch.io alpha live |

## Monetization

- **Free Alpha** (Itch.io): Week 4
- **Steam Early Access** ($4.99-$9.99): Month 3
- **Full Release** ($14.99-$19.99): Month 6

## Contributing

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Inspired by Hacknet, Uplink, and Thomas Was Alone
- Built with Godot 4 and Go
- Community feedback and support

---

*Last updated: 2026-07-10*
