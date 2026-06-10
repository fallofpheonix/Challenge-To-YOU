package ast

import (
	"chrysalis-engine/core/pscript/token"
)

// Node represents any element within the P-Script structural tree
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement represents an execution boundary that performs mutations without returning values
type Statement interface {
	Node
	statementNode()
}

// Expression represents an operation that resolves into data during execution
type Expression interface {
	Node
	expressionNode()
}

// Program acts as the root container node for the compiled bytecode sequence
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out string
	for _, s := range p.Statements {
		out += s.String()
	}
	return out
}

// LetStatement handles variable initialization: let battery = 100;
type LetStatement struct {
	Token token.Token // token.LET
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out string
	out += ls.TokenLiteral() + " " + ls.Name.String() + " = "
	if ls.Value != nil {
		out += ls.Value.String()
	}
	out += ";"
	return out
}

// Identifier maps variable names or function hooks
type Identifier struct {
	Token token.Token // token.IDENT
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral wraps fixed-point values for scaling math operations
type IntegerLiteral struct {
	Token token.Token // token.INT
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FunctionDeclaration handles isolated agent sub-protocols: fn main() { ... }
type FunctionDeclaration struct {
	Token      token.Token // token.FN
	Name       *Identifier
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fd *FunctionDeclaration) statementNode()       {}
func (fd *FunctionDeclaration) TokenLiteral() string { return fd.Token.Literal }
func (fd *FunctionDeclaration) String() string {
	var out string
	out += fd.TokenLiteral() + " " + fd.Name.String() + "("
	for idx, p := range fd.Parameters {
		out += p.String()
		if idx < len(fd.Parameters)-1 {
			out += ", "
		}
	}
	out += ") " + fd.Body.String()
	return out
}

// BlockStatement encapsulates consecutive scope executions inside braces
type BlockStatement struct {
	Token      token.Token // token.LBRACE
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out string
	out += "{ "
	for _, s := range bs.Statements {
		out += s.String() + " "
	}
	out += "}"
	return out
}

// CallExpression acts as our sensor and utility pipeline execution hook: HARVEST()
type CallExpression struct {
	Token     token.Token // token.IDENT
	Function  string      // Function name
	Arguments []Expression
}

func (ce *CallExpression) expressionNode() {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out string
	out += ce.Function + "("
	for idx, a := range ce.Arguments {
		out += a.String()
		if idx < len(ce.Arguments)-1 {
			out += ", "
		}
	}
	out += ")"
	return out
}

// ExpressionStatement permits isolated expressions to exist sequentially in blocks
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String() + ";"
	}
	return ""
}

// IfStatement structures our branching environmental logic pathing
type IfStatement struct {
	Token       token.Token // token.IF
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string {
	var out string
	out += "if " + is.Condition.String() + " " + is.Consequence.String()
	if is.Alternative != nil {
		out += "else " + is.Alternative.String()
	}
	return out
}

// WhileStatement binds our iterative telemetry polling loops
type WhileStatement struct {
	Token     token.Token // token.WHILE
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	return "while " + ws.Condition.String() + " " + ws.Body.String()
}

// InfixExpression represents operations like 5 + 5 or battery < 20
type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}
