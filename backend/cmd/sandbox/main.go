package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"challenge-to-you/backend/internal/ai"
	"challenge-to-you/backend/internal/db"
	"challenge-to-you/backend/internal/engine"
	"challenge-to-you/backend/internal/generator"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var oracle *ai.OracleClient

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

func main() {
	oracle = ai.NewOracleClient(os.Getenv("OLLAMA_URL"), os.Getenv("OLLAMA_MODEL"))

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "challenge.db"
	}
	if err := db.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	challengePath := os.Getenv("CHALLENGE_PATH")
	if challengePath == "" {
		challengePath = filepath.Join("challenges", "magitech_01.json")
	}

	def, err := engine.LoadChallenge(challengePath)
	if err != nil {
		log.Fatalf("Failed to load challenge %s: %v", challengePath, err)
	}

	fabric := def.BuildFabric()

	http.HandleFunc("/rift", func(w http.ResponseWriter, r *http.Request) {
		handleRift(w, r, def, fabric)
	})
	log.Printf("Axiomatic Fabric online — challenge: %s (%s)", def.ID, def.Name)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRift(w http.ResponseWriter, r *http.Request, def *engine.ChallengeDefinition, baseFabric *engine.AxiomaticFabric) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Rift connection failed:", err)
		return
	}
	defer conn.Close()

	profile, errProfile := db.GetOrCreateProfile()
	if errProfile != nil {
		log.Println("Failed to retrieve player profile:", errProfile)
	}

	// Parse query parameters for dynamic generation
	q := r.URL.Query()
	var fabric *engine.AxiomaticFabric
	var activeDef *engine.ChallengeDefinition = def

	if q.Get("seed") != "" && q.Get("luck") != "" && q.Get("paradigm") != "" {
		var seed int64
		var luck float64
		var paradigm engine.Paradigm

		_, errSeed := fmt.Sscanf(q.Get("seed"), "%d", &seed)
		_, errLuck := fmt.Sscanf(q.Get("luck"), "%f", &luck)
		paradigm = engine.Paradigm(q.Get("paradigm"))

		if errSeed == nil && errLuck == nil && (paradigm == engine.Magitech || paradigm == engine.Cyberpunk || paradigm == engine.Cosmic) {
			if profile != nil && !profile.IsParadigmUnlocked(string(paradigm)) {
				cost := 50
				if paradigm == engine.Cosmic {
					cost = 100
				}
				sendError(conn, &session{challenge: def, fabric: baseFabric}, fmt.Sprintf("LOCKED_PARADIGM: Paradigm %s is locked. Cost to unlock: %d Reputation.", paradigm, cost))
				return
			}

			log.Printf("Hydrating dynamic level: seed=%d, luck=%.2f, paradigm=%s", seed, luck, paradigm)
			genDef, errGen := generator.GenerateChallenge(seed, luck, paradigm)
			if errGen == nil {
				activeDef = genDef
				fabric = genDef.BuildFabric()
			} else {
				log.Printf("Procedural generation failed: %v", errGen)
			}
		}
	} else {
		if profile != nil && !profile.IsParadigmUnlocked(string(baseFabric.CurrentParadigm)) {
			cost := 50
			if baseFabric.CurrentParadigm == engine.Cosmic {
				cost = 100
			}
			sendError(conn, &session{challenge: def, fabric: baseFabric}, fmt.Sprintf("LOCKED_PARADIGM: Static paradigm %s is locked. Cost to unlock: %d Reputation.", baseFabric.CurrentParadigm, cost))
			return
		}
	}

	if fabric == nil {
		// Fallback to static level loaded at startup
		fabric = engine.NewAxiomaticFabric(baseFabric.CurrentParadigm, baseFabric.WinConditionKey, baseFabric.WinConditionVal)
		for k, v := range baseFabric.State {
			fabric.SetState(k, v)
		}
		for _, g := range baseFabric.Glitches {
			fabric.RegisterGlitch(g)
		}
	}

	s := &session{
		challenge: activeDef,
		fabric:    fabric,
	}

	eventChan := make(chan riftEvent, 10)
	done := make(chan struct{})

	// Goroutine 1: Read client messages and push them to eventChan
	go func() {
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
				sendError(conn, s, "MALFORMED_PAYLOAD: "+err.Error())
				continue
			}

			eventChan <- riftEvent{
				Type:       eventPlayerInput,
				InputEvent: payload.Event,
			}
		}
	}()

	// Goroutine 2: Ticker pushing tick events
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
					// Drop tick if queue is full
				}
			}
		}
	}()

	log.Println("New rift established.")
	sendSnapshot(conn, s, "", "Challenge loaded. Select a module to trigger.", false)

	// Main event loop (single thread, serializes all fabric state access!)
	for {
		select {
		case <-done:
			return
		case ev := <-eventChan:
			if ev.Type == eventPlayerInput {
				if ev.InputEvent == "" {
					sendError(conn, s, "MISSING_EVENT: send {\"event\": \"TRIGGER_...\"}")
					continue
				}

				if ev.InputEvent == "profile" {
					if p, errDB := db.GetOrCreateProfile(); errDB == nil {
						profile = p
					}
					if profile != nil {
						msg := fmt.Sprintf("USER STATS: reputation=%d, luck=%.2f, unlocked_paradigms=[%s]. Command: 'unlock <paradigm>' to progress.", profile.Reputation, profile.Luck, profile.UnlockedParadigms)
						sendSnapshot(conn, s, "", msg, false)
					} else {
						sendSnapshot(conn, s, "", "Offline profile mode. SQLite connection inactive.", false)
					}
					continue
				}

				if strings.HasPrefix(strings.ToLower(ev.InputEvent), "unlock ") {
					target := strings.ToUpper(strings.TrimSpace(strings.TrimPrefix(strings.ToLower(ev.InputEvent), "unlock ")))
					if target != "CYBERPUNK" && target != "COSMIC" {
						sendSnapshot(conn, s, "", "UNLOCK_ERROR: Available paradigms to unlock: CYBERPUNK (Cost: 50 rep), COSMIC (Cost: 100 rep)", false)
						continue
					}

					// Re-fetch database record to avoid stale memory
					if p, errDB := db.GetOrCreateProfile(); errDB == nil {
						profile = p
					}

					if profile == nil {
						sendSnapshot(conn, s, "", "UNLOCK_ERROR: Profile not loaded.", false)
						continue
					}

					if profile.IsParadigmUnlocked(target) {
						sendSnapshot(conn, s, "", fmt.Sprintf("UNLOCK_ERROR: Paradigm %s is already unlocked.", target), false)
						continue
					}

					cost := 50
					if target == "COSMIC" {
						cost = 100
					}

					if profile.Reputation < cost {
						sendSnapshot(conn, s, "", fmt.Sprintf("UNLOCK_ERROR: Insufficient reputation. Need %d, you have %d.", cost, profile.Reputation), false)
						continue
					}

					profile.Reputation -= cost
					profile.UnlockedParadigms += "," + target
					if errSave := db.SaveProfile(profile); errSave != nil {
						log.Println("Failed to save profile:", errSave)
					}

					sendSnapshot(conn, s, "", fmt.Sprintf("SUCCESS: Paradigm %s unlocked! Remaining Reputation: %d", target, profile.Reputation), false)
					continue
				}

				if ev.InputEvent == "mending_protocol" {
					entropyVal, _ := s.fabric.GetState("entropy")
					var entropy int
					switch e := entropyVal.(type) {
					case int:
						entropy = e
					case float64:
						entropy = int(e)
					case float32:
						entropy = int(e)
					}

					newEntropy := entropy + 30
					s.fabric.SetState("entropy", newEntropy)
					sendSnapshot(conn, s, "", "Mending protocol initiated. System entropy spiked by 30.", false)

					triggerAsyncRepair(s, eventChan, done)
					continue
				}

				cipher, complete, err := s.fabric.TriggerOntologicalShift(ev.InputEvent)
				if err != nil {
					sendError(conn, s, err.Error())
					triggerAsyncTaunt(s, ev.InputEvent, eventChan, done)
					if s.fabric.IsPurged() {
						return
					}
					continue
				}

				msgText := outcomeMessage(ev.InputEvent, cipher, complete)
				if complete && cipher != "" && profile != nil {
					isNew, errToken := db.RecordToken(cipher)
					if errToken == nil {
						if isNew {
							profile.Reputation += 20
							profile.Luck += 0.05
							if profile.Luck > 2.0 {
								profile.Luck = 2.0
							}
							if errSave := db.SaveProfile(profile); errSave == nil {
								msgText += fmt.Sprintf(" [NEW LOGOS CIPHER EXTRACTED: +20 Reputation, Luck improved! (Total Rep: %d)]", profile.Reputation)
							}
						} else {
							msgText += " [Logos cipher already registered. Reputation unchanged.]"
						}
					}
				}

				sendSnapshot(conn, s, cipher, msgText, complete)
				if !complete {
					triggerAsyncTaunt(s, ev.InputEvent, eventChan, done)
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
					s.fabric.SetState(k, v)
				}

				complete := s.fabric.CheckWinCondition()
				var cipher string
				msgText := "Mending execution complete. State repaired."
				if complete {
					cipher = s.challenge.LogosToken
					msgText = "Mending achieved confluence!"
					if cipher != "" && profile != nil {
						isNew, errToken := db.RecordToken(cipher)
						if errToken == nil {
							if isNew {
								profile.Reputation += 20
								profile.Luck += 0.05
								if profile.Luck > 2.0 {
									profile.Luck = 2.0
								}
								if errSave := db.SaveProfile(profile); errSave == nil {
									msgText += fmt.Sprintf(" [NEW LOGOS CIPHER EXTRACTED: +20 Reputation (Total Rep: %d)]", profile.Reputation)
								}
							} else {
								msgText += " [Logos cipher already registered. Reputation unchanged.]"
							}
						}
					}
				}
				sendSnapshot(conn, s, cipher, msgText, complete)
			} else if ev.Type == eventTick {
				// Read current entropy from fabric state
				entropyVal, hasEntropy := s.fabric.GetState("entropy")
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

				// If entropy > 0, the Archon's gaze intensifies
				if entropy > 0 {
					// Increase vigilance based on entropy
					s.fabric.ArchonVigilance += entropy * 0.002
					if s.fabric.ArchonVigilance > 1.0 {
						s.fabric.ArchonVigilance = 1.0
					}

					if s.fabric.IsPurged() {
						sendError(conn, s, "ONTOLOGICAL_PURGE: The Vigilant Archon has terminated execution due to high entropy!")
						return
					}

					sendSnapshot(conn, s, "", fmt.Sprintf("Archon vigilance rising! Current entropy: %.0f", entropy), false)
				}
			}
		}
	}
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

func sendSnapshot(conn *websocket.Conn, s *session, cipher, message string, complete bool) {
	snap := engine.Snapshot{
		ChallengeID:   s.challenge.ID,
		Paradigm:      string(s.challenge.Paradigm),
		Title:         s.challenge.Name,
		Description:   s.challenge.Description,
		Modules:       s.challenge.ToModules(),
		State:         s.fabric.State,
		Vigilance:     s.fabric.ArchonVigilance,
		Triggerable:   s.fabric.EvaluateAllGlitches(),
		LastCipher:    cipher,
		LevelComplete: complete,
		Message:       message,
	}
	if err := conn.WriteJSON(snap); err != nil {
		log.Println("Failed to send snapshot:", err)
	}
}

func sendError(conn *websocket.Conn, s *session, msg string) {
	snap := engine.Snapshot{
		ChallengeID:   s.challenge.ID,
		Paradigm:      string(s.challenge.Paradigm),
		Title:         "Error",
		Description:   s.challenge.Description,
		Modules:       s.challenge.ToModules(),
		State:         s.fabric.State,
		Vigilance:     s.fabric.ArchonVigilance,
		Triggerable:   s.fabric.EvaluateAllGlitches(),
		LastCipher:    "",
		LevelComplete: false,
		ErrorMessage:  msg,
	}
	if err := conn.WriteJSON(snap); err != nil {
		log.Println("Failed to send error:", err)
	}
}

func triggerAsyncTaunt(s *session, action string, eventChan chan riftEvent, done chan struct{}) {
	entropyVal, _ := s.fabric.GetState("entropy")
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

		taunt, err := oracle.GenerateTaunt(ctx, string(s.challenge.Paradigm), action, entropy)
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
			// Queue full
		}
	}()
}

func triggerAsyncRepair(s *session, eventChan chan riftEvent, done chan struct{}) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		patch, err := oracle.GenerateRepair(ctx, string(s.challenge.Paradigm), s.fabric.GetStateMap())
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
			// Drop
		}
	}()
}