---
name: chrysalis-networking
description: The WebSocket bridge between the Go core and Godot client — telemetry broadcast, inbound command injection, and the ChrysalisTelemetryPacket schema. Use when editing network/hub.go, network_bridge.gd, or the bridge contract.
---

# Chrysalis Networking Bridge

Go core is the WebSocket **server**; Godot is the **client** (ADR-005). Endpoint: `ws://127.0.0.1:8080/telemetry`. Decouples core from client, supports multiple observers and remote command/script injection at ~1ms/tick.

## Go side (`engine/core/network/hub.go`)
- `NetworkHub` (`NewNetworkHub()`), `Run()` — broadcast loop.
- `HandleConnections(w, r)` — upgrade HTTP → WebSocket.
- `StartReader(conn, commandChannel)` — reads `InboundCommand`s from a client onto a channel (bidirectional: telemetry out, commands/hot-patched scripts in).
- `InboundCommand` — the client→core command envelope.

## Godot side
`network_bridge.gd` (WebSocket client) → `GameHub` autoload routes packets to the 10 screens. See [[chrysalis-godot-ui]].

## Wire contract (`engine/bridge_schema.json`, draft-07)
`ChrysalisTelemetryPacket`:
- `packet_type` — const `"EMISSION_SNAPSHOT"`
- `tick` — int ≥ 0
- `payload` — required: `tick`, `drones[]`, `grid`, `hazards`, `aliens`, `colony_res`, `swarm_size`
- each drone: `id`, `x`, `y`, `state`, `inv`, `bat`, `comp`, …

## Rules
- All numeric sim values crossing the bridge are fixed-point (×10^6); the **client** divides for display, the core never sends floats.
- Any change to the packet shape must update `bridge_schema.json` and `network_bridge.gd` together — the schema is the contract.
- Telemetry is serialized from the EventBus/ECS at end-of-tick (see [[chrysalis-eventbus]], [[chrysalis-simulation]]).
