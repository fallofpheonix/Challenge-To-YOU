package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"challenge-to-you/backend/internal/engine"
	"challenge-to-you/backend/internal/obs"

	"github.com/gorilla/websocket"
)

func (s *Server) handleRift(ctx context.Context, r *http.Request, conn *websocket.Conn, def *engine.ChallengeDefinition, baseFabric *engine.AxiomaticFabric) {
	rlog := s.log.WithContext(ctx)

	profile, errProfile := s.db.GetOrCreateProfile()
	if errProfile != nil {
		rlog.Error("profile load failed", "error", errProfile, "class", obs.ClassPersistence)
	}

	q := r.URL.Query()
	var fabric *engine.AxiomaticFabric
	var activeDef *engine.ChallengeDefinition = def

	if q.Get("seed") != "" && q.Get("luck") != "" && q.Get("paradigm") != "" {
		var seed int64
		var luck float64
		paradigm := engine.Paradigm(q.Get("paradigm"))

		_, errSeed := fmt.Sscanf(q.Get("seed"), "%d", &seed)
		_, errLuck := fmt.Sscanf(q.Get("luck"), "%f", &luck)

		if errSeed == nil && errLuck == nil && (paradigm == engine.Magitech || paradigm == engine.Cyberpunk || paradigm == engine.Cosmic) {
			if profile != nil && !profile.IsParadigmUnlocked(string(paradigm)) {
				cost := 50
				if paradigm == engine.Cosmic {
					cost = 100
				}
				s.sendError(conn, def, baseFabric, fmt.Sprintf("LOCKED_PARADIGM: Paradigm %s is locked. Cost to unlock: %d Reputation.", paradigm, cost))
				return
			}

			genDef, genFabric := s.resolveChallenge(seed, luck, paradigm)
			if genDef != nil {
				activeDef = genDef
				fabric = genFabric
			} else {
				rlog.Warn("procedural generation failed", "seed", seed)
			}
		}
	} else {
		if profile != nil && !profile.IsParadigmUnlocked(string(baseFabric.CurrentParadigm)) {
			cost := 50
			if baseFabric.CurrentParadigm == engine.Cosmic {
				cost = 100
			}
			s.sendError(conn, def, baseFabric, fmt.Sprintf("LOCKED_PARADIGM: Static paradigm %s is locked. Cost to unlock: %d Reputation.", baseFabric.CurrentParadigm, cost))
			return
		}
	}

	if fabric == nil {
		fabric = s.createDefaultFabric(baseFabric)
	}

	sess := s.sessionManager.Create("", activeDef, fabric)
	defer s.sessionManager.Remove(sess.ID)

	ctx = obs.WithSessionID(ctx, string(sess.ID))
	rlog = s.log.WithContext(ctx)

	eventChan := make(chan riftEvent, 10)
	done := make(chan struct{})

	go s.readLoop(conn, activeDef, fabric, eventChan, done)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				select {
				case eventChan <- riftEvent{Type: eventTick}:
				default:
				}
			}
		}
	}()

	rlog.Lifecycle("SessionCreated", "session_id", string(sess.ID), "challenge_id", activeDef.ID)
	s.metrics.ActiveSessions.Add(1)
	defer s.metrics.ActiveSessions.Add(-1)

	s.sessionManager.Touch(sess.ID)
	s.sendSnapshotFromLoop(conn, activeDef, fabric, "", "Challenge loaded. Select a module to trigger.", false)

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case ev := <-eventChan:
			if ev.Type == eventPlayerInput {
				s.sessionManager.Touch(sess.ID)
				if ev.InputEvent == "" {
					s.sendError(conn, activeDef, fabric, "MISSING_EVENT: send {\"event\": \"TRIGGER_...\"}")
					continue
				}

				if ev.InputEvent == "profile" {
					if p, errDB := s.db.GetOrCreateProfile(); errDB == nil {
						profile = p
					}
					if profile != nil {
						msg := fmt.Sprintf("USER STATS: reputation=%d, luck=%.2f, unlocked_paradigms=[%s]. Command: 'unlock <paradigm>' to progress.", profile.Reputation, profile.Luck, profile.UnlockedParadigms)
						s.sendSnapshotFromLoop(conn, activeDef, fabric, "", msg, false)
					} else {
						s.sendSnapshotFromLoop(conn, activeDef, fabric, "", "Offline profile mode. SQLite connection inactive.", false)
					}
					continue
				}

				if strings.HasPrefix(strings.ToLower(ev.InputEvent), "unlock ") {
					target := strings.ToUpper(strings.TrimSpace(strings.TrimPrefix(strings.ToLower(ev.InputEvent), "unlock ")))
					if target != "CYBERPUNK" && target != "COSMIC" {
						s.sendSnapshotFromLoop(conn, activeDef, fabric, "", "UNLOCK_ERROR: Available paradigms to unlock: CYBERPUNK (Cost: 50 rep), COSMIC (Cost: 100 rep)", false)
						continue
					}

					if p, errDB := s.db.GetOrCreateProfile(); errDB == nil {
						profile = p
					}

					if profile == nil {
						s.sendSnapshotFromLoop(conn, activeDef, fabric, "", "UNLOCK_ERROR: Profile not loaded.", false)
						continue
					}

					if profile.IsParadigmUnlocked(target) {
						s.sendSnapshotFromLoop(conn, activeDef, fabric, "", fmt.Sprintf("UNLOCK_ERROR: Paradigm %s is already unlocked.", target), false)
						continue
					}

					cost := 50
					if target == "COSMIC" {
						cost = 100
					}

					if profile.Reputation < cost {
						s.sendSnapshotFromLoop(conn, activeDef, fabric, "", fmt.Sprintf("UNLOCK_ERROR: Insufficient reputation. Need %d, you have %d.", cost, profile.Reputation), false)
						continue
					}

					profile.Reputation -= cost
					profile.UnlockedParadigms += "," + target
					if errSave := s.db.SaveProfile(profile); errSave != nil {
						rlog.Error("profile save failed", "error", errSave, "class", obs.ClassPersistence)
					}

					s.sendSnapshotFromLoop(conn, activeDef, fabric, "", fmt.Sprintf("SUCCESS: Paradigm %s unlocked! Remaining Reputation: %d", target, profile.Reputation), false)
					continue
				}

				if ev.InputEvent == "mending_protocol" {
					entropyVal, _ := fabric.GetState("entropy")
					var entropy int
					switch e := entropyVal.(type) {
					case int:
						entropy = e
					case float64:
						entropy = int(e)
					case float32:
						entropy = int(e)
					}

					fabric.SetState("entropy", entropy+30)
					s.sendSnapshotFromLoop(conn, activeDef, fabric, "", "Mending protocol initiated. System entropy spiked by 30.", false)

					s.triggerAsyncRepairFromLoop(ctx, activeDef, fabric, eventChan, done)
					continue
				}

				cipher, complete, err := fabric.TriggerOntologicalShift(ev.InputEvent)
				if err != nil {
					s.sendError(conn, activeDef, fabric, err.Error())
					s.triggerAsyncTauntFromLoop(ctx, activeDef, fabric, ev.InputEvent, eventChan, done)
					if fabric.IsPurged() {
						return
					}
					continue
				}

				msgText := outcomeMessage(ev.InputEvent, cipher, complete)
				if complete {
					rlog.Lifecycle("ChallengeSolved", "cipher", cipher, "event", ev.InputEvent)
				}
				if complete && cipher != "" && profile != nil {
					isNew, errToken := s.db.RecordToken(cipher)
					if errToken == nil {
						if isNew {
							profile.Reputation += 20
							profile.Luck += 0.05
							if profile.Luck > 2.0 {
								profile.Luck = 2.0
							}
							if errSave := s.db.SaveProfile(profile); errSave == nil {
								msgText += fmt.Sprintf(" [NEW LOGOS CIPHER EXTRACTED: +20 Reputation, Luck improved! (Total Rep: %d)]", profile.Reputation)
							}
						} else {
							msgText += " [Logos cipher already registered. Reputation unchanged.]"
						}
					}
				}

				s.sendSnapshotFromLoop(conn, activeDef, fabric, cipher, msgText, complete)
				if !complete {
					s.triggerAsyncTauntFromLoop(ctx, activeDef, fabric, ev.InputEvent, eventChan, done)
				}
			} else if ev.Type == eventAIResponse {
				payload := map[string]interface{}{
					"type":    "archon_transmission",
					"message": ev.Text,
				}
				if err := conn.WriteJSON(payload); err != nil {
					rlog.Error("AI response send failed", "error", err, "class", obs.ClassTransport)
				}
			} else if ev.Type == eventAIRepairResponse {
				for k, v := range ev.RepairPatch {
					if k == "entropy" {
						continue
					}
					fabric.SetState(k, v)
				}

				complete := fabric.CheckWinCondition()
				var cipher string
				msgText := "Mending execution complete. State repaired."
				if complete {
					cipher = activeDef.LogosToken
					msgText = "Mending achieved confluence!"
					rlog.Lifecycle("ChallengeSolved", "cipher", cipher, "event", "mending_protocol")
					if cipher != "" && profile != nil {
						isNew, errToken := s.db.RecordToken(cipher)
						if errToken == nil {
							if isNew {
								profile.Reputation += 20
								profile.Luck += 0.05
								if profile.Luck > 2.0 {
									profile.Luck = 2.0
								}
								if errSave := s.db.SaveProfile(profile); errSave == nil {
									msgText += fmt.Sprintf(" [NEW LOGOS CIPHER EXTRACTED: +20 Reputation (Total Rep: %d)]", profile.Reputation)
								}
							} else {
								msgText += " [Logos cipher already registered. Reputation unchanged.]"
							}
						}
					}
				}
				s.sendSnapshotFromLoop(conn, activeDef, fabric, cipher, msgText, complete)
			} else if ev.Type == eventTick {
				s.metrics.Ticks.Add(1)
				entropyVal, hasEntropy := fabric.GetState("entropy")
				var entropy float64
				if hasEntropy {
					switch e := entropyVal.(type) {
					case int:
						entropy = float64(e)
					case float64:
						entropy = e
					case float32:
						entropy = float64(e)
					}
				}

				if entropy > 0 {
					fabric.ArchonVigilance += entropy * 0.002
					if fabric.ArchonVigilance > 1.0 {
						fabric.ArchonVigilance = 1.0
					}

					if fabric.IsPurged() {
						s.sendError(conn, activeDef, fabric, "ONTOLOGICAL_PURGE: The Vigilant Archon has terminated execution due to high entropy!")
						return
					}

					s.sendSnapshotFromLoop(conn, activeDef, fabric, "", fmt.Sprintf("Archon vigilance rising! Current entropy: %.0f", entropy), false)
				}
			}
		}
	}
}

func (s *Server) readLoop(conn interface {
	ReadMessage() (int, []byte, error)
	WriteJSON(interface{}) error
}, def *engine.ChallengeDefinition, fabric *engine.AxiomaticFabric, eventChan chan<- riftEvent, done chan struct{}) {
	defer close(done)
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var payload struct {
			Event string `json:"event"`
		}
		if err := json.Unmarshal(msg, &payload); err != nil {
			s.sendError(conn, def, fabric, "MALFORMED_PAYLOAD: "+err.Error())
			continue
		}

		eventChan <- riftEvent{
			Type:       eventPlayerInput,
			InputEvent: payload.Event,
		}
	}
}

func (s *Server) sendSnapshotFromLoop(conn interface {
	WriteJSON(interface{}) error
}, def *engine.ChallengeDefinition, fabric *engine.AxiomaticFabric, cipher, message string, complete bool) {
	snap := engine.Snapshot{
		ChallengeID:   def.ID,
		Paradigm:      string(def.Paradigm),
		Title:         def.Name,
		Description:   def.Description,
		Modules:       def.ToModules(),
		State:         fabric.State,
		Vigilance:     fabric.ArchonVigilance,
		Triggerable:   fabric.EvaluateAllGlitches(),
		LastCipher:    cipher,
		LevelComplete: complete,
		Message:       message,
	}
	if err := conn.WriteJSON(snap); err != nil {
		s.log.Error("snapshot send failed", "error", err, "class", obs.ClassTransport)
	}
}

func (s *Server) sendError(conn interface {
	WriteJSON(interface{}) error
}, def *engine.ChallengeDefinition, fabric *engine.AxiomaticFabric, msg string) {
	snap := engine.Snapshot{
		ChallengeID:   def.ID,
		Paradigm:      string(def.Paradigm),
		Title:         "Error",
		Description:   def.Description,
		Modules:       def.ToModules(),
		State:         fabric.State,
		Vigilance:     fabric.ArchonVigilance,
		Triggerable:   fabric.EvaluateAllGlitches(),
		LastCipher:    "",
		LevelComplete: false,
		ErrorMessage:  msg,
	}
	if err := conn.WriteJSON(snap); err != nil {
		s.log.Error("error snapshot send failed", "error", err, "class", obs.ClassTransport)
	}
}

func (s *Server) triggerAsyncTauntFromLoop(ctx context.Context, def *engine.ChallengeDefinition, fabric *engine.AxiomaticFabric, action string, eventChan chan<- riftEvent, done <-chan struct{}) {
	entropyVal, _ := fabric.GetState("entropy")
	var entropy int
	switch e := entropyVal.(type) {
	case int:
		entropy = e
	case float64:
		entropy = int(e)
	case float32:
		entropy = int(e)
	}

	go func() {
		s.metrics.OracleRequests.Add(1)
		defer s.metrics.OracleRequests.Add(-1)

		tctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		taunt, err := s.oracle.GenerateTaunt(tctx, string(def.Paradigm), action, entropy)
		if err != nil {
			s.log.Warn("AI taunt skipped", "error", err, "class", obs.ClassAI)
			return
		}

		select {
		case <-done:
			return
		case eventChan <- riftEvent{
			Type: eventAIResponse,
			Text: taunt,
		}:
		default:
		}
	}()
}

func (s *Server) triggerAsyncRepairFromLoop(ctx context.Context, def *engine.ChallengeDefinition, fabric *engine.AxiomaticFabric, eventChan chan<- riftEvent, done <-chan struct{}) {
	go func() {
		s.metrics.OracleRequests.Add(1)
		defer s.metrics.OracleRequests.Add(-1)

		tctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		patch, err := s.oracle.GenerateRepair(tctx, string(def.Paradigm), fabric.GetStateMap())
		if err != nil {
			s.log.Error("AI repair failed", "error", err, "class", obs.ClassAI)
			return
		}

		select {
		case <-done:
			return
		case eventChan <- riftEvent{
			Type:        eventAIRepairResponse,
			RepairPatch: patch,
		}:
		default:
		}
	}()
}
