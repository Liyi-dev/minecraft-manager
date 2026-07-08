package websocket

import (
	"encoding/json"
	"sync"

	"minecraft-manager/pkg/logger"
)

// Message represents a WebSocket message sent to clients.
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Hub manages all WebSocket connections.
type Hub struct {
	clients    map[*Client]bool
	logClients map[*Client]bool // Clients subscribed to log streaming
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		logClients: make(map[*Client]bool),
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client] = true
	logger.Info.Printf("WebSocket client connected. Total: %d", len(h.clients))
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		delete(h.logClients, client)
		close(client.Send)
		logger.Info.Printf("WebSocket client disconnected. Total: %d", len(h.clients))
	}
}

// SubscribeLogs subscribes a client to log streaming.
func (h *Hub) SubscribeLogs(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logClients[client] = true
}

// BroadcastPlayerJoin notifies all clients that a player joined.
func (h *Hub) BroadcastPlayerJoin(playerName string) {
	msg := Message{
		Type: "player_join",
		Data: map[string]string{"name": playerName},
	}
	h.broadcast(msg)
}

// BroadcastPlayerLeave notifies all clients that a player left.
func (h *Hub) BroadcastPlayerLeave(playerName string) {
	msg := Message{
		Type: "player_leave",
		Data: map[string]string{"name": playerName},
	}
	h.broadcast(msg)
}

// BroadcastCommandResult sends command execution result to all clients.
func (h *Hub) BroadcastCommandResult(username, command, result string) {
	msg := Message{
		Type: "command_result",
		Data: map[string]string{
			"username": username,
			"command":  command,
			"result":   result,
		},
	}
	h.broadcast(msg)
}

// BroadcastLog sends a log line to all subscribed clients.
func (h *Hub) BroadcastLog(line string) {
	msg := Message{
		Type: "log",
		Data: map[string]string{"line": line},
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for client := range h.logClients {
		select {
		case client.Send <- data:
		default:
			// Client buffer full, skip
		}
	}
}

func (h *Hub) broadcast(msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error.Printf("WebSocket marshal error: %v", err)
		return
	}

	for client := range h.clients {
		select {
		case client.Send <- data:
		default:
			// Client buffer full, skip
			go func(c *Client) {
				h.Unregister(c)
			}(client)
		}
	}
}

// BroadcastToClient sends a message to a specific client.
func (h *Hub) BroadcastToClient(client *Client, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case client.Send <- data:
	default:
	}
}
