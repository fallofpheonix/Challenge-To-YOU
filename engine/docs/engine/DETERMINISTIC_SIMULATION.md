---
Status: Active
Implementation: 100%
Confidence: Authoritative
---
# Game Engine — Deterministic Simulation

Ensures identical execution runs across any platform by eliminating floating-point non-determinism and enforcing strict state progression.

## Rules
- **Fixed-Point Math**: All physics and logic use the `crysmath` fixed-point library with a precision scale of `10^6`. Floating-point numbers are strictly forbidden in the simulation core.
- **Fixed Frame-Tick**: Execution is driven by a fixed 10Hz frame-tick loop (`BeginTick`, `stepDrones`, `CommitTick`).
- **Seeded RNG**: The random number generator is seeded per-engine instance (`DefaultWorldSeed`), ensuring identical hazard spawns and random movements across runs.
- **Double-Buffered Grid**: The spatial environment uses `CurrentCells` (read-only during tick) and `NextCells` (write-only during tick) with an atomic swap at the end of the tick. This eliminates evaluation order dependencies between drones.
