# ENTITY_MODEL.md

## Purpose

This document defines the core simulated entities of *Project Chrysalis*. Because the simulation operates on a strict Entity-Component-System (ECS) architecture, these definitions represent the pure data structures and states that the engine processes during every tick, rather than object-oriented classes.

---

### 1. Drone (Authoritative ECS Implementation)

The physical actuator of the swarm. Managed as contiguous component slices in `SwarmRegistry`.

* **State:** `Searching` (0) | `Returning` (1) | `Inert` (2) | `Compromised` (bool)
* **Components:**
    * `ID` (uint32)
    * `PositionX / PositionY` (FixedPoint)
    * `Battery` (int64) - Drains via movement and hazards.
    * `Inventory` (int32) - Payload carrying silicates.
    * `CorruptionFactor` (uint8) - Percent of logic manipulation [0-100].
    * `TrustScore` (int32) - Peer-to-peer reliability metric.

* **Lifecycle:** Fabricated at the Hub (5 silicates) -> Dispatched -> Loops P-Script -> Depletes battery or hits hazard -> Becomes `Inert` at zero power -> Compromised by Alien Nodes if nearby.

### 2. Resource (Silicates)

Physical materials existing on the grid, tracked in the `Cell` struct.

* **Properties:**
    * `ResourceCount` (int32) - Per-cell silicate density.
    * `GlobalSilicates` (int32) - Total pool stored at the Hub for fabrication.

* **Lifecycle:** Seeded during initialization -> Harvested by Drone (`HARVEST()`) -> Decrements cell count, increments Drone inventory -> Transported to Base (`IsBase`) -> Increments Global pool -> Consumed to spawn new drones.

### 3. Pheromones & Signals

Double-buffered grid layers for decentralized coordination.

* **Layers:**
    * `HomePheromone`: Gradient back to the Hub (Base).
    * `ResourcePheromone`: Gradient to detected silicate patches.
    * `AlienSignal`: Misleading purple trails emitted by `Compromised` drones.
* **Dynamics:**
    * `Intensity` (int32): Fixed-point density.
    * `Decay`: Fixed integer reduction per tick.
    * `Sensing`: Neighbors sampled via `SenseHighestGradient`.

### 4. Environmental Hazards

Dynamic spatial triggers that mutate drone components.

* **Type:** `HazardMagnetic` (Battery drain) | `HazardThermal` (Physical damage).
* **Properties:**
    * `Epicenter` (X, Y)
    * `Radius` (int32)
    * `Intensity` (int64) - Scaled effect per tick.

* **Loop:** Scans drones within radius -> Applies battery drain -> Forces state to `Inert` if empty.

### 5. Alien Network (Act III)

The adversarial intelligence competeing for control of the swarm.

* **Type:** `NodeInfector` (Logic Virus) | `NodeJammer` (Communication loss).
* **Dynamics:**
    * `InfectionRadius`: Area of wireless contagion.
    * `Contagion`: 5% probability per tick to flip `Compromised` status and increase `CorruptionFactor`.
    * `Spoofing`: Injects `AlienSignal` into the grid to reroute healthy drones.
