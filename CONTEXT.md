# Project Chrysalis: The Architect's Swarm — CONTEXT

## Domain Glossary

| Term | Definition |
|------|-----------|
| **Architect** | The player. Programs drone behavior via P-Script. Does not directly pilot drones. |
| **Swarm** | Collection of autonomous micro-drones executing decentralized logic. |
| **P-Script** | Custom DSL for programming drone behavior. Compiled to AST, tree-walk interpreted. |
| **ECS** | Entity Component System. Data-oriented layout using contiguous slices (SoA). |
| **SwarmRegistry** | The ECS data store. Holds PositionX/Y, Battery, State, Inventory, etc. as parallel slices. |
| **FixedPoint** | Deterministic math using integer scaling (Precision = 10^6). No floats in simulation. |
| **Double Buffering** | Grid uses CurrentCells/NextCells. Mutations stage in Next, commit atomically via SwapBuffers. |
| **Pheromone** | Chemical signal emulation. Home (trail back to base), Resource (trail to food), Alien (corruption). |
| **Hazard** | Environmental danger zone. Magnetic fields drain battery. Thermal (planned) deals physical damage. |
| **Alien Network** | Hostile nodes that spread logic virus (compromise) to nearby drones. |
| **Compromised** | Drone infected by alien virus. Deposits alien signals instead of home pheromone. |
| **TrustScore** | Peer-to-peer validation score (0-100). Drops on compromise. Used by quorum sensing. |
| **Quorum** | Consensus mechanism. Drones vote on neighbor trustworthiness. >50% = trustworthy. |
| **Fabrication** | Swarm self-replication. Costs 5 silicates. Creates new drone at base. |
| **Silicates** | Primary resource. Harvested by drones, deposited at base, used for fabrication. |
| **GameHub** | Autoloaded Godot singleton. Routes telemetry data to 10 specialized UI screens. |
| **NetworkBridge** | WebSocket client in Godot. Connects to Go core's telemetry server. |

## Architecture Decisions

### ADR-001: Fixed-Point Arithmetic
**Decision**: All simulation math uses `crysmath.FixedPoint` with `Precision = 10^6`.
**Rationale**: Floating-point is non-deterministic across platforms. Fixed-point guarantees bit-perfect reproducibility.
**Consequence**: All positions, battery levels, and pheromone values are integer-scaled. Client must divide by 10^6 for display.

### ADR-002: Data-Oriented ECS
**Decision**: Drone state stored as parallel slices (SoA) rather than array of structs (AoS).
**Rationale**: CPU cache lines favor contiguous access to individual components. With 1000+ drones, this matters.
**Consequence**: Adding/removing entities requires slice copying. No pointer indirection.

### ADR-003: Double-Buffered Grid
**Decision**: Grid uses CurrentCells/NextCells with atomic swap.
**Rationale**: Allows drones to read consistent state while writing to staging buffer. Prevents race conditions.
**Consequence**: One tick of latency between action and visibility. By design for determinism.

### ADR-004: Tree-Walk Interpreter (Not Bytecode VM)
**Decision**: P-Script is interpreted via AST walking, not compiled to bytecode.
**Rationale**: Simplicity for MVP. Hot-reload is instant (re-parse source). 100-iteration loop safety limit prevents hangs.
**Consequence**: Slower than bytecode. Acceptable for 100-500 drones at 10Hz. VM stub exists for future optimization.

### ADR-005: WebSocket for Core-Client Bridge
**Decision**: Go core runs WebSocket server, Godot connects as client.
**Rationale**: Decouples core from client. Multiple clients can observe same simulation. Supports remote command injection.
**Consequence**: Requires network bridge in Godot. Adds ~1ms latency per tick (negligible at 10Hz).

## System Boundaries

```
┌─────────────────────────────────────────────────────┐
│                   Godot Client                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │
│  │ main.gd  │→ │ GameHub  │→ │ 10 Screen Views  │  │
│  │ (spawn)  │  │ (router) │  │ (telemetry, etc) │  │
│  └────┬─────┘  └──────────┘  └──────────────────┘  │
│       │                                              │
│  ┌────▼─────────────┐                               │
│  │ NetworkBridge    │ ← WebSocket client            │
│  │ (WebSocket)      │                               │
│  └────┬─────────────┘                               │
└───────┼─────────────────────────────────────────────┘
        │ ws://127.0.0.1:8080/telemetry
┌───────┼─────────────────────────────────────────────┐
│  ┌────▼─────────────┐                               │
│  │ network/hub.go   │ ← WebSocket server            │
│  │ (broadcast)      │                               │
│  └────┬─────────────┘                               │
│       │                                              │
│  ┌────▼─────────────┐  ┌──────────────────────┐    │
│  │ main.go          │→ │ simulation.Engine    │    │
│  │ (10Hz loop)      │  │ (deterministic core) │    │
│  └────┬─────────────┘  └──────────────────────┘    │
│       │                                              │
│  ┌────▼─────────────┐  ┌──────────────────────┐    │
│  │ pscript/         │  │ simulation/          │    │
│  │ (lexer→parser→   │  │ (ECS, grid, hazards, │    │
│  │  interpreter)    │  │  aliens, pheromones) │    │
│  └──────────────────┘  └──────────────────────┘    │
│                   Go Core                            │
└─────────────────────────────────────────────────────┘
```

## P-Script Language Reference

### Keywords
`fn`, `let`, `if`, `else`, `while`, `return`, `true`, `false`

### Operators
`+`, `-`, `*`, `/`, `<`, `>`, `<=`, `>=`, `==`, `!=`, `=`, `!`

### Built-in Functions (Swarm API)

| Function | Returns | Description |
|----------|---------|-------------|
| `SENSE_RESOURCE()` | bool | True if resource pheromone detected nearby |
| `SENSE_HOME()` | bool | True if home pheromone detected nearby |
| `SENSE_CARGO()` | bool | True if drone is carrying silicates |
| `SENSE_BATTERY()` | int64 | Current battery level (scaled by 10^6) |
| `SENSE_TRUST()` | int64 | Peer trust score (0-100) |
| `SENSE_CORRUPTION()` | int64 | Corruption factor (0-100) |
| `SENSE_COMPROMISED()` | bool | True if infected by alien virus |
| `SENSE_ALIEN_SIGNAL()` | bool | True if alien signal detected nearby |
| `SENSE_SWARM_SIZE()` | int64 | Total drones in swarm |
| `SENSE_COLONY_RESOURCES()` | int64 | Total silicates in colony cache |
| `BROADCAST_VOTE()` | bool | Quorum consensus with neighbors |
| `HARVEST()` | bool | Harvest resource at current cell |
| `DROP_RESOURCE()` | bool | Deposit cargo at base |
| `MOVE_RANDOM()` | bool | Move to random adjacent cell |
| `MOVE_TOWARDS_RESOURCE()` | bool | Follow resource pheromone gradient |
| `MOVE_TOWARDS_HOME()` | bool | Follow home pheromone gradient |

### Example Script
```
fn main() {
    if (SENSE_BATTERY() < 25000000) {
        MOVE_TOWARDS_HOME()
    } else {
        if (SENSE_CARGO()) {
            DROP_RESOURCE()
            MOVE_TOWARDS_HOME()
        } else {
            HARVEST()
            if (SENSE_CARGO()) {
                MOVE_TOWARDS_HOME()
            } else {
                MOVE_TOWARDS_RESOURCE()
            }
        }
    }
}
```

## Simulation Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| Grid Size | 100x100 | World dimensions in cells |
| Tick Rate | 10 Hz | Simulation updates per second |
| Precision | 10^6 | Fixed-point scaling factor |
| Pheromone Decay | 5000/tick | Signal evaporation rate |
| Max Pheromone | 1000000 | Signal saturation ceiling |
| Battery Drain | 1000/tick | Movement cost per step |
| Hazard Drain | 1000000/tick | Magnetic field battery drain |
| Infection Radius | 3 cells | Alien virus spread range |
| Corruption Threshold | 100 | Full compromise at 100% |
| Fabrication Cost | 5 silicates | Resources needed for new drone |
| Max Swarm | 500 | Safety cap for MVP |
| While Loop Limit | 100 | Max iterations per while loop |

## File Map

### Go Core (`engine/core/`)
- `main.go` — Entry point, 10Hz loop, WebSocket server, hot-reload
- `main_test.go` — Test suite for the Go core entry points
- `.air.toml` — Live reload configuration for local development
- `.golangci.yml` — Linter configuration
- `simulation/simulation.go` — Engine lifecycle, drone AI, state serialization
- `simulation/ecs.go` — SwarmRegistry (SoA data layout)
- `simulation/grid.go` — Double-buffered 2D grid
- `simulation/pheromones.go` — Signal evaporation and gradient sensing
- `simulation/hazards.go` — Hazard zone system
- `simulation/alien.go` — Alien network and infection spreading
- `crysmath/fixedpoint.go` — Deterministic fixed-point arithmetic
- `pscript/token/token.go` — Token type definitions
- `pscript/lexer/lexer.go` — Lexical scanner
- `pscript/parser/parser.go` — Recursive-descent Pratt parser
- `pscript/ast/ast.go` — Abstract syntax tree nodes
- `pscript/interpreter/interpreter.go` — Tree-walk interpreter
- `pscript/vm/` — Stub for future bytecode VM implementation
- `network/hub.go` — WebSocket broadcast hub
- `state/` — Directory for replay state serialization (planned)
- `levels/` — Directory for JSON-based scenario files
- `scripts/agent.ps` — Default swarm behavior script

### Godot Client (`engine/client/`)
- `main.gd` / `main.tscn` — Main controller, Go core launcher
- `network_bridge.gd` — WebSocket client
- `ui/theme/chrysalis_theme.gd` — Design system (autoloaded)
- `ui/theme/chrysalis_colors.gd` — Color palette
- `ui/navigation/game_hub.gd` / `.tscn` — Screen router (autoloaded)
- `ui/screens/telemetry_dashboard.gd` — Tick/bandwidth/log display
- `ui/screens/drone_inspector.gd` — Drone list and detail panel
- `ui/screens/resource_logistics.gd` — Resource flow visualization
- `ui/screens/pheromone_view.gd` — Signal strength display
- `ui/screens/structure_manager.gd` — Blueprint catalog
- `ui/screens/hazard_monitor.gd` — Threat level display
- `ui/screens/alien_detector.gd` — Corruption monitoring
- `ui/screens/research_tree.gd` — Tech tree
- `ui/screens/uplink_terminal.gd` — Code deployment queue
- `ui/screens/replay_controls.gd` — Playback controls
- `ui/components/entity_row.tscn` — Reusable list row
- `ui/components/inspector_modal.gd` — Breakpoint inspector
- `ui/overlays/heatmap_overlay.gd` — Density visualization
- `ui/overlays/pheromone_overlay.gd` — Signal visualization
- `ui/overlays/hazard_overlay.gd` — Danger zone rendering
- `ui/overlays/alien_overlay.gd` — Corruption visualization
