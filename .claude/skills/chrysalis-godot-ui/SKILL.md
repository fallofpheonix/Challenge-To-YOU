---
name: chrysalis-godot-ui
description: The Godot 4 client — GameHub router, the 10 telemetry screens, overlays, theme system, and the presentation-only contract. Use when editing engine/client GDScript/scenes or adding UI.
---

# Chrysalis Godot Client (UI)

**Presentation only.** The client renders telemetry and sends commands; it holds no authoritative game state (that lives in the Go core — see [[chrysalis-engine-architecture]]). Godot **4.0**, project at `engine/client/`.

## Structure
- `main.gd` / `main.tscn` — main controller; launches the Go core.
- `network_bridge.gd` — WebSocket client (see [[chrysalis-networking]]).
- `GameHub` (`ui/navigation/game_hub.gd`) — **autoloaded singleton**; routes telemetry to the 10 screens.
- Theme (autoloaded): `ui/theme/chrysalis_theme.gd`, `ui/theme/chrysalis_colors.gd`.

## The 10 screens (`ui/screens/`)
telemetry_dashboard · drone_inspector · resource_logistics · pheromone_view · structure_manager · hazard_monitor · alien_detector · research_tree · uplink_terminal · replay_controls.
Overlays (`ui/overlays/`): heatmap, pheromone, hazard, alien. Reusable: `ui/components/entity_row.tscn`, `inspector_modal.gd`.

## Rules
- All incoming sim numbers are fixed-point (×10^6) — **divide by 10^6 for display**.
- Aggregate thousands of drones into heatmaps/overlays rather than one node each; target 60 FPS (Tier-3 testing feeds chaotic mock JSON to verify).
- GDScript conventions: `snake_case` files/functions/vars, `PascalCase` classes/nodes, **past-tense `snake_case` signals** (`connection_dropped`). See [[chrysalis-coding-standards]].
- Existing third-party UI kits live (empty scaffolds) under `engine/client/external/`; the godot-mcp editor addon is staged at `engine/client/addons/godot_mcp/` (enable in Project Settings → Plugins).
