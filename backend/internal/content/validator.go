package content

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("field '%s': %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	msgs := make([]string, len(e))
	for i, err := range e {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "; ")
}

func ValidateChallenge(data []byte) (*Challenge, ValidationErrors) {
	var ch Challenge
	if err := json.Unmarshal(data, &ch); err != nil {
		return nil, ValidationErrors{{Field: "root", Message: "invalid JSON: " + err.Error()}}
	}

	var errs ValidationErrors

	if ch.SchemaVersion == "" {
		ch.SchemaVersion = CurrentSchemaVersion
	}
	if ch.ID == "" {
		errs = append(errs, ValidationError{Field: "id", Message: "required"})
	}
	if ch.Title == "" {
		errs = append(errs, ValidationError{Field: "title", Message: "required"})
	}
	if ch.Description == "" {
		errs = append(errs, ValidationError{Field: "description", Message: "required"})
	}
	if ch.Category == "" {
		errs = append(errs, ValidationError{Field: "category", Message: "required"})
	}
	if ch.Difficulty < 0 || ch.Difficulty > 1 {
		errs = append(errs, ValidationError{Field: "difficulty", Message: "must be 0.0-1.0"})
	}
	if len(ch.SupportedLanguages) == 0 {
		errs = append(errs, ValidationError{Field: "supported_languages", Message: "at least one language required"})
	}
	if len(ch.VisibleTests) == 0 && len(ch.HiddenTests) == 0 && ch.Validator == "" {
		errs = append(errs, ValidationError{Field: "tests", Message: "at least one test case or validator required"})
	}
	if ch.XP <= 0 {
		errs = append(errs, ValidationError{Field: "xp", Message: "must be positive"})
	}
	if ch.TimeLimit <= 0 {
		ch.TimeLimit = 5000
	}
	if ch.MemoryLimit <= 0 {
		ch.MemoryLimit = 256 * 1024 * 1024
	}
	if ch.Version == "" {
		ch.Version = "1.0.0"
	}

	if len(errs) > 0 {
		return &ch, errs
	}
	return &ch, nil
}

func ValidateChallengeFile(path string) (*Challenge, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	ch, errs := ValidateChallenge(data)
	if len(errs) > 0 {
		return ch, errs
	}
	return ch, nil
}

func ValidatePack(data []byte) (*ChallengePack, error) {
	var pack ChallengePack
	if err := json.Unmarshal(data, &pack); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if pack.ID == "" {
		return nil, fmt.Errorf("pack id required")
	}
	if pack.Name == "" {
		return nil, fmt.Errorf("pack name required")
	}
	if len(pack.Challenges) == 0 {
		return nil, fmt.Errorf("pack must contain at least one challenge")
	}
	return &pack, nil
}
