package executionengine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"challenge-to-you/backend/internal/compiler"
	"challenge-to-you/backend/internal/sandbox"
)

type TestCase struct {
	ID             string `json:"id"`
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	Description    string `json:"description"`
	IsHidden       bool   `json:"is_hidden"`
	TimeoutMs      int    `json:"timeout_ms"`
	MemoryBytes    int    `json:"memory_bytes"`
}

type ValidationResult struct {
	AllPassed   bool         `json:"all_passed"`
	Results     []TestResult `json:"results"`
	Score       float64      `json:"score"`
	TotalTests  int          `json:"total_tests"`
	PassedTests int          `json:"passed_tests"`
	DurationMs  int          `json:"duration_ms"`
}

type TestResult struct {
	TestCaseID string `json:"test_case_id"`
	Passed     bool   `json:"passed"`
	Input      string `json:"input"`
	Expected   string `json:"expected_output"`
	Actual     string `json:"actual_output"`
	DurationMs int    `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

type Engine struct {
	compiler *compiler.Manager
	sandbox  *sandbox.ProcessSandbox
}

func NewEngine(compilerManager *compiler.Manager, sb *sandbox.ProcessSandbox) *Engine {
	return &Engine{
		compiler: compilerManager,
		sandbox:  sb,
	}
}

func (e *Engine) RunSingleTest(ctx context.Context, code string, test TestCase, lang string) (*TestResult, error) {
	start := time.Now()

	req := &sandbox.Request{
		Code:     code,
		Language: lang,
		Input:    test.Input,
		Config: sandbox.Config{
			TimeoutMs: 5000,
		},
	}

	if test.TimeoutMs > 0 {
		req.Config.TimeoutMs = test.TimeoutMs
	}

	resp, err := e.sandbox.Execute(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("sandbox execution failed: %w", err)
	}

	actual := strings.TrimSpace(resp.Output)
	expected := strings.TrimSpace(test.ExpectedOutput)
	passed := actual == expected && resp.Success

	return &TestResult{
		TestCaseID: test.ID,
		Passed:     passed,
		Input:      test.Input,
		Expected:   expected,
		Actual:     actual,
		DurationMs: int(time.Since(start).Milliseconds()),
		Error:      resp.Error,
	}, nil
}

func (e *Engine) RunAllTests(ctx context.Context, code string, tests []TestCase, lang string) (*ValidationResult, error) {
	start := time.Now()
	var results []TestResult
	passed := 0

	for _, test := range tests {
		result, err := e.RunSingleTest(ctx, code, test, lang)
		if err != nil {
			result = &TestResult{
				TestCaseID: test.ID,
				Passed:     false,
				Input:      test.Input,
				Expected:   test.ExpectedOutput,
				Error:      err.Error(),
			}
		}
		results = append(results, *result)
		if result.Passed {
			passed++
		}
	}

	total := len(tests)
	score := 0.0
	if total > 0 {
		score = float64(passed) / float64(total)
	}

	return &ValidationResult{
		AllPassed:   passed == total,
		Results:     results,
		Score:       score,
		TotalTests:  total,
		PassedTests: passed,
		DurationMs:  int(time.Since(start).Milliseconds()),
	}, nil
}

func (e *Engine) Validate(ctx context.Context, code string, tests []TestCase, lang string, validators []string) (*ValidationResult, error) {
	if len(validators) > 0 {
		return e.RunWithValidator(ctx, code, validators, tests, lang)
	}
	return e.RunAllTests(ctx, code, tests, lang)
}

func (e *Engine) RunWithValidator(ctx context.Context, code string, validators []string, tests []TestCase, lang string) (*ValidationResult, error) {
	start := time.Now()
	var results []TestResult
	passed := 0

	for _, test := range tests {
		validatorCode := code + "\n" + strings.Join(validators, "\n")
		result, err := e.RunSingleTest(ctx, validatorCode, test, lang)
		if err != nil {
			result = &TestResult{
				TestCaseID: test.ID,
				Passed:     false,
				Input:      test.Input,
				Expected:   test.ExpectedOutput,
				Error:      err.Error(),
			}
		}
		results = append(results, *result)
		if result.Passed {
			passed++
		}
	}

	total := len(tests)
	score := 0.0
	if total > 0 {
		score = float64(passed) / float64(total)
	}

	return &ValidationResult{
		AllPassed:   passed == total,
		Results:     results,
		Score:       score,
		TotalTests:  total,
		PassedTests: passed,
		DurationMs:  int(time.Since(start).Milliseconds()),
	}, nil
}
