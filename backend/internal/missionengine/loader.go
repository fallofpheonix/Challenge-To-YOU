package missionengine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MissionRegistry stores all loaded missions
type MissionRegistry struct {
	missions map[string]*Mission
	byEra    map[string][]string
	byAct    map[string]map[int][]string
}

// NewMissionRegistry creates an empty registry
func NewMissionRegistry() *MissionRegistry {
	return &MissionRegistry{
		missions: make(map[string]*Mission),
		byEra:    make(map[string][]string),
		byAct:    make(map[string]map[int][]string),
	}
}

// LoadMission loads a single mission from a JSON file
func LoadMission(path string) (*Mission, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read mission file %s: %w", path, err)
	}

	mission := &Mission{}
	if err := json.Unmarshal(data, mission); err != nil {
		return nil, fmt.Errorf("failed to parse mission file %s: %w", path, err)
	}

	if mission.ID == "" {
		return nil, fmt.Errorf("mission missing id in %s", path)
	}
	if mission.Title == "" {
		return nil, fmt.Errorf("mission %s missing title", mission.ID)
	}

	// Set defaults
	for i := range mission.Objectives {
		if mission.Objectives[i].IsRequired && !mission.Objectives[i].IsOptional {
			continue
		}
		if !mission.Objectives[i].IsOptional && !mission.Objectives[i].IsRequired {
			mission.Objectives[i].IsRequired = true
		}
	}

	return mission, nil
}

// LoadMissionsFromDir loads all missions from a directory
func (r *MissionRegistry) LoadMissionsFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read mission directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Recurse into subdirectories
			subDir := filepath.Join(dir, entry.Name())
			if err := r.LoadMissionsFromDir(subDir); err != nil {
				fmt.Printf("Warning: skipping subdirectory %s: %v\n", subDir, err)
			}
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		mission, err := LoadMission(path)
		if err != nil {
			fmt.Printf("Warning: skipping %s: %v\n", path, err)
			continue
		}

		r.Register(mission)
	}

	return nil
}

// Register adds a mission to the registry
func (r *MissionRegistry) Register(mission *Mission) {
	r.missions[mission.ID] = mission

	// Index by era
	r.byEra[mission.Era] = append(r.byEra[mission.Era], mission.ID)

	// Index by era + act
	if r.byAct[mission.Era] == nil {
		r.byAct[mission.Era] = make(map[int][]string)
	}
	r.byAct[mission.Era][mission.Act] = append(r.byAct[mission.Era][mission.Act], mission.ID)
}

// Get returns a mission by ID
func (r *MissionRegistry) Get(id string) (*Mission, bool) {
	mission, ok := r.missions[id]
	return mission, ok
}

// GetAll returns all missions
func (r *MissionRegistry) GetAll() []*Mission {
	result := make([]*Mission, 0, len(r.missions))
	for _, m := range r.missions {
		result = append(result, m)
	}
	return result
}

// GetByEra returns all missions for an era
func (r *MissionRegistry) GetByEra(era string) []*Mission {
	ids := r.byEra[era]
	result := make([]*Mission, 0, len(ids))
	for _, id := range ids {
		if m, ok := r.missions[id]; ok {
			result = append(result, m)
		}
	}
	return result
}

// GetByAct returns all missions for an era and act
func (r *MissionRegistry) GetByAct(era string, act int) []*Mission {
	eraAct, ok := r.byAct[era]
	if !ok {
		return nil
	}
	ids := eraAct[act]
	result := make([]*Mission, 0, len(ids))
	for _, id := range ids {
		if m, ok := r.missions[id]; ok {
			result = append(result, m)
		}
	}
	return result
}

// GetAvailable returns missions that can be started given completed missions
func (r *MissionRegistry) GetAvailable(completedMissions map[string]bool) []*Mission {
	var available []*Mission
	for _, m := range r.missions {
		if CanStart(m, completedMissions) {
			available = append(available, m)
		}
	}
	return available
}

// GetNextMission returns the next mission in a campaign
func (r *MissionRegistry) GetNextMission(currentID string) (*Mission, bool) {
	if _, ok := r.missions[currentID]; !ok {
		return nil, false
	}

	// Find missions that require the current mission
	for _, m := range r.missions {
		for _, req := range m.Requirements {
			if req == currentID {
				return m, true
			}
		}
	}

	return nil, false
}

// GetMissionsByDifficulty returns missions sorted by difficulty
func (r *MissionRegistry) GetMissionsByDifficulty(minDiff, maxDiff float64) []*Mission {
	var result []*Mission
	for _, m := range r.missions {
		if m.Difficulty >= minDiff && m.Difficulty <= maxDiff {
			result = append(result, m)
		}
	}
	return result
}

// GetMissionChain returns the full chain of missions from a starting point
func (r *MissionRegistry) GetMissionChain(startID string) []*Mission {
	var chain []*Mission
	visited := make(map[string]bool)

	current := startID
	for current != "" {
		if visited[current] {
			break
		}
		visited[current] = true

		mission, ok := r.missions[current]
		if !ok {
			break
		}

		chain = append(chain, mission)

		// Find next mission
		next, _ := r.GetNextMission(current)
		if next != nil {
			current = next.ID
		} else {
			break
		}
	}

	return chain
}

// Count returns the total number of loaded missions
func (r *MissionRegistry) Count() int {
	return len(r.missions)
}

// Eras returns all eras with missions
func (r *MissionRegistry) Eras() []string {
	eras := make([]string, 0, len(r.byEra))
	for era := range r.byEra {
		eras = append(eras, era)
	}
	return eras
}
