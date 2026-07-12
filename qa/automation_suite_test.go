package qa

import (
	"fmt"
	"os"
	"testing"
)

func TestQASuite(t *testing.T) {
	runner := NewRunner()

	// Milestone 1: Framework self-test
	runner.Add(Scenario{Name: "FrameworkSelfTest", Category: "framework", Fn: ScenarioFrameworkSelfTest})

	// Milestone 2: Backend Automation
	runner.Add(Scenario{Name: "BackendLaunchAndShutdown", Category: "backend", Fn: ScenarioBackendLaunchShutdown})
	runner.Add(Scenario{Name: "BackendHealthCheck", Category: "backend", Fn: ScenarioBackendHealthCheck})
	runner.Add(Scenario{Name: "BackendLogCapture", Category: "backend", Fn: ScenarioBackendLogCapture})

	// Milestone 3: WebSocket Automation
	runner.Add(Scenario{Name: "WebSocketConnect", Category: "websocket", Fn: ScenarioWSConnect})
	runner.Add(Scenario{Name: "WebSocketDisconnect", Category: "websocket", Fn: ScenarioWSDisconnect})
	runner.Add(Scenario{Name: "WebSocketReconnect", Category: "websocket", Fn: ScenarioWSReconnect})
	runner.Add(Scenario{Name: "WebSocketInvalidMessage", Category: "websocket", Fn: ScenarioWSInvalidMessage})
	runner.Add(Scenario{Name: "WebSocketTimeout", Category: "websocket", Fn: ScenarioWSTimeout})

	// Milestone 4: Mission Automation
	runner.Add(Scenario{Name: "MissionStart", Category: "mission", Fn: ScenarioMissionStart})
	runner.Add(Scenario{Name: "MissionObjectives", Category: "mission", Fn: ScenarioMissionObjectives})
	runner.Add(Scenario{Name: "MissionStateTransitions", Category: "mission", Fn: ScenarioMissionStateTransitions})
	runner.Add(Scenario{Name: "MissionCompletion", Category: "mission", Fn: ScenarioMissionCompletion})
	runner.Add(Scenario{Name: "MissionUnlocks", Category: "mission", Fn: ScenarioMissionUnlocks})

	// Milestone 5: Challenge Automation
	runner.Add(Scenario{Name: "ChallengeLoad", Category: "challenge", Fn: ScenarioChallengeLoad})
	runner.Add(Scenario{Name: "ChallengeCompile", Category: "challenge", Fn: ScenarioChallengeCompile})
	runner.Add(Scenario{Name: "ChallengeExecute", Category: "challenge", Fn: ScenarioChallengeExecute})
	runner.Add(Scenario{Name: "ChallengeScoring", Category: "challenge", Fn: ScenarioChallengeScoring})

	// Milestone 6: Save Verification
	runner.Add(Scenario{Name: "SaveCreateLoad", Category: "save", Fn: ScenarioSaveCreateLoad})
	runner.Add(Scenario{Name: "SaveRestartReload", Category: "save", Fn: ScenarioSaveRestartReload})
	runner.Add(Scenario{Name: "SaveCorruptedHandling", Category: "save", Fn: ScenarioSaveCorruptedHandling})

	// Milestone 7: Reward Verification
	runner.Add(Scenario{Name: "RewardXP", Category: "reward", Fn: ScenarioRewardXP})
	runner.Add(Scenario{Name: "RewardCredits", Category: "reward", Fn: ScenarioRewardCredits})
	runner.Add(Scenario{Name: "RewardReputation", Category: "reward", Fn: ScenarioRewardReputation})
	runner.Add(Scenario{Name: "RewardLuck", Category: "reward", Fn: ScenarioRewardLuck})
	runner.Add(Scenario{Name: "RewardUnlocks", Category: "reward", Fn: ScenarioRewardUnlocks})

	// Milestone 8: Regression Detection
	runner.Add(Scenario{Name: "RegressionDetection", Category: "regression", Fn: ScenarioRegressionDetection})

	// Milestone 10: Continuous Playthrough
	runner.Add(Scenario{Name: "ContinuousPlaythrough", Category: "e2e", Fn: ScenarioContinuousPlaythrough})

	// Run all
	report := runner.RunAll()

	// Write reports
	if err := WriteReport(report, "reports"); err != nil {
		t.Errorf("Failed to write reports: %v", err)
	}

	// Print summary
	fmt.Printf("\n═══════════════════════════════════════════════\n")
	fmt.Printf("  QA SUITE RESULTS: %d/%d PASSED\n", report.Passed, report.TotalScenarios)
	fmt.Printf("═══════════════════════════════════════════════\n")
	fmt.Printf("  Passed:  %d\n", report.Passed)
	fmt.Printf("  Failed:  %d\n", report.Failed)
	fmt.Printf("  Errors:  %d\n", report.Errors)
	fmt.Printf("  Skipped: %d\n", report.Skipped)
	fmt.Printf("  Duration: %dms\n\n", report.Duration.Milliseconds())

	if report.Failed > 0 || report.Errors > 0 {
		for _, s := range report.Scenarios {
			if s.Status == StatusFailed || s.Status == StatusError {
				fmt.Printf("  FAIL: %s (%s) - %s\n", s.Name, s.Category, s.Error)
			}
		}
		fmt.Println()
		t.Errorf("QA suite had %d failures and %d errors", report.Failed, report.Errors)
	} else {
		fmt.Println("  All scenarios passed!")
	}

	// Print report location
	absPath, _ := os.Getwd()
	fmt.Printf("\n  Reports: %s/reports/\n", absPath)
}
