package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Agent interface {
	Name() string
	Capabilities() []string
	Run(prompt string, model string) (string, error)
}

type LocalAgent struct {
	AgentName  string
	Caps       []string
	OllamaHost string
	HTTPClient *http.Client
}

func NewLocalAgent(name string, caps []string, host string) *LocalAgent {
	return &LocalAgent{
		AgentName:  name,
		Caps:       caps,
		OllamaHost: host,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (a *LocalAgent) Name() string {
	return a.AgentName
}

func (a *LocalAgent) Capabilities() []string {
	return a.Caps
}

func (a *LocalAgent) Run(prompt string, model string) (string, error) {
	url := fmt.Sprintf("%s/api/generate", a.OllamaHost)

	reqPayload := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.0,
			"top_p":       0.9,
		},
	}

	data, err := json.Marshal(reqPayload)
	if err != nil {
		return "", err
	}

	resp, err := a.HTTPClient.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("HTTP request to Ollama failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama API returned non-OK status: %d", resp.StatusCode)
	}

	var respPayload struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return "", err
	}

	return respPayload.Response, nil
}
