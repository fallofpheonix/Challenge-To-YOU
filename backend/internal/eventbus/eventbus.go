package eventbus

import (
	"fmt"
	"sync"
	"time"
)

type EventType string

const (
	EventChallengeLoaded    EventType = "challenge.loaded"
	EventChallengeCompleted EventType = "challenge.completed"
	EventSnapshot           EventType = "snapshot"
	EventArchonTransmission EventType = "archon.transmission"
	EventPurge              EventType = "purge"

	EventCodeSubmitted EventType = "code.submitted"
	EventCodeCompiled  EventType = "code.compiled"
	EventCodeExecuted  EventType = "code.executed"
	EventTestPassed    EventType = "test.passed"
	EventTestFailed    EventType = "test.failed"

	EventXPEarned            EventType = "xp.earned"
	EventLevelUp             EventType = "level.up"
	EventAchievementUnlocked EventType = "achievement.unlocked"

	EventMissionStarted   EventType = "mission.started"
	EventMissionCompleted EventType = "mission.completed"
	EventMissionFailed    EventType = "mission.failed"
	EventMissionExpired   EventType = "mission.expired"

	EventGameWon  EventType = "game.won"
	EventGameLost EventType = "game.lost"
	EventGameOver EventType = "game.over"

	EventTickCompleted EventType = "tick.completed"

	EventResourceGathered  EventType = "resource.gathered"
	EventResourceDelivered EventType = "resource.delivered"
	EventResourceConsumed  EventType = "resource.consumed"

	EventDialogueStarted EventType = "dialogue.started"
	EventDialogueChoice  EventType = "dialogue.choice"

	EventSaveLoaded  EventType = "save.loaded"
	EventSaveCreated EventType = "save.created"
)

type Event struct {
	Type      EventType
	Payload   interface{}
	Source    string
	Timestamp time.Time
}

type EventHandler struct {
	ID   string
	Func func(Event)
}

type EventBus struct {
	mu         sync.RWMutex
	handlers   map[EventType][]EventHandler
	history    []Event
	maxHistory int
	nextID     int
}

func NewEventBus(maxHistory int) *EventBus {
	if maxHistory <= 0 {
		maxHistory = 1000
	}
	return &EventBus{
		handlers:   make(map[EventType][]EventHandler),
		maxHistory: maxHistory,
	}
}

func (eb *EventBus) Subscribe(eventType EventType, handler func(Event)) string {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.nextID++
	id := fmt.Sprintf("h-%s-%d", eventType, eb.nextID)

	eb.handlers[eventType] = append(eb.handlers[eventType], EventHandler{
		ID:   id,
		Func: handler,
	})
	return id
}

func (eb *EventBus) Unsubscribe(eventType EventType, id string) bool {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	handlers := eb.handlers[eventType]
	for i, h := range handlers {
		if h.ID == id {
			eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			return true
		}
	}
	return false
}

func (eb *EventBus) Publish(event Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	eb.mu.Lock()
	eb.history = append(eb.history, event)
	if len(eb.history) > eb.maxHistory {
		eb.history = eb.history[len(eb.history)-eb.maxHistory:]
	}
	handlers := make([]EventHandler, len(eb.handlers[event.Type]))
	copy(handlers, eb.handlers[event.Type])
	eb.mu.Unlock()

	for _, h := range handlers {
		func() {
			defer func() { _ = recover() }()
			h.Func(event)
		}()
	}
}

func (eb *EventBus) GetHistory() []Event {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	result := make([]Event, len(eb.history))
	copy(result, eb.history)
	return result
}

func (eb *EventBus) GetHistoryByType(eventType EventType) []Event {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	var result []Event
	for _, e := range eb.history {
		if e.Type == eventType {
			result = append(result, e)
		}
	}
	return result
}

func (eb *EventBus) ClearHistory() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.history = nil
}

func (eb *EventBus) HandlerCount(eventType EventType) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.handlers[eventType])
}

func (eb *EventBus) TotalHandlerCount() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	total := 0
	for _, handlers := range eb.handlers {
		total += len(handlers)
	}
	return total
}
