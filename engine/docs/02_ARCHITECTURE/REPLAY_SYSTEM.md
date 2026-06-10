# REPLAY_SYSTEM.md

## Purpose

This document defines the architecture of the game's Save and Replay mechanisms. Because *Project Chrysalis* simulates massive entity counts, saving the positional and memory data of 50,000 individual drones 10 times a second is impossible. Instead, the game leverages its strict determinism to save *history* rather than *state*, allowing for time-travel debugging and incredibly lightweight save files.

## 1. Save Format (The Master Ledger)

A "Save File" does not contain a snapshot of the grid. It contains the exact ingredients required to recreate the grid from scratch.

* **The World Seed:** The cryptographic string used by the procedural generation algorithm to build the exact layout of caves, resource nodes, and dormant hazards for that specific campaign.
* **The Input Log:** An ordered, append-only ledger of every action the Architect ever took, permanently stamped with the exact simulation tick it occurred.
* **File Size:** Because it only stores text strings (code) and integers (ticks), a 50-hour campaign save file will rarely exceed a few megabytes.

## 2. Event Recording (The Inputs)

The simulation only cares about external forces that alter the logic state. The `Input_Log` records exactly three types of events:

* **Uplink Deployments:** The most common event. Recorded as `[Tick_Number, "CODE_DEPLOY", "String_Payload"]`. This logs the exact moment the Architect broadcasted a new Python script to the swarm.
* **Targeted Pings:** Recorded as `[Tick_Number, "DIAGNOSTIC_PING", "X,Y_Coordinates"]`. Logs when the Architect forced the telemetry UI to isolate a specific grid sector for debugging.
* **Meta-Events:** Any rare, game-level triggers, such as the Architect explicitly advancing to a new narrative Act or initiating a hard server reboot.

## 3. Replay Format (Playback & Debugging)

When the player loads a save file, or uses the "Diagnostic Ping" to rewind time to debug a traffic jam, the engine performs a "Headless Fast-Forward."

* **Background Reconstruction:** The Go backend spins up a hidden, secondary simulation instance. It initializes the grid using the World Seed.
* **Uncapped Execution:** Because this hidden instance does not need to send telemetry to the Godot visual client, it is not restricted to 10 Ticks Per Second. It runs at maximum CPU speed.
* **Sequential Injection:** The engine rapidly simulates time. As the tick counter hits the timestamps listed in the `Input_Log`, it injects the Architect's saved Python scripts into the Rule Engine exactly as they were deployed in the past.
* **The Catch-Up:** Once the hidden simulation reaches the target tick (either the player's current "Live" tick for a loaded save, or a specific past tick for debugging), it pauses, binds to the Godot client's WebSocket, and begins broadcasting the visual state. To the player, loading a massive late-game save takes only a few seconds of loading screen compute time.
