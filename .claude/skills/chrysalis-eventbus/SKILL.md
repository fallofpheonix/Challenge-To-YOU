---
name: chrysalis-eventbus
description: The simulation EventBus — the sealed Event type, its typed payloads, and the canonical event tags emitted each tick. Use when adding a new event kind, consuming events (replay, telemetry), or wiring emitters.
---

# Chrysalis EventBus

Immutable record of what happened each tick. `engine/core/simulation/events.go` + `events_emitters.go`. Consumed by the replay recorder and telemetry.

## Core types
- `EventBus` — `NewEventBus()`; collects events per tick.
- `Event` — `{ Type EventType; Tick; Data EventPayload }`, JSON-serializable; immutable.
- `EventPayload` — **sealed interface** via unexported marker `isEventPayload()`. Only in-package payload structs satisfy it, so `Event.Data` cannot be set to a foreign type — the compiler enforces it.

## Canonical event tags (`EventType`)
`drone_spawned`, `drone_died`, `drone_infected`, `harvested`, `deposited`, `trust_changed`, `mission_changed`, `fabricated`, `hazard_damage`.

## Payload structs
`SpawnedData`, `DiedData`, `InfectedData`, `HarvestData`, `DepositData`, `TrustData`, `HazardData`, `MissionData` — each implements `isEventPayload()`.

## Adding a new event (do all four)
1. Add an `EventType` const.
2. Define a `XxxData` payload struct.
3. Add `func (XxxData) isEventPayload() {}` to the sealed set.
4. Extend `Event.UnmarshalJSON` dispatch so it round-trips, and emit it from `events_emitters.go`.

Emitting the same events in the same order every tick is part of determinism — see [[chrysalis-simulation]]. Events feed [[chrysalis-replay-system]] and [[chrysalis-networking]]. Tests: `events_test.go`.
