---
name: chrysalis-replay-system
description: The deterministic replay recorder — checkpoints, per-frame event logs, world-hash verification, and seek. Use when editing engine/core/replay or building playback/scrubbing features in the client.
---

# Chrysalis Replay System

Records every tick's events plus periodic state checkpoints so a run can be reconstructed exactly. Works because the sim is deterministic (see [[chrysalis-simulation]]): checkpoint + replayed events = identical state. Files: `*.chrysalis_replay`; code: `engine/core/replay/recorder.go`.

## Data model
- `ReplayData` = `ReplayMeta` + `[]Checkpoint` + `[]FrameEvents`.
- `FrameEvents` — events emitted on one tick (from [[chrysalis-eventbus]]).
- `Checkpoint` — full serialized world state at a tick, with optional **`worldHash`** for verification.

## Recorder API
- `NewRecorder(initialState, checkpointEvery)` — checkpoint cadence in ticks.
- `Record(tick, events)` — append a frame.
- `Checkpoint(tick, state, worldHash...)` — snapshot; pass the world hash to enable drift detection.
- `NearestCheckpoint(targetTick)` / `EventsInRange(start, end)` — seek: jump to nearest checkpoint, replay forward.
- `Serialize()` / `Deserialize(raw, checkpointEvery)` — persist / load.
- Counters: `CheckpointCount()`, `TotalFrames()`, `TotalEvents()`.

## Seek algorithm
To reach tick T: `NearestCheckpoint(T)` → restore state → apply `EventsInRange(cp.Tick, T)`. If a stored `worldHash` diverges from the recomputed hash, the sim is non-deterministic — treat as a critical bug, not a replay bug.

Client playback UI: `ui/screens/replay_controls.gd`. Tests: `recorder_test.go`.
