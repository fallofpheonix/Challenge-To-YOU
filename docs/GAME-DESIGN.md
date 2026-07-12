# Game Design Document: Challenge To YOU

## Game Overview

**Challenge To YOU** is a roguelike coding puzzle game where players solve procedurally generated challenges across multiple fantasy/sci-fi eras. The core mechanic is **Emergent Multi-Layer Systems** — combining broken/unrelated code to create glitches, loopholes, and side-effects that produce passcodes.

### Genre
- Roguelike
- Puzzle
- Hacking Simulation
- Programming Game

### Target Audience
- Competitive programmers (Codeforces level)
- Casual gamers who code
- Fans of hacking/cyberpunk aesthetics
- Roguelike enthusiasts

### Unique Selling Points
1. **Multi-Era Progression**: From Medieval Magitech to Cyberpunk Neon
2. **Frankenstein Code**: Combine broken scripts to create exploits
3. **Dynamic Passcodes**: Different approaches produce different passcodes
4. **Luck Mechanic**: Roguelike replayability with volatility engine
5. **AI Integration**: Local AI analyzes your coding style

---

## Core Loop

```
┌─────────────────────────────────────────────────────────────┐
│                    Main Game Loop                            │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ 1. Receive   │→ │ 2. Analyze   │→ │ 3. Write Code    │  │
│  │   Challenge  │   │   Code Blocks│   │   (Editor)       │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
│                                                │            │
│                                                ▼            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ 6. Unlock    │← │ 5. Get       │← │ 4. Execute &     │  │
│  │   Next Era   │   │   Passcode   │   │   Analyze        │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Step-by-Step

1. **Receive Challenge**: Game generates random challenge from seed + luck
2. **Analyze Code Blocks**: Player reads the "junk code" modules
3. **Write Code**: Player writes solution in the editor
4. **Execute & Analyze**: Code runs in secure sandbox, AI analyzes approach
5. **Get Passcode**: System generates passcode from code interactions
6. **Unlock Next Era**: Passcode advances progression

---

## Era System

### Era Progression (v1 = 2 Eras)

| Tier | Era | Theme | Code Type | Unlock Requirement |
|------|-----|-------|-----------|-------------------|
| 1 | Medieval Magitech | Dark fantasy | Custom DSL (Runes) | Start |
| 2 | Cyberpunk Neon | Dystopian future | Real scripting (Python/JS) | Complete 10 Magitech challenges |

### Future Eras (Post-MVP)

| Tier | Era | Theme | Code Type |
|------|-----|-------|-----------|
| 3 | Dieselpunk Industrial | WW2-era mechanical | Punch cards, vacuum tubes |
| 4 | Cosmic Quantum Space | Alien technology | Quantum entanglement, reality manipulation |

---

### Era 1: Medieval Magitech

#### Aesthetic
- **Visuals**: Dark backgrounds, mystical fonts, rune symbols
- **Colors**: Deep purples, golds, ancient greens
- **UI Elements**: Stone textures, parchment backgrounds, magical glows
- **Audio**: Low drones, mystical chimes, ancient whispers

#### Code Type: Custom DSL (Runes)
```
// Magitech DSL Example
RUNE fire_rune = IGNITE(power: 100)
RUNE water_rune = FLOW(direction: NORTH)
RUNE earth_rune = STABILIZE(radius: 5)

// Combine runes to create effect
EFFECT {
    fire_rune.COMBINE(water_rune)
    earth_rune.ACTIVATE()
    // Passcode emerges from rune interaction
}
```

#### Challenges
- **Architect Mode**: Write new rune combinations
- **Ghost Mode**: Modify runes without alerting royal mages
- **Saboteur Mode**: Break rune sequences to cause magical failures

#### Narrative
You are a **Rune Hacker** in a medieval kingdom where magic is technology. The royal mages guard their spell-books jealously. You must rewrite the literal laws of magic to bypass security, steal secrets, and eventually overthrow the magical establishment.

---

### Era 2: Cyberpunk Neon

#### Aesthetic
- **Visuals**: Neon lights, terminal interfaces, holographic displays
- **Colors**: Electric blue, hot pink, neon green, deep black
- **UI Elements**: Glitch effects, scan lines, matrix rain
- **Audio**: Synthwave, electronic pulses, digital static

#### Code Type: Real Scripting (Python/JS)
```python
# Cyberpunk Python Example
import neon_lib as neon

# Create a decoy process
decoy = neon.Process(name="prime_calculator")
decoy.start(lambda: calculate_primes(1000000))

# While AI watches decoy, exploit race condition
exploit = neon.Exploit(target="password_server")
result = exploit.run(timing="microsecond_delay")

# Passcode emerges from exploit success
print(f"Passcode: {result.passcode}")
```

#### Challenges
- **Architect Mode**: Build new hacking tools
- **Ghost Mode**: Modify code under CPU monitoring threshold
- **Saboteur Mode**: Inject logic bombs that trigger later

#### Narrative
You are a **Data Ghost** in a neon-drenched mega-city. Mega-corporations control everything through code. You must hack their systems, steal their secrets, and expose their corruption — all while staying below the detection threshold of their AI monitoring systems.

---

## Gameplay Modes

### Mode 1: The Architect

#### Role
**Builder** — You create new systems from scratch.

#### Objective
Write clean, functional modules that automate tasks.

#### Skill Tested
- Algorithmic thinking
- System design
- Long-term planning

#### Gameplay
```
1. Receive blueprint for new system
2. Design architecture (multiple modules)
3. Write code for each module
4. Test integration between modules
5. Passcode emerges from successful integration
```

#### UI Elements
- **Blueprint Viewer**: Shows system requirements
- **Module Editor**: Write individual modules
- **Integration Tester**: Verify modules work together
- **Architecture Score**: Rates your design choices

#### Example Challenge
"Build an automated defense system for the castle. Create three modules: (1) Detection, (2) Alert, (3) Response. All three must work together to generate the passcode."

---

### Mode 2: The Ghost

#### Role
**Stealth Hacker** — You modify existing code without detection.

#### Objective
Make changes while keeping CPU usage below detection threshold.

#### Skill Tested
- Code optimization
- Stealth mindset
- Resource management

#### Gameplay
```
1. Receive existing codebase
2. Identify what needs to change
3. Write modifications
4. Keep detection meter below threshold
5. Passcode emerges from stealthy success
```

#### UI Elements
- **Detection Meter**: Shows AI awareness (0-100%)
- **CPU Monitor**: Tracks resource usage
- **Stealth Score**: Rates your invisibility
- **Timer**: Some missions are time-sensitive

#### Example Challenge
"Inject a backdoor into the corporate database. Keep CPU usage under 5% to avoid triggering the AI. If detection hits 100%, mission fails."

---

### Mode 3: The Saboteur

#### Role
**Chaos Agent** — You break code to cause chain reactions.

#### Objective
Break minimal code to cause maximum damage.

#### Skill Tested
- Judgment
- Critical thinking
- Understanding dependencies

#### Gameplay
```
1. Receive working system
2. Identify critical dependencies
3. Break minimal code
4. Watch chain reaction unfold
5. Passcode emerges from successful destruction
```

#### UI Elements
- **Dependency Graph**: Shows code relationships
- **Impact Preview**: Predicts chain reaction
- **Chaos Score**: Rates your destruction efficiency
- **Collateral Damage**: Tracks unintended side effects

#### Example Challenge
"Shut down the factory by breaking as little code as possible. Find the single point of failure that causes the most cascading effects."

---

## Procedural Generation System

### Seed-Based RNG

```python
# Example seed-based generation
seed = 12345  # From player luck + level
rng = random.Random(seed)

# Select modules based on luck
if luck > 0.7:  # High luck
    modules = select_easy_modules(rng)
elif luck > 0.3:  # Medium luck
    modules = select_medium_modules(rng)
else:  # Low luck
    modules = select_hard_modules(rng)

# Stitch modules together
code = stitch_modules(modules)

# Ensure at least one glitch exists
ensure_glitch_exists(code, modules)
```

### Luck Mechanic

| Luck Value | Level Difficulty | AI Monitoring | Glitch Availability |
|------------|------------------|---------------|---------------------|
| 0.0-0.3 | Hard | Aggressive | Rare |
| 0.3-0.7 | Medium | Normal | Moderate |
| 0.7-1.0 | Easy | Relaxed | Abundant |

### Module Types

| Type | Purpose | Example |
|------|---------|---------|
| INPUT | Process external data | `decode(input)` |
| CORE | Main logic | `transform(data)` |
| OUTPUT | Generate results | `encode(result)` |
| DECOY | Distract AI | `calculate_primes()` |
| EXPLOIT | Create glitches | `race_condition()` |

### Glitch Types

| Glitch | Description | Example |
|--------|-------------|---------|
| Race Condition | Timing-based exploit | Two scripts access same resource |
| Memory Leak | Allocation pattern | Loop allocates but never frees |
| Buffer Overflow | Data exceeds bounds | Write past array end |
| Logic Bomb | Delayed trigger | Code activates at specific time |

---

## Passcode System

### How Passcodes Are Generated

1. **Code Execution**: Player's code runs in secure sandbox
2. **Glitch Detection**: System identifies intentional exploits
3. **Style Analysis**: AI/AST analyzes coding approach
4. **Factor Combination**: Multiple factors combined into seed
5. **Hash Generation**: Seed hashed into 16-character passcode

### Passcode Sources

| Source | Description | Example |
|--------|-------------|---------|
| Error Logs | Hidden in stack traces | `TypeError at line 42: [PASSCODE]` |
| Memory Leaks | Allocation patterns | Specific allocation sequence |
| CPU Fluctuations | Timing-based | Execution time mod 1000 |
| Glitch Interactions | Frankenstein code | Two modules combine |

### Passcode Uniqueness

- Same code → Same passcode (deterministic)
- Different approach → Different passcode
- Same goal, different method → Different passcode
- No "correct" answer — any valid passcode works

### Passcode Usage

```python
# Example: Passcode unlocks content
passcode = "a1b2c3d4e5f6g7h8"

if passcode == expected_passcode:
    unlock_next_level()
    unlock_bonus_content()
    add_to_collection()
```

---

## Difficulty System

### Difficulty Tiers

| Tier | Description | Unlock Requirement |
|------|-------------|-------------------|
| Novice | Simple modules, obvious glitches | Start |
| Apprentice | Moderate complexity, hidden glitches | Complete 5 Novice challenges |
| Expert | High complexity, multiple glitches | Complete 10 Apprentice challenges |
| Master | Extreme complexity, emergent glitches | Complete 15 Expert challenges |

### Difficulty Modifiers

| Modifier | Effect | Example |
|----------|--------|---------|
| Code Obfuscation | Variable names hidden | `a1b2c3` instead of `fire_rune` |
| AI Monitoring | Detection threshold lower | 3% instead of 5% |
| Time Pressure | Timer added | 60 seconds to complete |
| Resource Limits | Memory/CPU reduced | 32MB instead of 64MB |

---

## Progression System

### Experience Points (XP)

| Action | XP Earned |
|--------|-----------|
| Complete challenge | 100 XP |
| Complete with high style score | +50 XP |
| Complete without detection (Ghost) | +25 XP |
| Minimal code break (Saboteur) | +25 XP |
| Discover new glitch type | +100 XP |

### Unlockables

| Unlock | Requirement | Cost |
|--------|-------------|------|
| Era 2 (Cyberpunk) | Complete 10 Magitech challenges | Free |
| New module types | Discover 5 glitches | 500 XP |
| Custom themes | Complete 50 challenges | 1000 XP |
| Challenge editor | Complete 100 challenges | 2000 XP |

### Leaderboards

| Category | Metric |
|----------|--------|
| Fastest Completion | Time to solve |
| Style Score | Code quality rating |
| Glitch Discovery | Number of unique glitches found |
| Era Mastery | Challenges completed per era |

---

## Special Events

### Boss Fights

#### AI Counter-Hack
- **Description**: An AI actively monitors and blocks your attempts
- **Mechanic**: Must write code that evolves to bypass AI detection
- **Reward**: Rare glitch type + bonus XP

#### System Meltdown
- **Description**: System is crashing; must fix before time runs out
- **Mechanic**: Time pressure + multiple simultaneous failures
- **Reward**: Emergency glitch type + bonus XP

### Meta-Progression

#### Hint Archive
- **Description**: Collect "Corrupted Data" from completed challenges
- **Mechanic**: Use hints to bypass security in special events
- **Reward**: Access to secret challenges

#### Glitch Collection
- **Description**: Catalog all discovered glitch types
- **Mechanic**: Different glitches unlock different content
- **Reward**: Permanent bonuses

---

## Monetization (Post-MVP)

### Free Alpha (Itch.io)
- 2 eras (Magitech, Cyberpunk)
- 3 modes (Architect, Ghost, Saboteur)
- 50+ procedurally generated challenges
- Local leaderboards

### Steam Early Access ($4.99-$9.99)
- All alpha content
- 4 eras (add Dieselpunk, Cosmic)
- 6 modes (add Forensic Analyst, Janitor)
- Online leaderboards
- Challenge editor

### Full Release ($14.99-$19.99)
- All Early Access content
- 100+ handcrafted challenges
- Multiplayer (cooperative hacking)
- Mod support
- Regular content updates

---

## Accessibility

### Visual
- Colorblind modes
- Adjustable text size
- High contrast themes
- Screen reader support

### Motor
- Customizable controls
- Adjustable timing
- Auto-save progress
- Pause functionality

### Cognitive
- Tutorial mode
- Hint system
- Difficulty adjustment
- Progress tracking

---

## Technical Requirements

### Minimum Specs
- **OS**: Windows 10 / macOS 10.15 / Ubuntu 18.04
- **CPU**: Intel i5 / AMD Ryzen 5
- **RAM**: 8 GB
- **GPU**: Integrated graphics sufficient
- **Storage**: 2 GB

### Recommended Specs
- **OS**: Windows 11 / macOS 12 / Ubuntu 20.04
- **CPU**: Intel i7 / AMD Ryzen 7
- **RAM**: 16 GB
- **GPU**: Dedicated GPU (for visual effects)
- **Storage**: 4 GB (SSD recommended)

---

*Last updated: 2026-07-10*
