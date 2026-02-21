package http

import (
	"net/http"
	opendisc "open_discord"
	"open_discord/internal/postgresql"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	MessageService postgresql.MessageService
}

func BindMessageRoutes(router *gin.Engine, messageHandler *MessageHandler) {
	router.POST("/messages", messageHandler.HandleCreateMessage)
}

func (h *MessageHandler) HandleCreateMessage(c *gin.Context) {
	var request opendisc.MessageCreateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}

	r, err := h.MessageService.Create(c.Request.Context(), request, username.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"message": r})
}
