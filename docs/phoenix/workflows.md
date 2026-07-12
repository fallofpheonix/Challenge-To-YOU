# Project Phoenix: Operational Workflows

This document defines the automated event-driven workflows that trigger code building, testing, self-repair, and documentation refreshes.

---

## 1. Incremental File Change Workflow

Triggered when a developer edits a source code file:

```
[File Change Detected]
         │
         ▼
[Determine Module] ──► (Backend: run 'go build')
         │         ──► (Client: run 'godot --check-only')
         ▼
[Run Unit Tests]   ──► (Success: finish workflow)
         │         ──► (Failure: Trigger Test Failure Workflow)
```

---

## 2. Build Failure Workflow

Triggered when compilation fails:
1. **Isolate Diagnostic**: Extract standard error compiler messages (identify file, line, and syntax error).
2. **Retrieve Context**: Inject the failing file code block and surrounding function structure.
3. **Query Brain**: Search `brain/build_failures/` for similar error diagnostics.
4. **Generate & Validate**: Execute the Repairer to generate a patch. Run the validator.
5. **Human Review**: If the patch fails compilation validation 3 times, halt the loop and ping the human developer with the evidence bundle.

---

## 3. Nightly Maintenance Workflow

Runs automatically every night to perform system cleanup:
- **Dependency Audit**: Run `govulncheck` to detect security vulnerabilities in imports.
- **Dead Code Sweep**: Execute `deadcode` on the backend packages, flagging unused variables or functions.
- **Index Optimization**: Rebuild vector database embeddings in the `brain/` database.
- **Documentation Refresh**: Automatically regenerate API and GoDoc documentation for newly added components.
- **Lore Graph Checks**: Scan all markdown files under `docs/` for broken references or timeline conflicts.
