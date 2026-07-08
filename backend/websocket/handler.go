package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"minecraft-manager/pkg/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for MVP; tighten in production
	},
}

// HandleWS handles the WebSocket upgrade request.
func HandleWS(hub *Hub, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error.Printf("WebSocket upgrade failed: %v", err)
			return
		}

		client := &Client{
			Hub:  hub,
			Conn: conn,
			Send: make(chan []byte, sendBufSize),
		}

		go client.WritePump()
		go client.ReadPump(jwtSecret)
	}
}
