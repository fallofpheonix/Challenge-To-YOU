---
Status: Active
Implementation: 100%
Confidence: Authoritative
---
# P-Script — Language Specification

A sandboxed, domain-specific scripting language designed for writing autonomous swarm agent behaviors.

## Overview
- **Interpreted**: Tree-walk interpreter operating directly on the AST.
- **Sandboxed**: No access to OS, file system, or network.
- **Bounded**: `while` loops have a hard limit of 100 iterations per tick to prevent infinite loops from halting the 10Hz engine.
- **Stateful**: Variables can be declared with `let` and reassigned.

## Syntax Example
```
fn main() {
    if (SENSE_BATTERY() < 25000000) {
        MOVE_TOWARDS_HOME()
    } else {
        if (SENSE_CARGO()) {
            DROP_RESOURCE()
            MOVE_TOWARDS_HOME()
        } else {
            HARVEST()
            if (SENSE_CARGO()) {
                MOVE_TOWARDS_HOME()
            } else {
                MOVE_TOWARDS_RESOURCE()
            }
        }
    }
}
```
