# GAMEPLAY_PILLARS.md

## Purpose

This document defines the core, unchangeable rules of the game's design. Every new feature, mechanic, or technical implementation must be tested against these pillars. If a proposed idea violates one of these principles, it must be discarded or aggressively redesigned.

## The Pillars

### 1. The Architect of Digital Destinies (No Direct Control)

The player is a master planner, never a pilot. There are no joysticks, no "attack-move" commands, and no manual overrides for individual units. You weave the fate of the swarm entirely through code, protocols, and decentralized rules. If a drone is about to be destroyed, you cannot click to save it; you can only write better survival logic for the next generation.

### 2. Local Decisions, Global Emergence

There is no centralized "brain" or omniscient pathfinding (like A* algorithms) guiding the swarm. A drone only knows what is in its immediate grid vicinity and what its current internal state dictates. Complex behaviors—like setting up a resource supply chain, surrounding an enemy node, or balancing loads across a network—must emerge purely from the interactions of thousands of localized decisions and pheromone gradients.

### 3. Imperfect Information & The Uplink Bottleneck

The player is isolated in orbit above a hostile, interference-heavy atmosphere. Telemetry is delayed, and communication bandwidth is finite. You cannot stream live patches continuously. You must queue your instructions and wait for Uplink Windows. This forces the player to design fault-tolerant, self-healing systems rather than relying on rapid micromanagement.

### 4. Observable Macro-Systems

Simulating thousands of autonomous agents can quickly turn into visual noise. To make this playable, every complex distributed system must be translated into readable visual spectacles. The player diagnoses network partitions, logic bottlenecks, and resource deficits through macro-level tools: density heatmaps, physical traffic jams, and visual logic flow overlays. The health of the entire ecosystem must be readable at a glance.

### 5. Diagnosable, Physical Failure

A logic bug does not crash the application; it physically manifests in the game world. When the player writes a flawed routing loop, the drones will literally march in endless circles until their power depletes. When a condition fails to account for a hazard, drones will physically pile up in a trench. Every failure is a measurable, physical event on the grid that the player can observe, pause, and step through to trace the root cause.
