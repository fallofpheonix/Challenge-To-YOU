package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"chrysalis-engine/core/network"
	"chrysalis-engine/core/pscript/ast"
	"chrysalis-engine/core/pscript/interpreter"
	"chrysalis-engine/core/pscript/lexer"
	"chrysalis-engine/core/pscript/parser"
	"chrysalis-engine/core/simulation"
	"time"
)

func main() {
	// 1. Initialize High-Performance Engine
	width, height := 100, 100
	droneCount := 100
	engine := simulation.NewEngine(width, height, droneCount)

	// 1.5 Setup Networking
	hub := network.NewNetworkHub()
	go hub.Run()
	http.HandleFunc("/telemetry", hub.HandleConnections)
	go func() {
		fmt.Fprintln(os.Stderr, "[NETWORK] Starting WebSocket server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Fprintf(os.Stderr, "[NETWORK ERROR] Server failed: %v\n", err)
		}
	}()

	// Seed some resources
	for i := 0; i < 5; i++ {
		rx, ry := 70+i, 70+i
		idx := engine.Grid.GetIndex(rx, ry)
		engine.Grid.CurrentCells[idx].ResourceCount = 500
		engine.Grid.NextCells[idx].ResourceCount = 500
	}

	// 2. Load and Parse P-Script
	scriptPath := os.Getenv("PHX_SCRIPT_PATH")
	if scriptPath == "" {
		scriptPath = "scripts/agent.ps"
	}
	
	program := loadScript(scriptPath)

	// 3. Setup Interpreter and Builtins
	builtins := map[string]interpreter.BuiltinFn{
		"SENSE_RESOURCE": func(e *simulation.Engine, i int) interface{} { return e.SenseResource(i) },
		"SENSE_HOME":     func(e *simulation.Engine, i int) interface{} { return e.SenseHome(i) },
		"SENSE_BATTERY":  func(e *simulation.Engine, i int) interface{} { return e.Registry.Battery[i] },
		"SENSE_TRUST":    func(e *simulation.Engine, i int) interface{} { return int64(e.Registry.TrustScore[i]) },
		"SENSE_CORRUPTION": func(e *simulation.Engine, i int) interface{} { return int64(e.Registry.CorruptionFactor[i]) },
		"SENSE_COMPROMISED": func(e *simulation.Engine, i int) interface{} { return e.Registry.Compromised[i] },
		"SENSE_ALIEN_SIGNAL": func(e *simulation.Engine, i int) interface{} { return e.SenseAlienSignal(i) },
		"BROADCAST_VOTE": func(e *simulation.Engine, i int) interface{} { return e.SenseQuorum(i) },
		"SENSE_SWARM_SIZE": func(e *simulation.Engine, i int) interface{} { return int64(e.Registry.Count) },
		"SENSE_COLONY_RESOURCES": func(e *simulation.Engine, i int) interface{} { return int64(e.GlobalSilicates) },
		"HARVEST":        func(e *simulation.Engine, i int) interface{} { e.Harvest(i); return true },
		"DROP_RESOURCE":  func(e *simulation.Engine, i int) interface{} { e.DropResource(i); return true },
		"MOVE_RANDOM":    func(e *simulation.Engine, i int) interface{} { e.MoveRandom(i); return true },
		"MOVE_TOWARDS_RESOURCE": func(e *simulation.Engine, i int) interface{} { e.MoveTowardsResource(i); return true },
		"MOVE_TOWARDS_HOME":     func(e *simulation.Engine, i int) interface{} { e.MoveTowardsHome(i); return true },
	}
	interp := interpreter.New(builtins)

	// 4. Main Simulation Loop (10 Hz as per spec)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	fmt.Fprintln(os.Stderr, "--- Project Chrysalis Go Core Started ---")

	var lastMod time.Time

	for {
		select {
		case <-ticker.C:
			// 4.1 Check for Script Reload
			info, err := os.Stat(scriptPath)
			if err == nil && info.ModTime().After(lastMod) {
				fmt.Fprintln(os.Stderr, "Reloading Architect script...")
				program = loadScript(scriptPath)
				lastMod = info.ModTime()
			}

			// 4.2 Step Simulation Environment
			engine.Grid.TickPheromones()
			
			// Reinforce Base Pheromone
			bIdx := engine.Grid.GetIndex(width/2, height/2)
			engine.Grid.NextCells[bIdx].HomePheromone = simulation.MaxPheromone

			// 4.3 Execute P-Script for every Drone in the Registry
			if program != nil {
				for i := 0; i < engine.Registry.Count; i++ {
					interp.Eval(program, engine, i)
				}
			}

			// 4.4 Commit mutations
			engine.Grid.SwapBuffers()
			engine.Tick++

			// 5. Emit state to Telemetry Bridge
			state := engine.GetState()
			
			packet := map[string]interface{}{
				"packet_type": "EMISSION_SNAPSHOT",
				"tick":        engine.Tick,
				"payload":     state,
			}

			data, err := json.Marshal(packet)
			if err != nil {
				fmt.Fprintf(os.Stderr, "JSON marshal error: %v\n", err)
				continue
			}
			
			// Broadcast to all connected Godot clients
			hub.Broadcast <- data
		}
	}
}

func loadScript(path string) *ast.Program {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading script %s: %v\n", path, err)
		return nil
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Parser errors in %s:\n", path)
		for _, msg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  - %s\n", msg)
		}
		return nil
	}
	
	return program
}
