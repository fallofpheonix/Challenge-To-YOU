package obs

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestRequestIDFrom_Missing(t *testing.T) {
	rid := RequestIDFrom(context.Background())
	if rid != "" {
		t.Fatalf("expected empty, got %q", rid)
	}
}

func TestRequestIDFrom_RoundTrip(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req-abc-123")
	rid := RequestIDFrom(ctx)
	if rid != "req-abc-123" {
		t.Fatalf("expected req-abc-123, got %q", rid)
	}
}

func TestSessionIDFrom_RoundTrip(t *testing.T) {
	ctx := WithSessionID(context.Background(), "sess-42")
	sid := SessionIDFrom(ctx)
	if sid != "sess-42" {
		t.Fatalf("expected sess-42, got %q", sid)
	}
}

func TestNewRequestID_Unique(t *testing.T) {
	seen := make(map[string]bool, 100)
	for i := 0; i < 100; i++ {
		id := NewRequestID()
		if seen[id] {
			t.Fatalf("duplicate request ID: %s", id)
		}
		seen[id] = true
		if len(id) != 32 {
			t.Fatalf("expected 32-char hex, got %d chars: %s", len(id), id)
		}
	}
}

func TestLogger_StructuredFields(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	l := &Logger{Logger: slog.New(h), component: "test"}

	l.Info("hello", "key", "value")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if entry["msg"] != "hello" {
		t.Fatalf("expected msg=hello, got %v", entry["msg"])
	}
	if entry["key"] != "value" {
		t.Fatalf("expected key=value, got %v", entry["key"])
	}
	if entry["level"] != "INFO" {
		t.Fatalf("expected level=INFO, got %v", entry["level"])
	}
}

func TestLogger_Component(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	l := &Logger{Logger: slog.New(h), component: "test"}

	l.Component("gameloop").Info("tick")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if entry["component"] != "gameloop" {
		t.Fatalf("expected component=gameloop, got %v", entry["component"])
	}
}

func TestLogger_SessionAndRequest(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	l := &Logger{Logger: slog.New(h), component: "test"}

	l.Session("sess-1").Request("req-2").Info("event")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if entry["session_id"] != "sess-1" {
		t.Fatalf("expected session_id=sess-1, got %v", entry["session_id"])
	}
	if entry["request_id"] != "req-2" {
		t.Fatalf("expected request_id=req-2, got %v", entry["request_id"])
	}
}

func TestLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	l := &Logger{Logger: slog.New(h), component: "test"}

	ctx := WithRequestID(context.Background(), "r-99")
	ctx = WithSessionID(ctx, "s-88")
	l.WithContext(ctx).Info("contextual")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if entry["request_id"] != "r-99" {
		t.Fatalf("expected request_id=r-99, got %v", entry["request_id"])
	}
	if entry["session_id"] != "s-88" {
		t.Fatalf("expected session_id=s-88, got %v", entry["session_id"])
	}
}

func TestLogger_Lifecycle(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	l := &Logger{Logger: slog.New(h), component: "test"}

	l.Lifecycle("ServerStarted", "port", 8080)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if entry["event"] != "ServerStarted" {
		t.Fatalf("expected event=ServerStarted, got %v", entry["event"])
	}
	if entry["port"] != float64(8080) {
		t.Fatalf("expected port=8080, got %v", entry["port"])
	}
}

func TestMetrics_CounterIncrement(t *testing.T) {
	m := NewMetrics()
	m.ActiveSessions.Add(5)
	m.ActiveWebSockets.Add(3)
	m.TotalRequests.Add(100)
	m.SandboxExecs.Add(10)
	m.OracleRequests.Add(7)

	snap := m.Snapshot()
	if snap.ActiveSessions != 5 {
		t.Fatalf("sessions: expected 5, got %d", snap.ActiveSessions)
	}
	if snap.ActiveWebSockets != 3 {
		t.Fatalf("websockets: expected 3, got %d", snap.ActiveWebSockets)
	}
	if snap.TotalRequests != 100 {
		t.Fatalf("requests: expected 100, got %d", snap.TotalRequests)
	}
	if snap.SandboxExecs != 10 {
		t.Fatalf("sandbox: expected 10, got %d", snap.SandboxExecs)
	}
	if snap.OracleRequests != 7 {
		t.Fatalf("oracle: expected 7, got %d", snap.OracleRequests)
	}
}

func TestMetrics_DebugHandler(t *testing.T) {
	m := NewMetrics()
	m.ActiveSessions.Store(42)

	req := httptest.NewRequest("GET", "/debug/runtime", nil)
	w := httptest.NewRecorder()

	DebugHandler(m)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var snap Snapshot
	if err := json.Unmarshal(w.Body.Bytes(), &snap); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if snap.Sessions != 42 {
		t.Fatalf("expected sessions=42, got %d", snap.Sessions)
	}
	if snap.Goroutines <= 0 {
		t.Fatalf("expected positive goroutine count, got %d", snap.Goroutines)
	}
	if snap.Uptime == "" {
		t.Fatal("expected non-empty uptime")
	}
}

func TestMetrics_MetricsHandler(t *testing.T) {
	m := NewMetrics()
	m.TotalRequests.Store(999)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	MetricsHandler(m)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var snap MetricsSnapshot
	if err := json.Unmarshal(w.Body.Bytes(), &snap); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if snap.TotalRequests != 999 {
		t.Fatalf("expected 999, got %d", snap.TotalRequests)
	}
}

func TestTickDurationTracker(t *testing.T) {
	var tt TickDurationTracker
	tt.Record(10 * time.Millisecond)
	tt.Record(20 * time.Millisecond)
	tt.Record(30 * time.Millisecond)

	avg := tt.Average()
	if avg != 20*time.Millisecond {
		t.Fatalf("expected 20ms, got %v", avg)
	}
}

func TestTickDurationTracker_Zero(t *testing.T) {
	var tt TickDurationTracker
	if avg := tt.Average(); avg != 0 {
		t.Fatalf("expected 0, got %v", avg)
	}
}

func TestConcurrentLogging(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	l := &Logger{Logger: slog.New(h), component: "test"}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			l.Info("concurrent", "i", n)
		}(i)
	}
	wg.Wait()

	dec := json.NewDecoder(&buf)
	count := 0
	for dec.More() {
		var entry map[string]any
		if err := dec.Decode(&entry); err != nil {
			t.Fatalf("decode error at entry %d: %v", count, err)
		}
		count++
	}
	if count != 100 {
		t.Fatalf("expected 100 log entries, got %d", count)
	}
}

func TestErrorClass(t *testing.T) {
	err := Classify(ClassSandbox, errTest)
	if err.Error() != "test error" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err.Class != ClassSandbox {
		t.Fatalf("expected class=sandbox, got %v", err.Class)
	}
	if err.Unwrap() != errTest {
		t.Fatal("unwrap returned wrong error")
	}
}

func TestClassify_Nil(t *testing.T) {
	if Classify(ClassAI, nil) != nil {
		t.Fatal("expected nil for nil error")
	}
}

func TestClassifyf(t *testing.T) {
	err := Classifyf(ClassPersistence, "table %s not found", "users")
	if err.Class != ClassPersistence {
		t.Fatalf("expected persistence, got %v", err.Class)
	}
	if err.Error() != "table users not found" {
		t.Fatalf("unexpected error: %v", err)
	}
}

var errTest = &testError{"test error"}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
