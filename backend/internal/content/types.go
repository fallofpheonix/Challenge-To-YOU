package content

import "time"

const CurrentSchemaVersion = "2.0.0"

type Challenge struct {
	SchemaVersion string `json:"schema_version"`
	ID            string `json:"id"`
	Version       string `json:"version"`

	Title       string `json:"title"`
	Description string `json:"description"`
	Story       string `json:"story"`
	Mission     string `json:"mission"`

	Difficulty float64  `json:"difficulty"`
	Category   string   `json:"category"`
	Tags       []string `json:"tags"`

	SupportedLanguages []string `json:"supported_languages"`
	StarterCode        string   `json:"starter_code"`
	ReferenceSolution  string   `json:"reference_solution"`
	AltSolutions       []string `json:"alt_solutions"`

	InputSpec   string   `json:"input_spec"`
	OutputSpec  string   `json:"output_spec"`
	Constraints []string `json:"constraints"`

	Examples []Example `json:"examples"`

	VisibleTests []TestCase `json:"visible_tests"`
	HiddenTests  []TestCase `json:"hidden_tests"`
	Validator    string     `json:"validator"`

	TimeLimit   int `json:"time_limit"`
	MemoryLimit int `json:"memory_limit"`

	XP      int           `json:"xp"`
	Rewards []Reward      `json:"rewards"`
	Unlock  []string      `json:"unlock_conditions"`
	Achieve []Achievement `json:"achievements"`

	DialogueID string         `json:"dialogue_id"`
	Hints      []Hint         `json:"hints"`
	Analytics  *AnalyticsMeta `json:"analytics,omitempty"`

	EstimatedTime      int      `json:"estimated_time"`
	LearningObjectives []string `json:"learning_objectives"`
	Complexity         string   `json:"complexity"`

	Metadata map[string]interface{} `json:"metadata,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Example struct {
	Input       string `json:"input"`
	Output      string `json:"output"`
	Description string `json:"description"`
}

type TestCase struct {
	ID             string `json:"id"`
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	Description    string `json:"description"`
	IsHidden       bool   `json:"is_hidden"`
	TimeLimit      int    `json:"time_limit,omitempty"`
	MemoryLimit    int    `json:"memory_limit,omitempty"`
}

type Reward struct {
	Type   string `json:"type"`
	Value  string `json:"value"`
	Amount int    `json:"amount"`
}

type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Hint struct {
	Level   int    `json:"level"`
	Cost    int    `json:"cost"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

type AnalyticsMeta struct {
	AvgCompletionTime int      `json:"avg_completion_time"`
	SuccessRate       float64  `json:"success_rate"`
	CommonErrors      []string `json:"common_errors"`
}

type ChallengePack struct {
	Version       int            `json:"version"`
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Era           string         `json:"era"`
	Tier          int            `json:"tier"`
	Prerequisites []string       `json:"prerequisites"`
	Unlocks       []string       `json:"unlocks"`
	LuckModifier  float64        `json:"luck_modifier"`
	Challenges    []ChallengeRef `json:"challenges"`
}

type ChallengeRef struct {
	ID         string   `json:"id"`
	File       string   `json:"file"`
	Difficulty float64  `json:"difficulty"`
	Category   string   `json:"category"`
	Tags       []string `json:"tags"`
}
