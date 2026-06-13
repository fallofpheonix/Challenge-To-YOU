---
Status: Active
Implementation: 100%
Confidence: Authoritative
---
# P-Script — Grammar Rules

Formal language grammar implemented via a recursive-descent Pratt parser.

## Syntax (EBNF)
```ebnf
Program             ::= Statement*
Statement           ::= LetStatement | ReturnStatement | ExpressionStatement | BlockStatement | IfStatement | WhileStatement
LetStatement        ::= "let" Identifier "=" Expression
ReturnStatement     ::= "return" Expression?
ExpressionStatement ::= Expression
BlockStatement      ::= "{" Statement* "}"
IfStatement         ::= "if" "(" Expression ")" BlockStatement ("else" BlockStatement)?
WhileStatement      ::= "while" "(" Expression ")" BlockStatement

Expression          ::= Identifier | Integer | Boolean | PrefixExpression | InfixExpression | CallExpression
PrefixExpression    ::= ("-" | "!") Expression
InfixExpression     ::= Expression ("+" | "-" | "*" | "/" | "<" | ">" | "<=" | ">=" | "==" | "!=") Expression
CallExpression      ::= Identifier "(" (Expression ("," Expression)*)? ")"

Boolean             ::= "true" | "false"
```
