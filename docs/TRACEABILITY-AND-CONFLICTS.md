# Traceability Matrix & Conflict Report

*Generated during the verification-and-completion pass. Authority order (highest → lowest):*
1. Current working code that builds and passes tests
2. `ARCHITECTURE-PHASE1.md`
3. `CHALLENGE-TO-YOU-PLAN.md`
4. `GAME-DESIGN.md`
5. `ARCHITECTURE.md` and older design docs (historical)

---

## 1. Documentation Conflicts (resolved per authority order)

| # | Topic | Older docs (`ARCHITECTURE.md`, PLAN early sections) | Reality (code + `ARCHITECTURE-PHASE1.md`) | Resolution |
|---|-------|--------------------------------|-------------------------------------------|------------|
| C1 | **Transport** | GDExtension native calls, *"no WebSocket for desktop"*, `libchallenge.so` | WebSocket server (`internal/server`, gorilla/websocket) + `client/scripts/network_bridge.gd` | Code wins. WebSocket is authoritative; GDExtension design is historical. |
| C2 | **Sandbox** | WASM (Extism/Wasmer) | `ARCHITECTURE-PHASE1.md` §7 shows Docker; code uses hardened host subprocess (`os/exec`) | Subprocess is the current implementation. WASM/Docker are **intentionally deferred** (PLAN itself: *"WASM for MVP: Skip for alpha, add for Steam"*). See D2. |
| C3 | **Backend packages** | `internal/{analyzer,passcode,narrative}`, `pscript/` | `internal/{ai,engine,compiler,executionengine,executor,gameloop,missionengine,eventbus,jobqueue,ratelimit,content,server,generator,sandbox,db}` | Code wins. Doc-named packages never existed; functionality lives in the packages above (e.g. passcode → `engine/matrix.go` "Logos Cipher"; analyzer → `ai` oracle + AST). |
| C4 | **Passcode naming** | "Passcode" | In-fiction "Logos Cipher" (`LogosCipher`) | Same concept; naming is intentional game fiction. No change. |

No conflict required a breaking architectural change. `ARCHITECTURE.md` has been annotated as historical.

---

## 2. Requirement → Implementation Traceability

Legend: ✅ Implemented · ⚠️ Partial · ⛔ Deferred (with reason) · ❌ Missing

| Documented requirement (PHASE1 / PLAN / GAME-DESIGN) | Package / file | Status | Notes |
|---|---|---|---|
| Seed-based procedural generation + luck | `internal/generator` | ✅ | Tests pass (seed reproducibility). |
| Challenge model (`ChallengeDefinition`) + evaluation | `internal/engine` (`challenge.go`, `evaluator.go`) | ✅ | Canonical engine model (per prior unification audit). |
| Emergent multi-layer / composite combination | `internal/engine/composite.go` | ✅ | Deterministic combiner (concat/state/pipe). |
| Dynamic passcode ("Logos Cipher") | `internal/engine/matrix.go`, `internal/server/challenge_service.go` | ✅ | Emerges from code interaction. |
| Mission engine + objectives | `internal/missionengine` | ✅ | Persists completion/active sessions via `*db.DB`. |
| Game loop (tick, replay, telemetry, resources) | `internal/gameloop` | ✅ | Replay + telemetry + inventory. |
| Canonical event system | `internal/eventbus` | ✅ | Mission/challenge/code/dialogue/game/level-up/achievement/archon events. |
| WebSocket protocol | `internal/server` | ✅ | Live transport (see C1). |
| Code execution / sandbox | `internal/sandbox`, `internal/executor/python`, `internal/executionengine` | ✅ | Hardened host subprocess (env isolation, process-group kill, argv exec). See D2 for isolation tier. |
| Compiler / language-plugin format | `internal/compiler` (`Language{CompileCmd,RunCmd}`) | ✅ | Data-driven language definitions. |
| AI oracle (Ollama) + deterministic fallback | `internal/ai` | ✅ | Local LLM with fallback. |
| Progression / XP curve / titles | `internal/db` (`ComputeLevel`, `TitleForLevel`) | ✅ | Curve `[0,100,250,…]`, titles Newcomer→Archon. |
| Persistent save (versioned, integrity-checked) | `internal/db` (`SaveVersion`, `VerifySave`, checksum) | ✅ | Single profile row (`id = 1`). |
| Rate limiting | `internal/ratelimit` | ✅ | |
| Async job queue | `internal/jobqueue` | ✅ | |
| Dialogue system (data + events) | `internal/content/types.go`, `internal/missionengine/types.go`, events | ✅ | Runtime data + events; client renders. |
| Godot client (editor, terminal, bridge) | `client/scripts/{main,network_bridge}.gd` | ✅ | Desktop client present. |
| Phoenix self-healing AI pipeline | `phoenix/` module | ✅ | Auxiliary tooling module. |
| **Multi-slot save + autosave triggers** (PHASE1 §8.2) | — | ⛔ | Single-slot versioned save exists; multiple slots not implemented. **Deferred**: not required for v1 alpha (single-profile roguelike); no conflict with current save layer. |
| **True WASM / Docker isolation** (ARCHITECTURE.md, PHASE1 §7) | — | ⛔ | **Deferred by design** (PLAN: WASM = post-alpha). Current subprocess path hardened this pass; upgrading to container/WASM isolation is a Steam-phase task. |
| Constraint solver (Milestone 5) | — | ⛔ | Explicit future roadmap milestone. |
| P-Script swarm / advanced VM (Milestone 6) | — | ⛔ | Explicit future roadmap milestone. |

---

## 3. Intentionally Deferred (with reason)

- **WASM/Docker sandbox** — PLAN designates WASM as post-alpha ("Skip for alpha, add for Steam"). The host-subprocess path was security-hardened instead (env isolation, process-group kill on timeout, direct-argv exec, goja VM interrupt).
- **Multi-slot / autosave** — v1 is a single-profile roguelike; the versioned, checksum-verified single save satisfies current requirements.
- **Milestones 5–9 & Phase 2/3** — the roadmap explicitly schedules these (constraint solver, swarm, content pipeline, polish, more eras, multiplayer) as future work.

These are deferrals, not defects: none is a stub, TODO, or broken path in the codebase.
