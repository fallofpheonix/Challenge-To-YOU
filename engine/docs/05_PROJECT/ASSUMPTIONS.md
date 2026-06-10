# ASSUMPTIONS.md

## Purpose
This document tracks the core hypotheses that underlie the game's design. If an assumption is proven false during prototyping, it triggers a strategic pivot.

| Assumption | Status | Validation Method |
| :--- | :--- | :--- |
| **Players enjoy debugging.** | Unvalidated | Observe if players find the Rewind/Step features satisfying or a chore during playtests. |
| **Pheromones are enough.** | **Validated** | Milestone 0 proved that simple local rules + pheromones create stable supply trails. |
| **Delay creates tension.** | Unvalidated | Test if the 15-second "Uplink Window" creates suspense or just frustration. |
| **Heatmaps are readable.** | Unvalidated | Test if players can diagnose a bottleneck by looking at a density fluid. |
| **Coding isn't a barrier.** | Unvalidated | Test with non-programmers using the "Act I: Teacher" simplified protocol. |
| **Go/Godot is performant.** | Unvalidated | Benchmark 10,000 entities on the Go core with a WebSocket broadcast. |
