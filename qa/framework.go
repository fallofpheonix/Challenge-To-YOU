package qa

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ScenarioStatus represents the outcome of a test scenario.
type ScenarioStatus string

const (
	StatusPassed  ScenarioStatus = "PASSED"
	StatusFailed  ScenarioStatus = "FAILED"
	StatusSkipped ScenarioStatus = "SKIPPED"
	StatusError   ScenarioStatus = "ERROR"
)

// StepResult records a single assertion or action within a scenario.
type StepResult struct {
	Name      string        `json:"name"`
	Status    ScenarioStatus `json:"status"`
	Duration  time.Duration `json:"duration_ms"`
	Error     string        `json:"error,omitempty"`
	Details   string        `json:"details,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// ScenarioResult holds the full outcome of one scenario.
type ScenarioResult struct {
	Name       string         `json:"name"`
	Category   string         `json:"category"`
	Status     ScenarioStatus `json:"status"`
	Steps      []StepResult   `json:"steps"`
	Duration   time.Duration  `json:"duration_ms"`
	Error      string         `json:"error,omitempty"`
	StartTime  time.Time      `json:"start_time"`
	EndTime    time.Time      `json:"end_time"`
}

// SuiteReport is the top-level report for a full QA run.
type SuiteReport struct {
	Project     string           `json:"project"`
	RunID       string           `json:"run_id"`
	StartTime   time.Time        `json:"start_time"`
	EndTime     time.Time        `json:"end_time"`
	Duration    time.Duration    `json:"duration_ms"`
	TotalScenarios int           `json:"total_scenarios"`
	Passed      int              `json:"passed"`
	Failed      int              `json:"failed"`
	Skipped     int              `json:"skipped"`
	Errors      int              `json:"errors"`
	Scenarios   []ScenarioResult `json:"scenarios"`
	GoVersion   string           `json:"go_version"`
	Platform    string           `json:"platform"`
}

// Scenario is a single test scenario to be executed.
type Scenario struct {
	Name     string
	Category string
	Fn       func(ctx *ScenarioContext) error
}

// ScenarioContext provides helpers and state for a running scenario.
type ScenarioContext struct {
	Name       string
	Category   string
	Steps      []StepResult
	StartTime  time.Time
	Error      error
	T          interface{ Errorf(string, ...any); Logf(string, ...any); FailNow() }
	FixturesDir string
	ReportsDir  string
	LogsDir     string
}

// Assert records a named assertion step.
func (ctx *ScenarioContext) Assert(name string, condition bool, detail string) {
	status := StatusPassed
	errMsg := ""
	if !condition {
		status = StatusFailed
		errMsg = detail
		ctx.Error = fmt.Errorf("assertion failed: %s: %s", name, detail)
	}
	ctx.Steps = append(ctx.Steps, StepResult{
		Name:      name,
		Status:    status,
		Error:     errMsg,
		Details:   detail,
		Timestamp: time.Now(),
	})
}

// Require is like Assert but marks the scenario for early exit on failure.
func (ctx *ScenarioContext) Require(name string, condition bool, detail string) {
	ctx.Assert(name, condition, detail)
	if !condition {
		ctx.Error = fmt.Errorf("require failed: %s: %s", name, detail)
	}
}

// Step records an informational step (always passes).
func (ctx *ScenarioContext) Step(name string, detail string) {
	ctx.Steps = append(ctx.Steps, StepResult{
		Name:      name,
		Status:    StatusPassed,
		Details:   detail,
		Timestamp: time.Now(),
	})
}

// Log records a log message for debugging.
func (ctx *ScenarioContext) Log(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	ctx.Steps = append(ctx.Steps, StepResult{
		Name:      "log",
		Status:    StatusPassed,
		Details:   msg,
		Timestamp: time.Now(),
	})
}

// Runner executes scenarios and produces a report.
type Runner struct {
	scenarios []Scenario
	report    SuiteReport
	mu        sync.Mutex
}

// NewRunner creates a new QA runner.
func NewRunner() *Runner {
	return &Runner{}
}

// Add registers a scenario with the runner.
func (r *Runner) Add(s Scenario) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scenarios = append(r.scenarios, s)
}

// RunAll executes all registered scenarios and returns the report.
func (r *Runner) RunAll() SuiteReport {
	r.report = SuiteReport{
		Project:   "Challenge To YOU",
		RunID:     fmt.Sprintf("qa_%d", time.Now().UnixMilli()),
		StartTime: time.Now(),
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	for _, s := range r.scenarios {
		result := r.runScenario(s)
		r.report.Scenarios = append(r.report.Scenarios, result)
		r.report.TotalScenarios++
		switch result.Status {
		case StatusPassed:
			r.report.Passed++
		case StatusFailed:
			r.report.Failed++
		case StatusSkipped:
			r.report.Skipped++
		case StatusError:
			r.report.Errors++
		}
	}

	r.report.EndTime = time.Now()
	r.report.Duration = r.report.EndTime.Sub(r.report.StartTime)
	return r.report
}

func (r *Runner) runScenario(s Scenario) (result ScenarioResult) {
	result = ScenarioResult{
		Name:      s.Name,
		Category:  s.Category,
		StartTime: time.Now(),
	}

	ctx := &ScenarioContext{
		Name:        s.Name,
		Category:    s.Category,
		FixturesDir: "fixtures",
		ReportsDir:  "reports",
		LogsDir:     "backend_logs",
	}

	defer func() {
		if rec := recover(); rec != nil {
			result.Status = StatusError
			result.Error = fmt.Sprintf("panic: %v", rec)
		}
		result.Steps = ctx.Steps
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		if result.Status == "" {
			if ctx.Error != nil {
				result.Status = StatusFailed
				result.Error = ctx.Error.Error()
			} else {
				result.Status = StatusPassed
			}
		}
	}()

	err := s.Fn(ctx)
	if err != nil {
		result.Status = StatusFailed
		result.Error = err.Error()
	}

	return
}

// WriteReport generates qa_report.json and qa_report.md in the reports directory.
func WriteReport(report SuiteReport, reportsDir string) error {
	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		return fmt.Errorf("create reports dir: %w", err)
	}

	jsonPath := filepath.Join(reportsDir, "qa_report.json")
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	if err := os.WriteFile(jsonPath, jsonData, 0o644); err != nil {
		return fmt.Errorf("write json: %w", err)
	}

	mdPath := filepath.Join(reportsDir, "qa_report.md")
	md := generateMarkdownReport(report)
	if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}

	return nil
}

func generateMarkdownReport(r SuiteReport) string {
	var b strings.Builder

	b.WriteString("# QA Automation Report\n\n")
	b.WriteString(fmt.Sprintf("**Project:** %s\n", r.Project))
	b.WriteString(fmt.Sprintf("**Run ID:** %s\n", r.RunID))
	b.WriteString(fmt.Sprintf("**Time:** %s — %s\n", r.StartTime.Format(time.RFC3339), r.EndTime.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("**Duration:** %dms\n\n", r.Duration.Milliseconds()))

	b.WriteString("## Summary\n\n")
	b.WriteString(fmt.Sprintf("| Metric | Count |\n"))
	b.WriteString(fmt.Sprintf("|--------|-------|\n"))
	b.WriteString(fmt.Sprintf("| Total | %d |\n", r.TotalScenarios))
	b.WriteString(fmt.Sprintf("| Passed | %d |\n", r.Passed))
	b.WriteString(fmt.Sprintf("| Failed | %d |\n", r.Failed))
	b.WriteString(fmt.Sprintf("| Skipped | %d |\n", r.Skipped))
	b.WriteString(fmt.Sprintf("| Errors | %d |\n", r.Errors))

	if r.Failed > 0 || r.Errors > 0 {
		b.WriteString(fmt.Sprintf("\n**Status: FAILING**\n\n"))
	} else {
		b.WriteString(fmt.Sprintf("\n**Status: ALL PASSING**\n\n"))
	}

	b.WriteString("## Scenarios\n\n")
	for _, s := range r.Scenarios {
		icon := "PASS"
		if s.Status == StatusFailed {
			icon = "FAIL"
		} else if s.Status == StatusError {
			icon = "ERR"
		} else if s.Status == StatusSkipped {
			icon = "SKIP"
		}
		b.WriteString(fmt.Sprintf("### [%s] %s (%s)\n\n", icon, s.Name, s.Category))
		b.WriteString(fmt.Sprintf("Duration: %dms\n\n", s.Duration.Milliseconds()))

		if s.Error != "" {
			b.WriteString(fmt.Sprintf("**Error:** %s\n\n", s.Error))
		}

		if len(s.Steps) > 0 {
			b.WriteString("| Step | Status | Details |\n")
			b.WriteString("|------|--------|--------|\n")
			for _, step := range s.Steps {
				details := step.Details
				if len(details) > 80 {
					details = details[:77] + "..."
				}
				details = strings.ReplaceAll(details, "|", "\\|")
				details = strings.ReplaceAll(details, "\n", " ")
				b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", step.Name, step.Status, details))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("Generated by QA Automation Suite | %s | %s\n", r.GoVersion, r.Platform))

	return b.String()
}
