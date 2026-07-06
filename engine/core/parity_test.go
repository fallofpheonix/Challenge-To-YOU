package main

import (
	"encoding/json"
	"testing"

	"chrysalis-engine/core/pscript/interpreter"
	"chrysalis-engine/core/pscript/lexer"
	"chrysalis-engine/core/pscript/parser"
	"chrysalis-engine/core/pscript/vm"
	"chrysalis-engine/core/simulation"
)

// assertBackendParity is the interpreter↔VM oracle guard (ADR-006). It compiles a
// script and runs it on two identical engines — one driven by the tree-walk
// interpreter, one by the bytecode VM — then asserts that after every tick both
// engines hold an identical canonical WorldHash AND emitted an identical event
// stream. Any divergence on a program that terminates within the shared execution
// budget fails the test. Both backends are exercised through the production
// stepEngine path (compiled != nil selects the VM; compiled == nil the interpreter).
func assertBackendParity(t *testing.T, script string, ticks int) {
	t.Helper()

	p := parser.New(lexer.New(script))
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	compiled := vm.NewCompiler().Compile(prog)
	if compiled == nil {
		t.Fatal("compilation failed")
	}

	const seed = 1234
	const w, h, drones = 30, 30, 6

	vmEng := simulation.NewEngineWithSeed(w, h, drones, seed)
	interpEng := simulation.NewEngineWithSeed(w, h, drones, seed)
	// Seed an identical resource node so drones have something to act on.
	for _, e := range []*simulation.Engine{vmEng, interpEng} {
		idx := e.Grid.GetIndex(w/2+1, h/2)
		e.Grid.CurrentCells[idx].ResourceCount = 5
		e.Grid.NextCells[idx].ResourceCount = 5
	}

	v := vm.NewVM(compiled, newBuiltinMap())
	interp := interpreter.New(newInterpreterBuiltins())

	for tick := 0; tick < ticks; tick++ {
		stepEngine(vmEng, prog, interp, v, compiled, -1) // VM path
		stepEngine(interpEng, prog, interp, v, nil, -1)  // interpreter path (compiled == nil)

		if vh, ih := vmEng.WorldHash(), interpEng.WorldHash(); vh != ih {
			t.Fatalf("tick %d: WorldHash diverged: vm=%016x interp=%016x", tick, vh, ih)
		}

		vmEv, _ := json.Marshal(vmEng.Bus.Events())
		inEv, _ := json.Marshal(interpEng.Bus.Events())
		if string(vmEv) != string(inEv) {
			t.Fatalf("tick %d: event stream diverged:\n vm=%s\n interp=%s", tick, vmEv, inEv)
		}
	}
}

// TestBackendParityRealisticAgent runs the canonical forage policy on both
// backends. This is the permanent regression guard: the two implementations of
// P-Script semantics must agree on real gameplay programs.
func TestBackendParityRealisticAgent(t *testing.T) {
	script := `fn main() {
    if (SENSE_BATTERY() < 25000000) {
        MOVE_TOWARDS_HOME()
    } else {
        if (SENSE_CARGO()) {
            DROP_RESOURCE()
            MOVE_TOWARDS_HOME()
        } else {
            HARVEST()
            if (SENSE_CARGO()) {
                MOVE_TOWARDS_HOME()
            } else {
                MOVE_TOWARDS_RESOURCE()
            }
        }
    }
}`
	assertBackendParity(t, script, 60)
}

// TestBackendParityWithinBudgetLoop exercises a loop well within the shared
// execution budget. Before PR2 the interpreter capped every while loop at 100
// iterations independently of the VM's aggregate budget; a loop-bearing policy
// that both backends can complete must now behave identically.
func TestBackendParityWithinBudgetLoop(t *testing.T) {
	script := `fn main() {
    let n = 0;
    while (n < 25) {
        n = n + 1
    }
    if (n >= 25) {
        HARVEST()
        MOVE_TOWARDS_RESOURCE()
    } else {
        MOVE_TOWARDS_HOME()
    }
}`
	assertBackendParity(t, script, 40)
}
