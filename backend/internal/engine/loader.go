package engine

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadFabric reads a challenge JSON file (in ChallengeDefinition format) and
// returns a hydrated AxiomaticFabric ready for the engine.
func LoadFabric(filepath string) (*AxiomaticFabric, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read challenge file %s: %w", filepath, err)
	}

	var def ChallengeDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse challenge JSON: %w", err)
	}

	if def.ID == "" {
		return nil, fmt.Errorf("challenge missing id")
	}
	if len(def.Flaws) == 0 {
		return nil, fmt.Errorf("challenge %s has no flaws", def.ID)
	}

	return def.BuildFabric(), nil
}
