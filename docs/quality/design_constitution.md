# Challenge To YOU v1.0 Design Constitution

## Scope

This document contains only stable rules that have survived validation.

Hypotheses, prototype results, rejected ideas, open questions, and evidence belong in [research_log.md](research_log.md).

## 1. Vision

> Challenge To YOU is a systems puzzle game where players make increasingly better decisions under uncertainty by observing, experimenting with, and mastering complex interactive systems.

Everything else is subordinate to this sentence.

## 2. Player Fantasy

Not:

> I am writing code.

Not:

> I am debugging.

Not:

> I am solving puzzles.

Instead:

> I understand this system better than anyone else, and I can make it behave exactly the way I want.

This fantasy should exist in every puzzle.

## 3. Atomic Interaction

Every five seconds the player should perform one tiny experiment.

```text
Observe

↓

Hypothesis

↓

Small Intervention

↓

Immediate System Response

↓

Explanation

↓

New Hypothesis
```

If a mechanic breaks this loop, remove it.

## 4. Core Gameplay Loop

```text
Observe

↓

Model

↓

Predict

↓

Experiment

↓

Observe

↓

Explain

↓

Master
```

Programming is not here.

Programming is merely one implementation of "Experiment."

## 5. Design Invariants

These are absolute.

### I1
Every puzzle must support multiple meaningful solutions.

### I2
Every surprising outcome must be explainable afterward.

### I3
Failure must reveal useful information.

### I4
Player improvement should primarily come from knowledge.

Not stats.

Not grinding.

Not upgrades.

### I5
Every mechanic must combine with existing mechanics.

Never add isolated mechanics.

### I6
Every mechanic must generate curiosity.

### I7
Every mechanic must be capable of creating a player story.

### I8
Every mechanic must increase experimentation.

Never decrease it.

## 6. Anti-Invariants

The game must never become:

❌ Coding interview practice

❌ Syntax memorization

❌ Trial-and-error guessing

❌ Random chaos

❌ One correct solution

❌ Debugging paperwork

❌ Artificial difficulty

❌ Information hidden without reason

## 7. Design Heuristics

These are defaults.

Prefer:

- interaction over explanation
- experiments over tutorials
- systems over scripted events
- composition over feature count
- discovery over exposition
- visible causality over hidden randomness
- depth over breadth
- player knowledge over character progression

## 8. Review Process

The constitution changes only when a repeated validation cycle proves a rule is wrong or incomplete.

When that happens:

1. Record the hypothesis and evidence in [research_log.md](research_log.md).
2. Update the constitution only after the new rule has survived validation.
3. Keep rejected ideas in the research log for historical context.
