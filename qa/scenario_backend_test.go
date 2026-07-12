package qa

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func getProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filename))
}

// ScenarioBackendLaunchShutdown verifies the server starts and stops cleanly.
func ScenarioBackendLaunchShutdown(ctx *ScenarioContext) error {
	root := getProjectRoot()
	tmpDir, err := os.MkdirTemp("", "qa_backend_launch_*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath, err := BuildBinary(root)
	if err != nil {
		return fmt.Errorf("build binary: %w", err)
	}
	ctx.Step("build_success", fmt.Sprintf("Binary built: %s", binaryPath))

	port, err := FindFreePort()
	if err != nil {
		return fmt.Errorf("find port: %w", err)
	}

	server := NewServer(ServerConfig{
		Port:    port,
		DBPath:  filepath.Join(tmpDir, "test.db"),
		WorkDir: root,
	})

	srvCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := server.StartWithPort(srvCtx, binaryPath, port); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	ctx.Step("server_started", fmt.Sprintf("Server started on port %d", port))

	uptime := server.Uptime()
	ctx.Assert("server_running", uptime > 0, fmt.Sprintf("Server uptime: %v", uptime))

	// Stop server
	server.Stop()
	time.Sleep(500 * time.Millisecond)

	// Verify port is free
	d := &netDialer{}
	err = d.dial(port)
	ctx.Assert("server_stopped", err != nil, "Port should be free after shutdown")

	return nil
}

// ScenarioBackendHealthCheck verifies the server health endpoint responds.
func ScenarioBackendHealthCheck(ctx *ScenarioContext) error {
	root := getProjectRoot()
	tmpDir, err := os.MkdirTemp("", "qa_backend_health_*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath, err := BuildBinary(root)
	if err != nil {
		return fmt.Errorf("build binary: %w", err)
	}

	port, err := FindFreePort()
	if err != nil {
		return fmt.Errorf("find port: %w", err)
	}

	server := NewServer(ServerConfig{
		Port:    port,
		DBPath:  filepath.Join(tmpDir, "test.db"),
		WorkDir: root,
	})

	srvCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := server.StartWithPort(srvCtx, binaryPath, port); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer server.Stop()

	// Health check via TCP
	conn, err := netDialTimeout(fmt.Sprintf("localhost:%d", port), 2*time.Second)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	conn.Close()
	ctx.Step("health_check", "TCP connection successful")

	// Test REST endpoint
	client := &httpGetClient{}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/api/languages", port))
	if err != nil {
		return fmt.Errorf("GET /api/languages: %w", err)
	}
	ctx.Assert("languages_endpoint", resp != "", "/api/languages should return data")
	ctx.Step("languages_response", fmt.Sprintf("Response length: %d bytes", len(resp)))

	return nil
}

// ScenarioBackendLogCapture verifies server logs are captured.
func ScenarioBackendLogCapture(ctx *ScenarioContext) error {
	root := getProjectRoot()
	tmpDir, err := os.MkdirTemp("", "qa_backend_logs_*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath, err := BuildBinary(root)
	if err != nil {
		return fmt.Errorf("build binary: %w", err)
	}

	port, err := FindFreePort()
	if err != nil {
		return fmt.Errorf("find port: %w", err)
	}

	server := NewServer(ServerConfig{
		Port:    port,
		DBPath:  filepath.Join(tmpDir, "test.db"),
		WorkDir: root,
	})

	srvCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := server.StartWithPort(srvCtx, binaryPath, port); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer server.Stop()

	// Give server time to log startup messages
	time.Sleep(1 * time.Second)

	logContent, err := server.ReadLog()
	if err != nil {
		return fmt.Errorf("read log: %w", err)
	}

	ctx.Assert("logs_captured", len(logContent) > 0, fmt.Sprintf("Log size: %d bytes", len(logContent)))
	ctx.Step("log_content", fmt.Sprintf("Captured %d bytes of server logs", len(logContent)))

	return nil
}
