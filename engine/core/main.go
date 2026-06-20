package main

import (
	"chrysalis-engine/core/network"
	"chrysalis-engine/core/pscript/ast"
	"chrysalis-engine/core/pscript/interpreter"
	"chrysalis-engine/core/pscript/lexer"
	"chrysalis-engine/core/pscript/parser"
	"chrysalis-engine/core/simulation"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

	commandChan := make(chan network.InboundCommand, 32)
	http.HandleFunc("/telemetry", func(w http.ResponseWriter, r *http.Request) {
		conn, err := network.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[NETWORK ERROR] Upgrade failed: %v\n", err)
			return
		}
		hub.Register <- conn
		go hub.StartReader(conn, commandChan)
	})

	go func() {
		fmt.Fprintln(os.Stderr, "[NETWORK] Starting WebSocket server on 127.0.0.1:8080")
		if err := http.ListenAndServe("127.0.0.1:8080", nil); err != nil {
			fmt.Fprintf(os.Stderr, "[NETWORK ERROR] Server failed: %v\n", err)
		}
	}()

	// Seed one guaranteed v0 resource node adjacent to base.
	resourceIdx := engine.Grid.GetIndex(51, 50)
	engine.Grid.CurrentCells[resourceIdx].ResourceCount = 500
	engine.Grid.NextCells[resourceIdx].ResourceCount = 500

	// 2. Load and Parse P-Script
	scriptPath := os.Getenv("PHX_SCRIPT_PATH")
	if scriptPath == "" {
		scriptPath = "scripts/agent.ps"
	}

	program := loadScript(scriptPath)

	// 3. Setup Interpreter and Builtins
	interp := interpreter.New(newBuiltins())

	// 4. Main Simulation Loop (10 Hz as per spec)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	fmt.Fprintln(os.Stderr, "--- Project Chrysalis Go Core Started ---")

	// Initialize lastMod to avoid double-load on first tick
	var lastMod time.Time
	if info, err := os.Stat(scriptPath); err == nil {
		lastMod = info.ModTime()
	}

	for {
		select {
		case cmd := <-commandChan:
			if cmd.Type == "COMMAND_INJECTION" {
				fmt.Fprintln(os.Stderr, "[NETWORK] Remote script override received. Parsing new AST tokens...")
				var payload struct {
					Code string `json:"code"`
				}
				if err := json.Unmarshal(cmd.Payload, &payload); err == nil {
					l := lexer.New(payload.Code)
					p := parser.New(l)
					newProg := p.ParseProgram()
					if len(p.Errors()) == 0 {
						program = newProg
						fmt.Fprintln(os.Stderr, "[NETWORK] Hot-patch applied successfully.")
					} else {
						fmt.Fprintf(os.Stderr, "[NETWORK ERROR] Patch failed validation: %v\n", p.Errors())
					}
				}
			}

		case <-ticker.C:
			info, err := os.Stat(scriptPath)
			if err == nil && info.ModTime().After(lastMod) {
				fmt.Fprintln(os.Stderr, "Reloading Architect script...")
				program = loadScript(scriptPath)
				lastMod = info.ModTime()
			}

			engine.BeginTick()
			if program != nil {
				for i := 0; i < engine.Registry.Count; i++ {
					interp.Eval(program, engine, i)
				}
			}
			engine.CommitTick()

			packet := map[string]interface{}{
				"packet_type": "EMISSION_SNAPSHOT",
				"tick":        engine.Tick,
				"payload":     engine.GetState(),
			}

			data, err := json.Marshal(packet)
			if err != nil {
				fmt.Fprintf(os.Stderr, "JSON marshal error: %v\n", err)
				continue
			}
			hub.Broadcast <- data
		}
	}
}

func newBuiltins() map[string]interpreter.BuiltinFn {
	return map[string]interpreter.BuiltinFn{
		"SENSE_RESOURCE":         func(e *simulation.Engine, i int) interface{} { return e.SenseResource(i) },
		"SENSE_HOME":             func(e *simulation.Engine, i int) interface{} { return e.SenseHome(i) },
		"SENSE_BATTERY":          func(e *simulation.Engine, i int) interface{} { return e.Registry.Battery[i] },
		"SENSE_TRUST":            func(e *simulation.Engine, i int) interface{} { return int64(e.Registry.TrustScore[i]) },
		"SENSE_CORRUPTION":       func(e *simulation.Engine, i int) interface{} { return int64(e.Registry.CorruptionFactor[i]) },
		"SENSE_COMPROMISED":      func(e *simulation.Engine, i int) interface{} { return e.Registry.Compromised[i] },
		"SENSE_ALIEN_SIGNAL":     func(e *simulation.Engine, i int) interface{} { return e.SenseAlienSignal(i) },
		"BROADCAST_VOTE":         func(e *simulation.Engine, i int) interface{} { return e.SenseQuorum(i) },
		"SENSE_SWARM_SIZE":       func(e *simulation.Engine, i int) interface{} { return int64(e.Registry.Count) },
		"SENSE_COLONY_RESOURCES": func(e *simulation.Engine, i int) interface{} { return int64(e.GlobalSilicates) },
		"SENSE_CARGO":            func(e *simulation.Engine, i int) interface{} { return e.SenseCargo(i) },
		"HARVEST":                func(e *simulation.Engine, i int) interface{} { e.Harvest(i); return true },
		"DROP_RESOURCE":          func(e *simulation.Engine, i int) interface{} { e.DropResource(i); return true },
		"MOVE_RANDOM":            func(e *simulation.Engine, i int) interface{} { e.MoveRandom(i); return true },
		"MOVE_TOWARDS_RESOURCE":  func(e *simulation.Engine, i int) interface{} { e.MoveTowardsResource(i); return true },
		"MOVE_TOWARDS_HOME":      func(e *simulation.Engine, i int) interface{} { e.MoveTowardsHome(i); return true },
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
