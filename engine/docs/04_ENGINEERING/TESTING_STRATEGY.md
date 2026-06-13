# Testing Strategy

## Required Checks

```bash
cd engine
make check
make test-race
```

`make check` runs Go tests, `go vet`, and Godot headless validation.

## Current Coverage

- Lexer, parser, infix expressions, conditionals, and builtins.
- Seeded deterministic simulation equivalence.
- Hazard and fabrication execution in the production tick lifecycle.
- Inert-drone action rejection.
- Harvest, home navigation, and resource deposit.
- Default `agent.ps` economy loop.
- Race detector across all Go packages.

## Missing Coverage

- WebSocket reconnect and malformed-frame integration.
- Godot UI interaction.
- Hot-patch rollback integration.
- Long-run state hash and replay.
- Collision behavior; the system is absent.
- 10,000+ drone performance budgets.

New authoritative behavior requires a deterministic Go regression test.
