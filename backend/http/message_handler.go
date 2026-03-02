package http

import (
	"backend/model"
	"backend/serverevent"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MessageHandler struct {
	ServerEventStore serverevent.ServerEventStore
}

func BindMessageRoutes(router *gin.Engine, messageHandler *MessageHandler) {
	router.POST("/messages", messageHandler.HandleCreateMessage)
}

func (h *MessageHandler) HandleCreateMessage(c *gin.Context) {
	// Get request body
	var request *model.MessageCreateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	newRequest := model.MessageCreateRequest{
		UserID:  userId.(uuid.UUID),
		RoomID:  request.RoomID,
		Message: request.Message,
	}

	serverEvent, err := h.ServerEventStore.Create(c, model.NewMessage, newRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"message": serverEvent})
}
