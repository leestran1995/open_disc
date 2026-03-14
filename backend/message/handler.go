package message

import (
	"backend/model"
	"backend/role"
	"backend/room"
	"backend/serverevent"
	"backend/user"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MessageHandler struct {
	ServerEventStore *serverevent.ServerEventStore
	UserService      *user.UserService
	RoomService      *room.RoomService
	MessageService   *Service
}

func NewMessageHandler(
	serverEventStore *serverevent.ServerEventStore,
	userService *user.UserService,
	roomService *room.RoomService,
	messageService *Service,
) *MessageHandler {
	return &MessageHandler{
		ServerEventStore: serverEventStore,
		UserService:      userService,
		RoomService:      roomService,
		MessageService:   messageService,
	}
}

func BindMessageRoutes(router *gin.Engine, messageHandler *MessageHandler) {
	router.POST("/messages", messageHandler.HandleCreateMessage)
	router.GET("/rooms/:roomId/messages", messageHandler.HandleGetRoomMessages)
}

func (h *MessageHandler) HandleGetRoomMessages(c *gin.Context) {
	roomId := c.Param("roomId")
	timestampStr := c.Query("timestamp")
	var cursorTimestamp *time.Time
	if timestampStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid timestamp format"})
			return
		}
		cursorTimestamp = &parsedTime
	} else {
		cursorTimestamp = nil
	}

	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	// Check if user has permission to view the room
	userRoles, err := h.UserService.GetUserRoles(c.Request.Context(), userId.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	roomRoles, err := h.RoomService.GetRolesForRoom(c, uuid.MustParse(roomId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !role.HasCommonRole(&userRoles, &roomRoles) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	message, err := h.MessageService.GetMessagesForRoom(c, uuid.MustParse(roomId), cursorTimestamp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": message})
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

	// Check if user has permission to post in the room
	userRoles, err := h.UserService.GetUserRoles(c.Request.Context(), userId.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	roomRoles, err := h.RoomService.GetRolesForRoom(c, request.RoomID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !role.HasCommonRole(&userRoles, &roomRoles) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// Done checking if user has permission

	newRequest := model.MessageCreateRequest{
		UserID:  userId.(uuid.UUID),
		RoomID:  request.RoomID,
		Message: request.Message,
	}

	msg, err := h.MessageService.CreateMessage(&newRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = h.ServerEventStore.Create(c, model.NewMessage, msg, &roomRoles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"message": msg})
}
