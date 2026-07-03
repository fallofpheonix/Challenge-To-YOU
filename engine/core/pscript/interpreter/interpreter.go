package interpreter

import (
	"chrysalis-engine/core/pscript/ast"
	"chrysalis-engine/core/simulation"
	"fmt"
	"strconv"
)

// BuiltinFn: A function defined in Go that can be called from P-Script.
// It receives the simulation engine and the index of the entity currently executing.
type BuiltinFn func(e *simulation.Engine, entityIndex int) interface{}

// traceBuilder accumulates DecisionSteps during a single Eval call.
// It is unexported; callers receive a completed DecisionFrame, never this struct.
type traceBuilder struct {
	steps []simulation.DecisionStep
}

func (b *traceBuilder) recordAction(name string, result interface{}) {
	b.steps = append(b.steps, simulation.DecisionStep{
		Kind:   "action",
		Name:   name,
		Result: formatResult(result),
	})
}

func (b *traceBuilder) recordCondition(name string, result interface{}, taken bool) {
	b.steps = append(b.steps, simulation.DecisionStep{
		Kind:   "condition",
		Name:   name,
		Result: formatResult(result),
		Taken:  taken,
	})
}

func formatResult(v interface{}) string {
	switch val := v.(type) {
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int64:
		return strconv.FormatInt(val, 10)
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Interpreter executes ASTs against the simulation engine.
type Interpreter struct {
	builtins  map[string]BuiltinFn
	variables map[string]interface{}
	builder   *traceBuilder // nil when not tracing
}

// New creates a new Interpreter with the given built-in functions.
func New(builtins map[string]BuiltinFn) *Interpreter {
	return &Interpreter{
		builtins:  builtins,
		variables: make(map[string]interface{}),
	}
}

// Eval walks the program and executes its statements for a specific entity.
func (i *Interpreter) Eval(program *ast.Program, e *simulation.Engine, entityIndex int) {
	clear(i.variables)
	for _, stmt := range program.Statements {
		i.evalStatement(stmt, e, entityIndex)
	}
}

// EvalTraced runs Eval and returns a completed, immutable DecisionFrame describing
// every condition and action the drone's policy executed. The builder is private;
// the returned frame is safe to share freely.
func (i *Interpreter) EvalTraced(program *ast.Program, e *simulation.Engine, entityIndex int, tick int64) *simulation.DecisionFrame {
	i.builder = &traceBuilder{steps: make([]simulation.DecisionStep, 0, 16)}
	i.Eval(program, e, entityIndex)
	frame := &simulation.DecisionFrame{
		DroneID: entityIndex,
		Tick:    tick,
		Steps:   i.builder.steps,
	}
	i.builder = nil
	return frame
}

func (i *Interpreter) evalStatement(stmt ast.Statement, e *simulation.Engine, entityIndex int) {
	switch s := stmt.(type) {
	case *ast.FunctionDeclaration:
		i.evalFunctionDeclaration(s, e, entityIndex)
	case *ast.ExpressionStatement:
		i.evalExpression(s.Expression, e, entityIndex)
	case *ast.IfStatement:
		i.evalIfStatement(s, e, entityIndex)
	case *ast.WhileStatement:
		i.evalWhileStatement(s, e, entityIndex)
	case *ast.LetStatement:
		i.evalLetStatement(s, e, entityIndex)
	case *ast.ReturnStatement:
		if s.ReturnValue != nil {
			i.evalExpression(s.ReturnValue, e, entityIndex)
		}
	case *ast.BlockStatement:
		for _, statement := range s.Statements {
			i.evalStatement(statement, e, entityIndex)
		}
	}
}

func (i *Interpreter) evalExpression(expr ast.Expression, e *simulation.Engine, entityIndex int) interface{} {
	switch e_node := expr.(type) {
	case *ast.CallExpression:
		return i.evalCallExpression(e_node, e, entityIndex)
	case *ast.Identifier:
		if val, ok := i.variables[e_node.Value]; ok {
			return val
		}
		return false
	case *ast.IntegerLiteral:
		return e_node.Value
	case *ast.InfixExpression:
		left := i.evalExpression(e_node.Left, e, entityIndex)
		right := i.evalExpression(e_node.Right, e, entityIndex)
		return i.evalInfixExpression(e_node.Operator, left, right)
	case *ast.PrefixExpression:
		right := i.evalExpression(e_node.Right, e, entityIndex)
		return i.evalPrefixExpression(e_node.Operator, right)
	}
	return false
}

func (i *Interpreter) evalInfixExpression(operator string, left, right interface{}) interface{} {
	switch {
	case operator == "+":
		lVal, lOk := left.(int64)
		rVal, rOk := right.(int64)
		if lOk && rOk {
			return lVal + rVal
		}
	case operator == "-":
		lVal, lOk := left.(int64)
		rVal, rOk := right.(int64)
		if lOk && rOk {
			return lVal - rVal
		}
	case operator == "*":
		lVal, lOk := left.(int64)
		rVal, rOk := right.(int64)
		if lOk && rOk {
			return lVal * rVal
		}
	case operator == "/":
		lVal, lOk := left.(int64)
		rVal, rOk := right.(int64)
		if lOk && rOk && rVal != 0 {
			return lVal / rVal
		}
	case operator == "<":
		lVal, lOk := left.(int64)
		rVal, rOk := right.(int64)
		if lOk && rOk {
			return lVal < rVal
		}
	case operator == ">":
		lVal, lOk := left.(int64)
		rVal, rOk := right.(int64)
		if lOk && rOk {
			return lVal > rVal
		}
	case operator == "<=":
		lVal, lOk := left.(int64)
		rVal, rOk := right.(int64)
		if lOk && rOk {
			return lVal <= rVal
		}
	case operator == ">=":
		lVal, lOk := left.(int64)
		rVal, rOk := right.(int64)
		if lOk && rOk {
			return lVal >= rVal
		}
	case operator == "==":
		return left == right
	case operator == "!=":
		return left != right
	}
	return false
}

func (i *Interpreter) evalPrefixExpression(operator string, right interface{}) interface{} {
	switch operator {
	case "-":
		if val, ok := right.(int64); ok {
			return -val
		}
	case "!":
		return !isTruthy(right)
	}
	return false
}

func (i *Interpreter) evalIfStatement(node *ast.IfStatement, e *simulation.Engine, entityIndex int) {
	cond := i.evalExpression(node.Condition, e, entityIndex)
	taken := isTruthy(cond)
	if i.builder != nil {
		i.builder.recordCondition(node.Condition.String(), cond, taken)
	}
	if taken {
		i.evalStatement(node.Consequence, e, entityIndex)
	} else if node.Alternative != nil {
		i.evalStatement(node.Alternative, e, entityIndex)
	}
}

func (i *Interpreter) evalWhileStatement(node *ast.WhileStatement, e *simulation.Engine, entityIndex int) {
	limit := 100
	for isTruthy(i.evalExpression(node.Condition, e, entityIndex)) && limit > 0 {
		i.evalStatement(node.Body, e, entityIndex)
		limit--
	}
}

func isTruthy(val interface{}) bool {
	switch v := val.(type) {
	case bool:
		return v
	case int64:
		return v != 0
	case nil:
		return false
	default:
		return true
	}
}

func (i *Interpreter) evalLetStatement(node *ast.LetStatement, e *simulation.Engine, entityIndex int) {
	if node.Value != nil {
		i.variables[node.Name.Value] = i.evalExpression(node.Value, e, entityIndex)
	}
}

func (i *Interpreter) evalFunctionDeclaration(node *ast.FunctionDeclaration, e *simulation.Engine, entityIndex int) {
	if node.Name.Value == "main" {
		i.evalStatement(node.Body, e, entityIndex)
	}
}

func (i *Interpreter) evalCallExpression(node *ast.CallExpression, e *simulation.Engine, entityIndex int) interface{} {
	if fn, ok := i.builtins[node.Function]; ok {
		result := fn(e, entityIndex)
		if i.builder != nil {
			i.builder.recordAction(node.Function, result)
		}
		return result
	}
	return false
}
