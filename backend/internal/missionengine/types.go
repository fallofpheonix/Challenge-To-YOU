package missionengine

import (
	"time"
)

type RoomDef struct {
	Description string   `json:"description"`
	Exits       []string `json:"exits"`
	Objects     []string `json:"objects"`
}

type WorldDef struct {
	StartRoom string             `json:"start_room"`
	Rooms     map[string]RoomDef `json:"rooms"`
}

// Mission represents a story-driven sequence of challenges
type Mission struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	World       *WorldDef `json:"world,omitempty"`
	Era         string    `json:"era"`
	Act         int       `json:"act"`
	Chapter     int       `json:"chapter"`

	Intro      *Cutscene          `json:"intro,omitempty"`
	Outro      *Cutscene          `json:"outro,omitempty"`
	Dialogue   []string           `json:"dialogue,omitempty"`
	Objectives []Objective        `json:"objectives"`
	Challenges []MissionChallenge `json:"challenges"`
	Reward     *MissionReward     `json:"reward,omitempty"`

	Requirements []string `json:"requirements,omitempty"`

	EstTimeMins int     `json:"est_time_mins"`
	Difficulty  float64 `json:"difficulty"`
	Paradigm    string  `json:"paradigm,omitempty"`

	IsSideQuest  bool `json:"is_side_quest,omitempty"`
	IsRepeatable bool `json:"is_repeatable,omitempty"`
	MaxReplays   int  `json:"max_replays,omitempty"`
}

// MissionChallenge links a challenge to a mission
type MissionChallenge struct {
	ID         string  `json:"id"`
	File       string  `json:"file,omitempty"`
	Difficulty float64 `json:"difficulty"`
	Mode       string  `json:"mode,omitempty"`
	OutputKey  string  `json:"output_key,omitempty"`
	IsOptional bool    `json:"is_optional,omitempty"`
}

// Objective represents a single task within a mission
type Objective struct {
	ID          string `json:"id"`
	Type        string `json:"type"` // "complete_challenge", "find_item", "talk_to_npc", "explore_area", "trigger_event"
	Target      string `json:"target"`
	Description string `json:"description"`
	IsOptional  bool   `json:"is_optional,omitempty"`
	IsRequired  bool   `json:"is_required,omitempty"`

	// State conditions
	RequiresState map[string]interface{} `json:"requires_state,omitempty"`
	SetsState     map[string]interface{} `json:"sets_state,omitempty"`
}

// Cutscene represents a non-interactive narrative sequence
type Cutscene struct {
	ID     string          `json:"id"`
	Type   string          `json:"type"` // "cinematic", "dialogue", "tutorial"
	Scenes []CutsceneScene `json:"scenes"`
}

// CutsceneScene is a single frame in a cutscene
type CutsceneScene struct {
	Background string         `json:"background,omitempty"`
	Characters []string       `json:"characters,omitempty"`
	Dialogue   []DialogueLine `json:"dialogue,omitempty"`
	Duration   int            `json:"duration_ms,omitempty"`
	Effect     string         `json:"effect,omitempty"`
}

// DialogueLine is a single line of dialogue
type DialogueLine struct {
	Speaker  string `json:"speaker"`
	Text     string `json:"text"`
	Emotion  string `json:"emotion,omitempty"`
	NextNode string `json:"next_node,omitempty"`
}

// MissionReward defines rewards for completing a mission
type MissionReward struct {
	XP           int      `json:"xp"`
	Credits      int      `json:"credits"`
	Unlocks      []string `json:"unlocks,omitempty"`
	Achievements []string `json:"achievements,omitempty"`
	Items        []string `json:"items,omitempty"`
}

// MissionStatus represents the current state of a mission
type MissionStatus int

const (
	MissionStatusLocked    MissionStatus = iota // Not yet available
	MissionStatusAvailable                      // Can be started
	MissionStatusActive                         // Currently in progress
	MissionStatusCompleted                      // Successfully finished
	MissionStatusFailed                         // Failed
	MissionStatusExpired                        // Time limit exceeded
)

func (s MissionStatus) String() string {
	switch s {
	case MissionStatusLocked:
		return "locked"
	case MissionStatusAvailable:
		return "available"
	case MissionStatusActive:
		return "active"
	case MissionStatusCompleted:
		return "completed"
	case MissionStatusFailed:
		return "failed"
	case MissionStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// ObjectiveStatus represents the state of an objective
type ObjectiveStatus int

const (
	ObjectiveStatusPending ObjectiveStatus = iota
	ObjectiveStatusActive
	ObjectiveStatusCompleted
	ObjectiveStatusFailed
	ObjectiveStatusSkipped
)

func (s ObjectiveStatus) String() string {
	switch s {
	case ObjectiveStatusPending:
		return "pending"
	case ObjectiveStatusActive:
		return "active"
	case ObjectiveStatusCompleted:
		return "completed"
	case ObjectiveStatusFailed:
		return "failed"
	case ObjectiveStatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// MissionSession tracks the runtime state of an active mission
type MissionSession struct {
	Mission             *Mission                   `json:"mission"`
	PlayerID            string                     `json:"player_id"`
	Status              MissionStatus              `json:"status"`
	CurrentStep         int                        `json:"current_step"`
	ObjectiveStatuses   map[string]ObjectiveStatus `json:"objective_statuses"`
	State               map[string]interface{}     `json:"state,omitempty"`
	FabricState         map[string]interface{}     `json:"fabric_state,omitempty"` // Serialized AxiomaticFabric state for persistence
	StartTime           time.Time                  `json:"start_time"`
	EndTime             *time.Time                 `json:"end_time,omitempty"`
	CompletedChallenges []string                   `json:"completed_challenges"`
	FailedObjectives    []string                   `json:"failed_objectives"`
}

// IsComplete returns true if all required objectives are completed
func (s *MissionSession) IsComplete() bool {
	for _, obj := range s.Mission.Objectives {
		if obj.IsRequired || !obj.IsOptional {
			status, ok := s.ObjectiveStatuses[obj.ID]
			if !ok || status != ObjectiveStatusCompleted {
				return false
			}
		}
	}
	return true
}

// GetProgress returns a float64 between 0 and 1 representing completion
func (s *MissionSession) GetProgress() float64 {
	if len(s.Mission.Objectives) == 0 {
		return 0
	}

	completed := 0
	for _, obj := range s.Mission.Objectives {
		if status, ok := s.ObjectiveStatuses[obj.ID]; ok {
			if status == ObjectiveStatusCompleted {
				completed++
			}
		}
	}

	return float64(completed) / float64(len(s.Mission.Objectives))
}

// CanStart checks if a mission can be started given completed missions
func CanStart(mission *Mission, completedMissions map[string]bool) bool {
	for _, req := range mission.Requirements {
		if !completedMissions[req] {
			return false
		}
	}
	return true
}

// GetActiveObjectives returns objectives that are currently active
func (s *MissionSession) GetActiveObjectives() []Objective {
	var active []Objective
	for _, obj := range s.Mission.Objectives {
		status := s.ObjectiveStatuses[obj.ID]
		if status == ObjectiveStatusPending || status == ObjectiveStatusActive {
			active = append(active, obj)
		}
	}
	return active
}
