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

// TestVariableReassignmentInLoop is a regression guard for a backend-divergence
// bug found by the interpreter↔VM parity oracle: the interpreter had no "="
// assignment case, so `n = n + 1` was a silent no-op and this loop never
// progressed (n stuck at 0). The VM handled it, so the two backends disagreed.
func TestVariableReassignmentInLoop(t *testing.T) {
	input := `fn main() {
	    let n = 0;
	    while (n < 5) { n = n + 1 }
	    if (n >= 5) { reached() }
	}`
	program := parser.New(lexer.New(input)).ParseProgram()
	engine := simulation.NewEngine(10, 10, 1)

	calls := 0
	builtins := map[string]BuiltinFn{
		"reached": func(e *simulation.Engine, i int) interface{} { calls++; return true },
	}
	interp := New(builtins)
	interp.Eval(program, engine, 0)

	if calls != 1 {
		t.Fatalf("reached() called %d times, want 1 — loop did not progress via reassignment", calls)
	}
}
