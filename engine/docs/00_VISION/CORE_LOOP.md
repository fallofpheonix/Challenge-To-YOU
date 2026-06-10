# CORE_LOOP.md

## Purpose

This document defines the moment-to-moment cognitive pipeline of the player and the hierarchical time loops that structure the game's progression. It dictates how the player interacts with the simulation and what keeps them engaged over short, medium, and long durations.

## The Cognitive Pipeline

The player executes this sequence continuously. This is the fundamental mechanical interaction model.

1. **Observe:** The player analyzes orbital telemetry. This involves reading density heatmaps, identifying traffic bottlenecks, locating resource nodes, and spotting anomalous drone behavior.
2. **Program:** The player drafts or modifies localized behavior scripts (`pscript`). This involves defining state checks, pheromone triggers, and conditional logic to address the observed state.
3. **Deploy:** The player commits the code. This action is gated by the Uplink Window and delayed by orbital latency, removing the ability to execute panic corrections.
4. **Analyze:** The player watches the swarm execute the new logic. The player must determine if the emergent behavior matches the intended outcome or if the localized rules created unintended global consequences.
5. **Adapt:** The player identifies logic flaws, deadlocks, or Byzantine faults within the swarm. The player adjusts the script to handle these edge cases and queues the next deployment.
6. **Expand:** Once a system is stable and self-sustaining, the player shifts focus to new grid sectors, pushing the swarm toward deeper resource veins or unknown anomalies.

## Hierarchical Time Loops

### The 5-Minute Loop: Debugging & Micro-Optimization

* **Focus:** Immediate problem solving and logic validation.
* **Action:** A supply chain halts because drones are trapped in a routing loop, or a node is compromised by a localized hazard. The player writes a 3-to-5 line logic patch, waits for the uplink broadcast, and observes the immediate physical result on the grid.
* **Reward:** The visual satisfaction of a physical traffic jam clearing and resource flow resuming.

### The 30-Minute Loop: Sector Expansion & Protocol Design

* **Focus:** Establishing new systems and adapting to local environment variables.
* **Action:** The player targets a new subterranean sector requiring a dedicated supply line. This requires writing specialized protocols. For example, programming a specialized sub-swarm to establish a relay chain for pheromone propagation over long distances, or designing a quarantine protocol for a highly volatile localized hazard.
* **Reward:** A fully automated, stable logistics sector that no longer requires active observation, increasing total global throughput.

### The 5-Hour Loop: Systemic Paradigm Shifts

* **Focus:** Advancing through narrative acts and restructuring global architecture.
* **Action:** The baseline environment fundamental rules change. The swarm encounters the alien intelligence (logic drift, phantom telemetry). Previously perfect logistics networks begin to fail globally. The player must research or unlock advanced architectural concepts, such as quorum consensus logic or cryptographic pheromone signing, completely rewriting their foundational codebase to survive the new threat model.
* **Reward:** Mastering a significantly more complex computer science concept and observing the swarm successfully out-compute a hostile distributed network.

---

### Fallback Strategy

If playtesting reveals that the 5-minute loop becomes tedious due to excessive uplink delays, implement a localized "Simulation Sandbox." This allows the player to test code on a small, isolated cluster of virtual drones instantly without orbital delay. If the simulation succeeds, the player then deploys it to the live planetary grid using the standard uplink mechanics.
