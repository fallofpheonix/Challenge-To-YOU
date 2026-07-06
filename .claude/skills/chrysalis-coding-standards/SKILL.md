---
name: chrysalis-coding-standards
description: Naming, structure, and testing rules across the Go core, P-Script, and Godot client. Use before writing or reviewing code in this repo to keep the deterministic split-stack consistent.
---

# Chrysalis Coding Standards

Split-stack (Go core + Godot client) with an enforced air-gap between simulation and presentation. Standards exist to prevent desyncs and keep the sim deterministic.

## Naming
**Go (core):** files `snake_case.go`; structs/interfaces `PascalCase`; `camelCase` internal / `PascalCase` exported. ECS pure-data structs end in `Component`; iterators/logic end in `System`.
**GDScript (client):** files `snake_case.gd`/`.tscn`; classes/nodes `PascalCase`; vars/functions `snake_case`; **signals past-tense `snake_case`** (`signal_received`, `connection_dropped`).
**P-Script (`.ps`):** built-ins `UPPER_SNAKE_CASE` (`SENSE_BATTERY`, `MOVE_TOWARDS_HOME`); see [[chrysalis-pscript]].

## Determinism (non-negotiable, core)
No floats — use `crysmath.FixedPoint` (×10^6). Seeded RNG only. No reliance on Go map iteration order. Mutate grid via `Next`+`SwapBuffers`. Details: [[chrysalis-simulation]].

## Testing tiers
- **Tier 1 — Determinism (Go, critical):** seed → inject script → run fixed ticks → hash final state. **Hash change = build fails.** Zero tolerance for flakes (a flake = race/float leak = critical bug).
- **Tier 2 — Sandbox guillotine (Go):** attack the rule engine — reject malicious/over-nested scripts; verify the VM 1000-step limit trips and the tick still finishes <100ms without crashing.
- **Tier 3 — Client telemetry (Godot, lightweight):** feed chaotic pre-recorded packets; assert 60 FPS while rendering thousands of nodes/heatmaps.

Tooling: `engine/core/Makefile`, `.golangci.yml`, `.air.toml` (live reload).

## ⚠️ Doc drift
`04_ENGINEERING/CODING_STANDARDS.md` and the folder layout it shows (`go_server/`, `godot_client/`, Python `sandbox/`) are **stale**. Reality: monorepo is `engine/core` (Go), `engine/client` (Godot), and the rule engine is **P-Script**, not Python. Trust `CONTEXT.md` + code. See [[chrysalis-engine-architecture]].
