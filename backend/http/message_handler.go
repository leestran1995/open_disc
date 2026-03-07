package http

import (
	"backend/model"
	"backend/role"
	"backend/room"
	"backend/serverevent"
	"backend/user"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MessageHandler struct {
	ServerEventStore *serverevent.ServerEventStore
	UserService      *user.UserService
	RoomService      *room.RoomService
}

func NewMessageHandler(
	serverEventStore *serverevent.ServerEventStore,
	userService *user.UserService,
	roomService *room.RoomService,
) *MessageHandler {
	return &MessageHandler{
		ServerEventStore: serverEventStore,
		UserService:      userService,
		RoomService:      roomService,
	}
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

	if !role.HasCommonRole(userRoles, roomRoles) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// Done checking if user has permission

	newRequest := model.MessageCreateRequest{
		UserID:  userId.(uuid.UUID),
		RoomID:  request.RoomID,
		Message: request.Message,
	}

	serverEvent, err := h.ServerEventStore.CreateAndBroadcast(c, model.NewMessage, newRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"message": serverEvent})
}
