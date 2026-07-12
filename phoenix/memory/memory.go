package memory

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FailureRecord struct {
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	Category   string    `json:"category"`
	Diagnostic string    `json:"diagnostic"`
	Diff       string    `json:"diff"`
	StackTrace string    `json:"stack_trace"`
}

type RepairRecord struct {
	ID                string   `json:"id"`
	AssociatedFailure string   `json:"associated_failure"`
	AppliedPatch      string   `json:"applied_patch"`
	DurationMs        int      `json:"duration_ms"`
	TestsPassed       bool     `json:"tests_passed"`
	Keywords          []string `json:"keywords"`
}

type MemoryManager struct {
	BrainDir string
}

func NewMemoryManager(brainDir string) *MemoryManager {
	// Initialize directories
	_ = os.MkdirAll(filepath.Join(brainDir, "build_failures"), 0755)
	_ = os.MkdirAll(filepath.Join(brainDir, "runtime_failures"), 0755)
	_ = os.MkdirAll(filepath.Join(brainDir, "successful_repairs"), 0755)
	_ = os.MkdirAll(filepath.Join(brainDir, "rejected_repairs"), 0755)

	return &MemoryManager{BrainDir: brainDir}
}

func (m *MemoryManager) SaveFailure(f *FailureRecord) error {
	f.Timestamp = time.Now()
	hash := sha256.Sum256([]byte(f.Diagnostic + f.StackTrace))
	f.ID = hex.EncodeToString(hash[:])[:8]

	dir := "build_failures"
	if f.Category == "runtime" {
		dir = "runtime_failures"
	}

	path := filepath.Join(m.BrainDir, dir, fmt.Sprintf("%d_%s.json", f.Timestamp.Unix(), f.ID))
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (m *MemoryManager) SaveRepair(r *RepairRecord, successful bool) error {
	hash := sha256.Sum256([]byte(r.AppliedPatch))
	r.ID = hex.EncodeToString(hash[:])[:8]

	dir := "rejected_repairs"
	if successful {
		dir = "successful_repairs"
	}

	path := filepath.Join(m.BrainDir, dir, fmt.Sprintf("%d_%s.json", time.Now().Unix(), r.ID))
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (m *MemoryManager) SearchSimilar(diagnostic string) ([]*RepairRecord, error) {
	var results []*RepairRecord
	dir := filepath.Join(m.BrainDir, "successful_repairs")

	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	diagLower := strings.ToLower(diagnostic)

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, file.Name())
		data, errRead := os.ReadFile(path)
		if errRead != nil {
			continue
		}
		var record RepairRecord
		if errUnmarshal := json.Unmarshal(data, &record); errUnmarshal != nil {
			continue
		}

		// Perform keyword intersection match
		match := false
		for _, kw := range record.Keywords {
			if strings.Contains(diagLower, strings.ToLower(kw)) {
				match = true
				break
			}
		}

		if match {
			results = append(results, &record)
		}
	}

	return results, nil
}
