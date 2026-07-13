#!/usr/bin/env python3
"""
Canonical JSON hasher for Challenge Bundle Specification v1.0.

Computes deterministic SHA-256 hashes of JSON files for regression testing.

Usage:
    python hash_utils.py <file.json>
    python hash_utils.py --verify <file.json> <expected_hash>
    python hash_utils.py --bundle <bundle_dir>
"""

import json
import hashlib
import sys
import os
from pathlib import Path


def canonical_json(obj):
    """
    Convert a Python object to canonical JSON string.
    
    Rules:
    - Sort all object keys lexicographically
    - Remove all whitespace
    - Use compact separators (no spaces)
    """
    if isinstance(obj, dict):
        items = sorted(obj.items())
        return "{" + ",".join(
            f'"{k}":{canonical_json(v)}'
            for k, v in items
        ) + "}"
    elif isinstance(obj, list):
        return "[" + ",".join(canonical_json(i) for i in obj) + "]"
    elif isinstance(obj, str):
        # Escape special characters and use double quotes
        escaped = obj.replace('\\', '\\\\').replace('"', '\\"').replace('\n', '\\n').replace('\r', '\\r').replace('\t', '\\t')
        return f'"{escaped}"'
    elif isinstance(obj, bool):
        return "true" if obj else "false"
    elif obj is None:
        return "null"
    elif isinstance(obj, int):
        return str(obj)
    elif isinstance(obj, float):
        # Ensure consistent float representation
        return f"{obj}"
    else:
        raise TypeError(f"Unknown type: {type(obj)}")


def compute_hash(data):
    """
    Compute SHA-256 hash of data with 'sha256:' prefix.
    """
    if isinstance(data, str):
        data = data.encode('utf-8')
    return "sha256:" + hashlib.sha256(data).hexdigest()


def compute_file_hash(filepath):
    """
    Compute canonical JSON hash of a JSON file.
    """
    with open(filepath, 'r', encoding='utf-8') as f:
        obj = json.load(f)
    canonical = canonical_json(obj)
    return compute_hash(canonical)


def compute_bundle_hash(bundle_dir):
    """
    Compute bundle hash by concatenating canonical JSON of all files.
    
    The bundle hash is computed from the actual file contents, not from the
    checksums stored in the manifest. This avoids circular dependencies.
    """
    bundle_path = Path(bundle_dir)
    
    # Load and canonicalize each component
    with open(bundle_path / "challenge.json", 'r', encoding='utf-8') as f:
        challenge_obj = json.load(f)
    challenge_canonical = canonical_json(challenge_obj)
    
    with open(bundle_path / "world.json", 'r', encoding='utf-8') as f:
        world_obj = json.load(f)
    world_canonical = canonical_json(world_obj)
    
    # Load manifest without checksums for hashing
    with open(bundle_path / "manifest.json", 'r', encoding='utf-8') as f:
        manifest_obj = json.load(f)
    
    # Create a copy without checksums for hashing
    manifest_for_hash = {k: v for k, v in manifest_obj.items() if k != 'checksums'}
    manifest_canonical = canonical_json(manifest_for_hash)
    
    # Bundle hash is hash of concatenated canonical JSON
    combined = f"{manifest_canonical}|{challenge_canonical}|{world_canonical}"
    
    # Also compute individual hashes for reference
    challenge_hash = compute_hash(challenge_canonical)
    world_hash = compute_hash(world_canonical)
    
    return compute_hash(combined), challenge_hash, world_hash


def verify_hash(filepath, expected_hash):
    """
    Verify that a file's hash matches the expected value.
    """
    actual_hash = compute_file_hash(filepath)
    return actual_hash == expected_hash, actual_hash, expected_hash


def verify_bundle(bundle_dir):
    """
    Verify all checksums in a bundle manifest.
    """
    bundle_path = Path(bundle_dir)
    manifest_path = bundle_path / "manifest.json"
    
    with open(manifest_path, 'r', encoding='utf-8') as f:
        manifest = json.load(f)
    
    results = []
    
    # Verify challenge hash
    challenge_path = bundle_path / manifest['content']['challenge']
    if challenge_path.exists():
        expected = manifest['checksums']['challenge']
        valid, actual, _ = verify_hash(challenge_path, expected)
        results.append(('challenge', valid, actual, expected))
    
    # Verify world hash
    world_path = bundle_path / manifest['content']['world']
    if world_path.exists():
        expected = manifest['checksums']['world']
        valid, actual, _ = verify_hash(world_path, expected)
        results.append(('world', valid, actual, expected))
    
    # Verify bundle hash
    bundle_actual, _, _ = compute_bundle_hash(bundle_dir)
    bundle_expected = manifest['checksums']['bundle']
    results.append(('bundle', bundle_actual == bundle_expected, bundle_actual, bundle_expected))
    
    return results


def update_manifest_hashes(bundle_dir):
    """
    Recompute and update all checksums in manifest.json.
    """
    bundle_path = Path(bundle_dir)
    manifest_path = bundle_path / "manifest.json"
    
    with open(manifest_path, 'r', encoding='utf-8') as f:
        manifest = json.load(f)
    
    # Compute hashes
    challenge_path = bundle_path / manifest['content']['challenge']
    world_path = bundle_path / manifest['content']['world']
    
    manifest['checksums']['challenge'] = compute_file_hash(challenge_path)
    manifest['checksums']['world'] = compute_file_hash(world_path)
    
    bundle_hash, _, _ = compute_bundle_hash(bundle_dir)
    manifest['checksums']['bundle'] = bundle_hash
    
    # Write updated manifest
    with open(manifest_path, 'w', encoding='utf-8') as f:
        json.dump(manifest, f, indent=2)
    
    return manifest['checksums']


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        sys.exit(1)
    
    if sys.argv[1] == "--verify":
        if len(sys.argv) != 4:
            print("Usage: python hash_utils.py --verify <file.json> <expected_hash>")
            sys.exit(1)
        filepath = sys.argv[2]
        expected = sys.argv[3]
        valid, actual, _ = verify_hash(filepath, expected)
        if valid:
            print(f"✓ Hash matches: {actual}")
        else:
            print(f"✗ Hash mismatch:")
            print(f"  Expected: {expected}")
            print(f"  Actual:   {actual}")
            sys.exit(1)
    
    elif sys.argv[1] == "--bundle":
        if len(sys.argv) != 3:
            print("Usage: python hash_utils.py --bundle <bundle_dir>")
            sys.exit(1)
        bundle_dir = sys.argv[2]
        results = verify_bundle(bundle_dir)
        all_valid = True
        for name, valid, actual, expected in results:
            if valid:
                print(f"✓ {name}: {actual}")
            else:
                print(f"✗ {name}: MISMATCH")
                print(f"  Expected: {expected}")
                print(f"  Actual:   {actual}")
                all_valid = False
        if not all_valid:
            sys.exit(1)
    
    elif sys.argv[1] == "--update":
        if len(sys.argv) != 3:
            print("Usage: python hash_utils.py --update <bundle_dir>")
            sys.exit(1)
        bundle_dir = sys.argv[2]
        hashes = update_manifest_hashes(bundle_dir)
        print("Updated manifest.json:")
        for key, value in hashes.items():
            print(f"  {key}: {value}")
    
    else:
        filepath = sys.argv[1]
        if not os.path.exists(filepath):
            print(f"Error: File not found: {filepath}")
            sys.exit(1)
        h = compute_file_hash(filepath)
        print(h)


if __name__ == "__main__":
    main()
