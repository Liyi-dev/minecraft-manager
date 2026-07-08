package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"minecraft-manager/model"
	"minecraft-manager/pkg/logger"
	"minecraft-manager/pkg/rcon"
	"minecraft-manager/websocket"
)

type ConsoleHandler struct {
	rcon *rcon.RCONClient
	db   *gorm.DB
	hub  *websocket.Hub
}

func NewConsoleHandler(rconClient *rcon.RCONClient, db *gorm.DB, hub *websocket.Hub) *ConsoleHandler {
	return &ConsoleHandler{rcon: rconClient, db: db, hub: hub}
}

func (h *ConsoleHandler) ExecCommand(c *gin.Context) {
	var req struct {
		Command string `json:"command" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "command is required"})
		return
	}

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	logger.Info.Printf("User %s executing: %s", username, req.Command)

	result, err := h.rcon.ExecuteWithRetry(req.Command, 2)
	if err != nil {
		logger.Error.Printf("RCON command failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log command
	h.db.Create(&model.CommandLog{
		UserID:  userID.(uint),
		Command: req.Command,
		Result:  result,
	})

	// Broadcast result via WebSocket
	if h.hub != nil {
		h.hub.BroadcastCommandResult(username.(string), req.Command, result)
	}

	c.JSON(http.StatusOK, gin.H{"command": req.Command, "result": result})
}
