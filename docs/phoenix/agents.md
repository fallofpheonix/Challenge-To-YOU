# Project Phoenix: Specialized Agent Roles

This document defines the core responsibilities, tool interfaces, and model configurations for the 9 specialized agents in the pipeline.

---

## 1. Agent Directory & Models

| Agent Role | Model Choice (Local) | Responsibilities | Target Commands |
| :--- | :--- | :--- | :--- |
| **Builder** | Qwen2.5-Coder 7B | Incremental compiling and checks | `go build ./...`, `godot --check-only` |
| **Tester** | Qwen2.5-Coder 32B | Running unit, integration, and replay tests | `go test ./...`, custom test runners |
| **Reviewer** | DeepSeek Coder 33B | Static analysis audits, style linting | `golangci-lint`, `go vet`, `gosec` |
| **Repairer** | DeepSeek Coder 33B | Generating unified git patches for fixes | `git apply` |
| **Optimizer** | Qwen2.5-Coder 32B | Performance profiling, latency, memory use | `go test -bench` |
| **Security** | Qwen2.5-Coder 7B | Vulnerability checks and secret detection | `govulncheck`, `gitleaks` |
| **Content** | Qwen2.5-Coder 7B | Validating JSON schemas and unlocks | Custom schema validation checks |
| **Architect** | Qwen3 32B | Enforcing public API signature boundaries | Go interface checks |
| **Knowledge** | Qwen3 32B | Indexing fixes, decisions, and history | Vector database queries |

---

## 2. Detailed Agent Specifications

### 2.1 Builder
- **Inputs**: Modified file paths list.
- **Tools**: CLI executor for target languages.
- **Workflow**: Run incremental compilation. If compile fails, capture standard error logs containing line numbers and file names, formatting it as a compile error case.

### 2.2 Tester
- **Inputs**: Build outputs.
- **Tools**: Go test harness, Godot headless check.
- **Workflow**: Run standard unit and integration test suites. If failures occur, collect stdout, stderr, stack traces, and execution timings.

### 2.3 Reviewer & Security Auditor
- **Inputs**: Git diff modifications.
- **Tools**: `golangci-lint`, `go vet`, `govulncheck`, `gosec`, `deadcode`.
- **Workflow**: Audit changed code blocks for syntax style, dead code channels, and cryptographic vulnerability bugs. Flag lines violating rules.

### 2.4 Repairer (Self-Repair Core)
- **Inputs**: Aggregated failure logs, compile error details, active diffs, source files, and design docs.
- **Tools**: Context compilation helper.
- **Workflow**: Search `brain/` memory logs for similar historical issues. Generate a minimal unified Git patch file that resolves the diagnostic error without changing APIs or public interfaces.

### 2.5 Optimizer
- **Inputs**: Active code execution.
- **Tools**: pprof, benchmarking runs.
- **Workflow**: Measure execution CPU cycles, heap memory allocations, and database query rows. Flag bottlenecks and propose localized index or memory pool adjustments.
