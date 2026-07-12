# Master Challenge Index & Template Schema

This document defines the base JSON schemes and world connection templates used by the Challenge Engine to load and evaluate game challenges.

---

## 1. Challenge Template Schema

All procedural and static challenges follow a standardized JSON structure loaded by the engine.

### JSON Schema Specification:
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ChallengeDefinition",
  "type": "object",
  "properties": {
    "id": { "type": "string" },
    "title": { "type": "string" },
    "description": { "type": "string" },
    "skill_type": { 
      "type": "string",
      "enum": ["recognize", "optimize", "write_from_spec", "state_machine"]
    },
    "difficulty": { "type": "integer", "minimum": 1, "maximum": 10 },
    "language": { "type": "string" },
    "ops_budget": { "type": "integer" },
    "memory_limit_bytes": { "type": "integer" },
    "code_skeleton": { "type": "string" },
    "visible_tests": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "input": { "type": "string" },
          "expected_output": { "type": "string" },
          "description": { "type": "string" }
        },
        "required": ["id", "input", "expected_output"]
      }
    },
    "analytics_tags": {
      "type": "array",
      "items": { "type": "string" }
    }
  },
  "required": ["id", "title", "skill_type", "difficulty", "code_skeleton"]
}
```

---

## 2. Validation & Testing Strategy

Before importing a challenge configuration:
- **Reference Solvers**: Every challenge must include a hidden reference script in python/javascript that passes all test assertions.
- **Cycle Assertions**: The simulator measures execution cycle counts. Puzzles are validated by checking if the reference script operates within 80% of the maximum cycle budget.
- **Anti-Cheat Validation**: Test cases include extreme boundary inputs (e.g. empty lists, overflow numbers, null targets) to verify that players cannot bypass checks using simple static array matches.

---

## 3. Multiverse Connection Nodes

The campaign maps level transitions based on node progression values.

```
[LEVEL M-01] ──► [LEVEL M-02] ──► [BOSS DEFEATED] ──► [UNLOCK PORTAL C-41]
 (Magitech)       (Magitech)       (Grand Compiler)      (Cyberpunk Net)
```

- **Rift Transitions**: Defeating a universe boss awards the player a **Kernel Key**.
- **BIOS Compilation**: Collecting all 14 keys allows the player to access the final portal in *The Kernel Beyond*, initializing the bios rebuild loop.
