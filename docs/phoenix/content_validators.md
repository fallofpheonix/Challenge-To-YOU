# Project Phoenix: Content & Lore Validation Specs

This document defines the validation schemas and logical check gates for challenge levels, dialogue strings, and lore timeline data.

---

## 1. Challenge & Level Validator

The Content Curator agent executes check scripts against JSON challenge templates:

### 1.1 JSON Schema Validation
- Ensures every puzzle configuration matches the master challenge schema (valid IDs, parameters, and keys).

### 1.2 Progression & Gating Verifications
- **Duplicate ID Check**: Audits the entire level directory to ensure no two challenges share the same unique identifier key.
- **Dependency Paths**: Verifies that all level unlock requirements map to existing target challenge nodes (preventing deadlocks in level selection).
- **Dialogue Mapping**: Scans the challenge configuration file for missing mentor hints or victory dialog text lines.
- **Reward Balance Check**: Verifies that standard clears do not award more than the maximum XP limit (500 XP per level), maintaining economic balances.

---

## 2. Lore & Narrative Validator

Ensures narrative consistency across all timeline documents:

### 2.1 Timeline Consistency
- Scans `docs/lore/lore.md` and `docs/eras/eras.md` to verify that chronological timeline events match exactly across files.
- **Broken Link Audits**: Crawls all markdown files to detect dead file schema references (broken `file:///` URLs).

### 2.2 Character & NPC Cross-Checks
- Verifies that all character keys referenced in story dialogue tables are defined inside `docs/characters/registry.md`.
- **Location Alignment**: Ensures that NPC locations are mapped to valid cities or landmarks registered in the universe encylopedias.
