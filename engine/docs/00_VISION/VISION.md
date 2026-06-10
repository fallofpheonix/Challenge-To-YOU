# Project Chrysalis: The Architect's Swarm

---

## One-Sentence Pitch

You are an orbital architect programming millions of autonomous micro-drones to explore, colonize, and survive on the hostile exoplanet Kepler-452b—where you write the rules, watch emergence unfold, and battle an alien intelligence that tries to hijack your swarm.

---

## Target Audience

### Primary (Core)
- **Programming enthusiasts** who love *Farmer Was Replaced*, *Bitburner*, *Shenzhen I/O*
- **Distributed systems engineers** fascinated by emergent behavior, swarm intelligence, pheromone coordination
- **Cybersecurity professionals** interested in adversarial AI, logic bombs, recursive defense
- **OS developers** who enjoy deterministic simulation, replay debugging, bit-perfect determinism

### Secondary (Extended)
- **DSA learners** wanting gamified debugging puzzles (broken code, log analysis)
- **Automation game fans** (*Factorio*, *Arison*) preferring programming over direct control
- **Sci-fi narrative lovers** drawn to emergent storytelling and "ghost in the machine" themes

### Who This Is NOT For
- Players wanting joystick-based action or direct character control
- Gamers seeking cinematic graphics or pre-scripted narrative
- Casual players preferring simple mechanics over deep programming

---

## Core Fantasy

### **You Are the Architect, Not the Pilot**

You don't control individual drones. You write their **behavioral protocols**—their digital "pheromones"—and watch simple rules cascade into complex, emergent supply chains.

| What You Do | What You Experience |
|-------------|---------------------|
| Write pscript code | Swarms autonomously harvest, replicate, defend |
| Deploy pods from orbit | Telemetry shows drones scatter, die, establish foothold |
| Refine logic after failures | Beautiful emergent supply chains form |
| Patch code mid-crisis | Swarm sacrifices few to save many, routes around danger |
| Write recursive counter-measures | You out-program the alien network in algorithmic battle |

### The Emotional Arc

**Act I**: *"I wrote 3 lines of code and my swarm built a colony."* → **Emergence wonder**

**Act II**: *"My drones are adapting to hazards I didn't program."* → **Pride in system design**

**Act III**: *"The alien AI is trying to overwrite my swarm—I need to write a better algorithm."* → **High-stakes code battle**

---

## Design Philosophy

### 1. **Programming Is the Gameplay**
- No joystick, no direct control—code is your only interface
- Every mechanic must be programmable via pscript
- Debugging is not a tax; it's the core loop (deterministic replay = replayability)

### 2. **Emergence Over Scripting**
- Simple rules → complex outcomes (ant colony principle)
- No pre-scripted events; everything emerges from player-written code
- Replayability comes from player creativity, not content volume

### 3. **Determinism Is Non-Negotiable**
- Bit-perfect replay for debugging (FixedPoint precision: 10⁶)
- Every failure must be reproducible and analyzable
- Replay system enables Act III's "algorithmic battle" analysis

### 4. **Narrative Through Telemetry**
- Story unfolds via orbital logs, drone death reports, anomaly detection
- No cutscenes; players infer narrative from system behavior
- Late-game "ghost in the swarm" revealed through code anomalies, not dialogue

### 5. **Educational Without Being Pedagogical**
- Broken-code puzzles teach real debugging skills (loops, conditionals, resource priority)
- DSA concepts emerge naturally (pathfinding = routing, recursion = counter-measures)
- No "homework"—learning happens through failure and iteration

### 6. **Systems Over Graphics**
- 2D top-down first (orbital telemetry view), 3D terrain optional
- Visual clarity > aesthetic fidelity (drones = dots, pheromones = glowing particles)
- Godot used for visualization, not cinematic rendering

---

## Success Criteria

### Quantitative (Measurable)

| Metric | Target | Why |
|--------|--------|-----|
| **Time to first emergence** | < 10 minutes | Players see swarm build colony within Act I opening |
| **Puzzle completion rate** | > 70% | Broken-code puzzles solvable without frustration |
| **Replay usage** | > 80% of failures | Deterministic rewind is core debugging loop |
| **Act I completion** | > 50% of players | Core loop engaging enough to continue |
| **Code complexity growth** | Players add 3-5 new functions by Act II | Game encourages iterative improvement |

### Qualitative (Player Experience)

| Experience | How We Know It Worked |
|------------|-----------------------|
| **"I am the Architect"** | Players describe themselves as "programming swarms" not "playing a game" |
| **"Emergence is beautiful"** | Players express awe at supply chains forming from simple rules |
| **"Debugging is satisfying"** | Players use replay voluntarily, not just when stuck |
| **"Act III feels epic"** | Players describe algorithmic battle as "high-stakes code war" |
| **"I learned real skills"** | Players mention debugging, DSA, distributed systems concepts |

### Technical (System Integrity)

| Criterion | Requirement |
|-----------|-------------|
| **Determinism** | 100% bit-perfect replay (no floating point, use FixedPoint) |
| **Performance** | 10,000+ drones simulatable at 60 FPS (Go backend, minimal Godot overhead) |
| **Extensibility** | pscript grammar can add recursion, self-modifying code without refactor |
| **Testability** | All VM ops have Go tests (deterministic proximity probe pattern) |

---

## Non-Goals

### ❌ What This Game Is NOT

| Non-Goal | Why |
|----------|-----|
| **Direct control game** | Core fantasy is "Architect, not Pilot"—joystick breaks immersion |
| **Cinematic narrative** | Story emerges from telemetry; no cutscenes or pre-scripted dialogue |
| **Graphics-focused** | Systems > visuals; 2D telemetry first, 3D optional |
| **Casual programming game** | Target is serious programmers; puzzles require real debugging skills |
| **Factorio clone** | No direct machine building—everything is programmable swarm behavior |
| **Cybersecurity simulator** | Act III is algorithmic battle, not "movie hacking" (real distributed systems) |
| **Educational tutorial** | Learning happens through failure; no forced lessons or quizzes |
| **Multiplayer competitive** | Single-player emergent gameplay; no PvP or co-op swarms |
| **Procedural content generator** | Puzzles are hand-designed (broken code, log analysis); not infinite levels |
| **Mobile game** | Desktop-first (Go + Godot); keyboard-heavy programming interface |

---

## Why This Game Exists

### The Problem It Solves

**Current programming games are either:**
1. **Too simple** (Bitburner: no emergent behavior, just resource loops)
2. **Too abstract** (Shenzhen I/O: no narrative context, pure puzzle)
3. **Too direct** (Factorio: joystick building, not programming)

**Project Chrysalis fills the gap:**
- **Emergent programming**: Simple rules → complex supply chains (ant colony)
- **Narrative context**: Orbital architect fighting alien AI (3-act story)
- **No direct control**: Code is your only interface (Architect, not Pilot)

### The Vision

> "Programming games should feel like **writing laws for a universe**, not **typing commands**."

When players finish Act I, they should say:
> *"I wrote 10 lines of code and my swarm built a colony. I didn't place a single drone."*

When players finish Act III, they should say:
> *"I out-programmed an alien intelligence. My recursive defense algorithm was better than its hijack attempt."*

That's the fantasy. That's why this game exists.

---

*Version: 0.1.0-alpha*  
*Last Updated: June 2026*  
*Author: Ujjwal Singh*
