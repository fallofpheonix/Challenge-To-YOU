# Current Implementation State

Status: playable engineering prototype, not a complete game.

## Implemented

- Go authoritative simulation at 10 Hz.
- Seeded deterministic RNG per engine instance.
- Fixed-point positions and integer authoritative state.
- Double-buffered spatial grid.
- Search/return AI using home and resource pheromones.
- Harvest, deposit, colony resources, and drone fabrication.
- Magnetic hazards, inert drones, alien infection, trust, and quorum sensing.
- Hardcoded v0 mission state with resource-target victory, infection loss, tick-limit loss, and Godot terminal banner.
- P-Script lexer, parser, bounded AST interpreter, hot reload, and per-drone variables.
- Loopback WebSocket telemetry and command injection.
- Godot telemetry, swarm, resource, pheromone, hazard, alien, and uplink screens.

## Verified

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- Godot headless project validation.
- Direct client/core binary launch, reconnect, telemetry handshake, and tick progression.
- End-to-end default P-Script resource harvest and return.

## Not Implemented

- Level JSON integration with the swarm simulation.
- Scoring, campaign progression, or saves.
- Replay serialization.
- Research effects, structures, multiplayer, leaderboard, or rewards.
- Thermal hazards.
- VM bytecode execution; the AST interpreter is active.
- Collision and occupancy constraints.
- Authentication; IPC is loopback-only.
- Production packaging.

See [KNOWN_LIMITATIONS.md](./KNOWN_LIMITATIONS.md).
