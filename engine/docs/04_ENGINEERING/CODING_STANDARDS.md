# CODING_STANDARDS.md

## Purpose

This document establishes the structural and stylistic rules for developing *Project Chrysalis*. Because the project utilizes a split-stack architecture (Go backend, Godot frontend) and relies on absolute deterministic simulation, strict adherence to these standards is required to prevent desyncs, memory leaks, and unmaintainable code.

## 1. Folder Structure (The Monorepo)

The repository is organized to enforce the strict air-gap between the simulation logic and the visual presentation.

```text
Project-Chrysalis/
├── docs/              # Architectural "North Star" documents
├── godot_client/      # The visual frontend
│   ├── scenes/        # UI layouts, Grid renderers, Heatmaps
│   ├── scripts/       # GDScript network listeners and visual logic
│   └── assets/        # Shaders, UI themes, generic sprites
├── go_server/         # The authoritative backend
│   ├── core/          # The tick loop, ECS manager, Spatial Hash Grid
│   ├── systems/       # Pheromone decay, collision, hazard execution
│   └── sandbox/       # The Python AST parser and strict execution environment
└── shared/            # Shared data contracts (JSON schemas/Protobufs)
```

## 2. Naming Conventions

Because the project spans multiple languages, naming conventions are strictly mapped to the specific domain and language being written.

### Go (The Simulation Core)

* **Files:** `snake_case.go` (e.g., `spatial_grid.go`).
* **Structs & Interfaces:** `PascalCase` (e.g., `DroneComponent`, `PheromoneSystem`).
* **Variables & Functions:** `camelCase` for internal, `PascalCase` for exported.
* **ECS Specifics:** Pure data structs must end in `Component` (e.g., `BatteryComponent`). Logic iterators must end in `System` (e.g., `MovementSystem`).

### GDScript (The Godot Client)

* **Files:** `snake_case.gd` and `snake_case.tscn`.
* **Classes & Nodes:** `PascalCase` (e.g., `TelemetryBroadcaster`, `GridHeatmap`).
* **Variables & Functions:** `snake_case` (e.g., `update_drone_position()`).
* **Signals:** `snake_case` in past tense (e.g., `signal_received`, `connection_dropped`).

### Python (The Architect's Sandbox)

* **Variables & Functions:** Strict PEP 8 `snake_case` (e.g., `drone.get_adjacent_cells()`).
* **Constants & Signals:** `UPPER_SNAKE_CASE` (e.g., `"RESOURCE_GRADIENT"`, `"NORTH"`).

## 3. Testing Standards

Given the complexity of thousands of interacting rules, the simulation backend must be aggressively tested. UI testing is secondary to engine determinism.

### Tier 1: Determinism Validation (Go)

This is the most critical test suite.

* **The Hash Test:** A test that initializes the ECS with a specific world seed, injects a static Python script for 1,000 drones, runs exactly 10,000 ticks, and computes a cryptographic hash of the final grid state. If the hash ever changes between commits, the build fails immediately.
* **No Flakes:** Determinism tests must run 100% reliably. Any test that occasionally fails indicates a race condition or floating-point leak and is treated as a critical bug.

### Tier 2: The Sandbox Guillotine (Go)

Tests designed exclusively to attack the Rule Engine.

* **Malicious Injection:** Tests that attempt to pass `import os`, recursive functions, or deeply nested logic to ensure the AST parser structurally rejects them.
* **The Infinite Loop Test:** Tests that intentionally submit `while True: pass` to verify that the Sandbox accurately increments its operation counter, throws the `SandboxTimeout` exception, and allows the master tick to finish under 100ms without crashing the engine.

### Tier 3: Client Telemetry (Godot)

Frontend testing is kept lightweight to avoid slowing down rapid UI iterations.

* **Mock State Ingestion:** Tests that feed massive, chaotic, pre-recorded JSON/WebSocket arrays into the Godot client to ensure the renderer maintains 60 FPS while updating thousands of visual nodes and heatmap shaders.
