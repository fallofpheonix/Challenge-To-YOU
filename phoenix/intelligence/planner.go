package intelligence

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ExecutionPlan struct {
	Problem          string   `json:"problem"`
	LikelyRootCauses []string `json:"likely_root_causes"`
	AffectedFiles    []string `json:"affected_files"`
	Tests            []string `json:"tests"`
	Risk             string   `json:"risk"`
	RepairPlan       []string `json:"repair_plan"`
}

type RepairPlanner struct{}

func NewRepairPlanner() *RepairPlanner {
	return &RepairPlanner{}
}

func (rp *RepairPlanner) GeneratePlan(problem string, failingFile string) (*ExecutionPlan, string, error) {
	plan := &ExecutionPlan{
		Problem:          problem,
		LikelyRootCauses: []string{"syntax error", "import scope issue", "unused variable definition"},
		AffectedFiles:    []string{failingFile},
		Tests:            []string{"go test ./..."},
		Risk:             "LOW",
		RepairPlan:       []string{"1. Identify the missing scope", "2. Inject required variable", "3. Re-build and run tests"},
	}

	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return nil, "", err
	}

	prompt := fmt.Sprintf("Analyze compilation/test errors and formulate a target repair strategy.\nTarget File: %s\nError Detail: %s\n\nReference Plan Skeleton:\n%s", failingFile, problem, string(data))
	return plan, prompt, nil
}

func (rp *RepairPlanner) ParsePlanJSON(response string) (*ExecutionPlan, error) {
	// Locate JSON blocks if returned
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("no JSON plan block found in response")
	}

	var plan ExecutionPlan
	if err := json.Unmarshal([]byte(response[start:end+1]), &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}
