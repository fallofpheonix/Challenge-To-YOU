# TESTING_STRATEGY.md

## Purpose

This document defines the strict testing protocols required to maintain the structural integrity of *Project Chrysalis*. Because the game relies entirely on a deterministic simulation engine running unverified player code, testing cannot be an afterthought. A single floating-point error or unhandled Sandbox escape will corrupt the global telemetry and destroy the replay system.

## 1. Unit Tests (The Foundation)

Unit tests isolate and validate the smallest individual components of the Go backend and Godot frontend, entirely independent of the master simulation tick.

* **AST Parser Validation (Go):** Tests feed malicious, malformed, and edge-case Python strings to the custom lexer. The test passes only if the parser correctly structures valid code and aggressively throws errors for `import` statements, unauthorized built-ins, or recursion.
* **Math & Grid Assertions (Go):** Validates all custom fixed-point arithmetic and spatial hash lookups. Tests assert that distance calculations and gradient diffusions yield the exact expected integer values.
* **UI Component Tests (Godot):** Validates that the visual layer correctly instantiates the required heatmap shaders and UI elements without memory leaks when fed mock JSON data.

## 2. Simulation Tests (Determinism & Logic)

These tests ensure that the interconnected ECS systems execute their rules flawlessly and predictably.

* **The Hash Test:** The core determinism check. The test initializes a grid with a fixed World Seed, injects a predefined Architect script, and advances the simulation by exactly 10,000 ticks. It then computes a cryptographic hash of the final grid array. If the hash differs from the baseline, the test fails immediately, flagging a determinism leak.
* **The Collision Arbiter Test:** Simulates 100 drones attempting to occupy the same single tile simultaneously. The test asserts that exactly one drone succeeds, the other 99 fail gracefully, and the game does not deadlock or crash.
* **The Sandbox Guillotine Test:** Injects a script containing `while True: pass`. Asserts that the engine correctly increments the operation counter, terminates that specific drone's logic execution via a `SandboxTimeout`, and allows the master tick to finish cleanly.

## 3. Performance Tests (The 100ms Budget)

Because the game runs at 10 Ticks Per Second (TPS), the engine must completely evaluate Phase 1 through Phase 5 in under 100 milliseconds, or the simulation will stutter.

* **Maximum Density Stress Test:** Initializes a map with 50,000 Drones, 10,000 active Pheromones, and 50 expanding Hazards. Asserts that a single tick processes in under 50ms on a standard CPU, leaving a 50ms buffer for WebSocket serialization and OS overhead.
* **Memory Allocation Profiling:** Monitors the Go garbage collector during prolonged high-traffic simulations. Asserts that the ECS reuses array memory rather than allocating/destroying objects per tick, preventing GC spikes that cause stuttering.

## 4. Regression Tests (Historical Preservation)

Updates to the game (e.g., adding a new API command or tweaking hazard damage) run the risk of breaking older Save Files, because old Replays rely on older logic to re-simulate properly.

* **Replay Compatibility Assertions:** A suite of massive, late-game save files from previous game versions are stored in the CI/CD pipeline. Before a new build is approved, the engine runs a headless fast-forward on all of them.
* **The Forking Rule:** If a game update alters fundamental movement or combat physics, the backend must implement "Versioned Logic" (e.g., executing `RuleEngine_v1` for old replay files and `RuleEngine_v2` for new saves). Regression tests ensure the engine correctly identifies and applies the correct physics version based on the save file's metadata.
