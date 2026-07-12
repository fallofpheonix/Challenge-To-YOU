package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"challenge-to-you/backend/internal/engine"
)

func (s *Server) handleRift(w http.ResponseWriter, r *http.Request, def *engine.ChallengeDefinition, baseFabric *engine.AxiomaticFabric) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Rift connection failed:", err)
		return
	}
	defer conn.Close()

	profile, errProfile := s.db.GetOrCreateProfile()
	if errProfile != nil {
		log.Println("Failed to retrieve player profile:", errProfile)
	}

	q := r.URL.Query()
	var fabric *engine.AxiomaticFabric
	var activeDef *engine.ChallengeDefinition = def

	// It can be done without switching for each paradigm — this is a single approach:
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
				s.sendError(conn, &session{challenge: def, fabric: baseFabric}, fmt.Sprintf("LOCKED_PARADIGM: Paradigm %s is locked. Cost to unlock: %d Reputation.", paradigm, cost))
				return
			}

			genDef, genFabric := s.resolveChallenge(seed, luck, paradigm)
			if genDef != nil {
				activeDef = genDef
				fabric = genFabric
			} else {
				log.Printf("Procedural generation failed for seed=%d", seed)
			}
		}
	} else {
		if profile != nil && !profile.IsParadigmUnlocked(string(baseFabric.CurrentParadigm)) {
			cost := 50
			if baseFabric.CurrentParadigm == engine.Cosmic {
				cost = 100
			}
			s.sendError(conn, &session{challenge: def, fabric: baseFabric}, fmt.Sprintf("LOCKED_PARADIGM: Static paradigm %s is locked. Cost to unlock: %d Reputation.", baseFabric.CurrentParadigm, cost))
			return
		}
	}

	if fabric == nil {
		fabric = s.createDefaultFabric(baseFabric)
	}

	sess := &session{
		challenge: activeDef,
		fabric:    fabric,
	}

	eventChan := make(chan riftEvent, 10)
	done := make(chan struct{})

	go s.readLoop(conn, sess, eventChan, done)

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

	log.Println("New rift established.")
	s.sendSnapshotFromLoop(conn, sess, "", "Challenge loaded. Select a module to trigger.", false)

	for {
		select {
		case <-done:
			return
		case ev := <-eventChan:
			if ev.Type == eventPlayerInput {
				if ev.InputEvent == "" {
					s.sendError(conn, sess, "MISSING_EVENT: send {\"event\": \"TRIGGER_...\"}")
					continue
				}

				if ev.InputEvent == "profile" {
					if p, errDB := s.db.GetOrCreateProfile(); errDB == nil {
						profile = p
					}
					if profile != nil {
						msg := fmt.Sprintf("USER STATS: reputation=%d, luck=%.2f, unlocked_paradigms=[%s]. Command: 'unlock <paradigm>' to progress.", profile.Reputation, profile.Luck, profile.UnlockedParadigms)
						s.sendSnapshotFromLoop(conn, sess, "", msg, false)
					} else {
						s.sendSnapshotFromLoop(conn, sess, "", "Offline profile mode. SQLite connection inactive.", false)
					}
					continue
				}

				if strings.HasPrefix(strings.ToLower(ev.InputEvent), "unlock ") {
					target := strings.ToUpper(strings.TrimSpace(strings.TrimPrefix(strings.ToLower(ev.InputEvent), "unlock ")))
					if target != "CYBERPUNK" && target != "COSMIC" {
						s.sendSnapshotFromLoop(conn, sess, "", "UNLOCK_ERROR: Available paradigms to unlock: CYBERPUNK (Cost: 50 rep), COSMIC (Cost: 100 rep)", false)
						continue
					}

					if p, errDB := s.db.GetOrCreateProfile(); errDB == nil {
						profile = p
					}

					if profile == nil {
						s.sendSnapshotFromLoop(conn, sess, "", "UNLOCK_ERROR: Profile not loaded.", false)
						continue
					}

					if profile.IsParadigmUnlocked(target) {
						s.sendSnapshotFromLoop(conn, sess, "", fmt.Sprintf("UNLOCK_ERROR: Paradigm %s is already unlocked.", target), false)
						continue
					}

					cost := 50
					if target == "COSMIC" {
						cost = 100
					}

					if profile.Reputation < cost {
						s.sendSnapshotFromLoop(conn, sess, "", fmt.Sprintf("UNLOCK_ERROR: Insufficient reputation. Need %d, you have %d.", cost, profile.Reputation), false)
						continue
					}

					profile.Reputation -= cost
					profile.UnlockedParadigms += "," + target
					if errSave := s.db.SaveProfile(profile); errSave != nil {
						log.Println("Failed to save profile:", errSave)
					}

					s.sendSnapshotFromLoop(conn, sess, "", fmt.Sprintf("SUCCESS: Paradigm %s unlocked! Remaining Reputation: %d", target, profile.Reputation), false)
					continue
				}

				if ev.InputEvent == "mending_protocol" {
					entropyVal, _ := sess.fabric.GetState("entropy")
					var entropy int
					switch e := entropyVal.(type) {
					case int:
						entropy = e
					case float64:
						entropy = int(e)
					case float32:
						entropy = int(e)
					}

					sess.fabric.SetState("entropy", entropy+30)
					s.sendSnapshotFromLoop(conn, sess, "", "Mending protocol initiated. System entropy spiked by 30.", false)

					s.triggerAsyncRepairFromLoop(sess, eventChan, done)
					continue
				}

				cipher, complete, err := sess.fabric.TriggerOntologicalShift(ev.InputEvent)
				if err != nil {
					s.sendError(conn, sess, err.Error())
					s.triggerAsyncTauntFromLoop(sess, ev.InputEvent, eventChan, done)
					if sess.fabric.IsPurged() {
						return
					}
					continue
				}

				msgText := outcomeMessage(ev.InputEvent, cipher, complete)
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

				s.sendSnapshotFromLoop(conn, sess, cipher, msgText, complete)
				if !complete {
					s.triggerAsyncTauntFromLoop(sess, ev.InputEvent, eventChan, done)
				}
			} else if ev.Type == eventAIResponse {
				payload := map[string]interface{}{
					"type":    "archon_transmission",
					"message": ev.Text,
				}
				if err := conn.WriteJSON(payload); err != nil {
					log.Println("Failed to send AI response:", err)
				}
			} else if ev.Type == eventAIRepairResponse {
				for k, v := range ev.RepairPatch {
					if k == "entropy" {
						continue
					}
					sess.fabric.SetState(k, v)
				}

				complete := sess.fabric.CheckWinCondition()
				var cipher string
				msgText := "Mending execution complete. State repaired."
				if complete {
					cipher = sess.challenge.LogosToken
					msgText = "Mending achieved confluence!"
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
				s.sendSnapshotFromLoop(conn, sess, cipher, msgText, complete)
			} else if ev.Type == eventTick {
				entropyVal, hasEntropy := sess.fabric.GetState("entropy")
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
					sess.fabric.ArchonVigilance += entropy * 0.002
					if sess.fabric.ArchonVigilance > 1.0 {
						sess.fabric.ArchonVigilance = 1.0
					}

					if sess.fabric.IsPurged() {
						s.sendError(conn, sess, "ONTOLOGICAL_PURGE: The Vigilant Archon has terminated execution due to high entropy!")
						return
					}

					s.sendSnapshotFromLoop(conn, sess, "", fmt.Sprintf("Archon vigilance rising! Current entropy: %.0f", entropy), false)
				}
			}
		}
	}
}

func (s *Server) readLoop(conn interface {
	ReadMessage() (int, []byte, error)
	WriteJSON(interface{}) error
}, sess *session, eventChan chan<- riftEvent, done chan struct{}) {
	defer close(done)
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Rift severed:", err)
			return
		}

		var payload struct {
			Event string `json:"event"`
		}
		if err := json.Unmarshal(msg, &payload); err != nil {
			s.sendError(conn, sess, "MALFORMED_PAYLOAD: "+err.Error())
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
}, sess *session, cipher, message string, complete bool) {
	snap := engine.Snapshot{
		ChallengeID:   sess.challenge.ID,
		Paradigm:      string(sess.challenge.Paradigm),
		Title:         sess.challenge.Name,
		Description:   sess.challenge.Description,
		Modules:       sess.challenge.ToModules(),
		State:         sess.fabric.State,
		Vigilance:     sess.fabric.ArchonVigilance,
		Triggerable:   sess.fabric.EvaluateAllGlitches(),
		LastCipher:    cipher,
		LevelComplete: complete,
		Message:       message,
	}
	if err := conn.WriteJSON(snap); err != nil {
		log.Println("Failed to send snapshot:", err)
	}
}

func (s *Server) sendError(conn interface {
	WriteJSON(interface{}) error
}, sess *session, msg string) {
	snap := engine.Snapshot{
		ChallengeID:   sess.challenge.ID,
		Paradigm:      string(sess.challenge.Paradigm),
		Title:         "Error",
		Description:   sess.challenge.Description,
		Modules:       sess.challenge.ToModules(),
		State:         sess.fabric.State,
		Vigilance:     sess.fabric.ArchonVigilance,
		Triggerable:   sess.fabric.EvaluateAllGlitches(),
		LastCipher:    "",
		LevelComplete: false,
		ErrorMessage:  msg,
	}
	if err := conn.WriteJSON(snap); err != nil {
		log.Println("Failed to send error:", err)
	}
}

func (s *Server) triggerAsyncTauntFromLoop(sess *session, action string, eventChan chan<- riftEvent, done <-chan struct{}) {
	entropyVal, _ := sess.fabric.GetState("entropy")
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
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		taunt, err := s.oracle.GenerateTaunt(ctx, string(sess.challenge.Paradigm), action, entropy)
		if err != nil {
			log.Println("AI taunt generation skipped:", err)
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

func (s *Server) triggerAsyncRepairFromLoop(sess *session, eventChan chan<- riftEvent, done <-chan struct{}) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		patch, err := s.oracle.GenerateRepair(ctx, string(sess.challenge.Paradigm), sess.fabric.GetStateMap())
		if err != nil {
			log.Println("AI repair generation failed:", err)
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
