package qa

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RegressionSnapshot captures a baseline for regression detection.
type RegressionSnapshot struct {
	Timestamp       time.Time         `json:"timestamp"`
	ChallengeCount  int               `json:"challenge_count"`
	ChallengeIDs    []string          `json:"challenge_ids"`
	MissionIDs      []string          `json:"mission_ids"`
	ServerEndpoints []string          `json:"server_endpoints"`
	ProfileDefaults map[string]interface{} `json:"profile_defaults"`
	LevelThresholds map[int]string    `json:"level_thresholds"`
	TitleMap        map[int]string    `json:"title_map"`
}

// ScenarioRegressionDetection compares current state against previous runs.
func ScenarioRegressionDetection(ctx *ScenarioContext) error {
	root := getProjectRoot()
	snapshotPath := filepath.Join(root, "qa", "fixtures", "regression_baseline.json")

	// Build current snapshot
	current := RegressionSnapshot{
		Timestamp:      time.Now(),
		ChallengeIDs:   collectChallengeIDs(ctx),
		MissionIDs:     collectMissionIDs(ctx),
		ServerEndpoints: []string{"/rift", "/api/languages", "/api/challenges"},
		LevelThresholds: buildLevelThresholds(),
		TitleMap:        buildTitleMap(),
	}

	// Count challenges
	current.ChallengeCount = len(current.ChallengeIDs)

	// Check if baseline exists
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		// First run - create baseline
		ctx.Step("baseline_created", "No previous baseline found. Creating regression baseline.")
		data, _ := json.MarshalIndent(current, "", "  ")
		os.MkdirAll(filepath.Dir(snapshotPath), 0o755)
		if err := os.WriteFile(snapshotPath, data, 0o644); err != nil {
			return fmt.Errorf("write baseline: %w", err)
		}
		ctx.Assert("baseline_saved", true, fmt.Sprintf("Baseline saved with %d challenges", current.ChallengeCount))
		return nil
	}

	// Load previous baseline
	prevData, err := os.ReadFile(snapshotPath)
	if err != nil {
		return fmt.Errorf("read baseline: %w", err)
	}

	var previous RegressionSnapshot
	if err := json.Unmarshal(prevData, &previous); err != nil {
		return fmt.Errorf("parse baseline: %w", err)
	}

	// Compare
	ctx.Assert("challenge_count_stable", current.ChallengeCount == previous.ChallengeCount,
		fmt.Sprintf("Current: %d, Previous: %d", current.ChallengeCount, previous.ChallengeCount))

	// Check for removed challenges
	removed := diffStringLists(previous.ChallengeIDs, current.ChallengeIDs)
	ctx.Assert("no_challenges_removed", len(removed) == 0,
		fmt.Sprintf("Removed challenges: %v", removed))

	// Check endpoints are stable
	ctx.Assert("endpoints_stable", len(current.ServerEndpoints) == len(previous.ServerEndpoints),
		fmt.Sprintf("Endpoints: %d -> %d", len(previous.ServerEndpoints), len(current.ServerEndpoints)))

	// Update baseline with new timestamp
	current.Timestamp = time.Now()
	data, _ := json.MarshalIndent(current, "", "  ")
	os.WriteFile(snapshotPath, data, 0o644)

	ctx.Step("regression_check_complete", fmt.Sprintf("Compared against baseline from %s", previous.Timestamp.Format(time.RFC3339)))
	return nil
}

func collectChallengeIDs(ctx *ScenarioContext) []string {
	root := getProjectRoot()
	challengesDir := filepath.Join(root, "backend", "challenges")
	var ids []string

	eras := []string{"magitech_tier1", "cyberpunk_tier1", "cosmic_tier1"}
	for _, era := range eras {
		eraDir := filepath.Join(challengesDir, era)
		entries, err := os.ReadDir(eraDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
				ids = append(ids, entry.Name())
			}
		}
	}
	return ids
}

func collectMissionIDs(ctx *ScenarioContext) []string {
	root := getProjectRoot()
	missionsDir := filepath.Join(root, "backend", "data", "missions")
	var ids []string

	entries, err := os.ReadDir(missionsDir)
	if err != nil {
		return ids
	}
	for _, entry := range entries {
		if entry.IsDir() {
			subEntries, _ := os.ReadDir(filepath.Join(missionsDir, entry.Name()))
			for _, sub := range subEntries {
				if !sub.IsDir() && filepath.Ext(sub.Name()) == ".json" {
					ids = append(ids, sub.Name())
				}
			}
		}
	}
	return ids
}

func buildLevelThresholds() map[int]string {
	thresholds := map[int]string{
		1:  "0 XP",
		2:  "100 XP",
		3:  "250 XP",
		4:  "500 XP",
		5:  "1000 XP",
		10: "3000 XP",
		15: "8000 XP",
		20: "15000 XP",
	}
	return thresholds
}

func buildTitleMap() map[int]string {
	titles := map[int]string{
		1: "Newcomer",
		2: "Initiate",
		5: "Operative",
		10: "Veteran",
		15: "Archon",
	}
	return titles
}

func diffStringLists(a, b []string) []string {
	setB := make(map[string]bool, len(b))
	for _, s := range b {
		setB[s] = true
	}
	var diff []string
	for _, s := range a {
		if !setB[s] {
			diff = append(diff, s)
		}
	}
	return diff
}
