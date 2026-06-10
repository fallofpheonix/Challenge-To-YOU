# P-Script Subsystem

> **Role**: Autonomous Agent Logic Language

P-Script is the domain-specific language used to program agent behaviors in Project Chrysalis.

## Implementation Progress
- [x] **Lexer**: Scans text into tokens.
- [x] **Parser**: Builds the Abstract Syntax Tree (AST).
- [x] **Interpreter**: Executes the AST within the simulation context.
- [ ] **VM**: High-performance bytecode execution (Planned).

## Packages
- `token/`: Lexical token definitions.
- `lexer/`: Lexical scanner implementation.
- `ast/`: AST node definitions.
- `parser/`: Recursive-descent parser.
- `interpreter/`: AST walk execution engine.
