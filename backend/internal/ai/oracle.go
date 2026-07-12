package ai

import (
	"bytes"
	"challenge-to-you/backend/internal/obs"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var log = obs.Default().Component("ai")

type OracleClient struct {
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

func NewOracleClient(baseURL, model string) *OracleClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "qwen2.5:1.5b-instruct"
	}
	return &OracleClient{
		BaseURL: baseURL,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system,omitempty"`
	Stream bool   `json:"stream"`
	Format string `json:"format,omitempty"`
}

type generateResponse struct {
	Response string `json:"response"`
}

var allowedRepairKeys = map[string]bool{
	"entropy":          true,
	"lock_mechanism":   true,
	"gate_state":       true,
	"circuit_status":   true,
	"seal_integrity":   true,
	"coolant_level":    true,
	"signal_noise":     true,
	"rune_charge":      true,
	"barrier_state":    true,
	"power_level":      true,
	"stability_index":  true,
	"connection_state": true,
	"override_code":    true,
}

func isAllowedRepairValue(v interface{}) bool {
	switch val := v.(type) {
	case int, int32, int64, float32, float64:
		return true
	case bool:
		return true
	case string:
		return len(val) <= 64 && !strings.ContainsAny(val, "\n\r\t")
	default:
		return false
	}
}

func validateRepairPatch(patch map[string]interface{}) (map[string]interface{}, []string) {
	validated := make(map[string]interface{})
	var rejected []string

	for k, v := range patch {
		if k == "entropy" {
			switch e := v.(type) {
			case float64:
				if e >= 0 && e <= 1000 {
					validated[k] = e
					continue
				}
			case int:
				if e >= 0 && e <= 1000 {
					validated[k] = float64(e)
					continue
				}
			}
			rejected = append(rejected, fmt.Sprintf("%s=out_of_range(%v)", k, v))
			continue
		}

		if !allowedRepairKeys[k] {
			rejected = append(rejected, fmt.Sprintf("%s=unknown_key", k))
			continue
		}

		if !isAllowedRepairValue(v) {
			rejected = append(rejected, fmt.Sprintf("%s=invalid_type(%T)", k, v))
			continue
		}

		validated[k] = v
	}

	return validated, rejected
}

func (c *OracleClient) GenerateTaunt(ctx context.Context, paradigm, action string, entropy int) (string, error) {
	systemPrompt := fmt.Sprintf("You are the hostile AI Archon supervising a high-security %s matrix. Keep responses under 15 words. Respond with pure text, no markdown, no quotes.", paradigm)
	userPrompt := fmt.Sprintf("The intruder player just triggered '%s' and spiked system entropy by %d. Taunt them contextually.", action, entropy)

	reqBody := generateRequest{
		Model:  c.Model,
		Prompt: userPrompt,
		System: systemPrompt,
		Stream: false,
	}

	resText, err := c.postGenerate(ctx, reqBody)
	if err != nil {
		return "", err
	}

	resText = sanitizeTaunt(resText)
	return resText, nil
}

func sanitizeTaunt(text string) string {
	cleaned := strings.TrimSpace(text)
	cleaned = strings.Trim(cleaned, "\"'`")
	if len(cleaned) > 200 {
		cleaned = cleaned[:200]
	}
	cleaned = strings.ReplaceAll(cleaned, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\r", "")
	return cleaned
}

func (c *OracleClient) GenerateRepair(ctx context.Context, paradigm string, currentState map[string]interface{}) (map[string]interface{}, error) {
	systemPrompt := "You are the security system mending protocol. You must output a JSON object containing state updates to repair/modify the database. Output ONLY valid JSON."

	stateBytes, _ := json.Marshal(currentState)
	userPrompt := fmt.Sprintf("The current state is %s. Generate 1 or 2 repair mutations in JSON format to reduce entropy or reset a locking mechanism. Example: {\"entropy\": 0}. Paradigm is %s.", string(stateBytes), paradigm)

	reqBody := generateRequest{
		Model:  c.Model,
		Prompt: userPrompt,
		System: systemPrompt,
		Stream: false,
		Format: "json",
	}

	resText, err := c.postGenerate(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	var rawPatch map[string]interface{}
	if err := json.Unmarshal([]byte(resText), &rawPatch); err != nil {
		return nil, fmt.Errorf("failed to parse JSON repair updates from AI response %q: %w", resText, err)
	}

	validated, rejected := validateRepairPatch(rawPatch)
	if len(rejected) > 0 {
		log.Warn("repair fields rejected", "count", len(rejected), "rejected", rejected)
	}

	if len(validated) == 0 {
		return nil, fmt.Errorf("all repair fields rejected by safety validation: %v", rejected)
	}

	return validated, nil
}

func (c *OracleClient) postGenerate(ctx context.Context, reqBody generateRequest) (string, error) {
	url := fmt.Sprintf("%s/api/generate", c.BaseURL)
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Ollama API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama returned status code %d", resp.StatusCode)
	}

	var resObj generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&resObj); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return resObj.Response, nil
}
