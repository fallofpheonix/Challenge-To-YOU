# SWARM_BEHAVIOR.md

## Purpose

This document defines the cognitive processing model of the individual drones and the macro-level phenomena that result from it. It outlines how a drone processes the Architect's instructions within the constraints of its physical sensors, and how those localized decisions aggregate into complex swarm intelligence.

## Decision Hierarchy

Drones process inputs in a strict, top-down hierarchy during their execution phase. A higher-tier imperative will always override a lower-tier one.

1. **Tier 1: Physics & Hardware (Involuntary):** The drone checks its battery state and physical integrity. If battery is `0` or physical damage is fatal, the drone halts all logic and transitions to `Inert`.
2. **Tier 2: The Architect's Protocol (Voluntary):** The drone evaluates the active code deployed during the last Uplink Window. It processes the conditional logic against its immediate surroundings.
3. **Tier 3: The Default State (Fallback):** If no conditions in the Architect's protocol are met, the drone defaults to `Idle` to conserve battery, waiting for a state change in its environment or a new Uplink patch.

## Rule Evaluation

Drones do not possess predictive logic or memory. They operate as simple finite state machines that evaluate rules instantaneously.

* **The Sensor Lock:** During Phase 3 of the simulation tick, the drone locks the data of its immediate 3x3 grid (Moore neighborhood) into its temporary memory.
* **Sequential Parsing:** The Architect's script is evaluated top-to-bottom against this locked data. The drone executes the *first* true condition it encounters and ignores the rest of the script for that tick.
* **Blind Actuation:** Once an action is chosen (e.g., `MOVE NORTH`), the drone attempts it blindly. If another drone moves into that space simultaneously, the simulation resolves the collision, potentially resulting in a failed action. The drone will not attempt to calculate a detour until the *next* tick.

## State Transitions

A drone shifts between functional states based on successful rule evaluations and physical triggers.

* **`Idle` → `Moving`:** Triggered when the Architect's script successfully matches a directional condition (e.g., `IF Resource_Gradient > 0 THEN MOVE`).
* **`Moving` → `Harvesting` / `Depositing`:** Triggered when a drone is physically adjacent to a Resource Node or Hub and executes an interaction command. This state locks the drone in place for a specific number of ticks based on extraction/deposit rates.
* **`*` → `Compromised` (Act III):** Triggered if an Alien Node successfully injects malicious logic. The drone drops the Architect's protocol and begins evaluating the Alien Network's protocol hierarchy.
* **`*` → `Inert`:** Triggered instantly by fatal hazard overlap or complete battery depletion. An `Inert` drone becomes a permanent physical obstacle unless salvaged.

## Emergent Behaviors

These are complex, undocumented macro-behaviors that arise naturally from the interaction of thousands of drones executing simple rules. The game is designed to encourage and reward these phenomena.

* **The Highway Effect:** Because drones drop `Return_Vector` pheromones when carrying resources, and because identical pheromones stack in intensity, initial chaotic scouting naturally condenses into highly efficient, high-density traffic lanes.
* **The Quarantine Wall:** If the Architect writes a rule to drop a `Hazard_Marker` before dying, drones walking into an anomaly will die and leave a scent. Subsequent drones read the scent, turn around, and drop their own markers. This organically generates a physical "wall" of signals that perfectly traces the outline of dynamic hazards.
* **Bifurcation (Swarm Splitting):** If a drone detects two equal `Resource_Gradients` pulling in opposite directions, the collision resolution engine causes the swarm to naturally split its workforce perfectly in half without any centralized load-balancing code.
* **Deadlocking:** A negative emergent behavior. If two dense highways intersect perpendicularly, and the Architect has not written right-of-way logic, the physical collision rules will cause a permanent, grid-wide traffic jam until intervention occurs.
