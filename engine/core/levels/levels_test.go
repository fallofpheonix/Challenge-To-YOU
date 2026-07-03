package levels

import (
	"os"
	"path/filepath"
	"testing"
)

// --- Level.Validate() tests ---

func validLevel() Level {
	return Level{
		SchemaVersion: 1,
		ID:            "test_level",
		World:         WorldDef{Width: 50, Height: 50},
		Drones:        DronesDef{InitialCount: 5},
		Mission: MissionDef{
			TargetResources:        10,
			MaxTicks:               1000,
			InfectionLossThreshold: 0.5,
		},
	}
}

func TestValidateAcceptsValidLevel(t *testing.T) {
	lvl := validLevel()
	if errs := lvl.Validate(); len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
}

func TestValidateRejectsFutureSchemaVersion(t *testing.T) {
	lvl := validLevel()
	lvl.SchemaVersion = CurrentSchemaVersion + 1
	errs := lvl.Validate()
	if len(errs) == 0 {
		t.Fatal("expected error for future schema_version, got none")
	}
}

func TestValidateRejectsEmptyID(t *testing.T) {
	lvl := validLevel()
	lvl.ID = ""
	errs := lvl.Validate()
	if len(errs) == 0 {
		t.Fatal("expected error for empty id, got none")
	}
}

func TestValidateRejectsZeroWorldDimensions(t *testing.T) {
	lvl := validLevel()
	lvl.World.Width = 0
	errs := lvl.Validate()
	if len(errs) == 0 {
		t.Fatal("expected error for zero world width, got none")
	}
}

func TestValidateRejectsNegativeDroneCount(t *testing.T) {
	lvl := validLevel()
	lvl.Drones.InitialCount = 0
	errs := lvl.Validate()
	if len(errs) == 0 {
		t.Fatal("expected error for zero initial_count, got none")
	}
}

func TestValidateRejectsOutOfBoundsResource(t *testing.T) {
	lvl := validLevel()
	lvl.Resources = []ResourceDef{{ID: "r1", X: 999, Y: 10, Count: 100}}
	errs := lvl.Validate()
	if len(errs) == 0 {
		t.Fatal("expected error for out-of-bounds resource x, got none")
	}
}

func TestValidateRejectsNegativeResourceCount(t *testing.T) {
	lvl := validLevel()
	lvl.Resources = []ResourceDef{{ID: "r1", X: 5, Y: 5, Count: -1}}
	errs := lvl.Validate()
	if len(errs) == 0 {
		t.Fatal("expected error for negative resource count, got none")
	}
}

func TestValidateRejectsDuplicateIDs(t *testing.T) {
	lvl := validLevel()
	lvl.Resources = []ResourceDef{
		{ID: "obj_1", X: 5, Y: 5, Count: 100},
		{ID: "obj_1", X: 6, Y: 6, Count: 100},
	}
	errs := lvl.Validate()
	found := false
	for _, e := range errs {
		if len(e) > 0 {
			found = true
		}
	}
	if !found {
		t.Fatal("expected duplicate ID error, got none")
	}
}

func TestValidateRejectsOutOfBoundsHazard(t *testing.T) {
	lvl := validLevel()
	lvl.Hazards = []HazardDef{{ID: "h1", Type: 0, X: -1, Y: 10, Radius: 5, Intensity: 1000000}}
	errs := lvl.Validate()
	if len(errs) == 0 {
		t.Fatal("expected error for out-of-bounds hazard x, got none")
	}
}

func TestValidateRejectsZeroHazardRadius(t *testing.T) {
	lvl := validLevel()
	lvl.Hazards = []HazardDef{{ID: "h1", Type: 0, X: 10, Y: 10, Radius: 0, Intensity: 1000000}}
	errs := lvl.Validate()
	if len(errs) == 0 {
		t.Fatal("expected error for zero hazard radius, got none")
	}
}

func TestValidateRejectsInfectionThresholdOutOfRange(t *testing.T) {
	lvl := validLevel()
	lvl.Mission.InfectionLossThreshold = 1.5
	errs := lvl.Validate()
	if len(errs) == 0 {
		t.Fatal("expected error for infection_loss_threshold > 1, got none")
	}
}

func TestValidateCollectsMultipleErrors(t *testing.T) {
	lvl := validLevel()
	lvl.World.Width = 0
	lvl.Mission.TargetResources = 0
	errs := lvl.Validate()
	if len(errs) < 2 {
		t.Fatalf("expected at least 2 errors, got %d: %v", len(errs), errs)
	}
}

// --- LoadLevel round-trip ---

func TestLoadLevelRoundTrip(t *testing.T) {
	lvl, err := LoadLevel("chrysalis_1.json")
	if err != nil {
		t.Fatalf("LoadLevel failed: %v", err)
	}
	if lvl.ID != "chrysalis_1" {
		t.Errorf("unexpected id: %s", lvl.ID)
	}
	if lvl.SchemaVersion != 1 {
		t.Errorf("unexpected schema_version: %d", lvl.SchemaVersion)
	}
	if len(lvl.Resources) == 0 {
		t.Error("expected at least one resource")
	}
	if lvl.Resources[0].ID == "" {
		t.Error("expected resource to have an id")
	}
}

func TestLoadLevelCreatesEngine(t *testing.T) {
	lvl, err := LoadLevel("chrysalis_1.json")
	if err != nil {
		t.Fatalf("LoadLevel failed: %v", err)
	}
	e := lvl.CreateEngine()
	if e == nil {
		t.Fatal("CreateEngine returned nil")
	}
	if e.Registry.Count != lvl.Drones.InitialCount {
		t.Errorf("drone count: want %d got %d", lvl.Drones.InitialCount, e.Registry.Count)
	}
	if e.Mission.TargetResources != lvl.Mission.TargetResources {
		t.Errorf("target resources: want %d got %d", lvl.Mission.TargetResources, e.Mission.TargetResources)
	}
	// Verify resource deposit was placed.
	r := lvl.Resources[0]
	idx := e.Grid.GetIndex(r.X, r.Y)
	if e.Grid.CurrentCells[idx].ResourceCount != r.Count {
		t.Errorf("resource count at (%d,%d): want %d got %d",
			r.X, r.Y, r.Count, e.Grid.CurrentCells[idx].ResourceCount)
	}
	// Verify hazard was placed.
	if len(lvl.Hazards) > 0 && !e.Hazards.Active[0] {
		t.Error("expected hazard slot 0 to be active")
	}
	// Verify alien node was placed.
	if len(lvl.Aliens) > 0 && !e.Aliens.Active[0] {
		t.Error("expected alien slot 0 to be active")
	}
}

// --- Campaign tests ---

func TestCampaignNextLevelEmptyProgress(t *testing.T) {
	c := Campaign{
		ID: "test_campaign",
		Levels: []CampaignEntry{
			{LevelFile: "chrysalis_1.json", LevelID: "chrysalis_1"},
		},
	}
	completed := map[string]bool{}
	entry := c.NextLevel(completed)
	if entry == nil {
		t.Fatal("expected a next level, got nil")
	}
	if entry.LevelID != "chrysalis_1" {
		t.Errorf("unexpected level id: %s", entry.LevelID)
	}
}

func TestCampaignNextLevelSkipsCompleted(t *testing.T) {
	c := Campaign{
		ID: "test_campaign",
		Levels: []CampaignEntry{
			{LevelFile: "a.json", LevelID: "level_a"},
			{LevelFile: "b.json", LevelID: "level_b", Requires: []string{"level_a"}},
		},
	}
	completed := map[string]bool{"level_a": true}
	entry := c.NextLevel(completed)
	if entry == nil {
		t.Fatal("expected level_b to be unlocked")
	}
	if entry.LevelID != "level_b" {
		t.Errorf("unexpected level id: %s", entry.LevelID)
	}
}

func TestCampaignNextLevelLockedByRequirement(t *testing.T) {
	c := Campaign{
		ID: "test_campaign",
		Levels: []CampaignEntry{
			{LevelFile: "b.json", LevelID: "level_b", Requires: []string{"level_a"}},
		},
	}
	completed := map[string]bool{}
	entry := c.NextLevel(completed)
	if entry != nil {
		t.Errorf("expected nil when requirements are unmet, got %s", entry.LevelID)
	}
}

func TestCampaignNextLevelAllComplete(t *testing.T) {
	c := Campaign{
		ID:     "test_campaign",
		Levels: []CampaignEntry{{LevelFile: "a.json", LevelID: "level_a"}},
	}
	completed := map[string]bool{"level_a": true}
	if c.NextLevel(completed) != nil {
		t.Error("expected nil when all levels are complete")
	}
}

// --- CampaignProgress save/load round-trip ---

func TestCampaignProgressSaveLoad(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "progress.json")

	p := NewCampaignProgress("test_campaign")
	p.Complete("chrysalis_1")
	if err := p.Save(tmp); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadProgress(tmp, "test_campaign")
	if err != nil {
		t.Fatalf("LoadProgress failed: %v", err)
	}
	if !loaded.Completed["chrysalis_1"] {
		t.Error("expected chrysalis_1 to be marked complete after load")
	}
}

func TestLoadProgressMissingFileReturnsEmpty(t *testing.T) {
	p, err := LoadProgress(filepath.Join(t.TempDir(), "nonexistent.json"), "test_campaign")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if len(p.Completed) != 0 {
		t.Error("expected empty progress for missing file")
	}
}

// --- LoadCampaign ---

func TestLoadCampaign(t *testing.T) {
	c, err := LoadCampaign("chrysalis_campaign.json")
	if err != nil {
		t.Fatalf("LoadCampaign failed: %v", err)
	}
	if c.ID != "chrysalis_act_1" {
		t.Errorf("unexpected campaign id: %s", c.ID)
	}
	if len(c.Levels) == 0 {
		t.Error("expected at least one campaign level")
	}
}

func TestLoadCampaignLoadNextLevel(t *testing.T) {
	c, err := LoadCampaign("chrysalis_campaign.json")
	if err != nil {
		t.Fatalf("LoadCampaign failed: %v", err)
	}
	lvl, err := c.LoadNextLevel(map[string]bool{})
	if err != nil {
		t.Fatalf("LoadNextLevel failed: %v", err)
	}
	if lvl == nil {
		t.Fatal("expected a level, got nil")
	}

	// Verify the loaded level is valid and runnable.
	e := lvl.CreateEngine()
	if e.Registry.Count == 0 {
		t.Error("engine has no drones")
	}
}

// --- LoadLevel from temp file for error cases ---

func TestLoadLevelRejectsInvalidJSON(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "bad.json")
	os.WriteFile(tmp, []byte(`{not valid json`), 0644)
	_, err := LoadLevel(tmp)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoadLevelRejectsFutureSchemaVersion(t *testing.T) {
	content := `{"schema_version":999,"id":"x","world":{"width":10,"height":10},"drones":{"initial_count":1},"mission":{"target_resources":1,"max_ticks":100,"infection_loss_threshold":0.5}}`
	tmp := filepath.Join(t.TempDir(), "future.json")
	os.WriteFile(tmp, []byte(content), 0644)
	_, err := LoadLevel(tmp)
	if err == nil {
		t.Fatal("expected error for unsupported schema_version, got nil")
	}
}
