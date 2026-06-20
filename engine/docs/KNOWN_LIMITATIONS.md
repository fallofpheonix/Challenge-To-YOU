# Known Limitations

## P0: Complete Game Blockers

- Scenario files do not configure the simulation.
- Victory/defeat is hardcoded for the v0 mission; no scoring, persistence, or campaign progression.
- Replay, research, and structure dashboards exceed the backend model.
- Drones may overlap because occupancy and collision are absent.
- Return-to-base has a deterministic fallback for unobstructed paths, but there is no obstacle/hazard-aware route planning yet.

## P1: Correctness And Security

- Hot patches are unauthenticated; exposure is restricted to loopback.
- P-Script has a per-loop cap but no aggregate instruction budget.
- Per-drone JSON telemetry will not scale to the stated 10,000+ entity target.
- Replay seed and RNG-state serialization are absent.
- Abrupt external termination should still be treated as risky, but tested normal and interrupt exits did not orphan `chrysalis-core`.

## P2: Tooling And Presentation

- `client/external/` contains large reference repositories (e.g. `GodotDynamicInventorySystem`, `godot-shadcn-ui-kit`, `willowui`) and triggers Godot warnings and slows down editor startup.
- Desktop visual QA passed through direct GUI launch and screenshot inspection; Computer Use accessibility remains unavailable, but OS screenshot capture works.
- The legacy editor/world scene is not integrated with scenario goals.

## Recommended Build Order

1. Versioned scenario schema loaded by the Go core.
2. Obstacle/hazard-aware return routing.
3. Scenario-driven mission state and win/loss evaluation.
4. Save/replay format including RNG state.
5. Occupancy and collision arbitration.
6. Aggregate density telemetry.
7. External repository cleanup.
