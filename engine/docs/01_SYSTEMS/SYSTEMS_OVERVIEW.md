# SYSTEMS_OVERVIEW.md

## Purpose

This document provides a high-level definition of the six primary systems that drive *Project Chrysalis*. It defines what these systems *are* and how they *behave* in the game space, stripped of underlying code or engine implementation.

---

### 1. The World (Kepler-452b)

The world is a persistent, grid-based, subterranean environment that acts as a hostile, dynamic ecosystem rather than a static map.

* **Vertical Layering:** The world is divided into depth strata. Deeper levels hold rarer resources but exhibit more aggressive environmental volatility.
* **Environmental Hazards:** Non-biological threats (thermal geysers, shifting magnetic anomalies, crust collapses) dynamically alter the terrain, forcing the swarm to physically reroute.
* **Orbital Weather:** Electromagnetic storms act as a meta-system, periodically dictating the frequency and duration of the Architect's "Uplink Windows" (the ability to deploy new code).

### 2. The Swarm

The swarm consists of thousands of disposable, autonomous micro-drones that act as the physical actuators of the Architect's logic.

* **Localized Awareness:** A drone has zero global awareness. It can only read the state of its immediate adjacent grid cells.
* **Strict Determinism:** Drones possess no inherent AI. They blindly execute the exact behavioral protocol downloaded during the last Uplink Window.
* **Physical Constraints:** Drones require solar or isotope power to function. If a drone's battery depletes or it sustains physical damage from a hazard, it becomes an inert obstacle on the grid.

### 3. Resources

Resources are strictly physical entities on the grid. There is no "global inventory" or magic storage pool.

* **Physical Logistics:** If a resource is mined, a drone must physically carry it to a Hub structure for it to be utilized for replication or research.
* **Silicates & Isotopes:** Basic silicates drive physical expansion and drone replication. Deep-core isotopes drive advanced computational unlocks and localized power grids.
* **System Bottlenecks:** Because resources are physical, supply chains are vulnerable to traffic jams, deadlocks, and physical severing.

### 4. Signals (Pheromones & Telemetry)

This is the communication backbone of the game. It bridges the gap between the drone's limited physical awareness and the need for global coordination.

* **Digital Pheromones:** Drones drop localized data markers (scents) on the grid. These markers evaporate over time and stack in intensity, allowing the swarm to dynamically create highways to resources or quarantine zones around hazards.
* **Telemetry Data:** This is the imperfect, delayed feedback loop sent to the Architect in orbit. It provides density heatmaps and error logs, but can be occluded by deep rock or corrupted by external interference.

### 5. Research

Research does not provide traditional video game stat boosts (e.g., "+10% mining speed"). It expands the player's computational vocabulary.

* **Algorithmic Unlocks:** Researching anomalous data unlocks new logic modules. Examples include the ability to write `WHILE` loops, implement cryptographic signal signing, or deploy quorum-based consensus checks.
* **Paradigm Shifts:** Research allows the Architect to transition the swarm from fragile, linear scripts to robust, fault-tolerant distributed networks capable of surviving Act III.

### 6. The Alien Network

The antagonist of the game. It is a dormant, subterranean distributed intelligence that operates using the exact same underlying logic framework as the player's swarm.

* **Algorithmic Warfare:** It does not use weapons. It attacks the integrity of the player's systems via Logic Drift (subtly altering drone scripts), Pheromone Spoofing (creating fake resource signals to cause mass starvation), and Byzantine Corruption (causing drones to report false telemetry).
* **Territorial Expansion:** It physically expands its own infected nodes across the grid, competing for the same bandwidth and physical space as the Architect's swarm.
