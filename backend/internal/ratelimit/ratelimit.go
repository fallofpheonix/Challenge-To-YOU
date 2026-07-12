package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

type Action string

const (
	ActionWebSocket  Action = "websocket"
	ActionChallenge  Action = "challenge"
	ActionAI         Action = "ai"
	ActionMission    Action = "mission"
	ActionCodeSubmit Action = "code_submit"
)

type Config struct {
	MessagesPerSecond float64
	BurstSize         int
	WindowDuration    time.Duration
}

var DefaultConfigs = map[Action]Config{
	ActionWebSocket:  {MessagesPerSecond: 20, BurstSize: 30, WindowDuration: time.Second},
	ActionChallenge:  {MessagesPerSecond: 2, BurstSize: 5, WindowDuration: time.Second},
	ActionAI:         {MessagesPerSecond: 1, BurstSize: 2, WindowDuration: time.Second},
	ActionMission:    {MessagesPerSecond: 5, BurstSize: 10, WindowDuration: time.Second},
	ActionCodeSubmit: {MessagesPerSecond: 1, BurstSize: 3, WindowDuration: time.Second},
}

type bucket struct {
	tokens    float64
	lastCheck time.Time
	config    Config
}

type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	maxAge  time.Duration
}

func New() *Limiter {
	return &Limiter{
		buckets: make(map[string]*bucket),
		maxAge:  10 * time.Minute,
	}
}

func (l *Limiter) Allow(clientID string, action Action) bool {
	return l.AllowN(clientID, action, 1)
}

func (l *Limiter) AllowN(clientID string, action Action, n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	config, ok := DefaultConfigs[action]
	if !ok {
		config = Config{MessagesPerSecond: 10, BurstSize: 20, WindowDuration: time.Second}
	}

	key := fmt.Sprintf("%s:%s", clientID, string(action))
	b, exists := l.buckets[key]
	if !exists {
		b = &bucket{
			tokens:    float64(config.BurstSize),
			lastCheck: time.Now(),
			config:    config,
		}
		l.buckets[key] = b
	}

	now := time.Now()
	elapsed := now.Sub(b.lastCheck).Seconds()
	b.lastCheck = now

	b.tokens += elapsed * config.MessagesPerSecond
	if b.tokens > float64(config.BurstSize) {
		b.tokens = float64(config.BurstSize)
	}

	if n > int(b.tokens) {
		return false
	}

	b.tokens -= float64(n)
	return true
}

func (l *Limiter) Remaining(clientID string, action Action) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	key := fmt.Sprintf("%s:%s", clientID, string(action))
	b, exists := l.buckets[key]
	if !exists {
		config := DefaultConfigs[action]
		return config.BurstSize
	}

	now := time.Now()
	elapsed := now.Sub(b.lastCheck).Seconds()
	tokens := b.tokens + elapsed*b.config.MessagesPerSecond
	if tokens > float64(b.config.BurstSize) {
		tokens = float64(b.config.BurstSize)
	}

	return int(tokens)
}

func (l *Limiter) Reset(clientID string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for key := range l.buckets {
		if len(key) > len(clientID) && key[:len(clientID)] == clientID+":" {
			delete(l.buckets, key)
		}
	}
}

func (l *Limiter) Cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := time.Now().Add(-l.maxAge)
	for key, b := range l.buckets {
		if b.lastCheck.Before(cutoff) {
			delete(l.buckets, key)
		}
	}
}
