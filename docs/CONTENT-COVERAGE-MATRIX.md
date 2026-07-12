# Content Requirement Coverage Matrix

*Verified against code and content files (not estimated). Generated during the
content-completion pass. Authority order: working code > ARCHITECTURE-PHASE1 >
PLAN > GAME-DESIGN > remaining docs.*

## Universe coverage (14 documented)

Legend: ✓ Implemented · ◐ Partial · ✗ Missing

| # | Universe | Engine paradigm? | Challenge files | Campaign pack | Missions | Status |
|---|----------|:----------------:|:---------------:|:-------------:|:--------:|:------:|
| 1 | Medieval Magitech | ✓ MAGITECH | 9 (+3 packs) | 7 wired | 1 | ◐ |
| 2 | Cyberpunk Neon | ✓ CYBERPUNK | 12 | 6 wired | 0 | ◐ |
| 3 | Cosmic Void | ✓ COSMIC | 8 | 8 wired | 0 | ◐ |
| 4 | Silicon Wastes | ✗ none | 0 | ✗ | 0 | ✗ |
| 5 | Neural Labyrinth | ✗ none | 0 | ✗ | 0 | ✗ |
| 6 | Chrono Registry | ✗ none | 0 | ✗ | 0 | ✗ |
| 7 | Quantum Nexus | ✗ none | 0 | ✗ | 0 | ✗ |
| 8 | BioForge Genome | ✗ none | 0 | ✗ | 0 | ✗ |
| 9 | Data Abyss | ✗ none | 0 | ✗ | 0 | ✗ |
| 10 | Cloud Dominion | ✗ none | 0 | ✗ | 0 | ✗ |
| 11 | Machine Cathedral | ✗ none | 0 | ✗ | 0 | ✗ |
| 12 | Cipher Realm | ✗ none | 0 | ✗ | 0 | ✗ |
| 13 | Fractal Dream | ✗ none | 0 | ✗ | 0 | ✗ |
| 14 | The Kernel Beyond | ✗ none | 0 | ✗ | 0 | ✗ |

**Universes implemented: 3 / 14.** Universes 4–14 have no engine paradigm, no
challenge files, no packs, no missions, no NPC/dialogue data.

## Challenge coverage

| Metric | Count |
|--------|-------|
| Documented target (`docs/problems/database.md`) | 560 |
| Playable engine challenges implemented | ~29 (magitech 9, cyberpunk 12, cosmic 8) |
| Additionally, generic algorithm challenges (`data/challenges/`) | 7 |
| **Coverage** | **~5%** |

### Documentation-depth reality (why "560 as documented" is not literally possible)

`docs/problems/database.md` names only **~4–5 challenges per universe** explicitly
(e.g. `M-01 Runic Initiation … M-05 Golem Pathing`), then **collapses the remaining
~35 per universe into a single "various" range row** (e.g. `M-06 to M-40 | Rune Logic
6-40 | various`). So:

- **~65 challenges** are named with a one-line concept.
- **~495 challenges** have **no individual specification at all** — only a thematic range.
- `docs/problems/specifications.md` gives real design detail for **~14 milestone
  challenges** (one per universe), as prose (objective/algorithm/mistakes), not engine JSON.

There is **no per-challenge puzzle definition** (initial_state / flaws / win_condition)
for the vast majority. Additionally, the implemented challenges do **not** map to the
documented IDs (e.g. `magitech_01_breach` ≠ documented `M-01 Runic Initiation`).

## Blocking conflicts (must be resolved before bulk implementation)

### B1 — Engine supports 3 paradigms; docs specify 14 universes
`internal/engine/matrix.go` defines exactly `MAGITECH`, `CYBERPUNK`, `COSMIC`, with
paradigm-specific state hydration hard-coded in `hydrator.go`. Universes 4–14 each
need a **new engine paradigm + hydration + generator support** to be playable.
That is engine extension — which directly conflicts with the instruction *"do not
redesign the engine / do not rewrite working architecture."* Per the standing rule
*"if a requirement needs a breaking architectural change, stop and report with a
migration plan,"* universes 4–14 cannot proceed without a decision.

### B2 — ~495 challenges are not individually specified
"Implement exactly as documented" cannot apply where the documentation only says
"various." Realizing these requires **authoring original puzzle designs**, which the
rules explicitly forbid (*"never invent functionality"*). This is design work, not
transcription, and spans far beyond the documented content.

## Proposed path (options for the owner to choose)

1. **Campaign spine (feasible now):** author the ~14 milestone challenges from
   `specifications.md` for the 3 existing paradigms, in the engine's real format,
   wired to progression + missions + tested. Deepens universes 1–3 toward their
   documented 40 using the milestone specs that *do* exist.
2. **Lean on the procedural generator:** `internal/generator` already produces
   infinite seed-based challenges per existing paradigm — the intended mechanism for
   content volume. Bulk "40 per universe" can be generator-backed rather than hand-authored.
3. **Universes 4–14:** require an approved engine-paradigm extension plan (new
   `Paradigm` constants + hydrator entries + generator cases per universe). This is a
   scoped engineering effort, not a rewrite, but it *is* engine work and needs sign-off
   because it was explicitly placed out of scope.

Nothing in universes 4–14 or the ~495 unspecified challenges can be implemented
"exactly as documented" today — the specification does not contain them, and the
engine does not model their paradigms.
