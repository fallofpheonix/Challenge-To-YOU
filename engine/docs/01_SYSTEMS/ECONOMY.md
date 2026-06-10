# ECONOMY.md

## Purpose

This document defines the macro-economic engine of *Project Chrysalis*. The game does not feature abstract currencies, global storage pools, or arbitrary timers. The economy is strictly physical and spatial. Wealth is defined by throughput—the swarm's ability to efficiently move matter and data across a volatile grid.

## 1. Resources (The Raw Materials)

There is no global inventory. If the Architect's terminal says the swarm possesses 1,000 Silicates, it means there are exactly 1,000 physical Silicate units sitting on Hub tiles across the map.

* **Silicates (Mass):** The fundamental building block. Found in massive, shallow crust deposits. Used exclusively for fabricating physical bodies: basic Drones, Relay Nodes, and Hubs.
* **Isotopes (Energy/Compute):** Rare, localized deposits found deeper in the crust. Highly volatile. Used to power advanced computational algorithms (Research) and to fabricate specialized hardware (Alien countermeasures).
* **Bandwidth (Meta-Resource):** The Architect's most precious commodity. Measured in Kilobytes per Uplink Window. Expanding physical Relay Nodes on the planet surface directly increases the orbital bandwidth, allowing the Architect to deploy larger, more complex Python scripts.

## 2. Costs (Friction & Consumption)

Costs in this economy are paid in physical materials, time, and spatial real estate.

* **Fabrication Cost:** Creating a new Drone or Structure consumes a fixed amount of deposited Silicates/Isotopes from a Hub.
* **Maintenance Cost (Battery Drain):** Drones do not exist for free. They consume battery power every tick they are active. They recharge passively via solar (slow, weather-dependent) or actively by consuming Isotopes at a Hub (fast, expensive).
* **The Traffic Tax:** The true cost of expansion is grid congestion. A resource node that is 100 tiles away "costs" significantly more to mine than one 10 tiles away, purely due to the travel time and the pheromone upkeep required to maintain the highway.

## 3. Production (The Logistics Chain)

Production is bottlenecked purely by the Architect's ability to optimize physical space.

* **Extraction:** Drones must spend ticks locked in the `Harvesting` state to break rock into carryable units.
* **Transportation:** The highest point of friction. Drones must physically walk the resource back to a Hub. If a highway is only 1 tile wide, production is hard-capped by physical collision limits, regardless of how many drones are assigned to the task.
* **Processing:** Resources dropped at a Hub are instantly converted into usable "Deposited" states, ready to be consumed by blueprints or the orbital research queue.

## 4. Replication (Swarm Growth)

The game relies on exponential growth managed by localized rules.

* **The Fabrication Loop:** The Architect does not click "Build 5 Drones." Instead, the Architect programs a Hub with an automated threshold (e.g., `IF Hub.Silicates > 10 THEN Fabricate_Drone`).
* **Self-Sustaining Expansion:** A successful logistic sector is one where drones mine silicates to build more drones, which then mine silicates faster.
* **The Diminishing Return:** Unchecked replication leads to gridlock. If a Hub fabricates 5,000 drones in a confined cave system, the resulting traffic jam will cause global production to plummet to zero. The Architect must design load-balancing protocols to migrate excess population to new frontiers.
