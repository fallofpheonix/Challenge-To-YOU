# SANDBOX_MODEL.md

## Purpose

This document defines the strict containment protocols for the game's execution environment. Because players submit actual Python code that runs on the game's servers (or within a highly concurrent local simulation), the Sandbox must guarantee three things: absolute security, strict determinism, and zero impact on the engine's 10 Ticks Per Second (TPS) master loop.

## 1. Security (The Air Gap)

The execution environment is completely severed from the host machine. Players cannot access the operating system, network, or file system.

* **AST Whitelisting:** The engine does not use Python's `eval()` or `exec()`. Instead, the player's code is parsed into an Abstract Syntax Tree (AST). The engine only compiles the script if it consists exclusively of approved nodes (e.g., `ast.If`, `ast.Compare`).
* **Import Blocking:** The `import` keyword is structurally banned at the parser level. Players cannot load `sys`, `os`, `math`, `time`, or any other standard library module.
* **Built-in Stripping:** Native Python functions that interact with memory or execution state (like `id()`, `globals()`, `locals()`, `open()`, `compile()`) are completely removed from the environment namespace.

## 2. Infinite Loop Prevention

A single `while True: pass` submitted by a player would instantly freeze the entire 10 TPS simulation loop. The Sandbox prevents this proactively.

* **Instruction Counting:** During the AST compilation phase, the engine secretly injects a counter into every loop and function call.
* **The Guillotine:** As the drone executes its logic during Phase 4, this counter increments. If the counter exceeds the Execution Limit (see below) before the script finishes, the engine immediately throws a silent `SandboxTimeout` exception.
* **Graceful Failure:** The main simulation catches this exception, terminates that specific drone's execution for the current tick, flags a `TIMEOUT` error in the Architect's telemetry log, and seamlessly moves on to the next drone. The master tick is never delayed.

## 3. Execution Limits

To ensure the simulation can process up to 50,000 drones in under 100 milliseconds, computational boundaries are strictly enforced per drone, per tick.

* **Operations Per Tick:** A drone's script is limited to a hard cap of Operations (e.g., 250 AST node evaluations per tick). An Operation is defined as a variable assignment, a mathematical calculation, or a condition check.
* **API Throttling:** A drone is allowed exactly one Primary Actuation call (`move()` or `interact()`) per tick. Subsequent calls are parsed but physically ignored by the Rule Engine.

## 4. Memory Limits

Drones are cheap, fragile actuators, not supercomputers. Their ability to store persistent data is heavily restricted to force players to use the grid (pheromones) as their primary memory bank.

* **The Dictionary Cap:** The persistent `drone.memory` dictionary is restricted to a maximum of 16 key-value pairs.
* **Data Type Enforcement:** Memory values can only be primitive data types (`int`, `bool`, `str`). Nested structures, objects, or multi-dimensional arrays will trigger a runtime error.
* **String Truncation:** To prevent memory bloat, any string stored in memory or used as a custom signal signature is hard-capped at 16 characters.
* **Payload Limit:** The total size of the raw text script deployed during an Uplink Window cannot exceed 4 Kilobytes.
