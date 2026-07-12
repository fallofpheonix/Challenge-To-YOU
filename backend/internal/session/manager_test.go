package session

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"challenge-to-you/backend/internal/engine"
)

func newTestSession() (*engine.ChallengeDefinition, *engine.AxiomaticFabric) {
	def := &engine.ChallengeDefinition{ID: "test_challenge", Paradigm: engine.Magitech}
	fab := engine.NewAxiomaticFabric(engine.Magitech, "done", true)
	return def, fab
}

// withClock replaces the manager clock with a controllable one and returns a
// pointer the test can advance.
func withClock(m *Manager) *time.Time {
	t := time.Unix(0, 0)
	m.now = func() time.Time { return t }
	return &t
}

func TestCreateAndGet(t *testing.T) {
	m := NewManager(time.Minute)
	def, fab := newTestSession()

	s := m.Create("player1", def, fab)
	if s.ID == "" {
		t.Fatal("Create must assign a non-empty ID")
	}
	if m.Count() != 1 {
		t.Errorf("Count = %d, want 1", m.Count())
	}

	got, ok := m.Get(s.ID)
	if !ok || got != s {
		t.Fatalf("Get(%q) did not return the created session", s.ID)
	}
	if got.PlayerID != "player1" || got.Challenge != def || got.Fabric != fab {
		t.Error("session fields not preserved")
	}

	if _, ok := m.Get("nonexistent"); ok {
		t.Error("Get of unknown id must return false")
	}
}

func TestUniqueIDs(t *testing.T) {
	m := NewManager(time.Minute)
	def, fab := newTestSession()
	seen := map[ID]bool{}
	for i := 0; i < 100; i++ {
		s := m.Create("p", def, fab)
		if seen[s.ID] {
			t.Fatalf("duplicate session ID: %s", s.ID)
		}
		seen[s.ID] = true
	}
	if m.Count() != 100 {
		t.Errorf("Count = %d, want 100", m.Count())
	}
}

func TestRemove(t *testing.T) {
	m := NewManager(time.Minute)
	def, fab := newTestSession()
	s := m.Create("p", def, fab)

	if !m.Remove(s.ID) {
		t.Error("Remove of existing session should return true")
	}
	if m.Remove(s.ID) {
		t.Error("Remove of already-removed session should return false")
	}
	if m.Count() != 0 {
		t.Errorf("Count = %d, want 0", m.Count())
	}
	if _, ok := m.Get(s.ID); ok {
		t.Error("session should be gone after Remove")
	}
}

func TestActive(t *testing.T) {
	m := NewManager(time.Minute)
	def, fab := newTestSession()
	a := m.Create("p", def, fab)
	b := m.Create("p", def, fab)

	active := m.Active()
	if len(active) != 2 {
		t.Fatalf("Active len = %d, want 2", len(active))
	}
	found := map[ID]bool{}
	for _, s := range active {
		found[s.ID] = true
	}
	if !found[a.ID] || !found[b.ID] {
		t.Error("Active did not return all live sessions")
	}
}

func TestEvictExpired(t *testing.T) {
	m := NewManager(30 * time.Second)
	clock := withClock(m)
	def, fab := newTestSession()

	old := m.Create("p", def, fab)
	*clock = clock.Add(20 * time.Second)
	fresh := m.Create("p", def, fab) // created 20s after old

	// Advance so old is 40s idle (expired) but fresh is only 20s idle (alive).
	*clock = clock.Add(20 * time.Second)

	if n := m.EvictExpired(); n != 1 {
		t.Fatalf("EvictExpired = %d, want 1", n)
	}
	if _, ok := m.Get(old.ID); ok {
		t.Error("expired session should have been evicted")
	}
	if _, ok := m.Get(fresh.ID); !ok {
		t.Error("fresh session should have survived")
	}
}

func TestTouchDefersEviction(t *testing.T) {
	m := NewManager(30 * time.Second)
	clock := withClock(m)
	def, fab := newTestSession()
	s := m.Create("p", def, fab)

	*clock = clock.Add(25 * time.Second)
	if !m.Touch(s.ID) {
		t.Fatal("Touch of live session should return true")
	}
	*clock = clock.Add(20 * time.Second) // 20s since touch < 30s timeout

	if n := m.EvictExpired(); n != 0 {
		t.Errorf("EvictExpired = %d, want 0 (touched session must survive)", n)
	}
	if _, ok := m.Get(s.ID); !ok {
		t.Error("touched session should still be present")
	}
	if m.Touch("nonexistent") {
		t.Error("Touch of unknown id must return false")
	}
}

func TestZeroTimeoutDisablesEviction(t *testing.T) {
	m := NewManager(0)
	clock := withClock(m)
	def, fab := newTestSession()
	m.Create("p", def, fab)

	*clock = clock.Add(24 * time.Hour)
	if n := m.EvictExpired(); n != 0 {
		t.Errorf("EvictExpired = %d, want 0 when timeout disabled", n)
	}
	if m.Count() != 1 {
		t.Errorf("Count = %d, want 1", m.Count())
	}
}

func TestConcurrentAccess(t *testing.T) {
	m := NewManager(time.Minute)
	def, fab := newTestSession()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			s := m.Create(fmt.Sprintf("p%d", n), def, fab)
			m.Touch(s.ID)
			_, _ = m.Get(s.ID)
			m.Active()
			m.Remove(s.ID)
		}(i)
	}
	wg.Wait()
	if m.Count() != 0 {
		t.Errorf("Count = %d, want 0 after all removed", m.Count())
	}
}

func TestReconnectLookup(t *testing.T) {
	m := NewManager(time.Minute)
	def, fab := newTestSession()
	s := m.Create("player1", def, fab)

	// Simulate original connection dropping — session still exists in manager.
	got, ok := m.Get(s.ID)
	if !ok {
		t.Fatal("session should survive after original connection drops")
	}
	if got.PlayerID != "player1" {
		t.Errorf("expected player1, got %s", got.PlayerID)
	}
	if got.Challenge != def {
		t.Error("expected same challenge pointer on reconnect")
	}
	if got.Fabric != fab {
		t.Error("expected same fabric pointer on reconnect")
	}

	// Touch on reconnect defers eviction.
	m.Touch(s.ID)
	if m.Count() != 1 {
		t.Errorf("Count = %d, want 1 after reconnect", m.Count())
	}
}

func TestConcurrentEvictionAndCreate(t *testing.T) {
	m := NewManager(time.Minute)
	def, fab := newTestSession()

	// Seed 50 sessions.
	for i := 0; i < 50; i++ {
		m.Create(fmt.Sprintf("p%d", i), def, fab)
	}

	// Advance time past timeout so all existing sessions expire.
	clock := withClock(m)
	*clock = clock.Add(2 * time.Minute)

	// Concurrently evict and create new sessions.
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.EvictExpired()
			m.Create(fmt.Sprintf("new_p%d", i), def, fab)
		}(i)
	}
	wg.Wait()

	// No panics, no data corruption. All remaining sessions must be valid.
	for _, s := range m.Active() {
		if s.ID == "" || s.PlayerID == "" {
			t.Error("found session with empty fields after concurrent eviction+create")
		}
	}
}
