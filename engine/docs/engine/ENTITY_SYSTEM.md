---
Status: Active
Implementation: 100%
Confidence: Authoritative
---
# Game Engine — Entity-Component-System (ECS)

Organizes gameplay objects using a strict Data-Oriented Design approach.

## ECS Structure
- **Registry (`SwarmRegistry`)**: The central data store for all drones.
- **Structure of Arrays (SoA)**: Components are stored as contiguous slices (e.g., `PositionX[]`, `PositionY[]`, `Battery[]`, `State[]`) rather than arrays of structs.
- **Cache-Friendly**: This SoA layout maximizes CPU cache utilization when iterating over thousands of drones.
- **Dynamic Expansion**: Slices automatically double in capacity when the swarm size exceeds current bounds.
- **No Entity Removal**: Inert drones remain in the registry but are marked with `StateInert`. This avoids slice copying/compaction on drone death.
