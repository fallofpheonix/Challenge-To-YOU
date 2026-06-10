package interpreter

import (
	"chrysalis-engine/core/pscript/lexer"
	"chrysalis-engine/core/pscript/parser"
	"chrysalis-engine/core/simulation"
	"testing"
)

func TestSwarmLogic(t *testing.T) {
	input := `
	fn main() {
		if (SENSE_RESOURCE()) {
			HARVEST()
		} else {
			MOVE()
		}
		
		while (SENSE_DANGER()) {
			RETREAT()
		}
	}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	harvestCalled := false
	moveCalled := false
	retreatCount := 0

	var builtins map[string]BuiltinFn
	builtins = map[string]BuiltinFn{
		"SENSE_RESOURCE": func(e *simulation.Engine, i int) interface{} { return true },
		"HARVEST": func(e *simulation.Engine, i int) interface{} {
			harvestCalled = true
			return true
		},
		"MOVE": func(e *simulation.Engine, i int) interface{} {
			moveCalled = true
			return true
		},
		"SENSE_DANGER": func(e *simulation.Engine, i int) interface{} {
			if retreatCount >= 3 {
				return false
			}
			return true
		},
		"RETREAT": func(e *simulation.Engine, i int) interface{} {
			retreatCount++
			return true
		},
	}

	i := New(builtins)
	engine := simulation.NewEngine(10, 10, 1)
	i.Eval(program, engine, 0)

	if !harvestCalled {
		t.Errorf("expected HARVEST to be called")
	}
	if moveCalled {
		t.Errorf("expected MOVE not to be called")
	}
	if retreatCount != 3 {
		t.Errorf("expected RETREAT to be called 3 times, got %d", retreatCount)
	}
}
