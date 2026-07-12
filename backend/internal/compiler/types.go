package compiler

import "context"

type Language struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Extensions  []string `json:"extensions"`
	CompileCmd  string   `json:"compile_cmd"`
	RunCmd      string   `json:"run_cmd"`
	TimeoutMs   int      `json:"timeout_ms"`
	MemoryBytes int      `json:"memory_bytes"`
	EnvVars     []string `json:"env_vars"`
}

type CompilationResult struct {
	Success    bool           `json:"success"`
	Output     string         `json:"output"`
	Errors     []CompileError `json:"errors"`
	Warnings   []string       `json:"warnings"`
	DurationMs int            `json:"duration_ms"`
}

type CompileError struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	File    string `json:"file"`
}

type ExecutionResult struct {
	Success    bool   `json:"success"`
	Output     string `json:"output"`
	Error      string `json:"error,omitempty"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int    `json:"duration_ms"`
	MemoryUsed int    `json:"memory_used_bytes"`
	TimedOut   bool   `json:"timed_out"`
}

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

type Executor interface {
	Compile(ctx context.Context, code string, lang *Language) (*CompilationResult, error)
	Execute(ctx context.Context, code string, input string, lang *Language) (*ExecutionResult, error)
}
