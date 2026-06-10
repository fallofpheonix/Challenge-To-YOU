# RULE_ENGINE.md

## Purpose

This document defines how the simulation core processes the Architect's Python-based behavioral scripts during every tick. It outlines the strict rules of engagement: how conditions are evaluated, how actions are queued, and how conflicts are resolved when logic clashes with physical reality.

## 1. Evaluation Order

The Rule Engine operates during Phase 4 of the Simulation Tick. It executes synchronously and strictly top-to-bottom for every active drone on the grid.

* **Simultaneous Processing:** While the engine processes drone logic sequentially in the backend (e.g., Drone 1 to Drone 50,000), the *gameplay effect* is simultaneous. No drone has a "first-mover advantage" in a single tick.
* **The State Snapshot:** When evaluation begins, all drones reference the exact same locked snapshot of the grid generated in Phase 3. If Drone A decides to move North, Drone B cannot detect Drone A's new position until the *next* tick.
* **Top-to-Bottom Execution:** The engine reads the Architect's Python script linearly. Control flow (`if`, `elif`, `else`) dictates the path. Once an actionable command is queued, the engine finishes reading the script to ensure no subsequent variables need updating, but will not queue a second conflicting action.

## 2. Conditions (The Triggers)

Conditions are the boolean gates that determine a drone's behavior. They rely entirely on the drone's immediate internal and external sensors.

* **Sensor Checks:** Evaluating the intensity of pheromones in the 3x3 adjacent grid (e.g., `drone.sense("HAZARD") > 50`).
* **Internal State:** Evaluating payload capacity or power levels (e.g., `drone.payload == "SILICATE"` or `drone.battery < 10`).
* **Memory State:** Evaluating custom variables set by the Architect in the drone's limited dictionary (e.g., `drone.memory["mode"] == "SCOUT"`).
* **Compound Logic:** Conditions can be chained using standard Python logical operators (`and`, `or`, `not`) to create highly specific triggers.

## 3. Actions (The Actuators)

Actions are the physical commands queued by the Rule Engine to be executed in Phase 5. A drone can only perform **one** primary physical action per tick.

* **Movement:** Stepping to an adjacent tile (`drone.move("NORTH")` or `drone.follow("RESOURCE_GRADIENT")`).
* **Interaction:** Mining a resource node or depositing a payload at a Hub (`drone.interact()`).
* **Signaling (Secondary Action):** Dropping a digital pheromone (`drone.emit("HAZARD", 255)`). *Note: Emitting a signal does not consume a drone's primary physical action for that tick; it can move and emit simultaneously.*

## 4. Priority (Conflict Resolution)

Because the Architect's code can be flawed, and the physical grid is chaotic, strict priority rules resolve conflicts.

* **Code-Level Priority (The First Match Wins):** Within the Python script, the first `if` or `elif` block that evaluates to `True` and contains an Action function locks in the drone's intent for that tick. If subsequent `if` blocks also contain Action functions, they are silently ignored.
* **Hardware Overrides Software:** The Tier 1 hardware checks (Battery = 0, or Fatal Damage) will always intercept and cancel a script's queued action before it executes.
* **Collision Resolution (The Physics Arbiter):** If Drone A and Drone B both evaluate rules that command them to move into the exact same empty tile, the simulation resolves this in Phase 5. The engine randomly (via deterministic PRNG) grants the tile to one drone, and the other drone's action fails, effectively losing its turn for that tick.
* **Deadlock Nullification:** If a drone attempts an impossible action (e.g., moving into solid rock, or interacting with empty space), the action simply fails. The drone does not crash, but it wastes its tick.
