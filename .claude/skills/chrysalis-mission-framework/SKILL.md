---
name: chrysalis-mission-framework
description: The mission/objective system — MissionState, status evaluation each tick, and JSON scenario levels. Use when adding objectives, win/lose conditions, or level definitions.
---

# Chrysalis Mission Framework

Tracks scenario objectives and win/lose state, evaluated deterministically each tick. Code: `engine/core/simulation/mission.go`; scenarios: `engine/core/levels/` (JSON).

## Model
- `MissionStatus` — status enum (e.g. in-progress / success / failure).
- `MissionState` — current objective state; `NewDefaultMissionState()` seeds the baseline.
- `(*Engine) EvaluateMission()` — called each tick; updates status from world state. Emits a `mission_changed` event on transition (see [[chrysalis-eventbus]]).
- `(*Engine) InfectedRatio() float64` — fraction of swarm compromised; a common fail-condition input (e.g. corruption overrun).

## Rules
- Evaluation must be a pure function of deterministic world state — no wall-clock, no unseeded randomness (see [[chrysalis-simulation]]). Same run → same mission outcome.
- Emit `mission_changed` only on actual status transitions so replays and telemetry stay clean.
- New scenarios go in `levels/*.json`; keep them data-only (initial world, objectives, thresholds), not logic.

Client surfaces mission state via telemetry; see [[chrysalis-networking]] and the relevant `ui/screens/`.
