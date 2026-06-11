---
Status: Active
Implementation: 40%
Confidence: Authoritative
---
# Game Architecture — Index

This directory establishes the architecture-first specifications of the Chrysalis Engine Layer. The documentation is continuously updated to reflect the live state of the Go core and Godot client.

## Document Directory

### Vision & System Core
- [CURRENT_STATE.md](./CURRENT_STATE.md): Captured reality of the active Go core & Godot client (UPDATED).
- [VISION.md](./VISION.md): Conceptual model of the gamified sandbox environment.
- [GAMEPLAY_LOOP.md](./GAMEPLAY_LOOP.md): Interactive cycle between agents, user, and sandbox.
- [PROGRESSION_SYSTEM.md](./PROGRESSION_SYSTEM.md): Dynamic curriculum design and level unlocking.
- [REWARD_SYSTEM.md](./REWARD_SYSTEM.md): Optimization reinforcement scoring functions.
- [LEADERBOARD_SYSTEM.md](./LEADERBOARD_SYSTEM.md): Verification and logging of high scores.
- [AI_HINT_SYSTEM.md](./AI_HINT_SYSTEM.md): Contextual hint generation using the LLM interface.
- [MULTIPLAYER_SYSTEM.md](./MULTIPLAYER_SYSTEM.md): Multiplayer state replication and sync.
- [ROADMAP.md](./ROADMAP.md): Development milestones (Milestones 0-3 COMPLETED).

### Engine Integration
- [Godot Architecture](./engine/GODOT_ARCHITECTURE.md): WebSocket-driven telemetry routing.
- [Entity System](./engine/ENTITY_SYSTEM.md): Data-oriented component mapping.
- [Deterministic Simulation](./engine/DETERMINISTIC_SIMULATION.md): Tick-based fixed-point execution.
- [Replay System](./engine/REPLAY_SYSTEM.md): Bit-perfect state reconstruction.
- [Physics Model](./engine/PHYSICS_MODEL.md): Dynamic spatial triggers & hazards.

### P-Script Subsystem
- [Language Spec](./pscript/LANGUAGE_SPEC.md): Pratt-parsed infix expressions and types.
- [Grammar](./pscript/GRAMMAR.md): Lexer and operator precedence specifications.
