# ENTITY_MODEL.md

## Purpose

This document defines the core simulated entities of *Project Chrysalis*. Because the simulation operates on a strict Entity-Component-System (ECS) architecture, these definitions represent the pure data structures and states that the engine processes during every tick, rather than object-oriented classes.

---

### 1. Drone

The physical actuator of the swarm. Highly fragile, strictly obedient, and possessing zero global awareness.

* **State:** `Idle` | `Executing_Instruction` | `Moving` | `Harvesting` | `Inert` (Out of power/Broken) | `Compromised` (Logic hijacked)
* **Properties:**
    * Unique ID
    * Grid Coordinates `(X, Y)`
    * Current Battery Level
    * Payload (Carrying `None` or `Resource ID`)
    * Active Protocol (The specific Architect script it is currently looping)

* **Lifecycle:** Fabricated at a Hub structure -> Dispatched onto the grid -> Loops assigned protocol -> Depletes battery or triggers a hazard -> Becomes an `Inert` physical obstacle -> Salvaged by other drones or permanently crushed by terrain.

### 2. Resource

Physical materials existing on the grid. They are strictly conserved; they cannot be teleported or globally pooled.

* **State:** `Embedded` (In rock) | `Carried` (By a drone) | `Stored` (At a Hub)
* **Properties:**
    * Type (`Silicate` for physical builds, `Isotope` for power/logic unlocks)
    * Grid Coordinates `(X, Y)` (If embedded or dropped)
    * Yield (Remaining extractable units)

* **Lifecycle:** Generated during world initialization or exposed by shifting crust hazards -> Extracted by a Drone -> Physically transported -> Deposited at a Hub -> Permanently consumed to fabricate new Drones, Structures, or Research.

### 3. Signal (Pheromone)

The localized, ephemeral data markers that facilitate decentralized swarm communication.

* **State:** `Active` | `Decaying`
* **Properties:**
    * Signature Type (`Resource_Found`, `Hazard_Warning`, `Rally_Point`, `Spoofed_Data`)
    * Grid Coordinates `(X, Y)`
    * Intensity `[Integer: 0 to 100]`

* **Lifecycle:** Emitted by a Drone onto a specific grid cell -> Intensity increases if other drones emit the same signature on that cell -> Automatically decays by a fixed integer rate every simulation tick -> Reaches `0` intensity -> Permanently wiped from the grid memory.

### 4. Structure

Immobile, multi-tile physical infrastructure built by the swarm to expand logistics.

* **State:** `Blueprint` | `Constructing` | `Operational` | `Offline` (Damaged/Unpowered)
* **Properties:**
    * Type (`Hub`, `Relay_Node`, `Storage_Cache`)
    * Grid Footprint (Occupied coordinates)
    * Structural Integrity
    * Contained Inventory

* **Lifecycle:** Drone script dictates a blueprint placement -> Drones deposit required Resources into the blueprint -> Transitions to `Operational` -> Facilitates replication or extends telemetry range -> Destroyed by hazards or Alien logic -> Reverts to `Inert` salvage.

### 5. Hazard

Non-biological environmental threats that dynamically alter the play space.

* **State:** `Dormant` | `Expanding` | `Active` | `Dissipating`
* **Properties:**
    * Threat Type (`Magnetic_Anomaly`, `Thermal_Geyser`, `Crust_Collapse`)
    * Epicenter Coordinates
    * Radius of Effect
    * Damage Value (Applied per tick to overlapping entities)

* **Lifecycle:** Triggered deterministically by the world seed -> Expands outward across the grid -> Applies immediate physical state changes to overlapping Drones/Structures -> Dissipates -> Leaves behind altered terrain or exposed Resources.

### 6. Alien Node

The physical hardware of the competing distributed intelligence. It obeys the exact same physical rules as player Structures.

* **State:** `Dormant` | `Awake` | `Broadcasting` | `Quarantined` (Neutralized)
* **Properties:**
    * Grid Coordinates
    * Corruption Radius
    * Spoofing Signature (The false signal it emits)
    * Injector Protocol (The malicious code it forces onto adjacent Drones)

* **Lifecycle:** Exists hidden beneath deep crust -> Awakens upon swarm proximity (Act III) -> Emits spoofed signals and hijacks adjacent Drone logic -> Competes for grid space and resources -> Is surrounded, physically isolated, or out-computed by the Architect's counter-protocols -> Rendered `Quarantined`.
