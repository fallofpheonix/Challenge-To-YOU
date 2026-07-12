package qa

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSMessage represents a WebSocket message sent or received.
type WSMessage struct {
	Event   string          `json:"event"`
	Payload string          `json:"payload,omitempty"`
	Raw     json.RawMessage `json:"-"`
}

// WSSnapshot represents the server's challenge snapshot response.
type WSSnapshot struct {
	ChallengeID   string                 `json:"challenge_id"`
	Paradigm      string                 `json:"paradigm"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	SkillType     string                 `json:"skill_type"`
	Modules       []WSModule             `json:"modules"`
	State         map[string]interface{} `json:"state"`
	Vigilance     float64                `json:"vigilance"`
	Triggerable   []string               `json:"triggerable"`
	LevelComplete bool                   `json:"level_complete"`
	LastCipher    string                 `json:"last_cipher"`
	Message       string                 `json:"message"`
	Profile       *WSProfile             `json:"profile,omitempty"`
	Error         string                 `json:"error_message,omitempty"`
	Dialogue      []map[string]string    `json:"dialogue,omitempty"`
	Intro         map[string]interface{} `json:"intro,omitempty"`
	World         map[string]interface{} `json:"world,omitempty"`
}

// WSModule represents a challenge module/trigger.
type WSModule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	InputEvent  string `json:"input_event"`
}

// WSProfile represents the player profile from the server.
type WSProfile struct {
	Reputation        int     `json:"reputation"`
	Luck              float64 `json:"luck"`
	UnlockedParadigms string  `json:"unlocked_paradigms"`
	XP                int     `json:"xp,omitempty"`
	Level             int     `json:"level,omitempty"`
	Title             string  `json:"title,omitempty"`
}

// WSClient is a WebSocket test client.
type WSClient struct {
	conn       *websocket.Conn
	mu         sync.Mutex
	messages   []WSMessage
	snapshots  []WSSnapshot
	serverMsgs []json.RawMessage
	port       int
	connected  bool
}

// NewWSClient creates a client that connects to the given port.
func NewWSClient(port int) *WSClient {
	return &WSClient{port: port}
}

// Connect establishes a WebSocket connection to the server.
func (c *WSClient) Connect() error {
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("localhost:%d", c.port),
		Path:   "/rift",
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	c.conn = conn
	c.connected = true
	return nil
}

// RequestInitialState sends start_game and waits for the initial snapshot.
func (c *WSClient) RequestInitialState(timeout time.Duration) (*WSSnapshot, error) {
	if err := c.Send("start_game", ""); err != nil {
		return nil, fmt.Errorf("send start_game: %w", err)
	}
	return c.ReadSnapshot(timeout)
}

// Send sends a JSON event to the server.
func (c *WSClient) Send(event string, payload string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	msg := WSMessage{Event: event, Payload: payload}
	c.messages = append(c.messages, msg)

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// ReadMessage reads the next message from the server with a timeout.
func (c *WSClient) ReadMessage(timeout time.Duration) (*WSMessage, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	c.conn.SetReadDeadline(time.Now().Add(timeout))
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	c.serverMsgs = append(c.serverMsgs, json.RawMessage(data))

	// Try to parse as snapshot first
	var snapshot WSSnapshot
	if err := json.Unmarshal(data, &snapshot); err == nil && snapshot.ChallengeID != "" {
		c.snapshots = append(c.snapshots, snapshot)
		return &WSMessage{Event: "snapshot", Raw: data}, nil
	}

	// Try to parse as event message
	var msg WSMessage
	if err := json.Unmarshal(data, &msg); err == nil && msg.Event != "" {
		return &msg, nil
	}

	// Return raw
	return &WSMessage{Event: "unknown", Raw: data}, nil
}

// ReadSnapshot reads until a snapshot message is received.
func (c *WSClient) ReadSnapshot(timeout time.Duration) (*WSSnapshot, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		msg, err := c.ReadMessage(remaining)
		if err != nil {
			return nil, err
		}
		if msg.Event == "snapshot" {
			var snap WSSnapshot
			if err := json.Unmarshal(msg.Raw, &snap); err == nil {
				return &snap, nil
			}
		}
	}
	return nil, fmt.Errorf("timeout waiting for snapshot")
}

// Disconnect closes the WebSocket connection.
func (c *WSClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.connected = false
		return c.conn.Close()
	}
	return nil
}

// IsConnected returns whether the client is connected.
func (c *WSClient) IsConnected() bool {
	return c.connected
}

// SentMessages returns all messages sent by this client.
func (c *WSClient) SentMessages() []WSMessage {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]WSMessage, len(c.messages))
	copy(out, c.messages)
	return out
}

// ReceivedSnapshots returns all snapshots received.
func (c *WSClient) ReceivedSnapshots() []WSSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]WSSnapshot, len(c.snapshots))
	copy(out, c.snapshots)
	return out
}

// LastSnapshot returns the most recent snapshot, or nil.
func (c *WSClient) LastSnapshot() *WSSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.snapshots) == 0 {
		return nil
	}
	s := c.snapshots[len(c.snapshots)-1]
	return &s
}

// RawMessages returns all raw server messages.
func (c *WSClient) RawMessages() []json.RawMessage {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]json.RawMessage, len(c.serverMsgs))
	copy(out, c.serverMsgs)
	return out
}

// SendRaw sends raw bytes to the server.
func (c *WSClient) SendRaw(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	return c.conn.WriteMessage(websocket.TextMessage, data)
}
