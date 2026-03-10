package sse

import (
	"backend/auth"
	"backend/logic"
	"backend/model"
	"backend/role"
	"backend/room"
	"backend/user"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SseHandler struct {
	RoomService    *room.RoomService
	Rooms          *map[uuid.UUID]*logic.Room
	TokenService   *auth.TokenService
	ClientRegistry *logic.ClientRegistry
	UserService    *user.UserService
}

func NewSseHandler(
	roomService *room.RoomService,
	Rooms *map[uuid.UUID]*logic.Room,
	tokenService *auth.TokenService,
	clientRegistry *logic.ClientRegistry,
	userService *user.UserService,
) *SseHandler {
	return &SseHandler{
		RoomService:    roomService,
		Rooms:          Rooms,
		TokenService:   tokenService,
		ClientRegistry: clientRegistry,
		UserService:    userService,
	}
}

func (s *SseHandler) EstablishSSEConnection(c *gin.Context) {

	slog.Info("Establishing new client connection")

	username := c.GetString("username")
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	sendChannel := make(chan model.ServerEvent, 50)

	roomClient := logic.RoomClient{
		UserID:      userId.(uuid.UUID),
		SendChannel: sendChannel,
	}

	s.ClientRegistry.Connect(&roomClient)

	// Set CORS headers to allow all origins. You may want to restrict this to specific origins in a production environment.
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	slog.Info("Established connection with " + username + ", waiting on messages to send them.")

	// If the client gets a message with a lastMessage ID they have not actually received, they know they missed something
	// and can re-sync with the backend
	clientMessageId := 0
	for {
		select {
		case <-c.Request.Context().Done():
			slog.Info("Closed client connection to ", username)
			s.ClientRegistry.Disconnect(roomClient)
			return

		case message := <-sendChannel:
			userRoles, err := s.UserService.GetUserRoles(c.Request.Context(), userId.(uuid.UUID))
			if err != nil {
				slog.Error("Error fetching user roles for SSE connection: ", err)
				continue
			}
			if role.HasCommonRole(&userRoles, message.Roles) {
				message.ClientMessageId = clientMessageId + 1
				clientMessageId++
				c.SSEvent(string(message.ServerEventType), message)
				c.Writer.Flush()
			}
		}
	}
}
