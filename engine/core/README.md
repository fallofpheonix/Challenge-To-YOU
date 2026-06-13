# Chrysalis Engine Core

> **Role**: Authoritative Simulation & Logic (The "Brains")

The Go-based simulation engine powering Project Chrysalis. Runs at 10Hz, executes P-Script autonomous logic, and emits deterministic state to the Godot client via WebSocket.

## Quick Start

```bash
# Build the core binary
make build-core

# Run the core directly
make run-core

# Run tests
make test-core

# Run the full game (Godot client + Go core)
make run-all
```

## Architecture

### ECS (Entity Component System)
- **SwarmRegistry** (`simulation/ecs.go`): SoA layout with contiguous slices for PositionX, PositionY, Battery, State, Inventory, Compromised, TrustScore, CorruptionFactor
- **Dynamic expansion**: Slices double in capacity when full
- **No entity removal**: Inert drones remain in registry (optimization opportunity)

### Grid System
- **Double-buffered**: `CurrentCells` (read) / `NextCells` (write) with atomic swap
- **Pheromone decay**: Home and Resource signals decay by 5000/tick, Alien signals by 10000/tick
- **Gradient sensing**: 3x3 neighborhood scan for strongest signal

### P-Script Language
- **Lexer**: Tokenizes keywords, identifiers, integers, operators
- **Parser**: Recursive-descent Pratt parser with precedence climbing
- **Interpreter**: Tree-walk evaluator with variable storage, infix/prefix expressions
- **Built-in functions**: 16 swarm API functions (sensors + actuators)

### Networking
- **WebSocket server**: Broadcasts full state snapshot at 10Hz
- **Command injection**: Remote script override via WebSocket packets
- **Hot-reload**: Watches `agent.ps` for file changes, re-parses automatically

## P-Script Language

### Operators
`+` `-` `*` `/` `<` `>` `<=` `>=` `==` `!=` `=` `!`

### Keywords
`fn` `let` `if` `else` `while` `return` `true` `false`

### Built-in Functions

| Function | Returns | Description |
|----------|---------|-------------|
| `SENSE_RESOURCE()` | bool | Resource pheromone nearby |
| `SENSE_HOME()` | bool | Home pheromone nearby |
| `SENSE_CARGO()` | bool | Carrying silicates |
| `SENSE_BATTERY()` | int64 | Battery level (scaled 10^6) |
| `SENSE_TRUST()` | int64 | Peer trust score (0-100) |
| `SENSE_CORRUPTION()` | int64 | Corruption factor (0-100) |
| `SENSE_COMPROMISED()` | bool | Infected by alien virus |
| `SENSE_ALIEN_SIGNAL()` | bool | Alien signal detected |
| `SENSE_SWARM_SIZE()` | int64 | Total drones in swarm |
| `SENSE_COLONY_RESOURCES()` | int64 | Total silicates in colony |
| `BROADCAST_VOTE()` | bool | Quorum consensus |
| `HARVEST()` | bool | Harvest resource at cell |
| `DROP_RESOURCE()` | bool | Deposit cargo at base |
| `MOVE_RANDOM()` | bool | Random adjacent move |
| `MOVE_TOWARDS_RESOURCE()` | bool | Follow resource gradient |
| `MOVE_TOWARDS_HOME()` | bool | Follow home gradient |

### Example
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

## State Serialization

The engine emits JSON packets at 10Hz with this structure:

```json
{
  "packet_type": "EMISSION_SNAPSHOT",
  "tick": 1234,
  "payload": {
    "tick": 1234,
    "drones": [
      {
        "id": 0,
        "x": 50000000,
        "y": 50000000,
        "state": 0,
        "inv": 0,
        "bat": 999000,
        "comp": false,
        "trust": 100,
        "corr": 0
      }
    ],
    "grid": [...],
    "hazards": [...],
    "aliens": [...],
    "colony_res": 42,
    "swarm_size": 100
  }
}
```

**Note**: All position values are scaled by `Precision` (10^6). Divide by 1000000 for display coordinates.

## Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./pscript/lexer/
go test ./pscript/parser/
go test ./pscript/interpreter/
go test ./simulation/
```

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `PHX_SCRIPT_PATH` | `scripts/agent.ps` | Path to P-Script agent behavior |

## Dependencies

- `github.com/gorilla/websocket` — WebSocket implementation
- Go 1.26+ (uses `clear()` builtin)
