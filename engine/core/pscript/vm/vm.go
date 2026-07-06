package vm

import (
	"chrysalis-engine/core/simulation"
	"fmt"
)

// OpCode defines a single VM instruction.
type OpCode uint8

const (
	OpConst    OpCode = iota // Push constant[arg] onto stack
	OpLoad                   // Push variables[arg] onto stack
	OpStore                  // Pop top → variables[arg]
	OpCall                   // Call builtins[arg], push result
	OpAdd                    // Pop b, a; push a + b
	OpSub                    // Pop b, a; push a - b
	OpMul                    // Pop b, a; push a * b
	OpDiv                    // Pop b, a; push a / b
	OpLt                     // Pop b, a; push a < b
	OpGt                     // Pop b, a; push a > b
	OpLte                    // Pop b, a; push a <= b
	OpGte                    // Pop b, a; push a >= b
	OpEq                     // Pop b, a; push a == b
	OpNeq                    // Pop b, a; push a != b
	OpNeg                    // Pop a; push -a
	OpNot                    // Pop a; push !a
	OpJump                   // PC += int16(arg)
	OpJumpIf                 // Pop cond; if !cond, PC += int16(arg)
	OpPop                    // Discard top of stack
	OpDone                   // Halt execution
)

// Instruction is a compact 4-byte VM operation.
// Op: 1 byte, Arg: 2 bytes (signed for jump offsets), Padding: 1 byte.
type Instruction struct {
	Op  OpCode
	Arg int16
}

// Program is the compiled bytecode plus its constant pool and symbol metadata.
type Program struct {
	Instructions []Instruction
	Constants    []int64
	FuncNames    []string // builtins referenced by index in FuncNames
	FuncIndex    map[string]int
	MaxVars      int
}

// MaxSteps is the safety limit for VM execution per tick (prevents infinite loops).
const MaxSteps = 1000

// BuiltinFn is the Go function signature for P-Script built-in calls.
type BuiltinFn func(e *simulation.Engine, entityIndex int) interface{}

// VM executes compiled P-Script bytecode against the simulation.
type VM struct {
	Program   *Program
	Stack     [256]interface{} // fixed stack — no allocation per tick
	SP        int              // stack pointer (next free slot)
	Variables [128]interface{} // named variable storage
	Builtins  map[string]BuiltinFn
}

// NewVM creates a VM ready to execute the given compiled program.
func NewVM(prog *Program, builtins map[string]BuiltinFn) *VM {
	return &VM{
		Program:  prog,
		Builtins: builtins,
	}
}

// Run executes the full program for a single drone. Returns immediately on OpDone.
func (v *VM) Run(e *simulation.Engine, entityIndex int) {
	v.SP = 0
	for i := range v.Variables {
		v.Variables[i] = nil
	}

	pc := 0
	prog := v.Program
	steps := 0
	for pc < len(prog.Instructions) && steps < MaxSteps {
		steps++
		ins := prog.Instructions[pc]
		pc++

		switch ins.Op {
		case OpConst:
			if int(ins.Arg) < 0 || int(ins.Arg) >= len(prog.Constants) {
				fmt.Printf("[VM ERROR] Const index %d out of range\n", ins.Arg)
				return
			}
			v.push(prog.Constants[ins.Arg])

		case OpLoad:
			if int(ins.Arg) < 0 || int(ins.Arg) >= len(v.Variables) {
				fmt.Printf("[VM ERROR] Var index %d out of range\n", ins.Arg)
				return
			}
			v.push(v.Variables[ins.Arg])

		case OpStore:
			if int(ins.Arg) < 0 || int(ins.Arg) >= len(v.Variables) {
				fmt.Printf("[VM ERROR] Var index %d out of range\n", ins.Arg)
				return
			}
			val := v.pop()
			v.Variables[ins.Arg] = val
			v.push(val) // leave value on stack for return/expression chaining

		case OpCall:
			if int(ins.Arg) < 0 || int(ins.Arg) >= len(prog.FuncNames) {
				fmt.Printf("[VM ERROR] Func index %d out of range\n", ins.Arg)
				return
			}
			name := prog.FuncNames[ins.Arg]
			if fn, ok := v.Builtins[name]; ok {
				result := fn(e, entityIndex)
				v.push(result)
			} else {
				v.push(false)
			}

		case OpAdd:
			b, a := v.popInt(), v.popInt()
			v.push(a + b)
		case OpSub:
			b, a := v.popInt(), v.popInt()
			v.push(a - b)
		case OpMul:
			b, a := v.popInt(), v.popInt()
			v.push(a * b)
		case OpDiv:
			b, a := v.popInt(), v.popInt()
			if b != 0 {
				v.push(a / b)
			} else {
				v.push(int64(0))
			}

		case OpLt:
			b, a := v.popInt(), v.popInt()
			v.push(a < b)
		case OpGt:
			b, a := v.popInt(), v.popInt()
			v.push(a > b)
		case OpLte:
			b, a := v.popInt(), v.popInt()
			v.push(a <= b)
		case OpGte:
			b, a := v.popInt(), v.popInt()
			v.push(a >= b)
		case OpEq:
			b, a := v.pop(), v.pop()
			v.push(a == b)
		case OpNeq:
			b, a := v.pop(), v.pop()
			v.push(a != b)

		case OpNeg:
			v.push(-v.popInt())
		case OpNot:
			v.push(!isTruthy(v.pop()))

		case OpJump:
			pc += int(ins.Arg)

		case OpJumpIf:
			cond := v.pop()
			if !isTruthy(cond) {
				pc += int(ins.Arg)
			}

		case OpPop:
			v.pop()

		case OpDone:
			return

		default:
			fmt.Printf("[VM ERROR] Unknown opcode: %d\n", ins.Op)
			return
		}
	}
}

// RunTraced executes the program and records every condition/action for the inspector.
func (v *VM) RunTraced(e *simulation.Engine, entityIndex int, tick int64) []TraceStep {
	v.SP = 0
	for i := range v.Variables {
		v.Variables[i] = nil
	}

	var trace []TraceStep
	pc := 0
	prog := v.Program
	steps := 0
	for pc < len(prog.Instructions) && steps < MaxSteps {
		steps++
		ins := prog.Instructions[pc]
		pc++

		switch ins.Op {
		case OpConst:
			if int(ins.Arg) < 0 || int(ins.Arg) >= len(prog.Constants) {
				fmt.Printf("[VM ERROR] Const index %d out of range\n", ins.Arg)
				return trace
			}
			v.push(prog.Constants[ins.Arg])
		case OpLoad:
			if int(ins.Arg) < 0 || int(ins.Arg) >= len(v.Variables) {
				fmt.Printf("[VM ERROR] Var index %d out of range\n", ins.Arg)
				return trace
			}
			v.push(v.Variables[ins.Arg])
		case OpStore:
			if int(ins.Arg) < 0 || int(ins.Arg) >= len(v.Variables) {
				fmt.Printf("[VM ERROR] Var index %d out of range\n", ins.Arg)
				return trace
			}
			val := v.pop()
			v.Variables[ins.Arg] = val
			v.push(val) // leave value on stack for return/expression chaining

		case OpCall:
			if int(ins.Arg) < 0 || int(ins.Arg) >= len(prog.FuncNames) {
				fmt.Printf("[VM ERROR] Func index %d out of range\n", ins.Arg)
				return trace
			}
			name := prog.FuncNames[ins.Arg]
			if fn, ok := v.Builtins[name]; ok {
				result := fn(e, entityIndex)
				trace = append(trace, TraceStep{Kind: "action", Name: name, Result: fmtResult(result)})
				v.push(result)
			} else {
				v.push(false)
			}

		case OpAdd:
			b, a := v.popInt(), v.popInt()
			v.push(a + b)
		case OpSub:
			b, a := v.popInt(), v.popInt()
			v.push(a - b)
		case OpMul:
			b, a := v.popInt(), v.popInt()
			v.push(a * b)
		case OpDiv:
			b, a := v.popInt(), v.popInt()
			if b != 0 {
				v.push(a / b)
			} else {
				v.push(int64(0))
			}

		case OpLt:
			b, a := v.popInt(), v.popInt()
			v.push(a < b)
		case OpGt:
			b, a := v.popInt(), v.popInt()
			v.push(a > b)
		case OpLte:
			b, a := v.popInt(), v.popInt()
			v.push(a <= b)
		case OpGte:
			b, a := v.popInt(), v.popInt()
			v.push(a >= b)
		case OpEq:
			b, a := v.pop(), v.pop()
			v.push(a == b)
		case OpNeq:
			b, a := v.pop(), v.pop()
			v.push(a != b)

		case OpNeg:
			v.push(-v.popInt())
		case OpNot:
			v.push(!isTruthy(v.pop()))

		case OpJump:
			pc += int(ins.Arg)
		case OpJumpIf:
			cond := v.pop()
			if !isTruthy(cond) {
				pc += int(ins.Arg)
			}

		case OpPop:
			v.pop()

		case OpDone:
			return trace
		}
	}
	return trace
}

func (v *VM) push(val interface{}) {
	if v.SP >= len(v.Stack) {
		fmt.Println("[VM ERROR] Stack overflow")
		return
	}
	v.Stack[v.SP] = val
	v.SP++
}

func (v *VM) pop() interface{} {
	if v.SP <= 0 {
		fmt.Println("[VM ERROR] Stack underflow")
		return int64(0)
	}
	v.SP--
	return v.Stack[v.SP]
}

func (v *VM) popInt() int64 {
	val := v.pop()
	if i, ok := val.(int64); ok {
		return i
	}
	if b, ok := val.(bool); ok {
		if b {
			return 1
		}
		return 0
	}
	return 0
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

func fmtResult(v interface{}) string {
	switch val := v.(type) {
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int64:
		return fmt.Sprintf("%d", val)
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// TraceStep mirrors simulation.DecisionStep for VM tracing.
type TraceStep struct {
	Kind   string
	Name   string
	Result string
	Taken  bool
}
