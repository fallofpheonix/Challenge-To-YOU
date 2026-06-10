# WORLD_GENERATION.md

## Purpose

This document defines the procedural generation rules for the subterranean grid of Kepler-452b. Because the game relies on a deterministic simulation, the world is not hand-crafted. It is mathematically generated from a master seed, ensuring that specific spatial challenges, chokepoints, and resource distributions dictate the swarm's required logistics.

## 1. Biomes (Depth Strata)

The map expands vertically downward. "Biomes" in *Project Chrysalis* are defined by depth, representing distinct geological and operational challenges.

* **The Crust (0m to -500m):** The staging ground. Characterized by highly stable rock, wide caverns, and dense Silicate deposits. Environmental hazards are rare. Designed to allow the Architect to establish basic logistics and replication loops safely.
* **The Mantle Transition (-500m to -2000m):** The friction zone. Characterized by dense, unmineable bedrock that creates severe natural chokepoints and winding tunnels. Introduces basic Isotopes. Highly volatile, with frequent Magnetic Anomalies and Thermal Geysers forcing dynamic rerouting.
* **The Core-Boundary (-2000m+):** The hostile frontier. Characterized by massive, open abysses and fractal cave networks. Contains Deep-Core Isotopes required for Act III/IV research. This is the primary territory of the dormant Alien Network.

## 2. Caves (Topology)

The physical layout of the grid dictates the swarm's traffic flow. The generation algorithm focuses on creating logistical friction rather than aesthetic landscapes.

* **Cellular Automata Carving:** The algorithm simulates geological erosion, resulting in organic, unpredictable cave systems rather than perfect geometric corridors.
* **Logistical Chokepoints:** The generation intentionally creates narrow, one-tile-wide tunnels connecting massive caverns. These are the primary structural challenges for the Architect, forcing the creation of right-of-way protocols to prevent gridlock.
* **Dead Ends:** Frequent false paths that require the swarm to drop `Return_Vector` pheromones to successfully navigate back out without starving their batteries.

## 3. Hazards (Environmental Triggers)

Hazards are not spawned randomly during gameplay; their epicenters are deterministically placed during world generation and lie dormant until triggered by time or proximity.

* **Placement Logic:** Hazards are aggressively seeded near high-value resource nodes and critical chokepoints to prevent trivial extraction.
* **Magnetic Anomalies:** Seeded heavily in the Mantle Transition. When active, they scramble local pheromone intensities, forcing the Architect to build physical relay structures to bypass the interference.
* **Crust Collapses:** Seeded in wide, unstable caverns. When triggered, they permanently alter the grid topology, turning open space into unmineable rock, severing established highways instantly.

## 4. Resource Placement

Resources are the gravitational centers of the swarm's expansion. Their placement dictates where highways will inevitably form.

* **Silicate Veins:** Generated as massive, sprawling clusters (blob generation) primarily in the Crust. They yield tens of thousands of units, meant to sustain long-term structural replication.
* **Isotope Nodes:** Generated as tiny, isolated points of extreme value (point generation) in deeper strata. Often completely surrounded by dormant Hazards or unmineable bedrock, requiring precision routing to extract.
* **Parasitic Alien Nodes (Act III):** Placed exclusively in the deep Mantle and Core-Boundary. The algorithm intentionally seeds these nodes directly adjacent to the richest Isotope clusters, guaranteeing that the swarm's most critical supply lines will be the first to suffer Byzantine corruption.
