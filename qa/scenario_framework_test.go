package qa

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// ScenarioFrameworkSelfTest verifies the QA framework itself works.
func ScenarioFrameworkSelfTest(ctx *ScenarioContext) error {
	ctx.Step("framework_init", "QA framework initialized")

	// Test assertions
	ctx.Assert("assert_true", true, "true should pass")
	ctx.Assert("assert_false_cond", 1+1 == 2, "math should work")

	// Test directory creation
	dirs := []string{"reports", "backend_logs", "fixtures", "screenshots"}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create dir %s: %w", d, err)
		}
		ctx.Assert("dir_exists_"+d, true, d+" directory created")
	}

	// Test report generation
	report := SuiteReport{
		Project:        "Test",
		RunID:          "test_run",
		StartTime:      time.Now(),
		EndTime:        time.Now(),
		Duration:       100 * time.Millisecond,
		TotalScenarios: 1,
		Passed:         1,
		GoVersion:      runtime.Version(),
		Platform:       fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Scenarios: []ScenarioResult{
			{Name: "test", Status: StatusPassed, Duration: 100 * time.Millisecond},
		},
	}

	if err := WriteReport(report, "reports"); err != nil {
		return fmt.Errorf("write report: %w", err)
	}

	jsonPath := filepath.Join("reports", "qa_report.json")
	mdPath := filepath.Join("reports", "qa_report.md")

	jsonExists := false
	mdExists := false
	if _, err := os.Stat(jsonPath); err == nil {
		jsonExists = true
	}
	if _, err := os.Stat(mdPath); err == nil {
		mdExists = true
	}

	ctx.Assert("report_json_exists", jsonExists, "qa_report.json should exist")
	ctx.Assert("report_md_exists", mdExists, "qa_report.md should exist")

	return nil
}
