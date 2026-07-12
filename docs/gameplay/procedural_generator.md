# Procedural Challenge Generator Architecture

This document defines the design and logic templates for the procedural challenge generator, capable of outputting 100,000+ unique code problems.

---

```
                       [PROCEDURAL GENERATOR PIPELINE]
                                      │
                                      ▼
                      ┌───────────────────────────────┐
                      │    1. Template Selection      │
                      │    - Load base code structure │
                      └───────────────┬───────────────┘
                                      │
                                      ▼
                      ┌───────────────────────────────┐
                      │    2. Parameter Mutation      │
                      │    - Set array sizes & names  │
                      └───────────────┬───────────────┘
                                      │
                                      ▼
                      ┌───────────────────────────────┐
                      │    3. Verification Run        │
                      │    - Run tests in sandbox     │
                      └───────────────┬───────────────┘
                                      │
                                      ▼
                      ┌───────────────────────────────┐
                      │    4. Adaptive Difficulty     │
                      │    - Adjust budget parameters │
                      └───────────────────────────────┘
```

---

## 1. Challenge Template System

The generator loads a base JSON template containing a functional challenge schema.

### Example Base Template:
```json
{
  "template_id": "array_search_01",
  "category": "algorithms",
  "variables": {
    "array_name": ["mana_nodes", "packet_queue", "process_ids", "qubit_states"],
    "target_name": ["target_val", "key_address", "auth_token"],
    "array_size": "range(10, 100)",
    "difficulty_modifier": "range(1, 5)"
  },
  "code_skeleton": {
    "python": "def solve(data_array, target):\n    # TODO: Write search logic\n    pass"
  }
}
```

---

## 2. Parameter Mutation & Story Randomization

The generator applies mutations based on a system seed:
- **Identifier Shuffling**: Variable names are mapped based on the universe (e.g. `mana_nodes` in Magitech, `packet_queue` in Cyberpunk).
- **Data Set Randomization**: Inputs and expected outputs are dynamically computed by executing a reference python solver during template generation.
- **Constraint Mutation**: Adds specific resource checks (e.g. no nested loops, maximum execution cycles) depending on difficulty multipliers.
- **Vigilance Scaling**: Steps limits and memory allowances scale down linearly with higher difficulty ratings.

---

## 3. Dynamic Test Case Verification & Replay Determinism

Before serving the challenge to the player:
- **Deterministic Seeding**: Puzzles generate using a 64-bit integer seed. Sharing this seed recreates the exact challenge structures, parameter arrays, and test cases.
- **Sandbox Evaluation**: The backend runs the reference solution against the generated test cases inside the sandbox.
- **Assertion Checks**: If tests do not pass or reference code execution exceeds cycle limits, the level is discarded and re-seeded.

---

## 4. Adaptive Learning & Hint Generation

The generator tracks player performance metrics:
- **Compilation Success Rate**: High success rates trigger harder parameter bounds.
- **Operations Density**: If players write highly optimized solutions in early challenges, the engine decreases cycle budgets in subsequent runs.
- **Hint Engine**: When multiple failures occur, the engine parses AST diffs and outputs hint directives (e.g. "Check your loop termination index" or "A recursion base case is missing").
- **Adaptive Prerequisite Gating**: Soft gates open lower-difficulty helper nodes when players struggle with advanced structures.
