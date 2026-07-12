package obs

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

// Snapshot holds runtime diagnostic counters exposed by the debug endpoint.
type Snapshot struct {
	Sessions        int64  `json:"sessions"`
	WebSockets      int64  `json:"websockets"`
	Requests        int64  `json:"requests"`
	SandboxExecs    int64  `json:"sandbox_executions"`
	OracleRequests  int64  `json:"oracle_requests"`
	Goroutines      int    `json:"goroutines"`
	Uptime          string `json:"uptime"`
	TickDurationAvg string `json:"avg_tick_duration,omitempty"`
}

var startTime = time.Now()

// DebugHandler returns an http.HandlerFunc that serves runtime diagnostics
// as JSON. It reads from the provided Metrics instance.
func DebugHandler(m *Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := Snapshot{
			Sessions:       m.ActiveSessions.Load(),
			WebSockets:     m.ActiveWebSockets.Load(),
			Requests:       m.TotalRequests.Load(),
			SandboxExecs:   m.SandboxExecs.Load(),
			OracleRequests: m.OracleRequests.Load(),
			Goroutines:     runtime.NumGoroutine(),
			Uptime:         time.Since(startTime).Round(time.Millisecond).String(),
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(snap)
	}
}

// TickDurationTracker tracks the rolling average tick duration.
type TickDurationTracker struct {
	total atomic.Int64
	count atomic.Int64
}

// Record adds a tick duration observation.
func (t *TickDurationTracker) Record(d time.Duration) {
	t.total.Add(int64(d))
	t.count.Add(1)
}

// Average returns the mean tick duration, or zero if no observations.
func (t *TickDurationTracker) Average() time.Duration {
	c := t.count.Load()
	if c == 0 {
		return 0
	}
	return time.Duration(t.total.Load() / c)
}
