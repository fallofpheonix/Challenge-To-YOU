---
name: chrysalis-simulation
description: The deterministic Go simulation core — ECS (SoA), double-buffered grid, fixed-point math, pheromones, hazards, aliens/trust, and the 10 Hz tick. Use when editing engine/core/simulation or anything that must stay bit-perfect reproducible.
---

# Chrysalis Simulation Core

Authoritative, deterministic world state. **Bit-perfect reproducibility is the prime directive** — same seed + same inputs must always yield the same state hash across platforms and commits.

## Determinism rules (do not break)
- **No floats.** All math via `crysmath.FixedPoint`, `Precision = 10^6`. Floats leak non-determinism.
- **Seeded RNG only.** Use the engine RNG (`simulation/rng.go`); never `math/rand` global or `time`-based seeds.
- **No map iteration order dependence.** Go map ranges are randomized — sort keys or use slices for ordered work.
- **Double-buffered grid** (`grid.go`): mutate `NextCells`, read `CurrentCells`, commit via `SwapBuffers`. One tick latency by design (ADR-003).
- **ECS is SoA** (`ecs.go`, `SwarmRegistry`): parallel slices `PositionX/Y`, `Battery`, `State`, `Inventory`… Entities are integer IDs. Add/remove copies slices.
- **Spatial hash** (`spatial.go`): O(1)-ish neighbor queries; never loop the global entity list for adjacency.

## Systems (`engine/core/simulation/`)
- `simulation.go` — engine lifecycle, drone AI, state serialization
- `pheromones.go` — Home / Resource / Alien signals; decay 5000/tick, max 10^6, gradient sensing
- `hazards.go` — magnetic drain 10^6/tick, thermal 2×10^6/tick
- `alien.go` — infection spread (radius 3), compromise, alien signal deposit
- trust/quorum — `TrustScore` 0-100, `BROADCAST_VOTE` consensus >50%
- `mission.go` — see [[chrysalis-mission-framework]]

## Key params
Grid 100×100 · 10 Hz · Pheromone decay 5000/tick, max 10^6 · Battery drain 1000/tick · Fabrication cost 5 silicates · Max swarm 500 · Inert grace 30 ticks · Resource respawn 100 ticks, max 5/cell · VM step limit 1000.

## Testing
Determinism is Tier-1: seed → inject script → run fixed ticks → hash final state; **hash change = build fails**. Determinism tests must never flake (a flake = race or float leak = critical bug). See `simulation_test.go`, `setstate_test.go`, `trace_test.go`. Related: [[chrysalis-replay-system]], [[chrysalis-eventbus]], [[chrysalis-coding-standards]].
