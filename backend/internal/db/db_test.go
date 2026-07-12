package db

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "challenge_db_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	dbPath := filepath.Join(tempDir, "test.db")
	d, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestGetOrCreateProfile_Defaults(t *testing.T) {
	d := setupTestDB(t)

	profile, err := d.GetOrCreateProfile()
	if err != nil {
		t.Fatalf("GetOrCreateProfile failed: %v", err)
	}
	if profile.Name != "Intruder" {
		t.Errorf("Name: got %q, want %q", profile.Name, "Intruder")
	}
	if profile.Luck != 1.0 {
		t.Errorf("Luck: got %f, want 1.0", profile.Luck)
	}
	if profile.Reputation != 0 {
		t.Errorf("Reputation: got %d, want 0", profile.Reputation)
	}
	if !profile.IsParadigmUnlocked("MAGITECH") {
		t.Error("MAGITECH should be unlocked by default")
	}
	if profile.IsParadigmUnlocked("CYBERPUNK") {
		t.Error("CYBERPUNK should be locked by default")
	}
}

func TestSaveProfile_Mutations(t *testing.T) {
	d := setupTestDB(t)

	profile, _ := d.GetOrCreateProfile()
	profile.Luck = 1.25
	profile.Reputation = 60
	profile.UnlockedParadigms = "MAGITECH,CYBERPUNK"
	if err := d.SaveProfile(profile); err != nil {
		t.Fatalf("SaveProfile failed: %v", err)
	}

	mutated, _ := d.GetOrCreateProfile()
	if mutated.Luck != 1.25 {
		t.Errorf("Luck: got %f, want 1.25", mutated.Luck)
	}
	if mutated.Reputation != 60 {
		t.Errorf("Reputation: got %d, want 60", mutated.Reputation)
	}
	if !mutated.IsParadigmUnlocked("CYBERPUNK") {
		t.Error("CYBERPUNK should be unlocked now")
	}
}

func TestRecordToken(t *testing.T) {
	d := setupTestDB(t)

	token := "TEST_LOGOS_TOKEN_XYZ"

	t.Run("new_token", func(t *testing.T) {
		isNew, err := d.RecordToken(token)
		if err != nil {
			t.Fatalf("RecordToken failed: %v", err)
		}
		if !isNew {
			t.Error("Expected token to be newly registered")
		}
	})

	t.Run("duplicate_token", func(t *testing.T) {
		// Ensure token exists first
		d.RecordToken(token)
		isNew, err := d.RecordToken(token)
		if err != nil {
			t.Fatalf("RecordToken duplicate check failed: %v", err)
		}
		if isNew {
			t.Error("Expected token to be marked as duplicate")
		}
	})
}

func TestComputeLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		xp       int
		expected int
	}{
		{"zero_xp", 0, 1},
		{"just_below_100", 99, 1},
		{"exactly_100", 100, 2},
		{"just_below_250", 249, 2},
		{"exactly_250", 250, 3},
		{"mid_range", 500, 4},
		{"high_xp", 1000, 5},
		{"max_level", 17300, 20},
		{"cap_at_20", 20000, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ComputeLevel(tt.xp)
			if got != tt.expected {
				t.Errorf("ComputeLevel(%d): got %d, want %d", tt.xp, got, tt.expected)
			}
		})
	}
}

func TestTitleForLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		level    int
		expected string
	}{
		{"level_1", 1, "Newcomer"},
		{"level_2", 2, "Initiate"},
		{"level_4", 4, "Initiate"},
		{"level_5", 5, "Operative"},
		{"level_9", 9, "Operative"},
		{"level_10", 10, "Veteran"},
		{"level_14", 14, "Veteran"},
		{"level_15", 15, "Archon"},
		{"level_20", 20, "Archon"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := TitleForLevel(tt.level)
			if got != tt.expected {
				t.Errorf("TitleForLevel(%d): got %q, want %q", tt.level, got, tt.expected)
			}
		})
	}
}

func TestAddXP(t *testing.T) {
	d := setupTestDB(t)

	if err := d.AddXP(150); err != nil {
		t.Fatalf("AddXP failed: %v", err)
	}
	p, _ := d.GetOrCreateProfile()
	if p.XP != 150 {
		t.Errorf("XP: got %d, want 150", p.XP)
	}
	if p.Level != 2 {
		t.Errorf("Level: got %d, want 2", p.Level)
	}
	if p.Title != "Initiate" {
		t.Errorf("Title: got %q, want %q", p.Title, "Initiate")
	}
}

func TestRecordChallengeCompletion(t *testing.T) {
	d := setupTestDB(t)

	t.Run("new_challenge", func(t *testing.T) {
		isNew, err := d.RecordChallengeCompletion("cyberpunk_01", 50)
		if err != nil {
			t.Fatalf("RecordChallengeCompletion failed: %v", err)
		}
		if !isNew {
			t.Error("Expected challenge to be newly recorded")
		}
	})

	t.Run("duplicate_returns_false", func(t *testing.T) {
		d.RecordChallengeCompletion("cyberpunk_02", 50)
		isNew, err := d.RecordChallengeCompletion("cyberpunk_02", 50)
		if err != nil {
			t.Fatalf("RecordChallengeCompletion duplicate check failed: %v", err)
		}
		if isNew {
			t.Error("Expected duplicate to return false")
		}
	})

	t.Run("is_completed_check", func(t *testing.T) {
		d.RecordChallengeCompletion("cyberpunk_03", 50)
		if !d.IsChallengeCompleted("cyberpunk_03") {
			t.Error("Expected challenge to be marked completed")
		}
		if d.IsChallengeCompleted("cyberpunk_04") {
			t.Error("Expected uncompleted challenge to return false")
		}
	})

	t.Run("get_list", func(t *testing.T) {
		d.RecordChallengeCompletion("cyberpunk_05", 50)
		completed, err := d.GetCompletedChallenges()
		if err != nil {
			t.Fatalf("GetCompletedChallenges failed: %v", err)
		}
		if len(completed) == 0 {
			t.Fatal("Expected at least 1 completed challenge")
		}
		found := false
		for _, c := range completed {
			if c.ChallengeID == "cyberpunk_05" && c.XPEarned == 50 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected cyberpunk_05 with XP 50 in completed list")
		}
	})
}

func TestRecordMissionCompletion(t *testing.T) {
	d := setupTestDB(t)

	t.Run("new_mission", func(t *testing.T) {
		isNew, err := d.RecordMissionCompletion("mission_magitech_01")
		if err != nil {
			t.Fatalf("RecordMissionCompletion failed: %v", err)
		}
		if !isNew {
			t.Error("Expected mission to be newly recorded")
		}
	})

	t.Run("duplicate_returns_false", func(t *testing.T) {
		d.RecordMissionCompletion("mission_magitech_02")
		isNew, err := d.RecordMissionCompletion("mission_magitech_02")
		if err != nil {
			t.Fatalf("RecordMissionCompletion duplicate check failed: %v", err)
		}
		if isNew {
			t.Error("Expected duplicate to return false")
		}
	})

	t.Run("is_completed_check", func(t *testing.T) {
		d.RecordMissionCompletion("mission_magitech_03")
		if !d.IsMissionCompleted("mission_magitech_03") {
			t.Error("Expected mission to be marked completed")
		}
		if d.IsMissionCompleted("mission_magitech_04") {
			t.Error("Expected uncompleted mission to return false")
		}
	})

	t.Run("get_list", func(t *testing.T) {
		d.RecordMissionCompletion("mission_magitech_05")
		completed, err := d.GetCompletedMissions()
		if err != nil {
			t.Fatalf("GetCompletedMissions failed: %v", err)
		}
		if len(completed) == 0 {
			t.Fatal("Expected at least 1 completed mission")
		}
		found := false
		for _, m := range completed {
			if m.MissionID == "mission_magitech_05" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected mission_magitech_05 in completed list")
		}
	})
}

func TestActiveMissionPersistence(t *testing.T) {
	d := setupTestDB(t)

	t.Run("save_and_retrieve", func(t *testing.T) {
		sessionData := map[string]interface{}{
			"status":     "active",
			"step":       2,
			"objectives": []string{"obj_1", "obj_2"},
		}
		if err := d.SaveActiveMission("mission_cyberpunk_01", "player1", sessionData); err != nil {
			t.Fatalf("SaveActiveMission failed: %v", err)
		}

		active, err := d.GetActiveMissions()
		if err != nil {
			t.Fatalf("GetActiveMissions failed: %v", err)
		}
		if len(active) != 1 {
			t.Fatalf("Expected 1 active mission, got %d", len(active))
		}
		if active[0].MissionID != "mission_cyberpunk_01" {
			t.Errorf("MissionID: got %q, want %q", active[0].MissionID, "mission_cyberpunk_01")
		}
		if active[0].PlayerID != "player1" {
			t.Errorf("PlayerID: got %q, want %q", active[0].PlayerID, "player1")
		}
		if active[0].SessionJSON == "" {
			t.Error("SessionJSON should not be empty")
		}
	})

	t.Run("remove", func(t *testing.T) {
		sessionData := map[string]interface{}{"status": "active"}
		d.SaveActiveMission("mission_cyberpunk_02", "player2", sessionData)

		if err := d.RemoveActiveMission("mission_cyberpunk_02"); err != nil {
			t.Fatalf("RemoveActiveMission failed: %v", err)
		}
		// Verify it was removed by checking it doesn't exist
		active, err := d.GetActiveMissions()
		if err != nil {
			t.Fatalf("GetActiveMissions failed: %v", err)
		}
		for _, m := range active {
			if m.MissionID == "mission_cyberpunk_02" {
				t.Error("Expected mission_cyberpunk_02 to be removed")
			}
		}
	})
}
