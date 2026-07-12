# Project Phoenix: Memory System Schema

This document defines the directories, metadata formats, and retrieval systems that structure the self-repair memory database under the `brain/` directory.

---

## 1. Directory Structure

Every execution loop results in a structured log to seed historical learning:

```
brain/
    build_failures/       # Captured compiler diagnostics and log output
        [timestamp]_[hash].json
    runtime_failures/     # Stack traces and unit test failures
        [timestamp]_[hash].json
    successful_repairs/   # Patches that successfully passed validation
        [timestamp]_[hash].json
    rejected_repairs/     # Patches that compiled but failed test audits
        [timestamp]_[hash].json
    performance/          # Benchmark history indexes
        [hash].json
```

---

## 2. Failure Case Metadata Schema

Each failure profile stores the complete system context:

```json
{
  "failure_id": "err_531b48d2",
  "timestamp": "2026-07-11T03:45:00Z",
  "category": "compilation",
  "target_modules": ["backend/cmd/sandbox/main.go"],
  "error_diagnostic": "use of internal package challenge-to-you/backend/internal/eventbus not allowed",
  "evidence": {
    "git_diff": "diff --git a/backend/verify_eventbus_verify.go b/backend/verify_eventbus_verify.go...",
    "stack_trace": "package command-line-arguments\\n\\t../../../../.gemini/antigravity-ide/brain/c4e154d7-2e40-48e3-adb8-48bb50a6553c/scratch/verify_eventbus.go:8:2: use of internal package challenge-to-you/backend/internal/eventbus not allowed"
  }
}
```

---

## 3. Successful Repair Schema

Stores patches that resolved verification runs:

```json
{
  "repair_id": "rep_a97b04f1",
  "associated_failure": "err_531b48d2",
  "applied_patch": "diff --git a/backend/verify_eventbus_verify.go b/backend/verify_eventbus_verify.go\n...",
  "validation_result": {
    "build_duration_ms": 1250,
    "tests_passed": true,
    "benchmarks": {
      "allocated_bytes_per_op": 128,
      "cpu_ns_per_op": 450
    }
  }
}
```

---

## 4. Search & Retrieval Pipeline

Before the LLM Repair agent drafts a patch:
1. **Semantic Querying**: The Knowledge Manager generates a vector embedding of the current compiler error string.
2. **K-Nearest Search**: Queries `brain/build_failures/` and retrieves the 3 most similar historical error blocks.
3. **Patch Injection**: If matches exist, the corresponding `successful_repairs/` patches are injected into the LLM context as few-shot examples, ensuring the agent uses pre-validated patterns.
