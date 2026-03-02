package http

import (
	"backend/auth"
	"backend/domain"
	"backend/logic"
	postgresql2 "backend/postgresql"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SseHandler struct {
	RoomService    *postgresql2.RoomService
	Rooms          *map[uuid.UUID]*logic.Room
	TokenService   *auth.TokenService
	ClientRegistry *logic.ClientRegistry
}

func (s *SseHandler) EstablishSSEConnection(c *gin.Context) {

	slog.Info("Establishing new client connection")

	username := c.GetString("username")
	sendChannel := make(chan domain.ServerEvent, 50)

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
			if message.ServerEventOrder == 0 {
				c.SSEvent(string(message.ServerEventType), message.Payload)
			} else {
				c.SSEvent(string(message.ServerEventType), message)
			}
			c.Writer.Flush()
		}
	}
}
