# Challenge To YOU Validation System

Validation results, open questions, and rejected ideas are tracked in [research_log.md](research_log.md).

## 1. Validation Gates

Every prototype must pass these gates in order.

### Gate 1: Intrinsic Enjoyment

Would someone voluntarily repeat this interaction?

### Gate 2: Curiosity

Does it naturally create "What happens if..."?

### Gate 3: Stories

Can players tell stories afterward?

### Gate 4: Knowledge

Does replaying produce genuine understanding?

### Gate 5: Mastery

Can experts still discover new things?

## 2. Curiosity Pipeline

```text
Curiosity

↓

Experiment

↓

Explanation

↓

Transfer

↓

Story
```

Every mechanic should support every stage.

## 3. Four-Layer Design Stack

Every mechanic should specify all four layers.

### Layer 1: Interaction

What does the player physically do?

### Layer 2: Cognition

What mental process occurs?

### Layer 3: Emotion

What should they feel?

### Layer 4: Memory

What story remains afterward?

## 4. Mechanic Card Template

Every mechanic gets exactly one page.

### Name

Latency Injection

### Purpose

Why does this exist?

### Player Interaction

Exactly what buttons does the player press?

### Player Thought

What are they thinking?

### Player Emotion

What should they feel?

### Player Story

What story will they tell later?

### Curiosity Trigger

What question naturally appears?

Example:

> What if I delay this packet?

### Learning Outcome

What concept is learned?

### Interactions

Which existing mechanics combine with this?

### Failure Modes

What mistakes teach something?

### Prototype Status

- Untested
- Testing
- Validated
- Rejected

## 5. Mechanic Fitness Score

Every mechanic scores itself.

| Metric | Question |
|--------|----------|
| Repeatability | Would players willingly use it hundreds of times? |
| Curiosity | Does it naturally invite experimentation? |
| Composability | Does it combine with existing mechanics? |
| Explainability | Can players understand surprising outcomes afterward? |
| Story Potential | Can it generate memorable anecdotes? |
| Mastery Depth | Does expertise meaningfully improve performance? |
| Spectator Value | Is it interesting to watch? |
| Elegance | Does it solve multiple design goals at once? |

Minimum average to enter production: 4.0 / 5

## 6. Hypothesis Template

Every prototype starts like this.

### Hypothesis

Players who can inject latency into a message bus will voluntarily perform at least five additional experiments beyond what is required.

### Success Metric

- Average Experiment Rate > 3
- Average Story Rate > 50%
- Players correctly explain the mechanic after solving

### Prototype

One mechanic.

One puzzle.

Nothing else.

### Observe

Record:

- time spent experimenting
- number of voluntary experiments
- wrong hypotheses
- corrected hypotheses
- smiles
- laughter
- verbal reactions
- frustration
- explanation quality

### Decision

- Keep
- Modify
- Reject

## 7. Research Log Workflow

Before a prototype starts, write a hypothesis to the research log.

After the prototype ends, record:

- result
- evidence
- decision
- open questions

The research log is the changeable record. The constitution is the stable record.

## 8. Prototype Roadmap

Spend the next 6-8 weeks validating the core interaction.

| Prototype | Goal | Success Condition |
|----------|------|-------------------|
| P1: Fault Injection | Is the core interaction fun by itself? | Players voluntarily retry experiments |
| P2: Imperfect Telemetry | Do players enjoy inference? | They revise hypotheses instead of guessing |
| P3: Signal Delay | Does timing create interesting decisions? | Multiple valid strategies emerge |
| P4: Visual Causality | Does the causal graph improve understanding without spoiling discovery? | Players explain outcomes more accurately |
| P5: Zero-Code | Is reasoning fun without scripting? | Players stay engaged using only observation and interventions |
| P6: Text Scripting | Does programming add depth rather than friction? | Experimentation increases instead of decreasing |
| P7: Hybrid Interface | Compare text + visual controls | Higher curiosity and lower frustration than either alone |

## 9. Weekly Review

Every Friday, do not ask:

- How much code did we write?
- How many features are done?
- How many bugs were fixed?

Ask:

1. What surprised players this week?
2. What experiment did every tester try voluntarily?
3. Which mechanic generated the best story?
4. Which mechanic nobody cared about?
5. Which hypothesis did we invalidate?
6. What should we remove next week?

If you cannot answer those questions, you are optimizing implementation rather than player experience.
