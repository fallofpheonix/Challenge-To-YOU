package missionengine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"challenge-to-you/backend/internal/db"
	"challenge-to-you/backend/internal/eventbus"

	"github.com/gorilla/websocket"
)

// newTestDB creates a temporary on-disk database for mission-engine tests.
func newTestDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.NewDB(filepath.Join(t.TempDir(), "mission_test.db"))
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestCampaignPlaythrough(t *testing.T) {
	bus := eventbus.NewEventBus(100)
	database := newTestDB(t)

	registry := NewMissionRegistry()
	mockMission := &Mission{
		ID:          "mission_magitech_act1_ch1",
		Title:       "The Fractured Ward",
		Description: "Repair the ward",
		Era:         "magitech",
		Objectives: []Objective{
			{
				ID:         "obj_complete_breach",
				Type:       "complete_challenge",
				Target:     "magitech_01_breach",
				IsRequired: true,
			},
		},
		Challenges: []MissionChallenge{
			{
				ID: "magitech_01_breach",
			},
		},
	}
	registry.Register(mockMission)
	manager := NewMissionManager(registry, bus, database)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		mSess, errStart := manager.StartMission("Intruder", "mission_magitech_act1_ch1")
		if errStart != nil {
			t.Errorf("failed to start mission: %v", errStart)
			return
		}

		snapPayload := map[string]interface{}{
			"challenge_id":   "magitech_01_breach",
			"level_complete": false,
		}
		_ = conn.WriteJSON(snapPayload)

		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var payload struct {
			Event   string `json:"event"`
			Payload string `json:"payload"`
		}
		_ = json.Unmarshal(msg, &payload)

		if payload.Event == "bypass" {
			_ = manager.CompleteChallenge("Intruder", mSess.Mission.ID, "magitech_01_breach")
			completeMission := manager.IsMissionCompleted("Intruder", mSess.Mission.ID)

			respPayload := map[string]interface{}{
				"challenge_id":   "magitech_01_breach",
				"level_complete": true,
				"message":        fmt.Sprintf("Mission Complete: %v", completeMission),
			}
			_ = conn.WriteJSON(respPayload)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	var initSnap map[string]interface{}
	errRead := conn.ReadJSON(&initSnap)
	if errRead != nil {
		t.Fatalf("failed to read: %v", errRead)
	}

	if initSnap["challenge_id"] != "magitech_01_breach" {
		t.Errorf("expected challenge magitech_01_breach, got %v", initSnap["challenge_id"])
	}

	msg := map[string]interface{}{
		"event":   "bypass",
		"payload": "",
	}
	_ = conn.WriteJSON(msg)

	var finishSnap map[string]interface{}
	_ = conn.ReadJSON(&finishSnap)

	msgText := finishSnap["message"].(string)
	if !strings.Contains(msgText, "Mission Complete: true") {
		t.Errorf("expected mission completed in output, got: %s", msgText)
	}
}
