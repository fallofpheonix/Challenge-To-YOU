package qa

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Server manages the lifecycle of a backend sandbox server for testing.
type Server struct {
	cmd        *exec.Cmd
	cancel     context.CancelFunc
	port       int
	dbPath     string
	workDir    string
	logFile    *os.File
	logPath    string
	startTime  time.Time
	healthURL  string
}

// ServerConfig holds configuration for launching the server.
type ServerConfig struct {
	Port          int
	DBPath        string
	BinaryPath    string
	ChallengePath string
	MissionDir    string
	WorkDir       string // Working directory for the server process
}

// DefaultServerConfig returns sensible defaults.
func DefaultServerConfig(tmpDir string) ServerConfig {
	return ServerConfig{
		Port:          0, // auto-assign
		DBPath:        filepath.Join(tmpDir, "test_challenge.db"),
		BinaryPath:    "", // will be built
		ChallengePath: "", // use default
		MissionDir:    "", // use default
	}
}

// NewServer creates a server manager but does not start it.
func NewServer(cfg ServerConfig) *Server {
	return &Server{
		port:      cfg.Port,
		dbPath:    cfg.DBPath,
		workDir:   cfg.WorkDir,
		healthURL: fmt.Sprintf("http://localhost:%d/api/languages", cfg.Port),
	}
}

// Start launches the backend server and waits until it is healthy.
func (s *Server) Start(ctx context.Context, binaryPath string) error {
	logPath := filepath.Join("backend_logs", fmt.Sprintf("server_%d.log", time.Now().UnixMilli()))
	os.MkdirAll("backend_logs", 0o755)

	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("create log file: %w", err)
	}
	s.logFile = logFile
	s.logPath = logPath

	args := []string{}
	if s.dbPath != "" {
		args = append(args, "--db="+s.dbPath)
	}

	cmdCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.cmd = exec.CommandContext(cmdCtx, binaryPath, args...)
	if s.workDir != "" {
		s.cmd.Dir = s.workDir
	}
	s.cmd.Stdout = io.MultiWriter(logFile, &logWriter{prefix: "[SERVER] "})
	s.cmd.Stderr = io.MultiWriter(logFile, &logWriter{prefix: "[SERVERERR] "})
	s.cmd.Env = append(os.Environ(),
		"DB_PATH="+s.dbPath,
		fmt.Sprintf("PORT=%d", s.port),
		"CHALLENGE_PATH="+filepath.Join(s.workDir, "backend", "challenges", "magitech_tier1", "magitech_01.json"),
	)

	if err := s.cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("start server: %w", err)
	}

	s.startTime = time.Now()

	// Wait for health check
	if err := s.waitForHealth(ctx, 30*time.Second); err != nil {
		s.Stop()
		return fmt.Errorf("server not healthy: %w", err)
	}

	return nil
}

// StartWithPort launches on a specific port.
func (s *Server) StartWithPort(ctx context.Context, binaryPath string, port int) error {
	s.port = port
	s.healthURL = fmt.Sprintf("http://localhost:%d/api/languages", port)
	logPath := filepath.Join("backend_logs", fmt.Sprintf("server_%d.log", time.Now().UnixMilli()))
	os.MkdirAll("backend_logs", 0o755)

	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("create log file: %w", err)
	}
	s.logFile = logFile
	s.logPath = logPath

	args := []string{}

	cmdCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.cmd = exec.CommandContext(cmdCtx, binaryPath, args...)
	if s.workDir != "" {
		s.cmd.Dir = s.workDir
	}
	s.cmd.Stdout = io.MultiWriter(logFile, &logWriter{prefix: "[SERVER] "})
	s.cmd.Stderr = io.MultiWriter(logFile, &logWriter{prefix: "[SERVERERR] "})
	s.cmd.Env = append(os.Environ(),
		"DB_PATH="+s.dbPath,
		fmt.Sprintf("PORT=%d", s.port),
		"CHALLENGE_PATH="+filepath.Join(s.workDir, "backend", "challenges", "magitech_tier1", "magitech_01.json"),
	)

	if err := s.cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("start server: %w", err)
	}

	s.startTime = time.Now()

	if err := s.waitForHealth(ctx, 30*time.Second); err != nil {
		s.Stop()
		return fmt.Errorf("server not healthy: %w", err)
	}

	return nil
}

// Stop gracefully terminates the server.
func (s *Server) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
		s.cmd.Wait()
	}
	if s.logFile != nil {
		s.logFile.Close()
	}
}

// Port returns the server's listening port.
func (s *Server) Port() int {
	return s.port
}

// LogPath returns the path to the server log file.
func (s *Server) LogPath() string {
	return s.logPath
}

// DBPath returns the path to the server's database.
func (s *Server) DBPath() string {
	return s.dbPath
}

// Uptime returns how long the server has been running.
func (s *Server) Uptime() time.Duration {
	return time.Since(s.startTime)
}

func (s *Server) waitForHealth(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", s.port), 500*time.Millisecond)
		if err == nil {
			conn.Close()
			// Port is open, give it a moment to fully initialize
			time.Sleep(200 * time.Millisecond)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for server on port %d", s.port)
}

// ReadLog returns the full server log contents.
func (s *Server) ReadLog() (string, error) {
	if s.logFile == nil {
		return "", nil
	}
	s.logFile.Sync()

	data, err := os.ReadFile(s.logPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// LogLines returns the server log as individual lines.
func (s *Server) LogLines() ([]string, error) {
	data, err := s.ReadLog()
	if err != nil {
		return nil, err
	}
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

// BuildBinary compiles the sandbox server binary and returns its path.
func BuildBinary(projectRoot string) (string, error) {
	binaryPath := filepath.Join(projectRoot, "backend", "sandbox_server_test")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/sandbox")
	cmd.Dir = filepath.Join(projectRoot, "backend")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build failed: %w\nOutput: %s", err, string(output))
	}
	return binaryPath, nil
}

// FindFreePort returns an available TCP port.
func FindFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

type logWriter struct {
	prefix string
}

func (lw *logWriter) Write(p []byte) (int, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(p)))
	for scanner.Scan() {
		fmt.Printf("%s%s\n", lw.prefix, scanner.Text())
	}
	return len(p), nil
}
