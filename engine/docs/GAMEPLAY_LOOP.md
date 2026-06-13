---
Status: Active
Implementation: 40%
Confidence: Authoritative
---
# Game Layer — Gameplay Loop

The core interaction loop of the simulation engine revolves around swarm economics and survival.

```mermaid
sequenceDiagram
    participant PScript as Swarm AI (P-Script)
    participant Engine as Engine (Go Core)
    participant Grid as Double-Buffered Grid
    participant Client as Godot Client

    Engine->>Engine: Begin Tick (10Hz)
    Engine->>Grid: Decay Pheromones & Process Hazards
    loop Every Drone
        Engine->>PScript: Execute `agent.ps` logic
        PScript->>Grid: Sense environment (Gradients, Resources)
        PScript->>Engine: Emit Action (Move, Harvest, Drop)
    end
    Engine->>Engine: Resolve Economics (Fabrication)
    Engine->>Grid: Swap Buffers (Commit State)
    Engine->>Client: Broadcast `EMISSION_SNAPSHOT` via WebSocket
    Client->>Client: Render UI & Telemetry
```

## Implemented Mechanics

- **Exploration**: Drones move randomly or follow gradients.
- **Economics**: Drones harvest silicates and return them to the base following the home pheromone gradient.
- **Fabrication**: When colony silicates reach the threshold (5), a new drone is fabricated.
- **Hazards & Contagion**: Magnetic fields drain battery. Alien nodes spread logic corruption to drones, turning them into hostile actors.
