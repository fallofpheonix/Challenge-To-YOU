# Challenge Bundle Specification v1.0

**Status:** Draft
**Date:** 2026-07-13
**Authors:** Phoenix Ecosystem
**Purpose:** Define the immutable, versioned contract for challenge scenarios shared across repositories.

---

## 1. Goals and Non-Goals

### Goals

- Define a **single, canonical format** for challenge definitions that all repositories consume.
- Enable **deterministic replay** of challenge generation (same inputs → same challenge).
- Enable **deterministic recording** of challenge playthroughs (same inputs → same outcome).
- Create a **durable, versioned artifact** that survives engine rewrites, renderer changes, and AI improvements.
- Support **regression testing** via golden bundles with expected hashes.
- Allow **cross-repository consumption** without reverse-engineering internal data structures.

### Non-Goals

- Replacing the internal runtime representation used by the Go backend.
- Defining the game client's rendering or UI format.
- Specifying the AI hint system or Ollama integration.
- Standardizing the database schema for player profiles.
- Supporting real-time multiplayer replay synchronization (future work).

---

## 2. Determinism Guarantees

The bundle format guarantees **generation determinism**, not **execution determinism**.

### 2.1 Generation Determinism (Guaranteed)

Given identical inputs, the challenge generation pipeline produces identical outputs.

```
seed + luck + paradigm + bundle_version → identical ChallengeBundle
```

This is achievable because:
- The procedural generator uses `math/rand` with a fixed seed.
- The hydrator uses `math/rand` with a derived seed.
- All vocabulary pools and flaw templates are versioned within the bundle.

### 2.2 Execution Determinism (Not Guaranteed in v1.0)

Challenge playthroughs involve real-time elements (wall clock, AI taunts, tick timing) that break determinism.

**v1.0 approach:** Record player actions as a sequence of discrete events with logical timestamps (tick numbers), not wall-clock time. This enables:
- Replaying the same action sequence against the same challenge.
- Comparing outcomes across repositories.
- Detecting regressions when the engine changes.

**Future work:** A deterministic execution mode (fixed tick rate, no AI, no wall clock) that guarantees bit-for-bit replay.

### 2.3 What "Deterministic" Means in This Spec

| Component | Deterministic? | Mechanism |
|-----------|----------------|-----------|
| Challenge generation | Yes | Seed-based RNG |
| Flaw injection | Yes | Seed-based RNG |
| Flaw ordering | Yes | Seed-based shuffle |
| Player actions | Recorded | Event sequence with tick stamps |
| Outcome evaluation | Yes | Same state + same events → same result |
| Vigilance decay | No (v1.0) | Wall-clock based |
| AI taunts | No | External service |
| Rendering | No | Client-specific |

---

## 3. Versioning Strategy

### 3.1 Semantic Versioning

The bundle format uses semantic versioning: `MAJOR.MINOR.PATCH`

| Change type | Version bump | Example |
|-------------|--------------|---------|
| Breaking change to schema | MAJOR | 1.0.0 → 2.0.0 |
| Backward-compatible addition | MINOR | 1.0.0 → 1.1.0 |
| Bug fix or clarification | PATCH | 1.0.0 → 1.0.1 |

### 3.2 Version Fields

Every bundle contains three version fields:

| Field | Purpose |
|-------|---------|
| `spec_version` | Version of this specification the bundle conforms to |
| `format_version` | Version of the bundle directory structure and manifest schema |
| `content_version` | Version of the challenge content (vocabulary, flaws, win conditions) |

### 3.3 Version Pinning

Consumers must pin to a `spec_version` major version. A v1.0 consumer must accept any v1.x bundle without errors.

---

## 4. Compatibility Policy

### 4.1 Forward Compatibility

A v1.0 consumer must handle v1.x bundles by ignoring unknown fields in the manifest and metadata.

### 4.2 Backward Compatibility

A v1.x consumer must handle v1.0 bundles by using default values for any fields added in v1.x.

### 4.3 Breaking Changes

A breaking change (MAJOR version bump) is required when:
- A required field is removed or renamed.
- The meaning of a field changes.
- The directory structure changes.
- The replay format binary encoding changes.

### 4.4 Deprecated Fields

Deprecated fields must:
- Be documented in the spec with a removal version.
- Be ignored by consumers (not cause errors).
- Be preserved in producers for at least one MAJOR version after deprecation.

---

## 5. Directory Layout

A challenge bundle is a directory (or archive) with the following structure:

```
bundle/
├── manifest.json           # Required: bundle metadata and content index
├── challenge.json          # Required: challenge definition
├── world.json              # Required: world/room definitions (if applicable)
├── replay/                 # Optional: recorded playthroughs
│   ├── replay_001.bin      # Binary replay data
│   ├── replay_001.json     # Replay metadata
│   └── ...
├── assets/                 # Optional: scenario-specific assets
│   ├── dialogue.json       # NPC dialogue (if applicable)
│   ├── npcs.json           # NPC definitions (if applicable)
│   └── ...
├── tests/                  # Required: validation data
│   ├── expected_hash.json  # Expected hashes for regression testing
│   ├── golden/             # Golden test bundles
│   │   └── golden_001.json
│   └── test_vectors.json   # Deterministic test cases
├── schema/                 # Optional: embedded schemas for validation
│   └── challenge.schema.json
└── README.md               # Optional: human-readable documentation
```

### 5.1 Archive Formats

Bundles may be distributed as:
- **Directory:** For development and testing.
- **ZIP archive:** For distribution (file extension `.ctb` — Challenge To Bundle).
- **TAR.GZ archive:** For large bundles with many assets.

The archive must preserve the directory structure relative to the archive root.

---

## 6. Bundle Manifest Schema

The `manifest.json` file is the entry point for all bundle consumers.

```json
{
  "spec_version": "1.0.0",
  "format_version": "1.0.0",
  "content_version": "1.0.0",
  "bundle_id": "cty-magitech-001",
  "bundle_name": "Magitech Challenge 001: The Breach",
  "description": "A beginner-level breach challenge in the Medieval Magitech era.",
  "era": "MAGITECH",
  "paradigm": "MAGITECH",
  "tier": 1,
  "difficulty": 0.3,
  "modes": ["ARCHITECT", "GHOST", "SABOTEUR"],
  "estimated_duration_seconds": 900,
  "required_engine_version": ">=1.0.0",
  "authors": ["Phoenix Ecosystem"],
  "license": "MIT",
  "created_at": "2026-07-13T00:00:00Z",
  "updated_at": "2026-07-13T00:00:00Z",
  "tags": ["beginner", "breach", "magitech"],
  "content": {
    "challenge": "challenge.json",
    "world": "world.json",
    "replays": ["replay/replay_001.bin"],
    "assets": ["assets/dialogue.json"],
    "tests": ["tests/expected_hash.json"]
  },
  "generation": {
    "seed": 42,
    "luck": 0.5,
    "paradigm": "MAGITECH",
    "generator_version": "1.0.0",
    "deterministic": true
  },
  "checksums": {
    "challenge": "sha256:abcdef1234567890...",
    "world": "sha256:abcdef1234567890...",
    "bundle": "sha256:abcdef1234567890..."
  },
  "metadata": {
    "schema_version": "2.0.0",
    "replayable": true,
    "variant_count": 1,
    "seed_sensitive": true,
    "luck_sensitive": true
  }
}
```

### 6.1 Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `spec_version` | string | Specification version |
| `format_version` | string | Bundle format version |
| `content_version` | string | Content version |
| `bundle_id` | string | Unique identifier (UUID or slug) |
| `bundle_name` | string | Human-readable name |
| `era` | string | Era enum: `MAGITECH`, `CYBERPUNK`, `COSMIC` |
| `paradigm` | string | Paradigm enum (same as era for v1) |
| `content` | object | References to bundle files |
| `checksums` | object | SHA-256 checksums for integrity validation |

### 6.2 Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Human-readable description |
| `tier` | integer | Difficulty tier (1-5) |
| `difficulty` | float | Difficulty rating (0.0-1.0) |
| `modes` | array | Supported gameplay modes |
| `estimated_duration_seconds` | integer | Expected completion time |
| `required_engine_version` | string | Semver range of compatible engines |
| `authors` | array | Content authors |
| `license` | string | Content license |
| `tags` | array | Searchable tags |
| `generation` | object | Procedural generation parameters |
| `metadata` | object | Extended metadata |

---

## 7. Challenge Definition Schema

The `challenge.json` file uses the existing v2.0.0 schema from `content/schema/challenge.schema.json`, with the following additions for bundle compatibility.

### 7.1 Bundle-Compatible Fields

```json
{
  "schema_version": "2.0.0",
  "id": "magitech_001",
  "name": "The Breach",
  "paradigm": "MAGITECH",
  "description": "Infiltrate the ancient firewall...",
  "skill_type": "breach",
  "logos_token": "BREACH_RUNE",
  "difficulty": 0.3,
  "modes": ["ARCHITECT", "GHOST", "SABOTEUR"],
  "initial_state": {
    "firewall_integrity": 100,
    "access_level": 0,
    "alert_status": "NORMAL"
  },
  "flaws": [
    {
      "id": "flaw_001",
      "trigger_event": "EXECUTE_RUNE",
      "name": "Overflow the Buffer",
      "flavor_text": "The ancient rune overflows...",
      "conditions": [
        {
          "state_key": "firewall_integrity",
          "operator": "EQUALS",
          "value": 100
        }
      ],
      "mutations": [
        {
          "target_state_key": "firewall_integrity",
          "mutation_value": -25
        }
      ],
      "fallback_mutations": [
        {
          "target_state_key": "alert_status",
          "mutation_value": "ELEVATED"
        }
      ]
    }
  ],
  "win_condition": {
    "target_state_key": "access_level",
    "expected_value": 3
  },
  "template_code": "# Your breach script here\n",
  "validation_script": "",
  "expected_answer": "",
  "bundle_metadata": {
    "bundle_id": "cty-magitech-001",
    "bundle_version": "1.0.0",
    "content_hash": "sha256:abcdef1234567890..."
  }
}
```

### 7.2 New Field: `bundle_metadata`

| Field | Type | Description |
|-------|------|-------------|
| `bundle_metadata.bundle_id` | string | References the containing bundle |
| `bundle_metadata.bundle_version` | string | Version of the bundle |
| `bundle_metadata.content_hash` | string | SHA-256 hash of the challenge definition |

This field is added by the bundle packaging tool, not by authors.

---

## 8. Replay Format Contract

### 8.1 Replay File Structure

A replay file (`replay_001.bin`) is a binary file with the following structure:

```
┌─────────────────────────────────────────────────────────┐
│ Header (fixed, 64 bytes)                                │
├─────────────────────────────────────────────────────────┤
│ Event Count (4 bytes, uint32, big-endian)               │
├─────────────────────────────────────────────────────────┤
│ Events (variable length)                                │
│   ┌───────────────────────────────────────────────────┐ │
│   │ Event 0                                           │ │
│   │   Tick: 4 bytes (uint32)                          │ │
│   │   Type: 2 bytes (uint16)                          │ │
│   │   Payload Length: 4 bytes (uint32)                 │ │
│   │   Payload: N bytes (JSON-encoded)                 │ │
│   └───────────────────────────────────────────────────┘ │
│   ┌───────────────────────────────────────────────────┐ │
│   │ Event 1                                           │ │
│   │   ...                                             │ │
│   └───────────────────────────────────────────────────┘ │
│   ...                                                   │
├─────────────────────────────────────────────────────────┤
│ Footer (32 bytes)                                       │
│   Event Hash: 32 bytes (SHA-256 of all events)          │
└─────────────────────────────────────────────────────────┘
```

### 8.2 Header Format (64 bytes)

| Offset | Size | Field | Description |
|--------|------|-------|-------------|
| 0 | 4 | Magic | `0x43545942` ("CTYB") |
| 4 | 2 | Version | `0x0100` (v1.0) |
| 6 | 2 | Flags | Reserved (0x0000) |
| 8 | 4 | Bundle ID Length | Length of bundle_id string |
| 12 | 32 | Bundle ID | Bundle ID (UTF-8, zero-padded) |
| 44 | 4 | Seed | RNG seed for challenge generation |
| 48 | 4 | Luck | Luck value as uint32 (float × 10000) |
| 52 | 4 | Created At | Unix timestamp |
| 56 | 4 | Player ID | Player identifier |
| 60 | 4 | Reserved | Zero |

### 8.3 Event Types

| Type Code | Name | Payload |
|-----------|------|---------|
| 0x0001 | CHALLENGE_LOADED | `{challenge_id, seed, luck}` |
| 0x0002 | ACTION_EXECUTED | `{action_type, target, parameters}` |
| 0x0003 | STATE_CHANGED | `{key, old_value, new_value}` |
| 0x0004 | FLAW_TRIGGERED | `{flaw_id, conditions_met}` |
| 0x0005 | GLITCH_OCCURRED | `{glitch_id, effects}` |
| 0x0006 | WIN_CONDITION_CHECK | `{state_key, expected, actual, passed}` |
| 0x0007 | VIGILANCE_TICK | `{level, entropy}` |
| 0x0008 | AI_TAUNT | `{message, source}` |
| 0x0009 | PLAYER_INPUT | `{command, raw_input}` |
| 0x000A | SESSION_END | `{outcome, final_state}` |
| 0x00FF | CUSTOM | `{type, data}` |

### 8.4 Event Payload Encoding

All event payloads are JSON-encoded strings, allowing cross-language consumption.

Example:
```json
{
  "action_type": "EXECUTE_RUNE",
  "target": "overflow_buffer",
  "parameters": {
    "rune_id": "BREACH_RUNE",
    "target_address": "0xDEADBEEF"
  }
}
```

### 8.5 Logical Timestamps

Events use **tick numbers** (uint32), not wall-clock time. The tick rate is defined by the engine (default: 2 ticks/second for v1.0).

This enables:
- Consistent replay across different hardware.
- Comparison of replays from different repositories.
- Regression testing with fixed tick progression.

### 8.6 Footer

The footer contains a SHA-256 hash of all event data (excluding the header and footer themselves). This enables:
- Integrity validation.
- Fast comparison of replay equivalence.
- Detection of corruption.

---

## 9. World Hash Definition

### 9.1 Purpose

The world hash enables regression testing. If the same bundle produces a different world hash, something changed in the generation pipeline.

### 9.2 Hash Computation

The world hash is computed over the **canonical JSON representation** of the challenge definition, with the following rules:

1. Serialize `challenge.json` to JSON.
2. Sort all object keys lexicographically.
3. Remove all whitespace.
4. Compute SHA-256 of the resulting bytes.
5. Prefix with `sha256:`.

### 9.3 Canonical JSON Algorithm

```python
def canonical_json(obj):
    if isinstance(obj, dict):
        return "{" + ",".join(
            f'"{k}":{canonical_json(v)}'
            for k, v in sorted(obj.items())
        ) + "}"
    elif isinstance(obj, list):
        return "[" + ",".join(canonical_json(i) for i in obj) + "]"
    elif isinstance(obj, str):
        return json.dumps(obj)
    elif isinstance(obj, (int, float)):
        return json.dumps(obj)
    elif isinstance(obj, bool):
        return "true" if obj else "false"
    elif obj is None:
        return "null"
    else:
        raise TypeError(f"Unknown type: {type(obj)}")
```

### 9.4 Regression Test

```
Given:  bundle/challenge.json
When:   Compute canonical JSON → SHA-256
Then:   Result equals bundle/tests/expected_hash.json.world_hash
```

---

## 10. Asset References

### 10.1 Relative Paths

All file references in the manifest are relative to the bundle root directory.

Example:
```json
{
  "content": {
    "challenge": "challenge.json",
    "world": "world.json",
    "replays": ["replay/replay_001.bin"],
    "assets": ["assets/dialogue.json"]
  }
}
```

### 10.2 Asset Types

| Type | Description | Required |
|------|-------------|----------|
| Challenge | Challenge definition | Yes |
| World | World/room definitions | Yes (if applicable) |
| Replay | Recorded playthrough | No |
| Dialogue | NPC dialogue | No |
| Asset | Scenario-specific data | No |
| Test | Validation data | Yes |

### 10.3 External References

Assets must not reference files outside the bundle directory. All required data must be contained within the bundle.

---

## 11. Validation Rules

### 11.1 Bundle Validation

A valid bundle must:

1. Contain a `manifest.json` at the root.
2. Contain a `challenge.json` referenced by the manifest.
3. Have valid JSON in all referenced files.
4. Have matching checksums for all referenced files.
5. Have a `challenge.json` that validates against the challenge schema.
6. Contain a `tests/expected_hash.json` with a valid world hash.

### 11.2 Replay Validation

A valid replay must:

1. Start with the magic bytes `0x43545942` ("CTYB").
2. Have a valid header (64 bytes).
3. Contain the correct number of events.
4. Have a footer with a matching SHA-256 hash.
5. Reference a valid bundle_id.

### 11.3 Error Handling

Validation errors must be reported with:

| Field | Description |
|-------|-------------|
| `error_type` | Category: `SCHEMA`, `CHECKSUM`, `MISSING`, `FORMAT`, `VERSION` |
| `file` | Path to the offending file |
| `field` | JSON path to the offending field (if applicable) |
| `message` | Human-readable error description |
| `severity` | `ERROR` (bundle invalid) or `WARNING` (bundle usable with caveats) |

---

## 12. Extension Mechanism

### 12.1 Custom Fields

Producers may add custom fields to any JSON object, prefixed with `x_`.

Example:
```json
{
  "id": "magitech_001",
  "x_my_custom_field": "value"
}
```

Consumers must ignore unknown fields (including `x_` prefixed fields).

### 12.2 Custom Event Types

Event types `0x0080`–`0x00FF` are reserved for custom use.

### 12.3 Custom Extensions

Future extensions must be proposed via the specification's change process and approved before implementation.

---

## 13. Security Considerations

### 13.1 Untrusted Bundles

Bundles may come from untrusted sources. Consumers must:

1. Validate checksums before loading.
2. Enforce size limits (see 13.2).
3. Reject bundles with invalid or missing checksums.
4. Not execute arbitrary code contained in replay payloads.

### 13.2 Size Limits

| Resource | Maximum |
|----------|---------|
| Bundle archive | 100 MB |
| Individual JSON file | 10 MB |
| Replay file | 50 MB |
| Event payload | 1 KB |
| Event count per replay | 100,000 |

### 13.3 Checksum Validation

All checksums use SHA-256 with the `sha256:` prefix.

Consumers must verify:
1. `manifest.checksums.challenge` matches `sha256(challenge.json)`.
2. `manifest.checksums.world` matches `sha256(world.json)`.
3. `manifest.checksums.bundle` matches `sha256(manifest.json + challenge.json + world.json)`.
4. Replay footer hash matches the SHA-256 of all events.

---

## 14. Test Vectors and Golden Bundles

### 14.1 Test Vector Format

The `tests/test_vectors.json` file defines deterministic test cases:

```json
{
  "vectors": [
    {
      "name": "generation_determinism",
      "description": "Same seed produces same challenge",
      "inputs": {
        "seed": 42,
        "luck": 0.5,
        "paradigm": "MAGITECH"
      },
      "expected": {
        "challenge_id": "magitech_001",
        "world_hash": "sha256:abcdef1234567890...",
        "flaw_count": 5,
        "win_condition_key": "access_level"
      }
    },
    {
      "name": "replay_determinism",
      "description": "Same replay produces same outcome",
      "inputs": {
        "bundle_id": "cty-magitech-001",
        "replay": "replay_001.bin"
      },
      "expected": {
        "final_state_hash": "sha256:abcdef1234567890...",
        "outcome": "WIN"
      }
    }
  ]
}
```

### 14.2 Golden Bundle

A golden bundle is a reference bundle used for regression testing. It must:

1. Be included in the repository under `tests/golden/`.
2. Have a fixed, known content hash.
3. Be updated only when the specification changes.
4. Be validated against all test vectors.

### 14.3 Test Runner

A test runner must:

1. Load the bundle.
2. Validate all checksums.
3. Execute all test vectors.
4. Compare results to expected values.
5. Report pass/fail with detailed diagnostics.

---

## 15. Implementation Notes

### 15.1 Serialization

All JSON serialization must use sorted keys and no whitespace for hash computation. Runtime serialization may use pretty-printing.

### 15.2 File Encoding

- JSON files: UTF-8 without BOM.
- Binary files: Big-endian byte order.
- Checksums: Lowercase hexadecimal with `sha256:` prefix.

### 15.3 Character Set

Bundle IDs and file names must use only:
- Alphanumeric characters (`a-z`, `A-Z`, `0-9`)
- Hyphens (`-`)
- Underscores (`_`)
- Dots (`.`)

---

## 16. Migration from Existing Formats

### 16.1 Current State

The CTY codebase currently has:

| Format | Location | Status |
|--------|----------|--------|
| Runtime challenge JSON | `backend/challenges/` | Active, minimal fields |
| Authoring v2 JSON | `data/challenges/` | Alternative format, richer |
| Content schema | `content/schema/` | Target format, comprehensive |

### 16.2 Migration Path

1. **v1.0 (current):** Bundle spec defines the canonical format.
2. **v1.1:** Add migration tool to convert existing challenges to bundles.
3. **v2.0:** Deprecate legacy formats, enforce bundle format.

### 16.3 Backward Compatibility

During the migration period:
- The Go backend must accept both legacy and bundle formats.
- The bundle format is preferred when available.
- Legacy formats are converted to bundles at load time (not at rest).

---

## 17. Repository Consumption

### 17.1 Challenge To YOU

| Use case | How |
|----------|-----|
| Load challenge | Read `challenge.json` from bundle |
| Play challenge | Load bundle, initialize fabric, process events |
| Record replay | Write events to replay binary format |
| Validate bundle | Check checksums, verify schema |
| Regression test | Compare world hash against expected |

### 17.2 Phoenix Champions

| Use case | How |
|----------|-----|
| Load AI benchmark | Read bundle, extract challenge and win condition |
| Run AI agent | Initialize fabric from challenge, process agent actions |
| Compare results | Compare replay events against human replays |
| Train on replays | Load replay files, extract state transitions |

### 17.3 PhoenixVirtualizer

| Use case | How |
|----------|-----|
| Visualize challenge | Parse challenge.json, render flaw graph |
| Analyze architecture | Extract dependency relationships from flaw conditions |
| Generate reports | Compute complexity metrics from challenge structure |

### 17.4 LLM Curriculum

| Use case | How |
|----------|-----|
| Generate lessons | Read challenge description, extract teaching points |
| Create tutorials | Use challenge structure to explain concepts |
| Evaluate comprehension | Compare LLM explanations against challenge metadata |

---

## 18. Open Questions

These are resolved before v1.0 is finalized:

| Question | Status | Resolution |
|----------|--------|------------|
| Should replays be JSON or binary? | Resolved | Binary for efficiency, JSON metadata for cross-language access |
| Should the bundle include the engine? | Resolved | No, bundles are data-only |
| Should the bundle be self-contained? | Resolved | Yes, all data must be inside the bundle |
| Should we support encrypted bundles? | Deferred | Not in v1.0 |
| Should we support signed bundles? | Deferred | Not in v1.0 |

---

## 19. Change History

| Version | Date | Change |
|---------|------|--------|
| 1.0.0 | 2026-07-13 | Initial specification |

---

## 20. References

| Document | Location |
|----------|----------|
| Challenge Schema (v2.0.0) | `content/schema/challenge.schema.json` |
| Mission Schema (v1.0.0) | `content/schema/mission.schema.json` |
| Analytics Schema | `content/analytics/analytics_schema.json` |
| CTY Architecture | `docs/ARCHITECTURE.md` |
| Go Backend Types | `backend/internal/engine/matrix.go` |

---

*End of specification.*
