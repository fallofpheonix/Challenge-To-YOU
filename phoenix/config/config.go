package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	OllamaHost         string            `json:"ollama_host"`
	ModelRouter        map[string]string `json:"model_router"`
	MaxRetries         int               `json:"max_retries"`
	VigilanceThreshold float64           `json:"vigilance_threshold"`
	WorkspaceRoot      string            `json:"workspace_root"`
	BrainDirectory     string            `json:"brain_directory"`
	WatchIntervalMs    int               `json:"watch_interval_ms"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	defaultBrain := filepath.Join(home, ".gemini/antigravity-ide/brain/c4e154d7-2e40-48e3-adb8-48bb50a6553c")

	return &Config{
		OllamaHost: "http://localhost:11434",
		ModelRouter: map[string]string{
			"planning":    "qwen2.5:32b-instruct",
			"coding":      "deepseek-coder:16b-lite",
			"reviewing":   "qwen2.5:32b-instruct",
			"embeddings":  "mxbai-embed-large",
			"quick_fixes": "qwen2.5-coder:7b",
		},
		MaxRetries:         3,
		VigilanceThreshold: 0.8,
		WorkspaceRoot:      "/Users/fallofpheonix/Project/game/Challenge-To-YOU",
		BrainDirectory:     defaultBrain,
		WatchIntervalMs:    1000,
	}
}

func LoadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := DefaultConfig()
		data, errMarshal := json.MarshalIndent(cfg, "", "  ")
		if errMarshal == nil {
			_ = os.WriteFile(path, data, 0600)
		}
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
