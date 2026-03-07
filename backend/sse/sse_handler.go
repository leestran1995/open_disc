package sse

import (
	"backend/auth"
	"backend/logic"
	"backend/model"
	"backend/redacter"
	"backend/room"
	"backend/user"
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SseHandler struct {
	RoomService    *room.RoomService
	Rooms          *map[uuid.UUID]*logic.Room
	TokenService   *auth.TokenService
	ClientRegistry *logic.ClientRegistry
	UsersService   *user.UserService
}

func NewSseHandler(roomService *room.RoomService, rooms *map[uuid.UUID]*logic.Room, tokenService *auth.TokenService, clientRegistry *logic.ClientRegistry, usersService *user.UserService) *SseHandler {
	return &SseHandler{
		RoomService:    roomService,
		Rooms:          rooms,
		TokenService:   tokenService,
		ClientRegistry: clientRegistry,
		UsersService:   usersService,
	}
}

func (s *SseHandler) EstablishSSEConnection(c *gin.Context) {

	slog.Info("Establishing new client connection")

	username := c.GetString("username")
	userId, exists := c.Get("user_id")
	if !exists {
		slog.Error("Failed to get user_id from context for username: " + username)
		c.JSON(500, gin.H{"error": "Failed to get user_id from context"})
		return
	}

	sendChannel := make(chan model.ServerEvent, 50)

	roomClient := logic.RoomClient{
		Username:    username,
		SendChannel: sendChannel,
	}

	s.ClientRegistry.Connect(&roomClient)

	// Set CORS headers to allow all origins. You may want to restrict this to specific origins in a production environment.
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	slog.Info("Established connection with " + username + ", waiting on messages to send them.")

	for {
		select {
		case <-c.Request.Context().Done():
			slog.Info("Closed client connection to ", username)
			s.ClientRegistry.Disconnect(roomClient)
			return

		case message := <-sendChannel:
			userRoles, err := s.UsersService.GetUserRoles(context.Background(), userId.(uuid.UUID))
			if err != nil {
				slog.Error("Failed to get user roles for user: "+username, "error", err)
				continue
			}

			redactResult := redacter.RedactServerEvent(message, userRoles)

			if message.ServerEventOrder == 0 {
				c.SSEvent(string(message.ServerEventType), redactResult.Payload)
			} else {
				c.SSEvent(string(message.ServerEventType), redactResult)
			}
			c.Writer.Flush()
		}
	}
}
