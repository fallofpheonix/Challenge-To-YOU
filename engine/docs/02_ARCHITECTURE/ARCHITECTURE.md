# ARCHITECTURE.md

## Purpose

This document defines the structural hierarchy of the software. To simulate thousands of autonomous entities and complex networking protocols without CPU bottlenecks, the simulation must remain strictly data-oriented and completely decoupled from the rendering layer.

## High-Level Pipeline

```text
  [ The Architect's Terminal / UI ]  <-- Player Input & Visuals
               ↓ ↑
      [ The Telemetry Layer ]        <-- State Serialization & Heatmaps
               ↓ ↑
      [ Gameplay Systems ]           <-- Hazards, Pheromones, Uplink Queue
               ↓ ↑
       [ The Rule Engine ]           <-- P-Script Parser & Execution Sandbox
               ↓ ↑
     [ The Simulation Core ]         <-- Deterministic Tick, ECS, Spatial Grid

```

## Layer Definitions

### 1. The Simulation Core (Data Layer)

The foundational engine that holds the true state of the world. It is completely blind to visuals, graphics, and the player.

* **Deterministic Tick Loop:** The master clock. It executes logic in fixed, sequential timesteps (e.g., 10 ticks per second) entirely independent of the visual framerate, ensuring reproducible simulation states.
* **Entity-Component-System (ECS):** Object-oriented programming is abandoned here. Drones, resources, and hazards are purely integer IDs mapped to contiguous arrays of data (Position, Battery, Status).
* **Spatial Hash Grid:** A highly optimized 2D lookup table. This is critical for performance, allowing thousands of drones to instantly query their immediate adjacent cells without looping through global entity lists.

### 2. The Rule Engine (Logic Layer)

The bridge between human-readable protocols and the machine executing them.

* **Instruction Parser:** Translates the Architect's localized behavioral scripts into an executable format (e.g., an Abstract Syntax Tree or bytecode).
* **Execution Sandbox:** The isolated environment that runs the Architect's (and the Alien Network's) logic against the ECS data during each tick. It strictly limits execution time to prevent user-created infinite loops from crashing the main core.

### 3. Gameplay Systems (Mechanics Layer)

The managers that enforce the specific rules of Kepler-452b.

* **Pheromone Manager:** Iterates over the grid every tick to process the placement, intensity stacking, and mathematical decay (evaporation) of digital scents.
* **Hazard Orchestrator:** Triggers planetary environmental shifts, dynamically updating the physical collision and hazard layers of the Spatial Grid.
* **Uplink Controller:** Buffers the Architect's incoming code deployments and delays their injection into the Rule Engine until the orbital bandwidth window opens.

### 4. The UI / Telemetry Layer (Presentation Layer)

The only layer the Architect interacts with. It translates the raw data of the Core into readable visual spectacles and interactive tools.

* **State Broadcaster:** Reads the ECS data at the end of a tick and translates it into a format the visual engine can render.
* **Heatmap Renderer:** Aggregates individual drone coordinates into fluid, macro-level density and threat visualizations to prevent visual noise.
* **The Architect's Terminal:** The frontend IDE where the player writes behavioral scripts, views error logs, and analyzes grid traffic.
