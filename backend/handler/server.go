package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"minecraft-manager/service"
)

type ServerHandler struct {
	svc *service.ServerService
}

func NewServerHandler(svc *service.ServerService) *ServerHandler {
	return &ServerHandler{svc: svc}
}

func (h *ServerHandler) GetStatus(c *gin.Context) {
	status := h.svc.GetStatus()
	c.JSON(http.StatusOK, status)
}
