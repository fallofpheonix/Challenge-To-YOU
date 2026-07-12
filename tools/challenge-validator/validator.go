package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type FlawDef struct {
	ID           string                 `json:"id"`
	TriggerEvent string                 `json:"trigger_event"`
	Name         string                 `json:"name"`
	FlavorText   string                 `json:"flavor_text"`
	Conditions   map[string]interface{} `json:"conditions"`
	Mutations    map[string]interface{} `json:"mutations"`
}

type WinCondDef struct {
	TargetStateKey string      `json:"target_state_key"`
	ExpectedValue  interface{} `json:"expected_value"`
}

type ChallengeDef struct {
	ID               string                 `json:"id"`
	Paradigm         string                 `json:"paradigm"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	InitialState     map[string]interface{} `json:"initial_state"`
	Flaws            []FlawDef              `json:"flaws"`
	WinCondition     WinCondDef             `json:"win_condition"`
	SkillType        string                 `json:"skill_type"`
	TemplateCode     string                 `json:"template_code"`
	ValidationScript string                 `json:"validation_script"`
}

type PackChallengeDef struct {
	ID         string  `json:"id"`
	File       string  `json:"file"`
	Difficulty float64 `json:"difficulty"`
	Mode       string  `json:"mode"`
}

type PackDef struct {
	Version       int                `json:"version"`
	Era           string             `json:"era"`
	Tier          int                `json:"tier"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	Prerequisites []string           `json:"prerequisites"`
	Unlocks       []string           `json:"unlocks"`
	Challenges    []PackChallengeDef `json:"challenges"`
}

type ValidationReport struct {
	Valid        bool     `json:"valid"`
	TotalScanned int      `json:"total_scanned"`
	Errors       []string `json:"errors"`
	Warnings     []string `json:"warnings"`
}

type ChallengeValidator struct {
	TargetDir string
}

func NewChallengeValidator(target string) *ChallengeValidator {
	return &ChallengeValidator{TargetDir: target}
}

func (cv *ChallengeValidator) Validate() (*ValidationReport, error) {
	report := &ValidationReport{
		Valid:        true,
		TotalScanned: 0,
		Errors:       make([]string, 0),
		Warnings:     make([]string, 0),
	}

	seenIDs := make(map[string]string)
	seenNames := make(map[string]string)

	validParadigms := map[string]bool{
		"MAGITECH": true, "CYBERPUNK": true, "COSMIC": true,
		"SILICON": true, "NEURAL": true, "CHRONO": true,
	}

	err := filepath.WalkDir(cv.TargetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}

		report.TotalScanned++
		data, errRead := os.ReadFile(path)
		if errRead != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: failed to read file: %v", path, errRead))
			report.Valid = false
			return nil
		}

		// Detect if it is a pack.json or a challenge file
		if strings.HasSuffix(d.Name(), "pack.json") {
			var pack PackDef
			if errUnmarshal := json.Unmarshal(data, &pack); errUnmarshal != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: failed to parse Pack JSON: %v", path, errUnmarshal))
				report.Valid = false
				return nil
			}

			// Validate pack fields
			if pack.Era == "" {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: missing 'era' field in pack", path))
				report.Valid = false
			}
			if len(pack.Challenges) == 0 {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: challenges list is empty in pack", path))
				report.Valid = false
			}

			// Validate each challenge reference in the pack
			for _, packChal := range pack.Challenges {
				if packChal.ID == "" {
					report.Errors = append(report.Errors, fmt.Sprintf("%s: pack contains a challenge reference with missing 'id'", path))
					report.Valid = false
				}
				if packChal.File == "" {
					report.Errors = append(report.Errors, fmt.Sprintf("%s: pack contains a challenge reference with missing 'file'", path))
					report.Valid = false
				}
				if packChal.Difficulty < 0.0 || packChal.Difficulty > 1.0 {
					report.Errors = append(report.Errors, fmt.Sprintf("%s: challenge reference '%s' difficulty %f out of range (0.0 - 1.0)", path, packChal.ID, packChal.Difficulty))
					report.Valid = false
				}
			}
		} else {
			var chal ChallengeDef
			if errUnmarshal := json.Unmarshal(data, &chal); errUnmarshal != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: failed to parse Challenge JSON: %v", path, errUnmarshal))
				report.Valid = false
				return nil
			}

			// Validate challenge fields
			if chal.ID == "" {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: missing unique 'id' field", path))
				report.Valid = false
			} else {
				if orig, exists := seenIDs[chal.ID]; exists {
					report.Errors = append(report.Errors, fmt.Sprintf("duplicate ID '%s' found in %s and %s", chal.ID, orig, path))
					report.Valid = false
				}
				seenIDs[chal.ID] = path
			}

			if chal.Name == "" {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: missing 'name' field", path))
				report.Valid = false
			} else {
				if orig, exists := seenNames[chal.Name]; exists {
					report.Warnings = append(report.Warnings, fmt.Sprintf("duplicate challenge name '%s' found in %s and %s", chal.Name, orig, path))
				}
				seenNames[chal.Name] = path
			}

			if chal.Paradigm == "" {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: missing 'paradigm' field", path))
				report.Valid = false
			} else if !validParadigms[strings.ToUpper(chal.Paradigm)] {
				report.Warnings = append(report.Warnings, fmt.Sprintf("%s: custom paradigm '%s' is not in standard list", path, chal.Paradigm))
			}

			if chal.SkillType != "composite" {
				if len(chal.Flaws) == 0 {
					report.Errors = append(report.Errors, fmt.Sprintf("%s: flaws array is empty", path))
					report.Valid = false
				}

				for _, flaw := range chal.Flaws {
					if flaw.ID == "" {
						report.Errors = append(report.Errors, fmt.Sprintf("%s: flaw contains missing 'id'", path))
						report.Valid = false
					}
					if flaw.TriggerEvent == "" {
						report.Errors = append(report.Errors, fmt.Sprintf("%s: flaw '%s' has missing 'trigger_event'", path, flaw.ID))
						report.Valid = false
					}
				}
			}

			if chal.WinCondition.TargetStateKey == "" {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: win_condition is missing 'target_state_key'", path))
				report.Valid = false
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return report, nil
}
