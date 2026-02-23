package http

import (
	"log/slog"
	opendisc "open_discord"
	"open_discord/internal/auth"
	"open_discord/internal/logic"
	postgresql2 "open_discord/internal/postgresql"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SseHandler struct {
	RoomService    *postgresql2.RoomService
	MessageService *postgresql2.MessageService
	Rooms          *map[uuid.UUID]*logic.Room
	TokenService   *auth.TokenService
	ClientRegistry *logic.ClientRegistry
}

func (s *SseHandler) HandleGinSseConnection(c *gin.Context) {

	slog.Info("Establishing new client connection")

	username := c.GetString("username")
	sendChannel := make(chan opendisc.RoomEvent, 50)

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
			c.SSEvent(string(message.RoomEventType), message.Payload)
			c.Writer.Flush()
		}
	}
}
