# Known Limitations

## P0: Complete Game Blockers

- Scenario files do not configure the simulation.
- No victory, defeat, scoring, persistence, or campaign progression.
- Replay, research, and structure dashboards exceed the backend model.
- Drones may overlap because occupancy and collision are absent.

## P1: Correctness And Security

- Hot patches are unauthenticated; exposure is restricted to loopback.
- P-Script has a per-loop cap but no aggregate instruction budget.
- Per-drone JSON telemetry will not scale to the stated 10,000+ entity target.
- Replay seed and RNG-state serialization are absent.
- Abrupt Godot termination can orphan the core process.

## P2: Tooling And Presentation

- `client/external/` contains large reference repositories (e.g. `GodotDynamicInventorySystem`, `godot-shadcn-ui-kit`, `willowui`) and triggers Godot warnings and slows down editor startup.
- Desktop visual QA was blocked by missing macOS Computer Use permission; headless runtime passed.
- The legacy editor/world scene is not integrated with scenario goals.

## Recommended Build Order

1. Versioned scenario schema loaded by the Go core.
2. Deterministic mission state and win/loss evaluation.
3. Save/replay format including RNG state.
4. Occupancy and collision arbitration.
5. Aggregate density telemetry.
6. External repository cleanup.
