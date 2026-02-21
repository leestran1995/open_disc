package http

import (
	"net/http"
	opendisc "open_discord"
	"open_discord/internal/postgresql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MessageHandler struct {
	MessageService postgresql.MessageService
}

func BindMessageRoutes(router *gin.Engine, messageHandler *MessageHandler) {
	router.POST("/messages", messageHandler.HandleCreateMessage)
	router.GET("/messages/:room_id", messageHandler.HandleGetMessages)
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

func (h *MessageHandler) HandleGetMessages(c *gin.Context) {
	roomId, err := uuid.Parse(c.Param("room_id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	timestampString := c.Query("timestamp")

	if timestampString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing timestamp"})
	}

	timestamp, err := time.Parse(time.RFC3339, timestampString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	result, err := h.MessageService.GetMessagesByTimestamp(c, roomId, timestamp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"messages": result})
}
