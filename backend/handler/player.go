package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"minecraft-manager/pkg/logger"
	"minecraft-manager/service"
	"minecraft-manager/websocket"
)

type PlayerHandler struct {
	svc *service.PlayerService
	hub *websocket.Hub
}

func NewPlayerHandler(svc *service.PlayerService, hub *websocket.Hub) *PlayerHandler {
	return &PlayerHandler{svc: svc, hub: hub}
}

func (h *PlayerHandler) GetPlayers(c *gin.Context) {
	players, err := h.svc.GetOnlinePlayers()
	if err != nil {
		logger.Warn.Printf("Failed to get players: %v", err)
		c.JSON(http.StatusOK, gin.H{"players": []interface{}{}, "count": 0})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"players": players,
		"count":   len(players),
	})
}

func (h *PlayerHandler) KickPlayer(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player name is required"})
		return
	}

	result, err := h.svc.KickPlayer(req.Name, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast player leave via WebSocket
	if h.hub != nil {
		h.hub.BroadcastPlayerLeave(req.Name)
	}

	c.JSON(http.StatusOK, gin.H{"message": "player kicked", "result": result})
}

func (h *PlayerHandler) BanPlayer(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player name is required"})
		return
	}

	result, err := h.svc.BanPlayer(req.Name, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast player leave via WebSocket
	if h.hub != nil {
		h.hub.BroadcastPlayerLeave(req.Name)
	}

	c.JSON(http.StatusOK, gin.H{"message": "player banned", "result": result})
}

func (h *PlayerHandler) OpPlayer(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player name is required"})
		return
	}

	result, err := h.svc.OpPlayer(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "player opped", "result": result})
}

func (h *PlayerHandler) DeopPlayer(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player name is required"})
		return
	}

	result, err := h.svc.DeopPlayer(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "player deopped", "result": result})
}
