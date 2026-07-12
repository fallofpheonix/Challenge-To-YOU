package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dop251/goja"
)

var (
	auditLogMu sync.Mutex
	auditLogID atomic.Int64
)

func nextAuditID() int64 {
	return auditLogID.Add(1)
}

func auditLog(entry string) {
	auditLogMu.Lock()
	defer auditLogMu.Unlock()
	log.Printf("[SANDBOX-AUDIT] %s", entry)
}

type Interface interface {
	Execute(ctx context.Context, req *Request) (*Response, error)
	Type() string
}

type Config struct {
	TimeoutMs     int  `json:"timeout_ms"`
	MemoryBytes   int  `json:"memory_bytes"`
	MaxOutputSize int  `json:"max_output_bytes"`
	NetworkAccess bool `json:"network_access"`
	FileSystem    bool `json:"filesystem_access"`
}

func DefaultConfig() Config {
	return Config{
		TimeoutMs:     5000,
		MemoryBytes:   256 * 1024 * 1024,
		MaxOutputSize: 1024 * 1024,
		NetworkAccess: false,
		FileSystem:    false,
	}
}

type Request struct {
	Code       string `json:"code"`
	Language   string `json:"language"`
	Input      string `json:"input"`
	Config     Config `json:"config"`
	WorkingDir string `json:"working_dir"`
}

type Response struct {
	Success    bool   `json:"success"`
	Output     string `json:"output"`
	Error      string `json:"error,omitempty"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int    `json:"duration_ms"`
	MemoryUsed int    `json:"memory_used_bytes"`
	TimedOut   bool   `json:"timed_out"`
}

type ProcessSandbox struct{}

func NewProcessSandbox() *ProcessSandbox {
	return &ProcessSandbox{}
}

func (s *ProcessSandbox) Type() string {
	return "native"
}

func (s *ProcessSandbox) Execute(ctx context.Context, req *Request) (*Response, error) {
	auditID := nextAuditID()
	_ = time.Now()

	if req.Config.TimeoutMs <= 0 {
		req.Config.TimeoutMs = 5000
	}
	if req.Config.MaxOutputSize <= 0 {
		req.Config.MaxOutputSize = 1024 * 1024
	}

	tmpDir, err := os.MkdirTemp("", "sandbox-*")
	if err != nil {
		auditLog(fmt.Sprintf("[%d] FAILED create_temp: %v", auditID, err))
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	var cleanupErr error
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			cleanupErr = fmt.Errorf("temp dir cleanup failed: %w", err)
		}
	}()

	var fileName string
	var runCmd string

	switch req.Language {
	case "python", "python3":
		fileName = "solution.py"
		runCmd = "python3 -I " + filepath.Join(tmpDir, fileName)
		// -I: isolated mode (no site-packages, no user site)
	case "go":
		fileName = "main.go"
		runCmd = "go run " + filepath.Join(tmpDir, fileName)
	case "javascript", "js", "node":
		fileName = "solution.js"
		runCmd = "node --experimental-policy= --experimental-disable-wasm --max-old-space-size=64 " + filepath.Join(tmpDir, fileName)
	case "java":
		fileName = "Main.java"
		runCmd = "javac " + filepath.Join(tmpDir, fileName) + " && java -cp " + tmpDir + " Main"
	default:
		fileName = "solution.txt"
		runCmd = "cat " + filepath.Join(tmpDir, fileName)
	}

	filePath := filepath.Join(tmpDir, fileName)
	if err := os.WriteFile(filePath, []byte(req.Code), 0600); err != nil {
		auditLog(fmt.Sprintf("[%d] FAILED write_file: %v", auditID, err))
		return nil, fmt.Errorf("failed to write code file: %w", err)
	}

	if req.Input != "" {
		inputFile := filepath.Join(tmpDir, "input.txt")
		if err := os.WriteFile(inputFile, []byte(req.Input), 0600); err != nil {
			return nil, fmt.Errorf("failed to write input file: %w", err)
		}
	}

	timeout := time.Duration(req.Config.TimeoutMs) * time.Millisecond
	execCtx, execCancel := context.WithTimeout(ctx, timeout)
	defer execCancel()

	parts := strings.Fields(runCmd)
	cmd := exec.CommandContext(execCtx, parts[0], parts[1:]...)
	cmd.Dir = tmpDir

	if req.Input != "" {
		cmd.Stdin = strings.NewReader(req.Input)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err = cmd.Run()
	duration := int(time.Since(start).Milliseconds())

	if execCtx.Err() == context.DeadlineExceeded {
		auditLog(fmt.Sprintf("[%d] TIMEOUT lang=%s duration=%dms timeout=%dms", auditID, req.Language, duration, req.Config.TimeoutMs))
		_ = strings.Join(cmd.Args, " ")
		if killErr := cmd.Cancel(); killErr != nil {
			_ = cmd.Wait()
		}
		return &Response{
			Success:    false,
			Error:      "execution timed out",
			DurationMs: duration,
			TimedOut:   true,
		}, nil
	}

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			auditLog(fmt.Sprintf("[%d] EXEC_ERROR lang=%s err=%v", auditID, req.Language, err))
			return nil, fmt.Errorf("execution failed: %w", err)
		}
	}

	output := stdout.String()
	if len(output) > req.Config.MaxOutputSize {
		output = output[:req.Config.MaxOutputSize]
		auditLog(fmt.Sprintf("[%d] OUTPUT_TRUNCATED lang=%s original=%d limit=%d", auditID, req.Language, len(stdout.String()), req.Config.MaxOutputSize))
	}

	if stderr.Len() > 0 {
		errOutput := stderr.String()
		if len(errOutput) > req.Config.MaxOutputSize {
			errOutput = errOutput[:req.Config.MaxOutputSize]
		}
		if output != "" {
			output += "\n"
		}
		output += errOutput
	}

	output = strings.TrimSpace(output)

	auditLog(fmt.Sprintf("[%d] SUCCESS lang=%s exit=%d duration=%dms output_size=%d", auditID, req.Language, exitCode, duration, len(output)))

	if cleanupErr != nil {
		auditLog(fmt.Sprintf("[%d] CLEANUP_WARN: %v", auditID, cleanupErr))
	}

	return &Response{
		Success:    exitCode == 0,
		Output:     output,
		Error:      stderr.String(),
		ExitCode:   exitCode,
		DurationMs: duration,
	}, nil
}

type Result struct {
	Output       string                 `json:"output"`
	MutatedState map[string]interface{} `json:"mutated_state"`
	DurationSecs float64                `json:"duration_secs"`
	OpsCount     int64                  `json:"ops_count"`
}

func Execute(code string, state map[string]interface{}, timeout time.Duration) (*Result, error) {
	auditID := nextAuditID()
	vm := goja.New()

	stateClone := make(map[string]interface{})
	for k, v := range state {
		stateClone[k] = v
	}

	vm.Set("getState", func(key string) interface{} {
		return stateClone[key]
	})

	vm.Set("setState", func(key string, value interface{}) {
		stateClone[key] = value
	})

	var opsCount int64
	vm.Set("incrementOp", func() {
		opsCount++
	})

	vm.Set("getOpsCount", func() int64 {
		return opsCount
	})

	var stdout bytes.Buffer
	vm.Set("console", map[string]interface{}{
		"log": func(args ...interface{}) {
			parts := make([]string, len(args))
			for i, arg := range args {
				parts[i] = fmt.Sprintf("%v", arg)
			}
			fmt.Fprintln(&stdout, strings.Join(parts, " "))
		},
	})

	var timeoutErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, timeoutErr = vm.RunString(code)
	}()

	select {
	case <-done:
		if timeoutErr != nil {
			return nil, fmt.Errorf("EXECUTION_ERROR: %v", timeoutErr)
		}
	case <-time.After(timeout):
		auditLog(fmt.Sprintf("[%d] JS_TIMEOUT timeout=%v", auditID, timeout))
		return nil, fmt.Errorf("TIMEOUT: execution exceeded %v", timeout)
	}

	auditLog(fmt.Sprintf("[%d] JS_SUCCESS ops=%d output_size=%d", auditID, opsCount, stdout.Len()))

	return &Result{
		Output:       strings.TrimSpace(stdout.String()),
		MutatedState: stateClone,
		DurationSecs: timeout.Seconds(),
		OpsCount:     opsCount,
	}, nil
}
