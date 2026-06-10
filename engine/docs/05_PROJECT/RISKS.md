# RISKS.md

## Purpose
This document tracks the primary systemic, technical, and psychological risks of Project Chrysalis. Every milestone execution must evaluate these vulnerabilities.

---

## 1. Systemic Performance Collapse
* **Risk:** Simulating 10,000+ individual drones executing conditional code trees at a 10Hz tick rate completely saturates the CPU cache, causing frame-drops and telemetry lag.
* **Mitigation:**
    * Keep the drone data packed sequentially in memory (ECS) to optimize CPU cache lines.
    * Offload global, stateless grid updates (like pheromone evaporation and diffusion) to parallel worker pools using Go channels.
    * Keep drone logic strictly sequential within the state-transition phase to guarantee 100% execution determinism.

## 2. Telemetry Signal Occlusion (Visual Chaos)
* **Risk:** Showing 10,000 active drone icons on screen results in incomprehensible visual noise, violating Design Pillar 4 (Observable Macro-Systems).
* **Mitigation:**
    * The frontend must never render individual drone objects from orbit.
    * Group drone coordinates into macro-level vector density arrays. 
    * Godot will render them as fluid velocity fields and heatmaps. Drones are discrete in the backend but a flowing collective from orbit.

## 3. The Coding Layer "Homework Trap"
* **Risk:** Forcing players to write complex syntax loops early on makes the game feel like a dry academic programming text rather than an elite systems engineering simulator.
* **Mitigation:**
    * Adopt the "Broken Code Puzzle" pattern.
    * Instead of a blank IDE, give the player a functional but flawed script (e.g., a routing algorithm with a subtle lockup) and task them with diagnosing and patching it.

## 4. Non-Deterministic Synchronization Drifts
* **Risk:** Tiny floating-point discrepancies or time-of-check to time-of-use (TOCTOU) state mutations break replay verification, ruining Act III's recursive debugging loop.
* **Mitigation:**
    * Enforce absolute bit-perfection. Native floating-point numbers (`float32`/`float64`) are banned across all simulation files.
    * All vectors, angles, distance measurements, and pheromone intensities must be encoded using the static integer `crysmath.FixedPoint` scaling system ($10^6$).

---

*Version: 1.0.0-M0*
*Status: Active Monitoring*
