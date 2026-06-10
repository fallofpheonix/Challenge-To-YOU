# PERFORMANCE_BUDGET.md

## Purpose

This document defines the uncompromising computational boundaries of *Project Chrysalis*. Because the game relies on simulating thousands of entities executing custom Python logic within a strict deterministic loop, any architectural decision that violates these budgets must be rejected. The engine must scale gracefully to prevent the game from turning into a slideshow during the climax.

## The Master Constraint

The simulation runs at exactly 10 Ticks Per Second (TPS).
Therefore, **1 Tick = 100 Milliseconds.**
The entire Go backend (Environment, Pheromones, AST Parsing, Drone Logic, Collision, and Serialization) must complete its cycle within this 100ms window on an average consumer CPU.

---

## Entity Scaling Targets

### 100 Drones (The Baseline)

* **Context:** Act I onboarding and localized testing environments.
* **Tick Budget Allocation:** `< 1 Millisecond`
* **Requirement:** At this scale, overhead should be statistically invisible. The parser, ECS, and spatial grid must initialize and execute without any noticeable CPU spike.

### 1,000 Drones (The Logistics Network)

* **Context:** Act II expansion. Multiple Hubs, established highways, and active environmental hazards.
* **Tick Budget Allocation:** `< 10 Milliseconds`
* **Requirement:** The primary cost here will be Phase 2 (Pheromone Decay). The engine must update and diffuse thousands of signal values across the spatial grid without iterating through empty cells. Spatial hashing efficiency is validated at this tier.

### 10,000 Drones (The Swarm Equilibrium)

* **Context:** Act III. The intended "healthy" late-game state. Massive cross-map supply chains and active skirmishes with the Alien Network.
* **Tick Budget Allocation:** `< 50 Milliseconds`
* **Requirement:** This is the critical design target. 50ms is allocated to the simulation core, leaving a 50ms buffer for OS overhead, garbage collection, and WebSocket JSON serialization for the Godot client. The custom Python AST Sandbox must prove it can evaluate 10,000 individual scripts simultaneously without breaching this limit.

### 100,000 Drones (The Hard Ceiling)

* **Context:** Act IV stress testing, unchecked replication loops, or global gridlock.
* **Tick Budget Allocation:** `<= 100 Milliseconds`
* **Requirement:** This is the architectural breaking point. If the swarm exceeds this count, the Go backend will begin dropping below 10 TPS, causing the game to slow down in real-time. This is treated as a gameplay feature, not a bug: it physically punishes the Architect for relying on brute-force replication instead of optimized, elegant logistics.

---

## Secondary Budgets

### 1. Memory Allocation (The GC Rule)

* **Constraint:** Zero allocations per tick during steady-state simulation.
* **Reasoning:** Go's Garbage Collector (GC) will cause unpredictable latency spikes if millions of short-lived objects are created every second.
* **Implementation:** The ECS must use pre-allocated, contiguous arrays (Object Pools) for Drones, Signals, and AST nodes. When a drone dies, its index is marked as reusable; it is never destroyed in memory.

### 2. Network Payload (The Telemetry Limit)

* **Constraint:** `< 5 Megabytes per Second (MB/s)` local WebSocket bandwidth.
* **Reasoning:** The Go server must broadcast the visual state to the Godot client 10 times a second.
* **Implementation:** The backend cannot send a JSON array of 50,000 individual drone coordinates. It must aggregate coordinate data into low-resolution density heatmaps and delta-compressed arrays before transmission.
