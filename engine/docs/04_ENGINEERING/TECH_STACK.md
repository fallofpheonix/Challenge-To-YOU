# TECH_STACK.md

## Purpose

This document defines the strict technological foundations of *Project Chrysalis*. Because the architecture completely decouples the heavy deterministic simulation from the visual presentation, a split tech stack is utilized to maximize CPU efficiency and rendering flexibility.

## 1. Frontend Engine (The Presentation Layer)

* **Framework:** Godot Engine 4.x
* **Language:** GDScript
* **Role:** A highly optimized "dumb terminal." It handles the rendering of 2D data grids, heatmap shaders, and the UI of the Architect's Terminal. It does not run any game physics, drone AI, or state logic. It merely listens for state updates from the backend and broadcasts the Architect's raw text scripts back.

## 2. Backend Engine (The Simulation Core)

* **Language:** Go (Golang)
* **Architecture:** Strict Entity-Component-System (ECS) with a decoupled, fixed-timestep loop.
* **Role:** The undisputed source of truth. Go was chosen specifically for its concurrency model (goroutines and channels). It handles the spatial hashing of 50,000+ drones, mathematical pheromone decay, collision resolution, and the secure sandboxed execution of the player's logic, all within a 100ms tick budget.

## 3. Libraries & Dependencies

To guarantee absolute cross-platform determinism, reliance on external black-box libraries is aggressively minimized.

* **Networking:** `gorilla/websocket` (Go) for persistent, low-latency, bidirectional telemetry streaming between the Go Core and the Godot Client.
* **Logic Parsing:** A custom, restricted Python Abstract Syntax Tree (AST) lexer/parser written natively in Go. No external Python interpreters (like CPython) are embedded, ensuring complete air-gapped security and instruction-level timeout control.
* **Math:** Custom Fixed-Point library (`phxmath`) to ensure cross-platform determinism without IEEE-754 drift.
* **Physics:** Zero external physics engines (no Box2D). All collision and movement logic is handled via custom fixed-point/integer spatial grid math natively in Go.

## 4. Build System & Deployment

* **Version Control:** Git, structured as a monorepo containing both the `godot_client` and `go_server` directories.
* **Continuous Integration:** GitHub Actions. Every commit triggers a strict suite of Go unit tests specifically validating the determinism of the tick engine (e.g., feeding a seed and a script, and asserting the exact grid state 1,000 ticks later).
* **Distribution:** The final build bundles the headless Go server executable alongside the Godot client binaries. Upon launch, Godot spins up the Go server as a background process and automatically connects via a local WebSocket port, allowing seamless offline play while preserving the client/server architecture for future orbital leaderboards.
