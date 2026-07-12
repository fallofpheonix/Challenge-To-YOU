// Package session owns the lifecycle of live gameplay sessions, per ADR-012:
// a single owner for session creation, lookup, activity tracking, eviction, and
// idle timeout — so transport code (the WebSocket handler) no longer creates or
// deletes sessions directly. It is intentionally transport-independent: it holds
// no connection, only the challenge and its mutable fabric.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"challenge-to-you/backend/internal/engine"
)

// ID uniquely identifies a live session.
type ID string

// Session is the owned unit of gameplay lifecycle for one player: the challenge
// being played and its mutable fabric state.
type Session struct {
	ID        ID
	PlayerID  string
	Challenge *engine.ChallengeDefinition
	Fabric    *engine.AxiomaticFabric
	CreatedAt time.Time
	LastSeen  time.Time
}

// Manager owns all live sessions. It is safe for concurrent use.
type Manager struct {
	mu       sync.RWMutex
	sessions map[ID]*Session
	timeout  time.Duration

	// now and idgen are injectable so tests can drive time and IDs deterministically.
	now   func() time.Time
	idgen func() ID
}

// NewManager returns a Manager that evicts sessions idle longer than timeout.
// A non-positive timeout disables idle eviction.
func NewManager(timeout time.Duration) *Manager {
	return &Manager{
		sessions: make(map[ID]*Session),
		timeout:  timeout,
		now:      time.Now,
		idgen:    randomID,
	}
}

func randomID() ID {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return ID(hex.EncodeToString(b))
}

// Create registers a new session for the given challenge/fabric and returns it.
func (m *Manager) Create(playerID string, challenge *engine.ChallengeDefinition, fabric *engine.AxiomaticFabric) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	s := &Session{
		ID:        m.idgen(),
		PlayerID:  playerID,
		Challenge: challenge,
		Fabric:    fabric,
		CreatedAt: now,
		LastSeen:  now,
	}
	m.sessions[s.ID] = s
	return s
}

// Get returns the session for id, if present (supports reconnect/lookup).
func (m *Manager) Get(id ID) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	return s, ok
}

// Touch marks a session as recently active, deferring idle eviction. Returns
// whether the session existed.
func (m *Manager) Touch(id ID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return false
	}
	s.LastSeen = m.now()
	return true
}

// Remove evicts a session by id (e.g. on disconnect). Returns whether it existed.
func (m *Manager) Remove(id ID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sessions[id]; !ok {
		return false
	}
	delete(m.sessions, id)
	return true
}

// Active returns a snapshot of all live sessions (order unspecified). Intended
// for graceful shutdown enumeration (Slice 5).
func (m *Manager) Active() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		out = append(out, s)
	}
	return out
}

// Count returns the number of live sessions.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// EvictExpired removes sessions idle longer than the configured timeout and
// returns the number evicted. A non-positive timeout is a no-op.
func (m *Manager) EvictExpired() int {
	if m.timeout <= 0 {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	evicted := 0
	for id, s := range m.sessions {
		if now.Sub(s.LastSeen) > m.timeout {
			delete(m.sessions, id)
			evicted++
		}
	}
	return evicted
}
