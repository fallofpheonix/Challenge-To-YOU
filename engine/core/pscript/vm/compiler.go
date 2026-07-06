package vm

import (
	"chrysalis-engine/core/pscript/ast"
	"fmt"
)

// Compiler translates a P-Script AST into a bytecode Program.
type Compiler struct {
	constants []int64
	code      []Instruction
	vars      []string
	varIndex  map[string]int
	funcs     []string
	funcIndex map[string]int
	errors    []string
}

// NewCompiler creates a fresh compiler instance.
func NewCompiler() *Compiler {
	return &Compiler{
		varIndex:  map[string]int{},
		funcIndex: map[string]int{},
	}
}

// Compile walks the AST and produces a Program. Returns nil on errors.
func (c *Compiler) Compile(program *ast.Program) *Program {
	for _, stmt := range program.Statements {
		c.compileStatement(stmt)
	}
	c.emit(OpDone, 0)

	if len(c.errors) > 0 {
		return nil
	}

	return &Program{
		Instructions: c.code,
		Constants:    c.constants,
		FuncNames:    c.funcs,
		FuncIndex:    c.funcIndex,
		MaxVars:      len(c.vars),
	}
}

// Errors returns compile-time errors.
func (c *Compiler) Errors() []string {
	return c.errors
}

func (c *Compiler) emit(op OpCode, arg int16) int {
	pos := len(c.code)
	c.code = append(c.code, Instruction{Op: op, Arg: arg})
	return pos
}

func (c *Compiler) patchJump(pos int) {
	offset := len(c.code) - pos - 1
	c.code[pos].Arg = int16(offset)
}

func (c *Compiler) addConstant(val int64) int16 {
	for i, v := range c.constants {
		if v == val {
			return int16(i)
		}
	}
	idx := int16(len(c.constants))
	c.constants = append(c.constants, val)
	return idx
}

func (c *Compiler) defineVar(name string) int {
	if idx, ok := c.varIndex[name]; ok {
		return idx
	}
	idx := len(c.vars)
	c.vars = append(c.vars, name)
	c.varIndex[name] = idx
	return idx
}

func (c *Compiler) getVar(name string) (int, bool) {
	idx, ok := c.varIndex[name]
	return idx, ok
}

func (c *Compiler) addFunc(name string) int16 {
	if idx, ok := c.funcIndex[name]; ok {
		return int16(idx)
	}
	idx := int16(len(c.funcs))
	c.funcs = append(c.funcs, name)
	c.funcIndex[name] = int(idx)
	return idx
}

func (c *Compiler) compileStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.FunctionDeclaration:
		c.compileFunctionDeclaration(s)
	case *ast.ExpressionStatement:
		c.compileExpression(s.Expression)
		c.emit(OpPop, 0) // discard result
	case *ast.IfStatement:
		c.compileIfStatement(s)
	case *ast.WhileStatement:
		c.compileWhileStatement(s)
	case *ast.LetStatement:
		c.compileLetStatement(s)
	case *ast.ReturnStatement:
		if s.ReturnValue != nil {
			c.compileExpression(s.ReturnValue)
		}
	case *ast.BlockStatement:
		for _, stmt := range s.Statements {
			c.compileStatement(stmt)
		}
	}
}

func (c *Compiler) compileFunctionDeclaration(node *ast.FunctionDeclaration) {
	if node.Name.Value == "main" {
		c.compileStatement(node.Body)
	}
}

func (c *Compiler) compileIfStatement(node *ast.IfStatement) {
	c.compileExpression(node.Condition)

	// JumpIf: pops condition; if false, jumps past consequence
	jumpIfPos := c.emit(OpJumpIf, 0)

	c.compileStatement(node.Consequence)

	if node.Alternative != nil {
		// Jump over alternative when consequence was taken
		jumpPos := c.emit(OpJump, 0)
		c.patchJump(jumpIfPos)
		c.compileStatement(node.Alternative)
		c.patchJump(jumpPos)
	} else {
		c.patchJump(jumpIfPos)
	}
}

func (c *Compiler) compileWhileStatement(node *ast.WhileStatement) {
	loopStart := len(c.code)

	c.compileExpression(node.Condition)

	// JumpIf: if condition false, exit loop
	jumpIfPos := c.emit(OpJumpIf, 0)

	c.compileStatement(node.Body)

	// Jump back to loop start
	c.emit(OpJump, int16(loopStart-len(c.code)-1))

	// Patch the exit jump
	c.patchJump(jumpIfPos)
}

func (c *Compiler) compileLetStatement(node *ast.LetStatement) {
	idx := c.defineVar(node.Name.Value)
	if node.Value != nil {
		c.compileExpression(node.Value)
	} else {
		c.emit(OpConst, c.addConstant(0))
	}
	c.emit(OpStore, int16(idx))
}

func (c *Compiler) compileExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		c.emit(OpConst, c.addConstant(e.Value))

	case *ast.Identifier:
		idx, ok := c.getVar(e.Value)
		if ok {
			c.emit(OpLoad, int16(idx))
		} else {
			c.emit(OpConst, c.addConstant(0))
		}

	case *ast.CallExpression:
		c.compileCallExpression(e)

	case *ast.InfixExpression:
		c.compileInfixExpression(e)

	case *ast.PrefixExpression:
		c.compilePrefixExpression(e)
	}
}

func (c *Compiler) compileCallExpression(node *ast.CallExpression) {
	fnIdx := c.addFunc(node.Function)
	c.emit(OpCall, fnIdx)
}

func (c *Compiler) compileInfixExpression(node *ast.InfixExpression) {
	// Handle reassignment: ident = expr
	if node.Operator == "=" {
		if ident, ok := node.Left.(*ast.Identifier); ok {
			idx, exists := c.getVar(ident.Value)
			if !exists {
				idx = c.defineVar(ident.Value)
			}
			c.compileExpression(node.Right)
			c.emit(OpStore, int16(idx))
			return
		}
	}

	c.compileExpression(node.Left)
	c.compileExpression(node.Right)

	switch node.Operator {
	case "+":
		c.emit(OpAdd, 0)
	case "-":
		c.emit(OpSub, 0)
	case "*":
		c.emit(OpMul, 0)
	case "/":
		c.emit(OpDiv, 0)
	case "<":
		c.emit(OpLt, 0)
	case ">":
		c.emit(OpGt, 0)
	case "<=":
		c.emit(OpLte, 0)
	case ">=":
		c.emit(OpGte, 0)
	case "==":
		c.emit(OpEq, 0)
	case "!=":
		c.emit(OpNeq, 0)
	default:
		c.errors = append(c.errors, fmt.Sprintf("unknown operator: %s", node.Operator))
	}
}

func (c *Compiler) compilePrefixExpression(node *ast.PrefixExpression) {
	c.compileExpression(node.Right)
	switch node.Operator {
	case "-":
		c.emit(OpNeg, 0)
	case "!":
		c.emit(OpNot, 0)
	}
}
