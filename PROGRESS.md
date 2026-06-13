Original prompt: build this game brick by brick and proper working ai and use it and find loopholes and fix them all and docuements everything

# Work Log

## 2026-06-12

- Audited Go core, P-Script, WebSocket bridge, and Godot client.
- Repaired omitted production systems and the default AI base deadlock.
- Added explicit tick lifecycle, seeded RNG, cargo sensing, inert guards, and tests.
- Isolated P-Script variables per drone.
- Switched Godot to the built core binary.
- Fixed GDScript compilation, telemetry mappings, dashboard visibility, and app icon.
- Restricted IPC to loopback and bounded inbound frames.
- Added current-state, runbook, testing, and limitation documents.

## Verification

- Go tests, race detector, and vet: pass.
- Godot headless validation: pass.
- Direct client/core handshake: pass.
- Desktop visual inspection: blocked by macOS Computer Use permissions.

## Next Work

- Scenario and mission state.
- Replay/save state.
- Occupancy and collision.
- WebSocket and Godot interaction tests.
