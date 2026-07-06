package main

import (
	"chrysalis-engine/core/levels"
	"chrysalis-engine/core/network"
	"chrysalis-engine/core/pscript/ast"
	"chrysalis-engine/core/pscript/interpreter"
	"chrysalis-engine/core/pscript/lexer"
	"chrysalis-engine/core/pscript/parser"
	"chrysalis-engine/core/pscript/vm"
	"chrysalis-engine/core/replay"
	"chrysalis-engine/core/simulation"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. Initialize Engine — from level JSON if PHX_LEVEL_PATH is set, else defaults.
	width, height := 100, 100
	droneCount := 10
	levelID := ""

	var engine *simulation.Engine

	if levelPath := os.Getenv("PHX_LEVEL_PATH"); levelPath != "" {
		lvl, err := levels.LoadLevel(levelPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[LEVEL ERROR] %v — falling back to defaults\n", err)
		} else {
			engine = lvl.CreateEngine()
			width = lvl.World.Width
			height = lvl.World.Height
			droneCount = lvl.Drones.InitialCount
			levelID = lvl.ID
			fmt.Fprintf(os.Stderr, "[LEVEL] Loaded: %s (%s)\n", lvl.Title, lvl.ID)
		}
	}

	if engine == nil {
		// Legacy default: 10 drones, one seeded resource node.
		engine = simulation.NewEngine(width, height, droneCount)
		resourceIdx := engine.Grid.GetIndex(51, 50)
		engine.Grid.CurrentCells[resourceIdx].ResourceCount = simulation.InitialResourceNode
		engine.Grid.NextCells[resourceIdx].ResourceCount = simulation.InitialResourceNode
	}

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

	// 2. Load and Parse P-Script, compile to bytecode
	scriptPath := os.Getenv("PHX_SCRIPT_PATH")
	if scriptPath == "" {
		scriptPath = "scripts/agent.ps"
	}

	builtins := newBuiltinMap()
	program, compiled := loadScript(scriptPath, builtins)

	// 3. Setup VM and Interpreter (interpreter is fallback for uncompiled scripts)
	v := vm.NewVM(compiled, builtins)
	interp := interpreter.New(newInterpreterBuiltins())

	// 3.5 Initialize replay recorder (always-on; pure consumer of the EventBus).
	recorder := replay.NewRecorder(engine.GetState(), 500)
	recorder.SetMeta(replay.ReplayMeta{
		Seed:       engine.Seed,
		Version:    "0.9.0",
		Width:      width,
		Height:     height,
		DroneCount: droneCount,
		StartedAt:  time.Now().Unix(),
		LevelID:    levelID,
	})

	// 4. Main Simulation Loop (10 Hz as per spec)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Graceful shutdown: listen for SIGINT/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Fprintln(os.Stderr, "--- Project Chrysalis Go Core Started ---")

	// Initialize lastMod to avoid double-load on first tick
	var lastMod time.Time
	if info, err := os.Stat(scriptPath); err == nil {
		lastMod = info.ModTime()
	}

	// inspectedDroneID is set by the Godot client via INSPECT_DRONE commands.
	// When >= 0, the interpreter collects a full behavior trace for that drone each tick.
	inspectedDroneID := -1

	running := true
	for running {
		select {
		case <-sigChan:
			fmt.Fprintln(os.Stderr, "\n[SHUTDOWN] Signal received. Saving replay and closing...")
			running = false
		case cmd := <-commandChan:
			switch cmd.Type {
			case "COMMAND_INJECTION":
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
						// Recompile to bytecode
						c := vm.NewCompiler()
						newCompiled := c.Compile(newProg)
						if newCompiled != nil {
							compiled = newCompiled
							v = vm.NewVM(compiled, builtins)
							fmt.Fprintln(os.Stderr, "[NETWORK] Hot-patch applied (bytecode).")
						} else {
							fmt.Fprintf(os.Stderr, "[NETWORK] Hot-patch AST OK, bytecode failed: %v — using interpreter fallback\n", c.Errors())
							compiled = nil
						}
					} else {
						fmt.Fprintf(os.Stderr, "[NETWORK ERROR] Patch failed validation: %v\n", p.Errors())
					}
				}

			case "INSPECT_DRONE":
				var payload struct {
					DroneID int `json:"drone_id"`
				}
				if err := json.Unmarshal(cmd.Payload, &payload); err == nil {
					inspectedDroneID = payload.DroneID
					fmt.Fprintf(os.Stderr, "[INSPECTOR] Tracing drone %d\n", inspectedDroneID)
				}

			case "INSPECT_CLEAR":
				inspectedDroneID = -1
				fmt.Fprintln(os.Stderr, "[INSPECTOR] Trace cleared")

			case "REPLAY_SEEK":
				var seekPayload struct {
					Tick int64 `json:"tick"`
				}
				if err := json.Unmarshal(cmd.Payload, &seekPayload); err == nil {
					cp := recorder.NearestCheckpoint(seekPayload.Tick)

					// Save live state so we can restore it after the seek simulation.
					liveState := engine.GetState()

					// Restore engine to checkpoint.
					engine.SetState(cp.State)

					// Validate: WorldHash after SetState must match what was recorded.
					// A mismatch means SetState has a coverage gap or the checkpoint is corrupt.
					if cp.WorldHash != 0 {
						restoredHash := engine.WorldHash()
						if restoredHash != cp.WorldHash {
							fmt.Fprintf(os.Stderr, "[REPLAY VALIDATION FAIL] SetState divergence at "+
								"checkpoint tick %d: want %016x got %016x — replay state may be incorrect\n",
								cp.Tick, cp.WorldHash, restoredHash)
						}
					}

					// Cap forward-simulation distance to prevent DoS via large seeks
					const MaxSeekDistance int64 = 10_000
					if seekPayload.Tick-cp.Tick > MaxSeekDistance {
						fmt.Fprintf(os.Stderr, "[REPLAY] Seek distance %d exceeds max %d — clamping\n",
							seekPayload.Tick-cp.Tick, MaxSeekDistance)
						seekPayload.Tick = cp.Tick + MaxSeekDistance
					}

					// Forward-simulate to target tick using the same stepEngine path as live mode.
					for engine.Tick < seekPayload.Tick {
						stepEngine(engine, program, interp, v, compiled, -1)
					}

					// Broadcast reconstructed state.
					seekState := engine.GetState()
					seekHash := engine.WorldHash()
					seekState["events"] = recorder.EventsInRange(cp.Tick, seekPayload.Tick)
					seekState["replay"] = map[string]interface{}{
						"recording":    false,
						"total_ticks":  recorder.TotalFrames(),
						"current_tick": seekPayload.Tick,
						"seek_to":      seekPayload.Tick,
						"world_hash":   seekHash,
						"events":       recorder.EventsInRange(cp.Tick, seekPayload.Tick),
					}
					seekPacket := map[string]interface{}{
						"packet_type": "EMISSION_SNAPSHOT",
						"tick":        seekPayload.Tick,
						"payload":     seekState,
					}
					if seekData, merr := json.Marshal(seekPacket); merr == nil {
						hub.Broadcast <- seekData
					}
					fmt.Fprintf(os.Stderr, "[REPLAY] Reconstructed tick %d from checkpoint %d (hash %016x)\n",
						seekPayload.Tick, cp.Tick, seekHash)

					// Restore live simulation state.
					engine.SetState(liveState)
				}

			case "REPLAY_SAVE":
				raw, serr := recorder.Serialize()
				if serr == nil {
					fname := fmt.Sprintf("replay_%d.chrysalis_replay", time.Now().Unix())
					if werr := os.WriteFile(fname, raw, 0644); werr != nil {
						fmt.Fprintf(os.Stderr, "[REPLAY ERROR] Save failed: %v\n", werr)
					} else {
						fmt.Fprintf(os.Stderr, "[REPLAY] Saved %d frames, %d checkpoints → %s\n",
							recorder.TotalFrames(), recorder.CheckpointCount(), fname)
					}
				}

			case "LOAD_LEVEL":
				// Hot-swap to a different level without restarting the process.
				// The client sends this after a mission completes (victory or defeat)
				// to advance campaign progression.
				var payload struct {
					LevelPath string `json:"level_path"`
				}
				if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
					fmt.Fprintf(os.Stderr, "[LEVEL ERROR] LOAD_LEVEL parse error: %v\n", err)
					break
				}
				lvl, err := levels.LoadLevel(payload.LevelPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "[LEVEL ERROR] LOAD_LEVEL failed: %v\n", err)
					break
				}
				engine = lvl.CreateEngine()
				width = lvl.World.Width
				height = lvl.World.Height
				droneCount = lvl.Drones.InitialCount
				levelID = lvl.ID
				inspectedDroneID = -1
				recorder = replay.NewRecorder(engine.GetState(), recorder.CheckpointEvery)
				recorder.SetMeta(replay.ReplayMeta{
					Seed:       engine.Seed,
					Version:    "0.9.0",
					Width:      width,
					Height:     height,
					DroneCount: droneCount,
					StartedAt:  time.Now().Unix(),
					LevelID:    levelID,
				})
				fmt.Fprintf(os.Stderr, "[LEVEL] Hot-loaded: %s (%s)\n", lvl.Title, lvl.ID)
				if notif, nerr := json.Marshal(map[string]interface{}{
					"packet_type": "LEVEL_LOADED",
					"payload": map[string]interface{}{
						"level_id":    lvl.ID,
						"title":       lvl.Title,
						"description": lvl.Description,
					},
				}); nerr == nil {
					hub.Broadcast <- notif
				}
			}

		case <-ticker.C:
			info, err := os.Stat(scriptPath)
			if err == nil && info.ModTime().After(lastMod) {
				fmt.Fprintln(os.Stderr, "Reloading Architect script...")
				program, compiled = loadScript(scriptPath, builtins)
				v = vm.NewVM(compiled, builtins)
				lastMod = info.ModTime()
			}

			activeFrame := stepEngine(engine, program, interp, v, compiled, inspectedDroneID)

			state := engine.GetState()

			// Record this tick before adding ephemeral fields (events/trace/replay)
			// so that checkpoints contain only durable world state.
			recorder.Record(engine.Tick, engine.Bus.Events())
			if engine.Tick%recorder.CheckpointEvery == 0 {
				recorder.Checkpoint(engine.Tick, state, engine.WorldHash())
			}

			state["events"] = engine.Bus.Events()
			if activeFrame != nil {
				state["trace"] = activeFrame
			}
			state["replay"] = map[string]interface{}{
				"recording":    true,
				"total_ticks":  engine.Tick,
				"current_tick": engine.Tick,
			}

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
			hub.Broadcast <- data
		}
	}

	// Shutdown: save replay, close connections
	ticker.Stop()
	raw, serr := recorder.Serialize()
	if serr == nil {
		fname := fmt.Sprintf("replay_%d.chrysalis_replay", time.Now().Unix())
		if werr := os.WriteFile(fname, raw, 0644); werr != nil {
			fmt.Fprintf(os.Stderr, "[SHUTDOWN] Replay save failed: %v\n", werr)
		} else {
			fmt.Fprintf(os.Stderr, "[SHUTDOWN] Replay saved: %s (%d frames, %d checkpoints)\n",
				fname, recorder.TotalFrames(), recorder.CheckpointCount())
		}
	}
	fmt.Fprintf(os.Stderr, "[SHUTDOWN] Core exited cleanly at tick %d\n", engine.Tick)
}

// newBuiltinMap returns builtins for the bytecode VM.
func newBuiltinMap() map[string]vm.BuiltinFn {
	return map[string]vm.BuiltinFn{
		"SENSE_RESOURCE":     func(e *simulation.Engine, i int) interface{} { return e.SenseResource(i) },
		"SENSE_HOME":         func(e *simulation.Engine, i int) interface{} { return e.SenseHome(i) },
		"SENSE_BATTERY": func(e *simulation.Engine, i int) interface{} {
			if i < 0 || i >= e.Registry.Count {
				return int64(0)
			}
			return e.Registry.Battery[i]
		},
		"SENSE_TRUST": func(e *simulation.Engine, i int) interface{} {
			if i < 0 || i >= e.Registry.Count {
				return int64(100)
			}
			return int64(e.Registry.TrustScore[i])
		},
		"SENSE_CORRUPTION": func(e *simulation.Engine, i int) interface{} {
			if i < 0 || i >= e.Registry.Count {
				return int64(0)
			}
			return int64(e.Registry.CorruptionFactor[i])
		},
		"SENSE_COMPROMISED": func(e *simulation.Engine, i int) interface{} {
			if i < 0 || i >= e.Registry.Count {
				return false
			}
			return e.Registry.Compromised[i]
		},
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

// newInterpreterBuiltins returns builtins for the tree-walk interpreter (fallback).
func newInterpreterBuiltins() map[string]interpreter.BuiltinFn {
	return map[string]interpreter.BuiltinFn{
		"SENSE_RESOURCE":     func(e *simulation.Engine, i int) interface{} { return e.SenseResource(i) },
		"SENSE_HOME":         func(e *simulation.Engine, i int) interface{} { return e.SenseHome(i) },
		"SENSE_BATTERY": func(e *simulation.Engine, i int) interface{} {
			if i < 0 || i >= e.Registry.Count {
				return int64(0)
			}
			return e.Registry.Battery[i]
		},
		"SENSE_TRUST": func(e *simulation.Engine, i int) interface{} {
			if i < 0 || i >= e.Registry.Count {
				return int64(100)
			}
			return int64(e.Registry.TrustScore[i])
		},
		"SENSE_CORRUPTION": func(e *simulation.Engine, i int) interface{} {
			if i < 0 || i >= e.Registry.Count {
				return int64(0)
			}
			return int64(e.Registry.CorruptionFactor[i])
		},
		"SENSE_COMPROMISED": func(e *simulation.Engine, i int) interface{} {
			if i < 0 || i >= e.Registry.Count {
				return false
			}
			return e.Registry.Compromised[i]
		},
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

// stepEngine runs one complete simulation tick. Uses the bytecode VM when
// available, falls back to the tree-walk interpreter for uncompiled scripts.
// Pass inspectID = -1 to skip behavior tracing; pass a drone index to collect
// a DecisionFrame for that drone (live mode only — replay passes -1).
func stepEngine(e *simulation.Engine, prog *ast.Program, interp *interpreter.Interpreter, v *vm.VM, compiled *vm.Program, inspectID int) *simulation.DecisionFrame {
	e.BeginTick()
	var frame *simulation.DecisionFrame

	if compiled != nil {
		// Bytecode VM path
		for i := 0; i < e.Registry.Count; i++ {
			if inspectID >= 0 && i == inspectID {
				traceSteps := v.RunTraced(e, i, e.Tick)
				frame = &simulation.DecisionFrame{
					DroneID: i,
					Tick:    e.Tick,
					Steps:   make([]simulation.DecisionStep, len(traceSteps)),
				}
				for k, ts := range traceSteps {
					frame.Steps[k] = simulation.DecisionStep{
						Kind:   ts.Kind,
						Name:   ts.Name,
						Result: ts.Result,
						Taken:  ts.Taken,
					}
				}
			} else {
				v.Run(e, i)
			}
		}
	} else if prog != nil {
		// Interpreter fallback
		for i := 0; i < e.Registry.Count; i++ {
			if inspectID >= 0 && i == inspectID {
				frame = interp.EvalTraced(prog, e, i, e.Tick)
			} else {
				interp.Eval(prog, e, i)
			}
		}
	}

	e.CommitTick()
	return frame
}

// loadScript parses a P-Script file and compiles it to bytecode.
// Returns the AST (for interpreter fallback) and compiled bytecode (may be nil).
func loadScript(path string, builtins map[string]vm.BuiltinFn) (*ast.Program, *vm.Program) {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading script %s: %v\n", path, err)
		return nil, nil
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Parser errors in %s:\n", path)
		for _, msg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  - %s\n", msg)
		}
		return nil, nil
	}

	// Compile to bytecode
	c := vm.NewCompiler()
	compiled := c.Compile(program)
	if compiled == nil {
		fmt.Fprintf(os.Stderr, "[VM] Compilation failed, using interpreter fallback: %v\n", c.Errors())
		return program, nil
	}

	fmt.Fprintf(os.Stderr, "[VM] Compiled %s → %d instructions, %d constants\n",
		path, len(compiled.Instructions), len(compiled.Constants))
	return program, compiled
}
