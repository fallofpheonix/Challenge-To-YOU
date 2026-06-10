# SIMULATION_MODEL.md

## Purpose

This document defines the strict rules governing how time and logic advance within *Project Chrysalis*. Because the game simulates massive entity counts and complex network behaviors, a rigid, deterministic execution pipeline is the only way to ensure the simulation does not collapse under its own weight or produce irreproducible bugs.

## The Execution Pipeline

### Tick Rate

The simulation is completely decoupled from the rendering frame rate.

* **Base Rate:** The engine operates at a fixed 10 Ticks Per Second (TPS).
* **Justification:** 10 TPS is fast enough to provide responsive, smooth logic for the swarm, but leaves a 100-millisecond buffer per tick for the CPU to process the ECS arrays of thousands of drones, hazard updates, and pheromone decay.

### Update Order

To prevent race conditions and ensure predictable outcomes, the simulation processes systems in a strict, unyielding sequence during every single tick. A phase must completely finish before the next begins.

1. **Phase 1: Environment & Hazards:** The world updates first. Magnetic anomalies shift, geysers erupt, and crust stability is calculated.
2. **Phase 2: Signal Propagation:** Pheromone markers decay by their mathematical evaporation rate. New scents dropped in the previous tick are added to the grid.
3. **Phase 3: Sensor Read (Perception):** All drones simultaneously read their immediate adjacent grid cells. They lock in their understanding of the world state for this tick.
4. **Phase 4: Rule Engine Execution (Logic):** The Architect's active scripts are processed. Every drone decides its next action based *only* on the data it locked in during Phase 3.
5. **Phase 5: State Write (Actuation):** All drones attempt to execute their chosen action (move, drop resource, emit signal). Collisions and traffic deadlocks are resolved here.

### Determinism

The simulation is strictly deterministic. If you supply the exact same initial state and the exact same Architect code, the swarm will behave identically 100% of the time, even millions of ticks later.

* **No Floating-Point Physics:** To avoid cross-platform rounding errors, all spatial logic, distance calculations, and pheromone intensities use integer math or strict fixed-point arithmetic.
* **Seeded Randomness:** True randomness does not exist. Any procedural generation or "random" hazard behavior uses a deterministic pseudo-random number generator (PRNG) initialized with a fixed map seed.

### The Replay System

Because the engine is strictly deterministic, the game does not need to save the individual data states of 50,000 active drones to load a save file.

* **Input Logging:** The game only records the initial world seed and a log of the Architect's code deployments paired with the exact tick they were executed (e.g., `[Tick 14052: Deploy Quarantine_Protocol_V2]`).
* **Debugging & Playback:** If a massive logic failure occurs, the player can drop a "Diagnostic Ping." The engine silently re-simulates the history from the last Uplink Window using the input log, allowing the player to seamlessly rewind time, step forward tick-by-tick, and pinpoint the exact moment a routing loop failed.
