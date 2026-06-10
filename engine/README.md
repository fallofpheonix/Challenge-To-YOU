# Chrysalis Engine Layer

This is the core gameplay layer for Project Chrysalis.

## Project Structure

- `core/`: Go-based simulation engine and P-Script VM.
- `client/`: Godot-based visual client.
- `docs/`: Architectural specifications and roadmap.
- `bin/`: Compiled binaries.

## Getting Started

### Prerequisites
- Go 1.26+
- Godot 4.x (installed in `/Applications/Godot.app`)

### Commands
Use the provided `Makefile` for common tasks:
- `make run-all`: Launches the Godot client (which spawns the Go core).
- `make run-core`: Runs only the Go simulation (check stdout).
- `make test-core`: Runs Go unit tests.
- `make build-core`: Compiles the Go core into `bin/`.

## Development Standards
- **Determinism:** All logic in `core/` must be deterministic. No floats in state paths.
- **Contract-First:** Changes to the JSON bridge must be updated in both `core/` and `client/`.
