package intelligence

import (
	"fmt"
	"strings"
)

type SafetyEngine struct {
	ProtectedDirs  []string
	ProtectedFiles []string
	ProtectedAPIs  []string
}

func NewSafetyEngine() *SafetyEngine {
	return &SafetyEngine{
		ProtectedDirs:  []string{"backend/internal/db", "client/scenes"},
		ProtectedFiles: []string{"backend/go.mod", "docs/phoenix/prompts.md"},
		ProtectedAPIs:  []string{"AxiomaticFabric", "RaiseLifecycleEvent"},
	}
}

func (se *SafetyEngine) ValidatePatchSafety(patchContent string) error {
	lines := strings.Split(patchContent, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			// Extract file paths
			for _, protected := range se.ProtectedFiles {
				if strings.Contains(line, protected) {
					return fmt.Errorf("safety violation: modification of protected file %s is blocked", protected)
				}
			}
			for _, protectedDir := range se.ProtectedDirs {
				if strings.Contains(line, protectedDir) {
					return fmt.Errorf("safety violation: modification in protected directory %s is blocked", protectedDir)
				}
			}
		}
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "+") {
			// Extract API deletions
			for _, api := range se.ProtectedAPIs {
				if strings.Contains(line, api) {
					return fmt.Errorf("safety violation: modification of protected API %s is blocked", api)
				}
			}
		}
	}
	return nil
}
