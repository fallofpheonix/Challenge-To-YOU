package interpreter

import (
	"chrysalis-engine/core/pscript/lexer"
	"chrysalis-engine/core/pscript/parser"
	"chrysalis-engine/core/simulation"
	"testing"
)

func TestEval(t *testing.T) {
	input := `fn main() { move_forward() }`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	engine := simulation.NewEngine(10, 10, 1)

	builtins := map[string]BuiltinFn{
		"move_forward": func(e *simulation.Engine, i int) interface{} {
			e.Registry.PositionX[i].V += 1000
			return true
		},
	}

	interp := New(builtins)
	interp.Eval(program, engine, 0)

	if engine.Registry.PositionX[0].V != 5001000 {
		t.Errorf("expected Agent.X to be 5001000, got %d", engine.Registry.PositionX[0].V)
	}
}
