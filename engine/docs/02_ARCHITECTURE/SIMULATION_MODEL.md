# SIMULATION_MODEL.md

## Purpose

This document defines the strict rules governing how time and logic advance within *Project Chrysalis*. Because the game simulates massive entity counts and complex network behaviors, a rigid, deterministic execution pipeline is the only way to ensure the simulation does not collapse under its own weight or produce irreproducible bugs.

---

## The Chrysalis Execution Pipeline (Go Core)

### 1. Authoritative Heartbeat (10Hz)

The simulation is completely decoupled from the rendering frame rate.

* **Base Rate:** The engine operates at a fixed 10 Ticks Per Second (100ms interval).
* **Justification:** Balancing real-time responsiveness with sufficient processing time for thousands of per-entity P-Script AST walks and double-buffered spatial updates.

### 2. Strict Sequential Tick Order

To prevent race conditions and ensure predictable outcomes, the simulation processes systems in a strict, unyielding sequence during every tick:

1. **Environment Pass:** Pheromone evaporation and diffusion occur first. 
2. **Hazard Pass:** Environmental threats (Magnetic Anomaly, Thermal Geyser) calculate spatial intersections and mutate drone components (Battery drain).
3. **Contagion Pass:** Alien Node radii and infected drones bleed `CorruptionFactor` and logic viruses to healthy neighbors.
4. **Economic Pass:** The Hub checks accumulated silicates; if >= 5, a fabrication event is triggered, reallocating ECS slices and spawning a new drone.
5. **Logic Execution (P-Script):** The interpreter walks the active `agent.ps` AST for each entity. Drones query their components via builtins (e.g., `SENSE_BATTERY()`, `BROADCAST_VOTE()`).
6. **Buffer Swap:** Final spatial mutations (moves, signal drops) are committed to the `CurrentCells` layer, ensuring a consistent world state for the next tick.

### 3. Bit-Perfect Determinism

Chrysalis guarantees identical cross-platform behavior down to the bit.

* **Fixed-Point Arithmetic:** All spatial coordinates, distances, and intensities use the `crysmath` library ($10^6$ precision). Floating-point math is strictly forbidden in authoritative paths.
* **Double-Buffered State:** The `Grid` uses a Read/Write buffer split (`CurrentCells` / `NextCells`) to ensure no drone perceives partial or mid-tick updates from its peers.

---

## The Network Boundary (WebSocket IPC)

The simulation core operates as a standalone server listening on `:8080/telemetry`.

* **Asynchronous Telemetry:** State snapshots (`EMISSION_SNAPSHOT`) are broadcast as sparse JSON packets over WebSockets, shielding the tick loop from I/O pressure.
* **Bi-Directional Command Center:** The Godot client can inject `COMMAND_INJECTION` packets to hot-patch drone logic or trigger breakpoints without stopping the Go process.
* **Remote Observability:** The Architect can connect any compatible terminal (web, Godot, or CLI) to the core hub for real-time diagnostics.
