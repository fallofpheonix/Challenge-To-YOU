# Chrysalis Engine Core

> **Role**: Authoritative Simulation & Logic (The "Brains")

This directory contains the Go-based simulation engine. It is responsible for:
- Maintaining the deterministic world state.
- Executing P-Script autonomous logic.
- Emitting state deltas to the visual client via JSON pipes.

## Subdirectories
- `pscript/`: The custom scripting language implementation.
- `state/`: World state definitions and serialization.
