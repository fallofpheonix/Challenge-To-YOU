# Challenge Writing Guidelines & Quality Standards

This document defines the quality gates, testing checklists, and naming conventions required for adding challenge content to the database.

---

## 1. Challenge Writing Guidelines

All challenges must align with standard formats:
- **Narrative Framing**: Problems must avoid raw, dry coding text (e.g. "Write an array search"). Instead, use systems lore (e.g. "Search the active memory bus to isolate corrupted register addresses").
- **Language Portability**: Keep code templates compatible across target scripting languages (e.g. avoid syntax structures that cannot be easily written in Python or JavaScript).

---

## 2. Naming Conventions

Puzzles follow strict ID formats:
- **Magitech**: `M-01` to `M-99`
- **Cyberpunk**: `C-01` to `C-99`
- **Void**: `V-01` to `V-99`
- **Embedded**: `S-01` to `S-99`
- **AI/ML**: `N-01` to `N-99`
- **Kernel**: `K-01` to `K-99`

---

## 3. Testing & Review Checklist

Before a challenge is merged into the master branch:

- [ ] **Verification**: Reference solution compiled and ran inside the sandbox with zero compile errors.
- [ ] **Boundary Tests**: Test suites include checks for empty arrays, overflow integers, and null pointers.
- [ ] **Cycle Assertion**: The reference code completes execution using under 80% of the maximum operations budget.
- [ ] **Curriculum Match**: The problem mapping matches an ACM/IEEE curriculum course.
- [ ] **Localization**: All narrative strings are extracted to translation catalogs.
- [ ] **Accessibility**: High-contrast normal maps are configured for graphics puzzles.

---

## 4. Version Control & Approval Workflow

1. Content designers write challenges on distinct branch forks.
2. The CI/CD pipeline compiles the backend, executes the reference solver against test matrices, and validates the schema.
3. Two separate peer reviews must approve the PR before the branch is merged into the master repository.
