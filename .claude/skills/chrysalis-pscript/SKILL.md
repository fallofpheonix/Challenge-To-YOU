---
name: chrysalis-pscript
description: P-Script, the DSL that programs Chrysalis drone behavior — grammar, the Swarm API built-ins, and the lexer→parser→compiler→VM pipeline. Use when editing the language, its compiler/VM, or writing agent scripts (.ps).
---

# P-Script

Custom DSL for drone behavior. Compiled to bytecode by a dedicated compiler, run by a stack-based VM; a tree-walk interpreter is the fallback. Default script: `engine/core/scripts/agent.ps`.

## Pipeline (Go, `engine/core/pscript/`)
`token/` → `lexer/` → `parser/` (recursive-descent Pratt) → `ast/` → `vm/compiler.go` (AST→bytecode) → `vm/vm.go` (18 opcodes, stack-based). Fallback: `interpreter/`.
Compile once per load/hot-reload; VM executes per tick with zero allocation. **VM step limit: 1000 instructions/drone/tick** — prevents infinite loops.

> Note: `pscript/README.md` marks the VM "planned" but it is implemented (`vm/`, ADR-004). Trust the code.

## Grammar
- Keywords: `fn`, `let`, `if`, `else`, `while`, `return`, `true`, `false`
- Operators: `+ - * / < > <= >= == != = !`
- Entry point is `fn main() { ... }`, evaluated every tick per drone.

## Swarm API (built-ins)
Sensors: `SENSE_RESOURCE()`, `SENSE_HOME()`, `SENSE_CARGO()`, `SENSE_BATTERY()` (int64, ×10^6), `SENSE_TRUST()` (0-100), `SENSE_CORRUPTION()` (0-100), `SENSE_COMPROMISED()`, `SENSE_ALIEN_SIGNAL()`, `SENSE_SWARM_SIZE()`, `SENSE_COLONY_RESOURCES()`, `BROADCAST_VOTE()` (quorum).
Actuators: `HARVEST()`, `DROP_RESOURCE()`, `MOVE_RANDOM()`, `MOVE_TOWARDS_RESOURCE()`, `MOVE_TOWARDS_HOME()`.

Battery values are fixed-point (×10^6): `SENSE_BATTERY() < 25000000` means <25 units.

## Example
```
fn main() {
    if (SENSE_BATTERY() < 25000000) { MOVE_TOWARDS_HOME() }
    else {
        if (SENSE_CARGO()) { DROP_RESOURCE() MOVE_TOWARDS_HOME() }
        else {
            HARVEST()
            if (SENSE_CARGO()) { MOVE_TOWARDS_HOME() } else { MOVE_TOWARDS_RESOURCE() }
        }
    }
}
```

When changing the language, update `token → lexer → parser → ast → compiler → vm` together and extend `vm/vm_test.go`. See [[chrysalis-simulation]], [[chrysalis-engine-architecture]].
