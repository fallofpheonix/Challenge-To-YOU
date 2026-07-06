---
name: chrysalis-engine-architecture
description: Layered architecture of Project Chrysalis — the Go deterministic core, P-Script rule engine, gameplay systems, and Godot telemetry client, plus the five ADRs. Use when reasoning about where code belongs, cross-layer boundaries, or the core↔client split.
---

# Chrysalis Engine Architecture

Split-stack: authoritative **Go core** owns simulation; **Godot client** is pure presentation. They communicate only over WebSocket. Never leak rendering concerns into the core, or authoritative state into the client.

## Layers (top = closest to player)
1. **UI / Telemetry** (Godot) — `main.gd` → `GameHub` router → 10 screen views + overlays. Reads state, sends commands. No game logic.
2. **Gameplay Systems** (Go `simulation/`) — pheromones, hazards, aliens, uplink queue, missions.
3. **Rule Engine** (Go `pscript/`) — lexer → parser → AST → compiler → bytecode VM (interpreter fallback). Sandboxed per-tick execution.
4. **Simulation Core** (Go `simulation/`) — deterministic 10 Hz tick, ECS (SoA), spatial hash grid.

## Architecture Decisions (ADRs)
- **ADR-001 Fixed-point math** — all sim math via `crysmath.FixedPoint`, `Precision = 10^6`. No floats in the core (floats desync across platforms). Client divides by 10^6 for display.
- **ADR-002 Data-oriented ECS** — `SwarmRegistry` stores components as parallel slices (SoA), not array-of-structs. Add/remove = slice copy; no pointer indirection.
- **ADR-003 Double-buffered grid** — `CurrentCells`/`NextCells`, atomic `SwapBuffers`. One tick of latency by design.
- **ADR-004 Bytecode VM** — 18-opcode stack VM, ~10x over interpreter for 500+ drones; compile on load/hot-reload, execute per-tick zero-alloc; 1000-step safety limit. Interpreter only if compilation fails.
- **ADR-005 WebSocket bridge** — Go serves, Godot connects; multiple observers + remote command injection; ~1ms/tick at 10 Hz.

## Boundary
`Godot NetworkBridge` ⇄ `ws://127.0.0.1:8080/telemetry` ⇄ `network/hub.go` ⇄ `main.go` (10 Hz loop) ⇄ `simulation.Engine`.

## Doc drift to be aware of
Older docs (`02_ARCHITECTURE/ARCHITECTURE.md`, `04_ENGINEERING/CODING_STANDARDS.md`) still describe a **Python sandbox**; the shipped rule engine is **P-Script in Go**. Trust `CONTEXT.md` + code over those. See [[chrysalis-pscript]], [[chrysalis-simulation]], [[chrysalis-coding-standards]].
