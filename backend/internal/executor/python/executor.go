package python

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"challenge-to-you/backend/internal/compiler"
)

type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) Compile(ctx context.Context, code string, lang *compiler.Language) (*compiler.CompilationResult, error) {
	start := time.Now()

	tmpDir, err := os.MkdirTemp("", "python-compile-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	pyFile := filepath.Join(tmpDir, "solution.py")
	if err := os.WriteFile(pyFile, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write python file: %w", err)
	}

	compileCmd := "python3 -m py_compile"
	if lang != nil && lang.CompileCmd != "" {
		compileCmd = lang.CompileCmd
	}
	compileCmd = strings.ReplaceAll(compileCmd, "{file}", pyFile)

	args := strings.Fields(compileCmd)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	duration := int(time.Since(start).Milliseconds())

	if err != nil {
		errors := parsePythonErrors(stderr.String(), pyFile)
		return &compiler.CompilationResult{
			Success:    false,
			Output:     stderr.String(),
			Errors:     errors,
			DurationMs: duration,
		}, nil
	}

	return &compiler.CompilationResult{
		Success:    true,
		Output:     stdout.String(),
		DurationMs: duration,
	}, nil
}

func (e *Executor) Execute(ctx context.Context, code string, input string, lang *compiler.Language) (*compiler.ExecutionResult, error) {
	start := time.Now()

	tmpDir, err := os.MkdirTemp("", "python-run-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	pyFile := filepath.Join(tmpDir, "solution.py")
	if err := os.WriteFile(pyFile, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write python file: %w", err)
	}

	if input != "" {
		inputFile := filepath.Join(tmpDir, "input.txt")
		if err := os.WriteFile(inputFile, []byte(input), 0644); err != nil {
			return nil, fmt.Errorf("failed to write input file: %w", err)
		}
	}

	runCmd := "python3 {file}"
	if lang != nil && lang.RunCmd != "" {
		runCmd = lang.RunCmd
	}
	runCmd = strings.ReplaceAll(runCmd, "{file}", pyFile)

	args := strings.Fields(runCmd)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	duration := int(time.Since(start).Milliseconds())

	if ctx.Err() == context.DeadlineExceeded {
		return &compiler.ExecutionResult{
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
		}
	}

	return &compiler.ExecutionResult{
		Success:    exitCode == 0,
		Output:     stdout.String(),
		Error:      stderr.String(),
		ExitCode:   exitCode,
		DurationMs: duration,
	}, nil
}

func parsePythonErrors(stderr string, file string) []compiler.CompileError {
	var errors []compiler.CompileError
	lines := strings.Split(stderr, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "File \"") && strings.Contains(line, ", line ") {
			var err compiler.CompileError
			err.File = file

			parts := strings.Split(line, ", line ")
			if len(parts) >= 2 {
				lineNum := strings.Trim(parts[1], "\"")
				fmt.Sscanf(lineNum, "%d", &err.Line)
			}

			if idx := strings.Index(line, "SyntaxError: "); idx != -1 {
				err.Message = line[idx:]
			} else if idx := strings.Index(line, "IndentationError: "); idx != -1 {
				err.Message = line[idx:]
			} else if idx := strings.Index(line, "NameError: "); idx != -1 {
				err.Message = line[idx:]
			} else if idx := strings.Index(line, "TypeError: "); idx != -1 {
				err.Message = line[idx:]
			} else {
				err.Message = line
			}

			errors = append(errors, err)
		}
	}

	if len(errors) == 0 && stderr != "" {
		errors = append(errors, compiler.CompileError{
			Message: stderr,
		})
	}

	return errors
}
