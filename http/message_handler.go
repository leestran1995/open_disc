package http

import (
	"net/http"
	opendisc "open_discord"
	"open_discord/postgresql"

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

	r, err := h.MessageService.Create(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"message": r})
}
