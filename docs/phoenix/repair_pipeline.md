# Project Phoenix: Self-Repair Execution Pipeline

This document defines the step-by-step lifecycle of the self-repair loop, executing from error detection to patch validation.

---

## 1. Pipeline Execution Phases

```
[DETECT] ──► [REPRODUCE] ──► [ANALYZE] ──► [GENERATE] ──► [COMPILE] ──► [TEST & BENCH] ──► [COMMIT]
```

### Phase 1: Detect
- The File Change Detector identifies file mutations.
- The task queue listens for compile faults, failed test runs, or static analysis alerts.

### Phase 2: Reproduce
- The pipeline isolates the issue by running the specific compiler or test command that failed on a clean branch copy of the workspace.
- Verification ensures that the issue reproduces deterministically before launching LLM repair prompts.

### Phase 3: Analyze
- Gathers logs, stack traces, active diff changes, and historical repair metadata.
- Prepares the aggregated context payload and identifies context code files.

### Phase 4: Generate Patch
- The Repairer agent processes the context payload, generating a minimal unified git patch block (`patch.diff`).

### Phase 5: Compile
- Applies the patch using `git apply patch.diff`.
- Runs incremental compilation checks: `go build` and `godot --check-only`.
- If compilation fails, the compiler error logs are routed back to Phase 3 for a retry check.

### Phase 6: Test & Benchmark
- Executes the test suite: `go test ./...`.
- Runs memory and CPU benchmarks. If execution latency or byte consumption increases by more than 10%, the patch is rejected.

### Phase 7: Commit Proposal
- If all checks pass, the validator commits the changes to a temporary `repair/[ticket_id]` branch and opens a Pull Request for human review, logging the success to `brain/successful_repairs/`.
