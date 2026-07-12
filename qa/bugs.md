# QA Bug Log

This document records the integration bugs discovered and resolved during build verification:

---

## 🐛 Discovered & Resolved Issues

### 1. Challenge Test File Paths
- **Impact**: Backend tests in `internal/engine/` failed to locate standard challenge configs because `magitech_01.json` duplicate was removed from root directory.
- **Resolution**: Updated `challenge_test.go` to look in `challenges/magitech_tier1/magitech_01.json`.

### 2. Sandbox DB Relative Path Warning
- **Impact**: Launching the server with relative `DB_PATH=qa/test_playthrough.db` failed when executing from inside `backend/cmd/sandbox/` due to missing directory structure.
- **Resolution**: Changed startup configurations to utilize absolute paths or run from the workspace root.
