package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PackDefinition represents a collection of challenges (a "chapter" or "tier").
type PackDefinition struct {
	Version       int            `json:"version"`
	Era           string         `json:"era"`
	Tier          int            `json:"tier"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Prerequisites []string       `json:"prerequisites,omitempty"`
	Unlocks       []string       `json:"unlocks,omitempty"`
	Challenges    []ChallengeRef `json:"challenges"`
	LuckModifier  float64        `json:"luck_modifier,omitempty"`
}

// ChallengeRef is a reference to a challenge file within the pack.
type ChallengeRef struct {
	ID         string  `json:"id"`
	File       string  `json:"file"`
	Difficulty float64 `json:"difficulty"`
	Mode       string  `json:"mode,omitempty"` // "architect", "ghost", "saboteur"
}

// LoadPack loads a challenge pack from a directory.
func LoadPack(packPath string) (*PackDefinition, error) {
	// Read pack.json
	packFile := filepath.Join(packPath, "pack.json")
	data, err := os.ReadFile(packFile)
	if err != nil {
		return nil, fmt.Errorf("read pack.json: %w", err)
	}

	var pack PackDefinition
	if err := json.Unmarshal(data, &pack); err != nil {
		return nil, fmt.Errorf("parse pack.json: %w", err)
	}

	if pack.Version != 1 {
		return nil, fmt.Errorf("unsupported pack version: %d", pack.Version)
	}

	// Load each challenge
	var loadedChallenges []*ChallengeDefinition
	for _, ref := range pack.Challenges {
		challengePath := filepath.Join(packPath, ref.File)
		def, err := LoadChallenge(challengePath)
		if err != nil {
			return nil, fmt.Errorf("load challenge %s (%s): %w", ref.ID, ref.File, err)
		}
		if def.ID != ref.ID {
			return nil, fmt.Errorf("challenge ID mismatch: pack says %s, file has %s", ref.ID, def.ID)
		}
		loadedChallenges = append(loadedChallenges, def) //nolint:staticcheck // retained for validation side-effects during pack load
	}

	return &pack, nil
}

// LoadChallengePack is an alias for LoadPack for clarity.
func LoadChallengePack(packPath string) (*PackDefinition, error) {
	return LoadPack(packPath)
}

// ChallengePack is the runtime representation of a loaded pack.
type ChallengePack struct {
	Definition *PackDefinition
	Challenges map[string]*ChallengeDefinition // id -> challenge
	Fabrics    map[string]*AxiomaticFabric     // id -> fabric
}

// NewChallengePack creates a runtime pack from a definition.
func NewChallengePack(def *PackDefinition, packPath string) (*ChallengePack, error) {
	pack := &ChallengePack{
		Definition: def,
		Challenges: make(map[string]*ChallengeDefinition),
		Fabrics:    make(map[string]*AxiomaticFabric),
	}

	for _, ref := range def.Challenges {
		challengePath := filepath.Join(packPath, ref.File)
		// We already loaded and validated in LoadPack, but rebuild fabrics here
		def, err := LoadChallenge(challengePath)
		if err != nil {
			return nil, fmt.Errorf("load challenge %s: %w", ref.ID, err)
		}
		pack.Challenges[def.ID] = def
		pack.Fabrics[def.ID] = def.BuildFabric()
	}

	return pack, nil
}

// GetChallenge returns a challenge by ID.
func (p *ChallengePack) GetChallenge(id string) (*ChallengeDefinition, bool) {
	c, ok := p.Challenges[id]
	return c, ok
}

// GetFabric returns a fresh fabric for a challenge by ID.
func (p *ChallengePack) GetFabric(id string) (*AxiomaticFabric, bool) {
	f, ok := p.Fabrics[id]
	if !ok {
		return nil, false
	}
	// Return a copy so each playthrough is independent
	fabric := NewAxiomaticFabric(f.CurrentParadigm, f.WinConditionKey, f.WinConditionVal)
	for k, v := range f.State {
		fabric.SetState(k, v)
	}
	for _, g := range f.Glitches {
		fabric.RegisterGlitch(g)
	}
	return fabric, true
}

// ListChallenges returns all challenge IDs in order.
func (p *ChallengePack) ListChallenges() []string {
	var ids []string
	for _, ref := range p.Definition.Challenges {
		ids = append(ids, ref.ID)
	}
	return ids
}

// GetAvailableChallenges returns challenge IDs whose prerequisites are met.
// This is the runtime dependency check.
func (p *ChallengePack) GetAvailableChallenges(completedIDs map[string]bool) []string {
	var available []string
	for _, ref := range p.Definition.Challenges {
		if completedIDs[ref.ID] {
			continue // already completed
		}
		allMet := true
		for _, prereq := range p.Definition.Prerequisites {
			if !completedIDs[prereq] {
				allMet = false
				break
			}
		}
		if allMet {
			available = append(available, ref.ID)
		}
	}
	return available
}
