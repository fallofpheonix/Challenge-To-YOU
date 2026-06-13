# Runtime Runbook

## Build And Verify

Requirements: macOS, Go 1.26+, and Godot 4.x at `/Applications/Godot.app`.

```bash
cd engine
make check
make test-race
make build-core
```

## Run

```bash
cd engine
make run-all
```

Godot starts `bin/chrysalis-core`. The core reads `PHX_SCRIPT_PATH` and listens on `127.0.0.1:8080`.

## Standalone Core

```bash
cd engine
make run-core
```

WebSocket endpoint: `ws://127.0.0.1:8080/telemetry`.

## P-Script Deployment

The active file is `core/scripts/agent.ps`. Godot writes it on deployment; the core reloads it on modification.

```json
{
  "packet_type": "COMMAND_INJECTION",
  "tick": 0,
  "payload": {
    "code": "fn main() { MOVE_TOWARDS_RESOURCE() }"
  }
}
```

Invalid patches retain the last valid program.

## Failure Checks

```bash
lsof -nP -iTCP:8080 -sTCP:LISTEN
make build-core
make check
```

Abruptly killing Godot can leave the child core running. Terminate the process bound to port 8080 before restarting.
