// Package levels loads mission definitions from JSON and produces a configured
// simulation.Engine. Adding a new mission requires only a JSON file; no Go
// changes are needed.
//
// Schema reference: see chrysalis_1.json for a complete annotated example.
package levels

import (
	"chrysalis-engine/core/simulation"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// CurrentSchemaVersion is the level JSON schema version this build understands.
// Bump this whenever the schema adds incompatible fields; LoadLevel rejects older
// files whose schema_version exceeds this value.
const CurrentSchemaVersion = 1

// WorldDef defines map dimensions and the deterministic seed.
type WorldDef struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Seed   int64 `json:"seed"`
}

// DronesDef defines the initial swarm population.
type DronesDef struct {
	InitialCount int `json:"initial_count"`
}

// ResourceDef places a resource deposit at a grid cell.
type ResourceDef struct {
	ID    string `json:"id,omitempty"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Count int32  `json:"count"`
}

// HazardDef places an environmental hazard.
// Intensity is a raw crysmath.Precision-scaled value (1_000_000 = 1 unit/tick).
// Type: 0 = HazardMagnetic (battery drain), 1 = HazardThermal (reserved).
type HazardDef struct {
	ID        string `json:"id,omitempty"`
	Type      int    `json:"type"`
	X         int32  `json:"x"`
	Y         int32  `json:"y"`
	Radius    int32  `json:"radius"`
	Intensity int64  `json:"intensity"`
}

// AlienDef places an alien network node.
// Type: 0 = NodeInfector (spreads logic virus), 1 = NodeJammer (reserved).
type AlienDef struct {
	ID     string `json:"id,omitempty"`
	Type   int    `json:"type"`
	X      int32  `json:"x"`
	Y      int32  `json:"y"`
	Radius int32  `json:"radius"`
}

// MissionDef defines victory and defeat conditions.
type MissionDef struct {
	TargetResources        int     `json:"target_resources"`
	MaxTicks               int64   `json:"max_ticks"`
	InfectionLossThreshold float64 `json:"infection_loss_threshold"`
}

// NarrativeDef holds display strings for mission outcomes.
type NarrativeDef struct {
	Victory         string `json:"victory"`
	DefeatInfection string `json:"defeat_infection"`
	DefeatTickLimit string `json:"defeat_tick_limit"`
}

// Level is the complete mission definition loaded from a JSON file.
type Level struct {
	SchemaVersion    int           `json:"schema_version"`
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	Description      string        `json:"description"`
	Narrative        NarrativeDef  `json:"narrative"`
	World            WorldDef      `json:"world"`
	Drones           DronesDef     `json:"drones"`
	Resources        []ResourceDef `json:"resources"`
	Hazards          []HazardDef   `json:"hazards"`
	Aliens           []AlienDef    `json:"aliens"`
	Mission          MissionDef    `json:"mission"`
	UnlockedBuiltins []string      `json:"unlocked_builtins"`
}

// LoadLevel reads, parses, applies defaults, and validates a level JSON file.
// Returns an error if the file cannot be read, cannot be parsed, or fails
// validation. Missing optional fields are filled with engine defaults.
func LoadLevel(path string) (*Level, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lvl Level
	if err := json.Unmarshal(data, &lvl); err != nil {
		return nil, err
	}

	// Apply engine defaults for any unset fields before validation so rules
	// that check for positive values see the resolved values, not zero.
	if lvl.World.Width == 0 {
		lvl.World.Width = 100
	}
	if lvl.World.Height == 0 {
		lvl.World.Height = 100
	}
	if lvl.Drones.InitialCount == 0 {
		lvl.Drones.InitialCount = 10
	}
	if lvl.Mission.TargetResources == 0 {
		lvl.Mission.TargetResources = simulation.DefaultMissionTargetResources
	}
	if lvl.Mission.MaxTicks == 0 {
		lvl.Mission.MaxTicks = simulation.DefaultMissionMaxTicks
	}
	if lvl.Mission.InfectionLossThreshold == 0 {
		lvl.Mission.InfectionLossThreshold = simulation.DefaultMissionInfectionLoss
	}

	if errs := lvl.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("level %q validation failed:\n  %s", lvl.ID, strings.Join(errs, "\n  "))
	}

	return &lvl, nil
}

// Validate returns a list of human-readable errors found in the level definition.
// An empty slice means the level is valid. Callers can display all errors at once
// rather than discovering them one at a time during simulation startup.
func (l *Level) Validate() []string {
	var errs []string

	// Schema version — reject future formats this binary cannot interpret.
	if l.SchemaVersion > CurrentSchemaVersion {
		errs = append(errs, fmt.Sprintf(
			"schema_version %d is not supported (this build supports up to %d); upgrade the engine",
			l.SchemaVersion, CurrentSchemaVersion,
		))
	}

	if l.ID == "" {
		errs = append(errs, "id must not be empty")
	}

	// World dimensions.
	if l.World.Width <= 0 || l.World.Height <= 0 {
		errs = append(errs, "world.width and world.height must be positive")
	}

	// Swarm population.
	if l.Drones.InitialCount <= 0 {
		errs = append(errs, "drones.initial_count must be positive")
	}

	// Mission conditions.
	if l.Mission.TargetResources <= 0 {
		errs = append(errs, "mission.target_resources must be positive")
	}
	if l.Mission.MaxTicks <= 0 {
		errs = append(errs, "mission.max_ticks must be positive")
	}
	if l.Mission.InfectionLossThreshold < 0 || l.Mission.InfectionLossThreshold > 1 {
		errs = append(errs, "mission.infection_loss_threshold must be in [0, 1]")
	}

	// Track IDs globally across all placed objects.
	seenIDs := map[string]bool{}
	checkID := func(id, context string) {
		if id == "" {
			return
		}
		if seenIDs[id] {
			errs = append(errs, fmt.Sprintf("duplicate id %q on %s", id, context))
		}
		seenIDs[id] = true
	}

	// Resources.
	for i, r := range l.Resources {
		label := fmt.Sprintf("resources[%d]", i)
		checkID(r.ID, label)
		if l.World.Width > 0 && (r.X < 0 || r.X >= l.World.Width) {
			errs = append(errs, fmt.Sprintf("%s: x=%d out of world bounds [0,%d)", label, r.X, l.World.Width))
		}
		if l.World.Height > 0 && (r.Y < 0 || r.Y >= l.World.Height) {
			errs = append(errs, fmt.Sprintf("%s: y=%d out of world bounds [0,%d)", label, r.Y, l.World.Height))
		}
		if r.Count < 0 {
			errs = append(errs, fmt.Sprintf("%s: count must be non-negative", label))
		}
	}

	// Hazards.
	for i, h := range l.Hazards {
		label := fmt.Sprintf("hazards[%d]", i)
		checkID(h.ID, label)
		if l.World.Width > 0 && (int(h.X) < 0 || int(h.X) >= l.World.Width) {
			errs = append(errs, fmt.Sprintf("%s: x=%d out of world bounds [0,%d)", label, h.X, l.World.Width))
		}
		if l.World.Height > 0 && (int(h.Y) < 0 || int(h.Y) >= l.World.Height) {
			errs = append(errs, fmt.Sprintf("%s: y=%d out of world bounds [0,%d)", label, h.Y, l.World.Height))
		}
		if h.Radius <= 0 {
			errs = append(errs, fmt.Sprintf("%s: radius must be positive", label))
		}
		if h.Intensity <= 0 {
			errs = append(errs, fmt.Sprintf("%s: intensity must be positive", label))
		}
	}

	// Alien nodes.
	for i, a := range l.Aliens {
		label := fmt.Sprintf("aliens[%d]", i)
		checkID(a.ID, label)
		if l.World.Width > 0 && (int(a.X) < 0 || int(a.X) >= l.World.Width) {
			errs = append(errs, fmt.Sprintf("%s: x=%d out of world bounds [0,%d)", label, a.X, l.World.Width))
		}
		if l.World.Height > 0 && (int(a.Y) < 0 || int(a.Y) >= l.World.Height) {
			errs = append(errs, fmt.Sprintf("%s: y=%d out of world bounds [0,%d)", label, a.Y, l.World.Height))
		}
		if a.Radius <= 0 {
			errs = append(errs, fmt.Sprintf("%s: radius must be positive", label))
		}
	}

	return errs
}

// CreateEngine builds a fully configured simulation.Engine for this level.
// Resources, hazards, and alien nodes are placed exactly as specified in the
// level JSON. The engine is ready to tick immediately.
func (l *Level) CreateEngine() *simulation.Engine {
	seed := l.World.Seed
	if seed == 0 {
		seed = simulation.DefaultWorldSeed
	}

	e := simulation.NewBaseEngineWithSeed(l.World.Width, l.World.Height, l.Drones.InitialCount, seed)

	e.Mission.TargetResources = l.Mission.TargetResources
	e.Mission.MaxTicks = l.Mission.MaxTicks
	e.Mission.InfectionLossThreshold = l.Mission.InfectionLossThreshold

	for _, r := range l.Resources {
		idx := e.Grid.GetIndex(r.X, r.Y)
		e.Grid.CurrentCells[idx].ResourceCount = r.Count
		e.Grid.NextCells[idx].ResourceCount = r.Count
	}

	for _, h := range l.Hazards {
		e.Hazards.Add(simulation.HazardType(h.Type), h.X, h.Y, h.Radius, h.Intensity)
	}

	for _, a := range l.Aliens {
		e.Aliens.Add(simulation.AlienNodeType(a.Type), a.X, a.Y, a.Radius)
	}

	return e
}
