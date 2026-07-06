package main

import (
	"testing"

	"chrysalis-engine/core/pscript/interpreter"
	"chrysalis-engine/core/pscript/vm"
	"chrysalis-engine/core/simulation"
)

func TestDefaultAgentCompletesResourceLoop(t *testing.T) {
	engine := simulation.NewEngineWithSeed(21, 21, 16, 123)
	engine.Hazards = simulation.NewHazardSystem(0)
	engine.Aliens = simulation.NewAlienNetwork(0)

	resourceIdx := engine.Grid.GetIndex(13, 10)
	engine.Grid.CurrentCells[resourceIdx].ResourceCount = 50
	engine.Grid.NextCells[resourceIdx].ResourceCount = 50

	builtins := newBuiltinMap()
	program, compiled := loadScript("scripts/agent.ps", builtins)
	if program == nil {
		t.Fatal("default agent script failed to parse")
	}
	v := vm.NewVM(compiled, builtins)
	interp := interpreter.New(newInterpreterBuiltins())

	for tick := 0; tick < 2_000 && engine.GlobalSilicates == 0; tick++ {
		engine.BeginTick()
		for i := 0; i < engine.Registry.Count; i++ {
			if compiled != nil {
				v.Run(engine, i)
			} else {
				interp.Eval(program, engine, i)
			}
		}
		engine.CommitTick()
	}

	if engine.GlobalSilicates == 0 {
		t.Fatal("default agent failed to harvest and return a nearby resource in 2,000 ticks")
	}
}

func TestDefaultMissionReachesVictory(t *testing.T) {
	engine := simulation.NewEngineWithSeed(100, 100, 100, simulation.DefaultWorldSeed)

	resourceIdx := engine.Grid.GetIndex(51, 50)
	engine.Grid.CurrentCells[resourceIdx].ResourceCount = 500
	engine.Grid.NextCells[resourceIdx].ResourceCount = 500

	builtins := newBuiltinMap()
	program, compiled := loadScript("scripts/agent.ps", builtins)
	if program == nil {
		t.Fatal("default agent script failed to parse")
	}
	v := vm.NewVM(compiled, builtins)
	interp := interpreter.New(newInterpreterBuiltins())

	for engine.Mission.Status == simulation.MissionRunning {
		engine.BeginTick()
		for i := 0; i < engine.Registry.Count; i++ {
			if compiled != nil {
				v.Run(engine, i)
			} else {
				interp.Eval(program, engine, i)
			}
		}
		engine.CommitTick()
	}

	if engine.Mission.Status != simulation.MissionVictory {
		carrying := 0
		for i := 0; i < engine.Registry.Count; i++ {
			if engine.Registry.Inventory[i] > 0 {
				carrying++
			}
		}
		t.Fatalf("default mission did not reach victory: status=%s reason=%s tick=%d deposited=%d carrying=%d infected_ratio=%.2f",
			engine.Mission.Status,
			engine.Mission.Reason,
			engine.Mission.Tick,
			engine.Mission.ResourcesDeposited,
			carrying,
			engine.Mission.InfectedRatio,
		)
	}
}
