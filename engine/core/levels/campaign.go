package levels

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// CampaignEntry is one slot in the campaign level sequence.
// Requires lists level IDs that must be completed before this entry unlocks.
// An empty Requires slice means the level is available from the start.
type CampaignEntry struct {
	LevelFile string   `json:"level_file"` // path relative to the campaign JSON directory
	LevelID   string   `json:"level_id"`   // id declared inside the level JSON, used for progress tracking
	Requires  []string `json:"requires,omitempty"`
}

// Campaign defines an ordered mission sequence with an unlock DAG.
// Levels are evaluated in declaration order; the first unlocked, incomplete
// level is treated as the active mission.
type Campaign struct {
	ID     string          `json:"id"`
	Title  string          `json:"title"`
	Levels []CampaignEntry `json:"levels"`

	dir string // directory of the campaign file; set by LoadCampaign
}

// LoadCampaign reads and parses a campaign JSON file.
func LoadCampaign(path string) (*Campaign, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Campaign
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	c.dir = filepath.Dir(path)
	return &c, nil
}

// NextLevel returns the first entry whose requirements are satisfied by the
// completed set and which has not itself been completed. Returns nil when
// the entire campaign is finished or all remaining levels are locked.
func (c *Campaign) NextLevel(completed map[string]bool) *CampaignEntry {
	for i := range c.Levels {
		entry := &c.Levels[i]
		if completed[entry.LevelID] {
			continue
		}
		unlocked := true
		for _, req := range entry.Requires {
			if !completed[req] {
				unlocked = false
				break
			}
		}
		if unlocked {
			return entry
		}
	}
	return nil
}

// LoadNextLevel resolves the path for the next incomplete level and loads it.
// Returns (nil, nil) when the campaign is finished.
func (c *Campaign) LoadNextLevel(completed map[string]bool) (*Level, error) {
	entry := c.NextLevel(completed)
	if entry == nil {
		return nil, nil
	}
	return LoadLevel(filepath.Join(c.dir, entry.LevelFile))
}

// CampaignProgress tracks which levels have been completed. It is serialized
// to disk alongside the campaign file so progress persists between sessions.
type CampaignProgress struct {
	CampaignID string          `json:"campaign_id"`
	Completed  map[string]bool `json:"completed"`
}

// NewCampaignProgress creates an empty progress record for the given campaign.
func NewCampaignProgress(campaignID string) *CampaignProgress {
	return &CampaignProgress{
		CampaignID: campaignID,
		Completed:  make(map[string]bool),
	}
}

// LoadProgress reads a progress JSON file written by Save. If the file does
// not exist it returns an empty progress record for campaignID.
func LoadProgress(path, campaignID string) (*CampaignProgress, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return NewCampaignProgress(campaignID), nil
	}
	if err != nil {
		return nil, err
	}
	var p CampaignProgress
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	if p.Completed == nil {
		p.Completed = make(map[string]bool)
	}
	return &p, nil
}

// Complete marks levelID as finished.
func (p *CampaignProgress) Complete(levelID string) {
	p.Completed[levelID] = true
}

// Save writes the progress record to path as JSON.
func (p *CampaignProgress) Save(path string) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
