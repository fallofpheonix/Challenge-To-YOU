# Current Implementation State

Status: playable v0 engineering prototype, not a complete game.

## Implemented

- Go authoritative simulation at 10 Hz.
- Seeded deterministic RNG per engine instance.
- Fixed-point positions and integer authoritative state.
- Double-buffered spatial grid.
- Search/return AI using home and resource pheromones.
- Harvest, deposit, colony resources, and drone fabrication.
- Deterministic return-to-base fallback for loaded drones on unobstructed paths; stale home trails that do not improve distance-to-base are ignored.
- Magnetic hazards, inert drones, alien infection, trust, and quorum sensing.
- Hardcoded v0 mission state with resource-target victory, infection loss, tick-limit loss, and Godot terminal banner.
- P-Script lexer, parser, bounded AST interpreter, hot reload, and per-drone variables.
- Loopback WebSocket telemetry and command injection.
- Godot telemetry, swarm, resource, pheromone, hazard, alien, and uplink screens.
- Split validation targets: fast Go-only checks, Godot smoke launch, and full Godot headless editor validation.

## Verified

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- Godot headless project validation.
- Direct client/core binary launch, reconnect, telemetry handshake, and tick progression.
- End-to-end default P-Script resource harvest and return.
- Default v0 mission reaches victory with the shipped P-Script and seeded adjacent resource node.
- Visual GUI verification of the terminal banner: `VICTORY` / `RESOURCE TARGET REACHED` renders and remains latched.
- Godot reconnect timer cleanup verified; prior transient `SceneTreeTimer` ObjectDB leak is fixed.
- No orphaned `Godot` or `chrysalis-core` process after tested exits.

## Not Implemented

- Level JSON integration with the swarm simulation.
- Scoring, campaign progression, or saves.
- Replay serialization.
- Research effects, structures, multiplayer, leaderboard, or rewards.
- Thermal hazards.
- VM bytecode execution; the AST interpreter is active.
- Collision and occupancy constraints.
- Obstacle/hazard-aware route planning.
- Authentication; IPC is loopback-only.
- Production packaging.

See [KNOWN_LIMITATIONS.md](./KNOWN_LIMITATIONS.md).
