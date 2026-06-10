---
Status: Planned
Implementation: 0%
Confidence: Conceptual
---
# P-Script — Grammar Rules

Formal language grammar.

## Syntax (EBNF)
```ebnf
Program ::= Statement*
Statement ::= LetStatement | FnDeclaration | ExpressionStatement
LetStatement ::= "let" Ident "=" Expression ";"
```
