package validators

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ValidationEngine struct {
	WorkspaceRoot string
}

func NewValidationEngine(root string) *ValidationEngine {
	return &ValidationEngine{WorkspaceRoot: root}
}

func (v *ValidationEngine) CompileGo() (string, error) {
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = v.WorkspaceRoot
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func (v *ValidationEngine) RunTests() (string, error) {
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = v.WorkspaceRoot
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func (v *ValidationEngine) RunLinter() (string, error) {
	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = v.WorkspaceRoot
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func (v *ValidationEngine) ValidateChallengeContent() error {
	dir := filepath.Join(v.WorkspaceRoot, "challenges")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// If challenges folder is missing, skip or return nil
		return nil
	}

	seenIDs := make(map[string]bool)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}

		data, errRead := os.ReadFile(path)
		if errRead != nil {
			return errRead
		}

		var chal struct {
			ID        string `json:"id"`
			SkillType string `json:"skill_type"`
		}
		if errUnmarshal := json.Unmarshal(data, &chal); errUnmarshal != nil {
			return fmt.Errorf("JSON parse error in %s: %w", info.Name(), errUnmarshal)
		}

		if chal.ID == "" {
			return fmt.Errorf("empty ID in %s", info.Name())
		}
		if seenIDs[chal.ID] {
			return fmt.Errorf("duplicate ID %s found in %s", chal.ID, info.Name())
		}
		seenIDs[chal.ID] = true
		return nil
	})

	return err
}

func (v *ValidationEngine) VerifyAll() error {
	if logs, err := v.CompileGo(); err != nil {
		return fmt.Errorf("compile failure:\n%s", logs)
	}
	if logs, err := v.RunTests(); err != nil {
		return fmt.Errorf("test failure:\n%s", logs)
	}
	if logs, err := v.RunLinter(); err != nil {
		return fmt.Errorf("lint failure:\n%s", logs)
	}
	if err := v.ValidateChallengeContent(); err != nil {
		return fmt.Errorf("challenge validation failure: %w", err)
	}
	return nil
}
