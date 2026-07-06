package network

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const MaxInboundMessageBytes int64 = 64 * 1024

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		// Accept any origin from localhost/127.0.0.1 (with or without port)
		for _, prefix := range []string{"http://127.0.0.1", "http://localhost", "ws://127.0.0.1", "ws://localhost"} {
			if len(origin) >= len(prefix) && origin[:len(prefix)] == prefix {
				return true
			}
		}
		return false
	},
}

type InboundCommand struct {
	Type    string          `json:"packet_type"`
	Tick    int64           `json:"tick"`
	Payload json.RawMessage `json:"payload"`
}

type NetworkHub struct {
	Clients    map[*websocket.Conn]bool
	Broadcast  chan []byte
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
	mu         sync.Mutex
}

func NewNetworkHub() *NetworkHub {
	return &NetworkHub{
		Clients:    make(map[*websocket.Conn]bool),
		Broadcast:  make(chan []byte, 256), // Buffered to handle rapid tick bursts
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
	}
}

func (h *NetworkHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
			fmt.Println("[NETWORK] Orbital Relay link established.")
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				client.Close()
			}
			h.mu.Unlock()
			fmt.Println("[NETWORK] Orbital Relay link severed.")
		case message := <-h.Broadcast:
			h.mu.Lock()
			for client := range h.Clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					client.Close()
					delete(h.Clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *NetworkHub) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[NETWORK ERROR] Upgrade failed: %v\n", err)
		return
	}
	h.Register <- conn
}

// StartReader Listening Loop runs per connected client thread
func (h *NetworkHub) StartReader(conn *websocket.Conn, commandChannel chan<- InboundCommand) {
	conn.SetReadLimit(MaxInboundMessageBytes)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			h.Unregister <- conn
			break
		}

		var cmd InboundCommand
		if err := json.Unmarshal(message, &cmd); err == nil {
			if cmd.Type == "" {
				fmt.Println("[NETWORK ERROR] Rejected command without packet_type.")
				continue
			}
			commandChannel <- cmd
		} else {
			fmt.Printf("[NETWORK ERROR] Failed to unmarshal client frame: %v\n", err)
		}
	}
}
