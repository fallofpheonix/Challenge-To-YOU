# Player Progression & Skill trees

This document defines the progression curves, skill trees, and gating systems that govern player progression in **Challenge To YOU**.

---

```
                       [PLAYER EXPERIENCE LEVEL CURVE]
                                      │
                                      ▼
                      ┌───────────────────────────────┐
                      │  Level 1-10: Scribe (Era I)   │
                      │  - Basic DSL & loop unlocks   │
                      └───────────────┬───────────────┘
                                      │ (1,000 XP/lvl)
                                      ▼
                      ┌───────────────────────────────┐
                      │ Level 11-30: Hacker (Era II)  │
                      │  - Concurrency & sockets      │
                      └───────────────┬───────────────┘
                                      │ (5,000 XP/lvl)
                                      ▼
                      ┌───────────────────────────────┐
                      │ Level 31+: Ascendant (Era III)│
                      │  - Kernel & assembly control  │
                      └───────────────────────────────┘
```

---

## 1. XP Economy & Level Curves

Players earn experience points (XP) by solving challenges:
- **Challenge Clear**: `100 * Difficulty` XP.
- **Optimization Bonus**: Up to `50 * Difficulty` XP based on cycle reduction performance.
- **Perfect Run Bonus**: `50` XP for zero execution failures on first submission.
- **Level Leveling Formula**: `XP_Required = Level * 1000` (capped at 50,000 XP per level).

---

## 2. Skill Trees & Role Unlocks

The player allocates skill points across three distinct branches:

### 2.1 The Runic Scribe (Algorithmic Branch)
- **Node 1: Constant Binding**: Unlocks compile-time optimizations (reduces cycle cost of assignments).
- **Node 2: Loop Invariant**: Highlights loop invariants inside the editor.
- **Node 3: Recursion Shield**: Prevents call stack overflow on deep recursion, converting it into iterative structures behind the scenes.

### 2.2 The Net Intruder (System Branch)
- **Node 1: Async Shunt**: Allows swapping thread contexts in multi-threaded terminals.
- **Node 2: Memory Lock**: Prevents the garbage collector from reclaiming active variables.
- **Node 3: Port Bypass**: Spoofs connection requests, skipping minor proxy router handshakes.

### 2.3 The Bare-Metal Smith (Hardware Branch)
- **Node 1: Direct IO**: Maps variables directly to hardware registers.
- **Node 2: Cache Forward**: Bypasses write delays in simulated cache lines.
- **Node 3: Interrupt Vector Hook**: Custom ISR injection to override system schedulers.

---

## 3. Gating Systems

Progression is managed by soft and hard gates:
- **Soft Gating (Vigilance Limit)**: Challenges in Era II/III require optimized execution cycles. Players must purchase RAM upgrades and code-optimizing modules using Cyber Credits to expand their cycle budget.
- **Hard Gating (Kernel Keys)**: Gateways between universes are locked. Players must defeat the universe boss and acquire their **Kernel Key** to unlock subsequent worlds.

---

## 4. Prestige System

Upon reaching Level 50 and completing the final boot loop challenge:
- **Prestige Reset**: The player reboots their system profiles, resetting Level to 1 and clearing skill trees.
- **Perk Retention**: Retains all collected lore artifacts and titles.
- **Prestige Rewards**: Unlocks exclusive gold terminal themes, developer IDE fonts, and access to **Infinite Mainframe** priority tiers.
