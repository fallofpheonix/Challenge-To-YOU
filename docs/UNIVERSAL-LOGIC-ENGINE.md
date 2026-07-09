# Universal Logic Evaluation Engine — Technical Specification

## Overview

The **Universal Logic Evaluation Engine** (ULEE) is the core backend system that treats all game elements — magic runes, code scripts, physical gears — as the same data structure: **Events, Conditions, and Effects** in a Directed Acyclic Graph (DAG).

---

## Core Concept: The Cause-and-Effect Graph

```
[INPUT EVENT] ───► [CONDITION/STATE CHECK] ───► [OUTPUT EFFECT]
```

To the engine, everything is:
- **Event**: Something that happened (spell cast, code executed, gear jammed)
- **Condition**: A state that must be true for the effect to trigger
- **Effect**: The result when event + condition are satisfied

---

## Data Structures

### 1. GameNode (Universal Element)

```go
// GameNode represents any element in the game world
type GameNode struct {
    ID          string            `json:"id"`
    Type        NodeType          `json:"type"`        // EVENT, CONDITION, EFFECT
    Theme       string            `json:"theme"`       // magitech, cyberpunk, industrial
    Name        string            `json:"name"`        // "Fire Rune", "Audio Driver"
    Description string            `json:"description"`
    Properties  map[string]interface{} `json:"properties"` // Theme-specific data
    State       NodeState         `json:"state"`       // Current state
    Connections []string          `json:"connections"` // IDs of connected nodes
}

type NodeType int
const (
    NodeTypeEvent NodeType = iota
    NodeTypeCondition
    NodeTypeEffect
)

type NodeState int
const (
    StateInactive NodeState = iota
    StateActive
    StateTriggered
    StateBlocked
)
```

### 2. GameGraph (Cause-and-Effect Network)

```go
// GameGraph is the complete logic network for a level
type GameGraph struct {
    ID           string                 `json:"id"`
    Theme        string                 `json:"theme"`
    Nodes        map[string]*GameNode   `json:"nodes"`
    Edges        []*GameEdge            `json:"edges"`
    WinCondition *WinCondition          `json:"win_condition"`
    State        map[string]interface{} `json:"state"` // Global state variables
}

type GameEdge struct {
    From       string `json:"from"`       // Source node ID
    To         string `json:"to"`         // Target node ID
    Weight     float64 `json:"weight"`    // Connection strength (0.0-1.0)
    Condition  string `json:"condition"`  // Optional condition expression
}

type WinCondition struct {
    Type       string                 `json:"type"`       // "state", "event", "chain"
    Target     string                 `json:"target"`     // Node ID or state key
    Operator   string                 `json:"operator"`   // "==", ">", "<", "contains"
    Value      interface{}            `json:"value"`      // Expected value
    Required   bool                   `json:"required"`   // Must be satisfied to win
}
```

### 3. GlitchDetector (Emergent Solution Finder)

```go
// GlitchDetector finds unscripted player solutions
type GlitchDetector struct {
    RuleMatrix map[string][]Rule `json:"rule_matrix"` // Theme-specific rules
}

type Rule struct {
    Event     string `json:"event"`     // Trigger event type
    Condition string `json:"condition"` // Required condition
    Effect    string `json:"effect"`    // Resulting effect
    Score     float64 `json:"score"`    // How "creative" this solution is
}
```

---

## Theme Abstraction

### How Different Themes Map to the Same Structure

#### Theme 1: Medieval Magitech

```json
{
  "id": "fire_rune_001",
  "type": "EVENT",
  "theme": "magitech",
  "name": "Fire Rune",
  "properties": {
    "element": "fire",
    "power": 100,
    "school": "destruction",
    "affinity": "heat"
  },
  "connections": ["water_rune_001", "earth_totem_001"]
}
```

**Player sees**: A glowing rune on a spell-book  
**Engine sees**: Node with `element=fire, power=100`

#### Theme 2: Cyberpunk Code

```json
{
  "id": "audio_driver_001",
  "type": "EVENT",
  "theme": "cyberpunk",
  "name": "Audio Driver",
  "properties": {
    "language": "python",
    "code": "def play_audio(): buffer_overflow()",
    "dependencies": ["sound_card", "memory_allocator"],
    "port": 80
  },
  "connections": ["firewall_001", "security_system_001"]
}
```

**Player sees**: A terminal with Python code  
**Engine sees**: Node with `language=python, port=80`

#### Theme 3: Industrial Mechanical

```json
{
  "id": "gear_001",
  "type": "EVENT",
  "theme": "industrial",
  "name": "Conveyor Belt Gear",
  "properties": {
    "material": "steel",
    "voltage_rating": 220,
    "rpm": 100,
    "connected_to": "power_grid"
  },
  "connections": ["circuit_breaker_001", "door_lock_001"]
}
```

**Player sees**: A metal gear on a conveyor belt  
**Engine sees**: Node with `voltage_rating=220, rpm=100`

---

## Emergent Loophole Detection

### Rule Matrix Example

```go
var magitechRules = []Rule{
    {
        Event:     "ice_spell_cast",
        Condition: "near_heat_source",
        Effect:    "thermal_explosion",
        Score:     0.8, // Creative solution
    },
    {
        Event:     "fire_rune_activated",
        Condition: "water_rune_present",
        Effect:    "steam_explosion",
        Score:     0.6,
    },
}

var cyberpunkRules = []Rule{
    {
        Event:     "audio_buffer_overflow",
        Condition: "firewall_port_80_open",
        Effect:    "password_hash_leak",
        Score:     0.9, // Very creative
    },
    {
        Event:     "race_condition_triggered",
        Condition: "concurrent_access_enabled",
        Effect:    "data_corruption",
        Score:     0.7,
    },
}

var industrialRules = []Rule{
    {
        Event:     "gear_jammed",
        Condition: "voltage_exceeds_220v",
        Effect:    "circuit_breaker_blown",
        Score:     0.5,
    },
    {
        Event:     "power_surge",
        Condition: "ground_fault_present",
        Effect:    "electronic_door_unlocked",
        Score:     0.8,
    },
}
```

### How Glitch Detection Works

```go
func (gd *GlitchDetector) DetectGlitches(graph *GameGraph) []Glitch {
    var glitches []Glitch
    
    // 1. Find all active events
    activeEvents := graph.GetActiveEvents()
    
    // 2. For each event, check all possible conditions
    for _, event := range activeEvents {
        for _, rule := range gd.RuleMatrix[event.Type] {
            // 3. Check if condition is satisfied
            if graph.CheckCondition(rule.Condition) {
                // 4. Check if effect is NOT already triggered
                if !graph.IsEffectTriggered(rule.Effect) {
                    // 5. Found a potential glitch!
                    glitches = append(glitches, Glitch{
                        Event:     event,
                        Rule:      rule,
                        Creativity: rule.Score,
                        Feasibility: calculateFeasibility(graph, rule),
                    })
                }
            }
        }
    }
    
    // 6. Sort by creativity score
    sort.Slice(glitches, func(i, j int) bool {
        return glitches[i].Creativity > glitches[j].Creativity
    })
    
    return glitches
}
```

---

## State Verification

### Win Condition Types

#### Type 1: State-Based Win

```go
type StateWinCondition struct {
    StateKey  string      `json:"state_key"`  // "security_level"
    Operator  string      `json:"operator"`   // "==", ">", "<"
    Value     interface{} `json:"value"`      // 0
}

// Example: "Security level drops to 0"
{
    "type": "state",
    "state_key": "system_security_level",
    "operator": "==",
    "value": 0
}
```

#### Type 2: Event-Based Win

```go
type EventWinCondition struct {
    EventID   string `json:"event_id"`   // "password_leaked"
    Triggered bool   `json:"triggered"`  // true
}

// Example: "Password hash is leaked"
{
    "type": "event",
    "event_id": "password_hash_leak",
    "triggered": true
}
```

#### Type 3: Chain-Based Win

```go
type ChainWinCondition struct {
    ChainID   string `json:"chain_id"`   // "explosion_chain"
    MinLength int    `json:"min_length"` // 3
    MaxTime   int    `json:"max_time"`   // 10 (seconds)
}

// Example: "Chain reaction of at least 3 effects within 10 seconds"
{
    "type": "chain",
    "chain_id": "explosion_chain",
    "min_length": 3,
    "max_time": 10
}
```

---

## Level Configuration (JSON)

### Designer-Friendly Format

```json
{
  "id": "magitech_level_001",
  "theme": "magitech",
  "title": "The Frozen Library",
  "description": "Bypass the royal mages' spell-ward by combining runes they think are useless.",
  "nodes": [
    {
      "id": "fire_rune",
      "type": "EVENT",
      "name": "Fire Rune",
      "properties": {
        "element": "fire",
        "power": 50,
        "description": "A basic fire rune. Seems harmless alone."
      }
    },
    {
      "id": "ice_rune",
      "type": "EVENT",
      "name": "Ice Rune",
      "properties": {
        "element": "ice",
        "power": 30,
        "description": "A weak ice rune. Why would you need this?"
      }
    },
    {
      "id": "heat_source",
      "type": "CONDITION",
      "name": "Near Heat Source",
      "properties": {
        "radius": 5,
        "description": "The fireplace in the library corner."
      }
    },
    {
      "id": "thermal_explosion",
      "type": "EFFECT",
      "name": "Thermal Explosion",
      "properties": {
        "damage": 200,
        "radius": 10,
        "description": "Creates a shockwave that disables all wards."
      }
    }
  ],
  "edges": [
    {
      "from": "fire_rune",
      "to": "heat_source",
      "weight": 0.8
    },
    {
      "from": "ice_rune",
      "to": "thermal_explosion",
      "weight": 1.0,
      "condition": "near_heat_source"
    }
  ],
  "win_condition": {
    "type": "state",
    "state_key": "library_wards_disabled",
    "operator": "==",
    "value": true
  },
  "hints": [
    "The mages think ice is useless against their wards.",
    "What happens when ice meets intense heat?",
    "Sometimes opposites create the biggest explosions."
  ]
}
```

---

## Engine Evaluation Flow

### Step-by-Step Process

```
1. LOAD LEVEL
   └── Parse JSON configuration
   └── Build GameGraph with nodes and edges
   └── Initialize global state

2. PLAYER ACTION
   └── Player connects/modifies nodes
   └── Send action to engine

3. EVALUATE GRAPH
   └── Check all active events
   └── Evaluate conditions against global state
   └── Trigger valid effects
   └── Update global state

4. DETECT GLITCHES
   └── Run GlitchDetector on current graph
   └── Find unscripted solutions
   └── Calculate creativity scores

5. CHECK WIN CONDITION
   └── Evaluate win condition against new state
   └── If satisfied → Level complete!
   └── If not → Continue playing

6. GENERATE PASSCODE
   └── Hash: graph_state + glitches_found + player_approach
   └── Return 16-character passcode
```

---

## Go Backend Implementation

### Core Engine

```go
package engine

type UniversalLogicEngine struct {
    graphs      map[string]*GameGraph
    detectors   map[string]*GlitchDetector
    currentState map[string]interface{}
}

func NewUniversalLogicEngine() *UniversalLogicEngine {
    return &UniversalLogicEngine{
        graphs:      make(map[string]*GameGraph),
        detectors:   make(map[string]*GlitchDetector),
        currentState: make(map[string]interface{}),
    }
}

func (ule *UniversalLogicEngine) LoadLevel(config []byte) error {
    var graph GameGraph
    if err := json.Unmarshal(config, &graph); err != nil {
        return err
    }
    
    ule.graphs[graph.ID] = &graph
    ule.detectors[graph.ID] = NewGlitchDetector(graph.Theme)
    
    return nil
}

func (ule *UniversalLogicEngine) PlayerAction(levelID string, action PlayerAction) (*ActionResult, error) {
    graph, exists := ule.graphs[levelID]
    if !exists {
        return nil, fmt.Errorf("level not found: %s", levelID)
    }
    
    // 1. Apply player action to graph
    if err := graph.ApplyAction(action); err != nil {
        return nil, err
    }
    
    // 2. Evaluate all nodes
    graph.Evaluate()
    
    // 3. Detect glitches
    detector := ule.detectors[levelID]
    glitches := detector.DetectGlitches(graph)
    
    // 4. Check win condition
    won := graph.CheckWinCondition()
    
    // 5. Generate passcode
    passcode := generatePasscode(graph, glitches, action)
    
    return &ActionResult{
        Graph:    graph,
        Glitches: glitches,
        Won:      won,
        Passcode: passcode,
    }, nil
}
```

---

## Benefits of This Architecture

### 1. Single Engine for All Themes
- Magic, code, physics all use the same graph structure
- No separate engines needed

### 2. Designer-Friendly
- Edit JSON to create new levels
- No programming required

### 3. Emergent Solutions
- GlitchDetector finds unscripted solutions
- Encourages creative thinking

### 4. Scalable
- Add new themes by adding new rules to RuleMatrix
- Add new effects by adding new node types

### 5. Performant
- Graph evaluation is O(V + E)
- Glitch detection is O(R × E) where R = rules

---

## Implementation Timeline

### Week 1
- [ ] Implement GameNode and GameGraph structures
- [ ] Build basic graph evaluation
- [ ] Create JSON level loader

### Week 2
- [ ] Implement GlitchDetector with rule matrix
- [ ] Add magitech theme rules
- [ ] Add cyberpunk theme rules

### Week 3
- [ ] Build win condition evaluator
- [ ] Implement passcode generation
- [ ] Add industrial theme rules

### Week 4
- [ ] Polish and optimize
- [ ] Add level editor UI
- [ ] Test with 100+ levels

---

*Last updated: 2026-07-10*
