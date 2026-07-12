package obs

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// Metrics holds all lightweight runtime counters. All fields are safe for
// concurrent use via atomic operations.
type Metrics struct {
	ActiveSessions   atomic.Int64
	ActiveWebSockets atomic.Int64
	TotalRequests    atomic.Int64
	SandboxExecs     atomic.Int64
	OracleRequests   atomic.Int64
	Ticks            atomic.Int64
	TickDuration     TickDurationTracker
	ShutdownDuration atomic.Int64 // nanoseconds
}

// NewMetrics returns a zero-valued Metrics ready for use.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// Snapshot returns a point-in-time copy of all counters.
func (m *Metrics) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		ActiveSessions:   m.ActiveSessions.Load(),
		ActiveWebSockets: m.ActiveWebSockets.Load(),
		TotalRequests:    m.TotalRequests.Load(),
		SandboxExecs:     m.SandboxExecs.Load(),
		OracleRequests:   m.OracleRequests.Load(),
		Ticks:            m.Ticks.Load(),
		AvgTickDuration:  m.TickDuration.Average().String(),
		ShutdownDuration: time.Duration(m.ShutdownDuration.Load()).String(),
	}
}

// MetricsSnapshot is a JSON-serializable point-in-time view of Metrics.
type MetricsSnapshot struct {
	ActiveSessions   int64  `json:"active_sessions"`
	ActiveWebSockets int64  `json:"active_websockets"`
	TotalRequests    int64  `json:"total_requests"`
	SandboxExecs     int64  `json:"sandbox_executions"`
	OracleRequests   int64  `json:"oracle_requests"`
	Ticks            int64  `json:"ticks"`
	AvgTickDuration  string `json:"avg_tick_duration"`
	ShutdownDuration string `json:"shutdown_duration"`
}

// MetricsHandler returns an http.HandlerFunc that serves metrics as JSON.
func MetricsHandler(m *Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := m.Snapshot()
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(snap)
	}
}
