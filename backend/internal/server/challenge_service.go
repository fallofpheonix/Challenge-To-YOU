package server

import (
	"challenge-to-you/backend/internal/engine"
	"challenge-to-you/backend/internal/generator"
)

type session struct {
	challenge *engine.ChallengeDefinition
	fabric    *engine.AxiomaticFabric
}

type eventType int

const (
	eventPlayerInput eventType = iota
	eventTick
	eventAIResponse
	eventAIRepairResponse
)

type riftEvent struct {
	Type        eventType
	InputEvent  string
	Text        string
	RepairPatch map[string]interface{}
}

func (s *Server) resolveChallenge(seed int64, luck float64, paradigm engine.Paradigm) (*engine.ChallengeDefinition, *engine.AxiomaticFabric) {
	genDef, errGen := generator.GenerateChallenge(seed, luck, paradigm)
	if errGen != nil {
		return nil, nil
	}
	return genDef, genDef.BuildFabric()
}

func (s *Server) createDefaultFabric(baseFabric *engine.AxiomaticFabric) *engine.AxiomaticFabric {
	fabric := engine.NewAxiomaticFabric(baseFabric.CurrentParadigm, baseFabric.WinConditionKey, baseFabric.WinConditionVal)
	for k, v := range baseFabric.State {
		fabric.SetState(k, v)
	}
	for _, g := range baseFabric.Glitches {
		fabric.RegisterGlitch(g)
	}
	return fabric
}

func outcomeMessage(event, cipher string, complete bool) string {
	if complete && cipher != "" {
		return "Confluence achieved. Passcode: " + cipher
	}
	if cipher != "" {
		return "Partial confluence. Cipher fragment: " + cipher
	}
	return "Event " + event + " processed. Conditions may not be met."
}
