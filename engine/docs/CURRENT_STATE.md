# Project Chrysalis — Current Implementation State

*Project Chrysalis* is a deterministic, high-performance swarm simulation where players (Architects) program decentralized agent behaviors. The engine is built in Go for bit-perfect determinism and visualized in Godot 4 via an asynchronous WebSocket IPC bridge.

---

## 1. Project Directory Structure

The project is organized under the `Project-Chrysalis-The-Architect-s-Swarm/engine/` directory:

- **`core/`**: The Go-based simulation engine and P-Script subsystem.
  - **`simulation/`**: The authoritative ECS engine.
    - **`ecs.go`**: Swarm Registry using contiguous memory slices with dynamic expansion.
    - **`grid.go`**: Double-buffered spatial grid with fixed-point pheromone and alien signal layers.
    - **`hazards.go`**: Environmental threat system (Magnetic Anomalies, Thermal Geysers).
    - **`alien.go`**: Viral logic contagion and alien node management.
  - **`crysmath/`**: Bit-perfect **Fixed-Point Arithmetic** library (`Precision = 10^6`).
  - **`pscript/`**: Pratt parsing engine and per-entity AST interpreter.
  - **`network/`**: Asynchronous, multi-threaded WebSocket hub (port `:8080`).
- **`client/`**: The Godot 4 visual relay and Architect command terminal.

---

## 2. Hardened Infrastructure (Milestones 0 - 3 COMPLETED)

The foundational architecture is now 100% stable and verified.

### Key Achievements
- **Deterministic ECS Data Layer**: High-performance `SwarmRegistry` with contiguous slices ensuring CPU cache efficiency. Supports dynamic reallocation up to 500+ entities.
- **Bit-Perfect Math**: Canonical `FixedPoint` scaling ($10^6$) used for all spatial and intensity vectors, eliminating floating-point drift.
- **P-Script Infix Engine**: Hardened Pratt parser supporting complex algebraic expressions (e.g., `if (SENSE_BATTERY() < 25000000)`).
- **Asynchronous WebSocket IPC**: Bi-directional command/response bridge between Go and Godot. Supports zero-latency telemetry and remote logic hot-patches.
- **Trust Mesh & Contagion**: Multi-stage logic infection system with Byzantine Fault Tolerance (BFT) quorum voting (`BROADCAST_VOTE()`).
- **Replica Matrix**: Economic resource loop where gathered silicates are consumed to fabricate new autonomous units at the Hub.

---

## 3. Visual & Diagnostic Layer

The Godot 4 client acts as the "Architect's Terminal":
- **10-Screen Dashboard**: Real-time monitoring of hazards, pheromones, population growth, and swarm security.
- **Swarm Inspector**: Modal diagnostic window for real-time logic analysis and surgical AST patching.
- **Viral Visualization**: Real-time color-grading from authoritative Cyan to viral Purple based on corruption levels.

---

## 4. Verification & Deployment

```bash
# Run the complete technical validation suite
cd engine
make test-core
```

The system is currently active and idling at 10Hz, awaiting strategic logic deployment.
