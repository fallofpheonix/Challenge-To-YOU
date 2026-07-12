package intelligence

import (
	"challenge-to-you/phoenix/config"
	"fmt"
	"net/http"
	"time"
)

type ModelRouter struct {
	Config *config.Config
}

func NewModelRouter(cfg *config.Config) *ModelRouter {
	return &ModelRouter{Config: cfg}
}

func (mr *ModelRouter) Route(task string) (string, error) {
	preferredModel, exists := mr.Config.ModelRouter[task]
	if !exists {
		preferredModel = "qwen2.5:32b-instruct"
	}

	// Verify Ollama has the preferred model or fallback
	if err := mr.checkModelStatus(preferredModel); err != nil {
		fallback := "qwen2.5-coder:7b"
		if preferredModel == fallback {
			return "", fmt.Errorf("both preferred model %s and fallback are unreachable: %w", preferredModel, err)
		}
		return fallback, nil
	}

	return preferredModel, nil
}

func (mr *ModelRouter) checkModelStatus(model string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	url := fmt.Sprintf("%s/api/tags", mr.Config.OllamaHost)
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}
	return nil
}
