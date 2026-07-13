# Challenge Bundle Example

This is an example bundle conforming to the Challenge Bundle Specification v1.0.

## Structure

```
example-bundle/
├── manifest.json           # Bundle metadata and content index
├── challenge.json          # Challenge definition (v2.0.0 format)
├── world.json              # World/room definitions
├── replay/                 # Recorded playthroughs (empty in example)
├── assets/                 # Scenario-specific assets (empty in example)
├── tests/
│   ├── expected_hash.json  # Expected hashes for regression testing
│   ├── test_vectors.json   # Deterministic test cases
│   └── golden/
│       └── golden_001.json # Golden test bundle
├── schema/                 # Embedded schemas (empty in example)
└── README.md               # This file
```

## Usage

This example bundle demonstrates:

1. **Manifest format** — How to describe a bundle's contents and metadata.
2. **Challenge definition** — A complete challenge with flaws, conditions, and mutations.
3. **World definition** — Rooms, exits, and objects for spatial navigation.
4. **Test vectors** — Deterministic test cases for validation.
5. **Golden bundle** — Reference bundle for regression testing.

## Validation

To validate this bundle:

```bash
# Check manifest structure
cat manifest.json | python -m json.tool

# Verify challenge schema
cat challenge.json | python -m json.tool

# Compute world hash (after implementing canonical JSON)
python -c "import json; print(json.dumps(json.load(open('challenge.json')), sort_keys=True, separators=(',', ':')))"
```

## Migration

This bundle format is designed to replace the existing challenge formats:

| Existing Format | Location | Migration |
|-----------------|----------|-----------|
| Runtime JSON | `backend/challenges/` | Wrap in bundle structure |
| Authoring v2 | `data/challenges/` | Add manifest and metadata |
| Content schema | `content/schema/` | Already compatible |

## Next Steps

1. Implement canonical JSON hashing in Go.
2. Implement bundle loader in the Go backend.
3. Implement replay recorder.
4. Create golden bundles from existing challenges.
5. Add bundle validation to CI/CD pipeline.
