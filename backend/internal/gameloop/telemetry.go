package gameloop

import (
	"sync"
	"time"
)

type TelemetryCollector struct {
	mu        sync.RWMutex
	points    []TelemetryPoint
	maxPoints int
	labels    map[string]string
}

func NewTelemetryCollector(maxPoints int) *TelemetryCollector {
	if maxPoints <= 0 {
		maxPoints = 1000
	}
	return &TelemetryCollector{
		points:    make([]TelemetryPoint, 0, maxPoints),
		maxPoints: maxPoints,
		labels:    make(map[string]string),
	}
}

func (tc *TelemetryCollector) Record(tick int, state GameState, vigilance, entropy float64, snapshot map[string]interface{}, metrics map[string]float64) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if metrics == nil {
		metrics = make(map[string]float64)
	}
	metrics["vigilance"] = vigilance
	metrics["entropy"] = entropy

	point := TelemetryPoint{
		Tick:      tick,
		Vigilance: vigilance,
		Entropy:   entropy,
		State:     state,
		Metrics:   metrics,
		Snapshot:  snapshot,
		Timestamp: time.Now(),
	}

	tc.points = append(tc.points, point)
	if len(tc.points) > tc.maxPoints {
		tc.points = tc.points[len(tc.points)-tc.maxPoints:]
	}
}

func (tc *TelemetryCollector) Points() []TelemetryPoint {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	result := make([]TelemetryPoint, len(tc.points))
	copy(result, tc.points)
	return result
}

func (tc *TelemetryCollector) Recent(n int) []TelemetryPoint {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	if n > len(tc.points) {
		n = len(tc.points)
	}
	result := make([]TelemetryPoint, n)
	copy(result, tc.points[len(tc.points)-n:])
	return result
}

func (tc *TelemetryCollector) Latest() *TelemetryPoint {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	if len(tc.points) == 0 {
		return nil
	}
	p := tc.points[len(tc.points)-1]
	return &p
}

func (tc *TelemetryCollector) SetLabel(key, value string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.labels[key] = value
}

func (tc *TelemetryCollector) Labels() map[string]string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	result := make(map[string]string, len(tc.labels))
	for k, v := range tc.labels {
		result[k] = v
	}
	return result
}

func (tc *TelemetryCollector) Reset() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.points = make([]TelemetryPoint, 0, tc.maxPoints)
	tc.labels = make(map[string]string)
}

func (tc *TelemetryCollector) Count() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return len(tc.points)
}
