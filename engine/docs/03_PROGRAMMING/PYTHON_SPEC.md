# PYTHON_SPEC.md

## Purpose

This document defines the restricted Python 3 environment (The Sandbox) used by the Architect to program the swarm. To ensure the deterministic simulation engine remains performant and cannot be crashed by infinite loops, the drones execute a highly constrained subset of standard Python.

## 1. Syntax

The environment uses standard Python 3 syntax. Indentation, colons, and standard operator logic apply exactly as they do in standard Python.

* **Allowed:** Basic data structures, standard math operators (`+`, `-`, `*`, `/`, `%`), and logical operators (`and`, `or`, `not`).
* **Forbidden:** `import` statements of any kind. The standard library (including `math`, `time`, `random`) is completely inaccessible to prevent breaking determinism or accessing host systems.

## 2. Variables & State

Drones are generally stateless FSMs that re-evaluate their surroundings every tick, but they possess a small, persistent memory dictionary to track multi-tick goals.

* **Data Types:** Restricted to `int`, `bool`, `str` (strictly for predefined signal IDs like `"SILICATE"` or `"HAZARD"`), and basic one-dimensional `list`s.
* **Persistent Memory:** Variables declared globally reset every tick. To persist data across simulation ticks, the Architect must use the predefined `drone.memory` dictionary.
```python
# Persists across ticks
if "steps_taken" not in drone.memory:
    drone.memory["steps_taken"] = 0
```

## 3. Conditions

Control flow relies on standard Python `if`, `elif`, and `else` statements. This is the primary driver of swarm logic.

* **Evaluation:** Conditions are evaluated synchronously during Phase 4 of the tick loop.
```python
if drone.sense("HAZARD_MARKER") > 50:
    drone.move("SOUTH")
    drone.emit("HAZARD_MARKER", 100)
elif drone.payload == "SILICATE":
    drone.follow("RETURN_VECTOR")
else:
    drone.follow("RESOURCE_GRADIENT")
```

## 4. Loops

Loops are heavily restricted. Because the game simulation runs 10 times per second, a script that hangs in a loop will freeze the entire planetary engine.

* **The Tick is the Loop:** The Architect's script is naturally called once per tick. `while True:` is rarely necessary and actively dangerous.
* **Allowed:** `for` loops used for iterating through limited arrays or adjacent grid cell data.
```python
# Allowed: Checking adjacent cells
for cell in drone.get_adjacent_cells():
    if cell.type == "OBSTACLE":
        drone.memory["blocked"] = True
```
* **Restricted:** `while` loops are monitored by the AST parser. If a loop exceeds the operational limit (see *Limits*), the engine terminates the drone's logic for that tick and flags a `TIMEOUT` error in the Architect's telemetry.

## 5. Functions

The Architect can define helper functions to organize complex protocols. The environment also provides a strict, immutable API of built-in functions.

* **Custom Functions:** Defined via `def`. Recursion is strictly forbidden and will be blocked by the parser.
* **Core Built-ins (Actuation & Sensors):**
    * `drone.move(direction)`: Attempts movement (e.g., `"NORTH"`, `"SOUTH"`).
    * `drone.sense(signal_type)`: Returns the integer intensity of a pheromone on the current tile.
    * `drone.emit(signal_type, intensity)`: Drops a pheromone marker on the current tile.
    * `drone.interact()`: Mines a resource or deposits a payload.
    * `drone.get_adjacent_cells()`: Returns a list of the 8 surrounding grid states.

## 6. Limits (The Sandbox Constraints)

To maintain the 10 TPS simulation speed for 50,000+ drones, execution constraints are enforced at the engine level.

* **Instruction Limit (Ops/Tick):** A single drone script cannot exceed 100 operations (comparisons, assignments, or API calls) per tick. Exceeding this causes the drone to "stutter" and lose its turn.
* **Memory Limit:** The `drone.memory` dictionary cannot exceed 16 key-value pairs per drone.
* **Deployment Payload:** A single broadcasted Python script cannot exceed 4 Kilobytes of text. This forces the Architect to write elegant, optimized logic rather than massive, bloated rule trees.
* **API Action Cap:** A drone can only execute one physical Actuation function (e.g., `move()` OR `interact()`) per tick. Multiple calls will result in only the first being executed.
