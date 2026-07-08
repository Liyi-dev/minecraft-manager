package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"

	"minecraft-manager/pkg/jwt"
	"minecraft-manager/pkg/logger"
	"minecraft-manager/pkg/redis"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
	sendBufSize    = 256
)

// Client represents a single WebSocket connection.
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   uint
	Username string
}

// ClientMessage is the JSON message received from the client.
type ClientMessage struct {
	Type  string `json:"type"`
	Token string `json:"token,omitempty"`
}

// readPump reads messages from the WebSocket connection.
func (c *Client) ReadPump(jwtSecret string) {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	authenticated := false

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logger.Error.Printf("WebSocket read error: %v", err)
			}
			break
		}

		var msg ClientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.Warn.Printf("WebSocket invalid message: %s", string(message))
			continue
		}

		switch msg.Type {
		case "auth":
			if msg.Token == "" {
				c.sendError("token required for auth")
				return
			}

			// Check blacklist
			if redis.IsTokenBlacklisted(msg.Token) {
				c.sendError("token revoked")
				return
			}

			claims, err := jwt.ParseToken(msg.Token, jwtSecret)
			if err != nil {
				c.sendError("invalid token")
				return
			}

			c.UserID = claims.UserID
			c.Username = claims.Username
			authenticated = true

			c.Hub.Register(c)
			c.sendMessage(Message{Type: "auth_ok", Data: map[string]interface{}{
				"username": claims.Username,
				"message":  "authenticated",
			}})
			logger.Info.Printf("WebSocket auth: %s", claims.Username)

		case "subscribe_logs":
			if !authenticated {
				c.sendError("authenticate first")
				continue
			}
			c.Hub.SubscribeLogs(c)
			c.sendMessage(Message{Type: "subscribed_logs", Data: map[string]string{
				"message": "subscribed to log stream",
			}})

		case "ping":
			c.sendMessage(Message{Type: "pong", Data: nil})

		default:
			logger.Warn.Printf("WebSocket unknown message type: %s", msg.Type)
		}
	}
}

// writePump writes messages to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Error.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendMessage(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case c.Send <- data:
	default:
	}
}

func (c *Client) sendError(message string) {
	c.sendMessage(Message{
		Type: "error",
		Data: map[string]string{"message": message},
	})
}
