package main

import (
	"testing"

	"chrysalis-engine/core/pscript/interpreter"
	"chrysalis-engine/core/simulation"
)

func TestDefaultAgentCompletesResourceLoop(t *testing.T) {
	engine := simulation.NewEngineWithSeed(21, 21, 16, 123)
	engine.Hazards = simulation.NewHazardSystem(0)
	engine.Aliens = simulation.NewAlienNetwork(0)

	resourceIdx := engine.Grid.GetIndex(13, 10)
	engine.Grid.CurrentCells[resourceIdx].ResourceCount = 50
	engine.Grid.NextCells[resourceIdx].ResourceCount = 50

	program := loadScript("scripts/agent.ps")
	if program == nil {
		t.Fatal("default agent script failed to parse")
	}
	interp := interpreter.New(newBuiltins())

	for tick := 0; tick < 2_000 && engine.GlobalSilicates == 0; tick++ {
		engine.BeginTick()
		for i := 0; i < engine.Registry.Count; i++ {
			interp.Eval(program, engine, i)
		}
		engine.CommitTick()
	}

	if engine.GlobalSilicates == 0 {
		t.Fatal("default agent failed to harvest and return a nearby resource in 2,000 ticks")
	}
}
