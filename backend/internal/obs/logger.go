// Package obs provides structured observability for the Challenge-To-YOU
// backend: structured logging, request correlation, metrics, and error
// classification. It wraps the stdlib log/slog with no external dependencies.
package obs

import (
	"context"
	"log/slog"
	"os"
	"sync/atomic"
)

// Level constants matching slog levels.
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Logger is a structured logger that carries component, session, and request
// context. It wraps *slog.Logger and is safe for concurrent use.
type Logger struct {
	*slog.Logger
	component string
}

// global is the default application logger. Initialized by Init.
var global atomic.Pointer[Logger]

func init() {
	global.Store(New("app"))
}

// New creates a Logger writing to stderr as JSON with the given component name.
func New(component string) *Logger {
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return &Logger{
		Logger:    slog.New(h),
		component: component,
	}
}

// Init sets the global default logger. Call once at startup.
func Init(component string) {
	global.Store(New(component))
}

// Default returns the global logger.
func Default() *Logger {
	return global.Load()
}

// With returns a new Logger with the given slog attributes appended. The
// original logger is not modified.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger:    l.Logger.With(args...),
		component: l.component,
	}
}

// Component returns a new Logger scoped to the named sub-component while
// preserving all existing attributes.
func (l *Logger) Component(name string) *Logger {
	return &Logger{
		Logger:    l.Logger.With("component", name),
		component: name,
	}
}

// Session returns a new Logger that includes the given session ID in all
// subsequent log entries.
func (l *Logger) Session(id string) *Logger {
	return &Logger{
		Logger:    l.Logger.With("session_id", id),
		component: l.component,
	}
}

// Request returns a new Logger that includes the given request ID in all
// subsequent log entries.
func (l *Logger) Request(id string) *Logger {
	return &Logger{
		Logger:    l.Logger.With("request_id", id),
		component: l.component,
	}
}

// Player returns a new Logger that includes the given player ID.
func (l *Logger) Player(id string) *Logger {
	return &Logger{
		Logger:    l.Logger.With("player_id", id),
		component: l.component,
	}
}

// Mission returns a new Logger that includes the given mission ID.
func (l *Logger) Mission(id string) *Logger {
	return &Logger{
		Logger:    l.Logger.With("mission_id", id),
		component: l.component,
	}
}

// Tick returns a new Logger that includes the given tick number.
func (l *Logger) Tick(n int64) *Logger {
	return &Logger{
		Logger:    l.Logger.With("tick", n),
		component: l.component,
	}
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Info logs an informational message.
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// WithContext extracts request_id and session_id from ctx and returns a Logger
// with those fields pre-set. Missing keys are ignored.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	out := l
	if rid := RequestIDFrom(ctx); rid != "" {
		out = out.Request(rid)
	}
	if sid := SessionIDFrom(ctx); sid != "" {
		out = out.Session(sid)
	}
	return out
}

// Lifecycle emits a structured lifecycle event. Use this for major state
// transitions (ServerStarted, ClientConnected, SessionCreated, etc.).
func (l *Logger) Lifecycle(event string, args ...any) {
	all := make([]any, 0, len(args)+2)
	all = append(all, "event", event)
	all = append(all, args...)
	l.Logger.Info("lifecycle", all...)
}

// ActiveSessions is an atomic counter for active sessions.
var ActiveSessions atomic.Int64

// ActiveWebSockets is an atomic counter for active WebSocket connections.
var ActiveWebSockets atomic.Int64

// TotalRequests is an atomic counter for total HTTP requests.
var TotalRequests atomic.Int64

// SandboxExecutions is an atomic counter for sandbox executions.
var SandboxExecutions atomic.Int64

// OracleRequests is an atomic counter for AI oracle requests.
var OracleRequests atomic.Int64
