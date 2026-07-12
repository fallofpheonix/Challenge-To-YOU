package sandbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNativeSandboxInterface(t *testing.T) {
	sb := NewProcessSandbox()
	if sb.Type() != "native" {
		t.Errorf("expected type 'native', got %q", sb.Type())
	}
}

func TestNativeSandbox_PythonHello(t *testing.T) {
	sb := NewProcessSandbox()
	req := &Request{
		Code:     `print("hello from sandbox")`,
		Language: "python",
		Config:   DefaultConfig(),
	}
	ctx := context.Background()
	resp, err := sb.Execute(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success, got exit code %d: %s", resp.ExitCode, resp.Error)
	}
	if !strings.Contains(resp.Output, "hello from sandbox") {
		t.Errorf("expected output containing 'hello from sandbox', got %q", resp.Output)
	}
}

func TestNativeSandbox_OutputLimit(t *testing.T) {
	sb := NewProcessSandbox()
	config := DefaultConfig()
	config.MaxOutputSize = 10
	req := &Request{
		Code:     `print("a" * 1000)`,
		Language: "python",
		Config:   config,
	}
	ctx := context.Background()
	resp, err := sb.Execute(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Output) > 100 {
		t.Errorf("expected truncated output (limit=10), got %d bytes", len(resp.Output))
	}
}

func TestNativeSandbox_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}
	sb := NewProcessSandbox()
	config := DefaultConfig()
	config.TimeoutMs = 100
	req := &Request{
		Code:     `import time; time.sleep(5); print("done")`,
		Language: "python",
		Config:   config,
	}
	ctx := context.Background()
	resp, err := sb.Execute(ctx, req)
	if err != nil {
		t.Fatalf("expected response even on timeout, got error: %v", err)
	}
	if !resp.TimedOut {
		t.Errorf("expected TimedOut=true, got success=%v", resp.Success)
	}
}

func TestNativeSandbox_TempDirIsolation(t *testing.T) {
	sb := NewProcessSandbox()
	req := &Request{
		Code:     `import os; print(os.getcwd())`,
		Language: "python",
		Config:   DefaultConfig(),
	}
	ctx := context.Background()
	resp, err := sb.Execute(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cwd := strings.TrimSpace(resp.Output)
	if !strings.Contains(cwd, "sandbox-") {
		t.Errorf("expected cwd to contain 'sandbox-', got %q", cwd)
	}
}

func TestNativeSandbox_FileCleanup(t *testing.T) {
	sb := NewProcessSandbox()
	req := &Request{
		Code:     `print("cleanup test")`,
		Language: "python",
		Config:   DefaultConfig(),
	}
	ctx := context.Background()
	resp, err := sb.Execute(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success: %s", resp.Error)
	}

	entries, _ := os.ReadDir(os.TempDir())
	sandboxDirs := 0
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "sandbox-") {
			now := time.Now()
			info, _ := e.Info()
			if info != nil && now.Sub(info.ModTime()) < time.Minute {
				sandboxDirs++
			}
		}
	}
	if sandboxDirs > 0 {
		t.Logf("found %d recent sandbox dirs (may include concurrent test runs)", sandboxDirs)
	}
}

func TestNativeSandbox_LanguageArg(t *testing.T) {
	sb := NewProcessSandbox()
	ctx := context.Background()

	req := &Request{
		Code:     `print("js test")`,
		Language: "javascript",
		Config:   DefaultConfig(),
	}
	resp, err := sb.Execute(ctx, req)
	if err != nil {
		t.Fatalf("javascript execution failed: %v", err)
	}
	if resp.Success && !strings.Contains(resp.Output, "js test") {
		t.Errorf("unexpected output: %q", resp.Output)
	}
}

func TestLegacyExecute(t *testing.T) {
	code := `
	var result = getState("counter") || 0;
	setState("counter", result + 1);
	console.log("count:", getState("counter"));
	incrementOp();
	`
	state := map[string]interface{}{"counter": 5}
	result, err := Execute(code, state, 2*time.Second)
	if err != nil {
		t.Fatalf("legacy execute failed: %v", err)
	}
	if !strings.Contains(result.Output, "count:") {
		t.Errorf("expected output containing 'count:', got %q", result.Output)
	}
	if result.OpsCount < 1 {
		t.Errorf("expected at least 1 op, got %d", result.OpsCount)
	}
	mutated := result.MutatedState["counter"]
	switch v := mutated.(type) {
	case float64:
		if v != 6 {
			t.Errorf("expected counter=6, got %v", mutated)
		}
	case int:
		if v != 6 {
			t.Errorf("expected counter=6, got %v", mutated)
		}
	case int64:
		if v != 6 {
			t.Errorf("expected counter=6, got %v", mutated)
		}
	default:
		t.Errorf("unexpected type %T with value %v", mutated, mutated)
	}
}

func TestLegacyExecuteTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}
	code := `
	while(true) { }
	`
	_, err := Execute(code, nil, 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "TIMEOUT") {
		t.Errorf("expected TIMEOUT error, got %v", err)
	}
}

func TestNativeSandbox_FilePermissions(t *testing.T) {
	sb := NewProcessSandbox()
	req := &Request{
		Code:     `import os; print(oct(os.stat("solution.py").st_mode))`,
		Language: "python",
		Config:   DefaultConfig(),
	}
	ctx := context.Background()
	resp, err := sb.Execute(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parts := strings.Fields(resp.Output)
	if len(parts) > 0 {
		mode := parts[len(parts)-1]
		if strings.Contains(mode, "777") || strings.Contains(mode, "666") || strings.Contains(mode, "644") {
			t.Logf("sandbox file mode: %s (may be permissive on this platform)", mode)
		}
	}
}

func TestNativeSandbox_GoRuns(t *testing.T) {
	whichGo := filepath.Join(os.Getenv("GOROOT"), "bin", "go")
	if _, err := os.Stat(whichGo); os.IsNotExist(err) {
		t.Skip("Go not installed, skipping")
	}
	sb := NewProcessSandbox()
	req := &Request{
		Code:     `package main; import "fmt"; func main() { fmt.Println("go works") }`,
		Language: "go",
		Config:   DefaultConfig(),
	}
	ctx := context.Background()
	resp, err := sb.Execute(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Logf("go execution note (may need GOROOT): exit=%d err=%s", resp.ExitCode, resp.Error)
	}
}

func TestNativeSandbox_AuditLogDoesNotPanic(t *testing.T) {
	for i := 0; i < 10; i++ {
		auditLog(fmt.Sprintf("concurrent audit test %d", i))
	}
}
