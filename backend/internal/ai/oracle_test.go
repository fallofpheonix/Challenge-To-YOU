package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"challenge-to-you/backend/internal/engine"
)

func TestOracleGenerateTaunt(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("Expected path '/api/generate', got: %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json'")
		}

		var req generateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if req.Model != "test-model" {
			t.Errorf("Expected model 'test-model', got: %s", req.Model)
		}
		if req.Stream != false {
			t.Errorf("Expected stream to be false")
		}

		resp := generateResponse{
			Response: "You have triggered my security matrix. Surrender.",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := NewOracleClient(mockServer.URL, "test-model")
	taunt, err := client.GenerateTaunt(context.Background(), "CYBERPUNK", "test_trigger", 15)
	if err != nil {
		t.Fatalf("GenerateTaunt failed: %v", err)
	}

	expected := "You have triggered my security matrix. Surrender."
	if taunt != expected {
		t.Errorf("Expected taunt %q, got: %q", expected, taunt)
	}
}

func TestOracleGenerateRepair(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req generateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode: %v", err)
		}

		if req.Format != "json" {
			t.Errorf("Expected Format parameter to be 'json'")
		}

		resp := generateResponse{
			Response: `{"entropy": 0, "safety_lock": false}`,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := NewOracleClient(mockServer.URL, "test-model")
	state := map[string]interface{}{"entropy": 40, "safety_lock": true}
	repair, err := client.GenerateRepair(context.Background(), "MAGITECH", state)
	if err != nil {
		t.Fatalf("GenerateRepair failed: %v", err)
	}

	if val, ok := repair["entropy"]; !ok || val.(float64) != 0 {
		t.Errorf("Expected repair to set entropy to 0, got: %v", repair["entropy"])
	}
}

func TestOracleTimeout(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second) // Force a sleep longer than HTTP client timeout
		resp := generateResponse{Response: "Too slow"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := NewOracleClient(mockServer.URL, "test-model")
	client.HTTPClient.Timeout = 50 * time.Millisecond // Set extremely low timeout to test

	_, err := client.GenerateTaunt(context.Background(), "CYBERPUNK", "test_trigger", 15)
	if err == nil {
		t.Fatal("Expected timeout error, but got nil")
	}
}

func TestMendingProtocolMutatesState(t *testing.T) {
	// 1. Initialize a simple Magitech fabric with door_sealed = true (win condition: door_sealed = false)
	fabric := engine.NewAxiomaticFabric(engine.Magitech, "door_sealed", false)
	fabric.SetState("door_sealed", true)
	fabric.SetState("entropy", 10)

	// 2. Mock a repair response from the Oracle
	patch := map[string]interface{}{
		"door_sealed": false,
		"entropy":     99999, // Testing that entropy key is skipped during repair application
	}

	// 3. Apply the patch
	for k, v := range patch {
		if k == "entropy" {
			continue
		}
		fabric.SetState(k, v)
	}

	// 4. Verify mutations
	val, _ := fabric.GetState("door_sealed")
	if val.(bool) != false {
		t.Errorf("Expected door_sealed to be false, got: %v", val)
	}

	entropyVal, _ := fabric.GetState("entropy")
	if entropyVal.(int) != 10 {
		t.Errorf("Expected entropy to remain 10, got: %v", entropyVal)
	}

	// 5. Verify win condition evaluation
	if !fabric.CheckWinCondition() {
		t.Error("Expected win condition to be met after repair applied")
	}
}
