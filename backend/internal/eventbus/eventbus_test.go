package eventbus

import (
	"fmt"
	"sync"
	"testing"
)

func TestSubscribeAndPublish(t *testing.T) {
	bus := NewEventBus(100)
	var mu sync.Mutex
	var received []Event

	id := bus.Subscribe(EventChallengeCompleted, func(e Event) {
		mu.Lock()
		received = append(received, e)
		mu.Unlock()
	})

	if id == "" {
		t.Fatal("expected non-empty handler ID")
	}

	bus.Publish(Event{
		Type:    EventChallengeCompleted,
		Payload: "ch_01",
		Source:  "test",
	})

	mu.Lock()
	if len(received) != 1 {
		t.Fatalf("expected 1 event, got %d", len(received))
	}
	if received[0].Payload != "ch_01" {
		t.Errorf("expected payload 'ch_01', got %v", received[0].Payload)
	}
	mu.Unlock()
}

func TestSubscribeOnlyReceivesOwnType(t *testing.T) {
	bus := NewEventBus(100)
	var received int

	bus.Subscribe(EventChallengeCompleted, func(e Event) {
		received++
	})

	bus.Publish(Event{Type: EventChallengeLoaded, Payload: "loaded"})
	if received != 0 {
		t.Errorf("expected 0 events for different type, got %d", received)
	}

	bus.Publish(Event{Type: EventChallengeCompleted, Payload: "completed"})
	if received != 1 {
		t.Errorf("expected 1 event, got %d", received)
	}

	bus.Publish(Event{Type: EventChallengeCompleted, Payload: "completed2"})
	if received != 2 {
		t.Errorf("expected 2 events, got %d", received)
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := NewEventBus(100)
	var received int

	id := bus.Subscribe(EventChallengeCompleted, func(e Event) {
		received++
	})

	bus.Publish(Event{Type: EventChallengeCompleted, Payload: "first"})
	if received != 1 {
		t.Fatalf("expected 1 event, got %d", received)
	}

	ok := bus.Unsubscribe(EventChallengeCompleted, id)
	if !ok {
		t.Error("expected Unsubscribe to return true")
	}

	bus.Publish(Event{Type: EventChallengeCompleted, Payload: "second"})
	if received != 1 {
		t.Errorf("expected still 1 event after unsubscribe, got %d", received)
	}
}

func TestUnsubscribeNonexistent(t *testing.T) {
	bus := NewEventBus(100)
	ok := bus.Unsubscribe(EventChallengeCompleted, "nonexistent")
	if ok {
		t.Error("expected Unsubscribe to return false for nonexistent handler")
	}
}

func TestUnsubscribeWrongType(t *testing.T) {
	bus := NewEventBus(100)
	id := bus.Subscribe(EventChallengeCompleted, func(e Event) {})

	ok := bus.Unsubscribe(EventPurge, id)
	if ok {
		t.Error("expected Unsubscribe for wrong type to return false")
	}
}

func TestMultipleHandlers(t *testing.T) {
	bus := NewEventBus(100)
	var count int
	var mu sync.Mutex

	bus.Subscribe(EventChallengeCompleted, func(e Event) {
		mu.Lock()
		count++
		mu.Unlock()
	})
	bus.Subscribe(EventChallengeCompleted, func(e Event) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	bus.Publish(Event{Type: EventChallengeCompleted, Payload: "multi"})

	mu.Lock()
	if count != 2 {
		t.Errorf("expected 2 handler invocations, got %d", count)
	}
	mu.Unlock()
}

func TestMultipleHandlersOneUnsubscribes(t *testing.T) {
	bus := NewEventBus(100)
	var count int
	var mu sync.Mutex

	id := bus.Subscribe(EventChallengeCompleted, func(e Event) {
		mu.Lock()
		count++
		mu.Unlock()
	})
	bus.Subscribe(EventChallengeCompleted, func(e Event) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	bus.Unsubscribe(EventChallengeCompleted, id)

	bus.Publish(Event{Type: EventChallengeCompleted, Payload: "after_unsub"})

	mu.Lock()
	if count != 1 {
		t.Errorf("expected 1 handler invocation after removing one, got %d", count)
	}
	mu.Unlock()
}

func TestHistoryLimit(t *testing.T) {
	bus := NewEventBus(5)
	for i := 0; i < 10; i++ {
		bus.Publish(Event{Type: EventChallengeCompleted, Payload: i})
	}
	history := bus.GetHistory()
	if len(history) != 5 {
		t.Errorf("expected 5 events in history, got %d", len(history))
	}
}

func TestHistoryByType(t *testing.T) {
	bus := NewEventBus(100)
	bus.Publish(Event{Type: EventChallengeCompleted, Payload: "ch"})
	bus.Publish(Event{Type: EventPurge, Payload: "purge"})
	bus.Publish(Event{Type: EventChallengeCompleted, Payload: "ch2"})

	completed := bus.GetHistoryByType(EventChallengeCompleted)
	if len(completed) != 2 {
		t.Errorf("expected 2 challenge.completed events, got %d", len(completed))
	}

	purges := bus.GetHistoryByType(EventPurge)
	if len(purges) != 1 {
		t.Errorf("expected 1 purge event, got %d", len(purges))
	}
}

func TestClearHistory(t *testing.T) {
	bus := NewEventBus(100)
	bus.Publish(Event{Type: EventChallengeCompleted, Payload: "ch"})
	bus.ClearHistory()
	if len(bus.GetHistory()) != 0 {
		t.Error("expected empty history after clear")
	}
}

func TestTimestampSet(t *testing.T) {
	bus := NewEventBus(100)
	bus.Publish(Event{Type: EventChallengeCompleted, Payload: "ch"})
	history := bus.GetHistory()
	if len(history) > 0 && history[0].Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
}

func TestConcurrentPublishAndSubscribe(t *testing.T) {
	bus := NewEventBus(1000)
	var wg sync.WaitGroup
	var received sync.Map

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			bus.Subscribe(EventChallengeCompleted, func(e Event) {
				received.Store(e.Payload.(string), true)
			})
		}(i)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			bus.Publish(Event{
				Type:    EventChallengeCompleted,
				Payload: fmt.Sprintf("ev_%d", idx),
				Source:  "concurrent",
			})
		}(i)
	}

	wg.Wait()
}

func TestHandlerCount(t *testing.T) {
	bus := NewEventBus(100)
	if bus.HandlerCount(EventChallengeCompleted) != 0 {
		t.Error("expected 0 handlers initially")
	}

	bus.Subscribe(EventChallengeCompleted, func(e Event) {})
	if bus.HandlerCount(EventChallengeCompleted) != 1 {
		t.Errorf("expected 1 handler, got %d", bus.HandlerCount(EventChallengeCompleted))
	}

	bus.Subscribe(EventChallengeCompleted, func(e Event) {})
	if bus.HandlerCount(EventChallengeCompleted) != 2 {
		t.Errorf("expected 2 handlers, got %d", bus.HandlerCount(EventChallengeCompleted))
	}
}

func TestPanicInHandlerDoesNotAffectOthers(t *testing.T) {
	bus := NewEventBus(100)
	var count int
	var mu sync.Mutex

	bus.Subscribe(EventChallengeCompleted, func(e Event) {
		panic("handler panic")
	})
	bus.Subscribe(EventChallengeCompleted, func(e Event) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	bus.Publish(Event{Type: EventChallengeCompleted, Payload: "panic_test"})

	mu.Lock()
	if count != 1 {
		t.Errorf("expected 1 successful handler call despite panic, got %d", count)
	}
	mu.Unlock()
}
