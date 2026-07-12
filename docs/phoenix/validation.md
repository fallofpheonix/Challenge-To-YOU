# Project Phoenix: Patch Validation & Safety Specifications

This document defines the validation loop, trigger rules, safety constraints, and rollback protocols that govern code generation.

---

## 1. Safety Rules & Gating Constraints

To prevent self-repair agents from causing structural regression or data loss, the validator enforces strict blocks.

### 1.1 Forbidden Operations (AI Safety Gate)
The pipeline will immediately reject and discard patches containing:
- **Directory Deletions**: Deletion of workspace subdirectories (`rm -rf`).
- **Git History Rewriting**: Commands matching `git commit --amend`, `git rebase -i`, or `git push --force`.
- **Public API Modifications**: Alterations to public Go structures or interface types inside `backend/internal/`.
- **Schema Mutation**: Edits to JSON challenge definitions or database schema layouts.
- **Security Check Removal**: Deletions of signature filters, encryption routines, or certificate verification blocks.

---

## 2. Trigger Conditions

The validation pipeline runs automatically under target states:

- **Incremental Change**: File edits detected in the Go codebase or Godot script directories.
- **Pre-Commit Hook**: Triggers a local scan before git commits are finalized.
- **CI Pull Request**: Triggers an automated build/test/lint sweep on target repair branches.
- **Nightly Run**: Triggers a full project regression and compilation check.

---

## 3. Patch Validation & Rollback Loop

When the Repairer outputs a Git patch:

```
[patch generated] ──► [apply patch] ──► [verify build] ──► [run tests] ──► [audit lint]
                                                                               │
                                                                               ▼
[commit branch] ◄── [success evaluation] ◄── [run benchmarks] ◄── [no regressions]
```

### 3.1 Validation Process:
1. **Apply**: Execute `git apply patch.diff` on an isolated branch.
2. **Build**: Run `go build ./...` and `godot --headless --check-only`.
3. **Test**: Run `go test ./...`. If unit test assertions fail, the validation fails.
4. **Lint**: Run static analyzers. New code must output zero linter findings.
5. **Benchmark**: Execute `go test -bench ./...`. If performance regressions exceed 10% on CPU cycles or memory allocation, the patch is discarded.

### 3.2 Rollback Execution:
If any validation step fails:
- The system automatically runs `git checkout .` to reset the repository state.
- Logs the failure to `brain/rejected_repairs/` and notifies the development team.
