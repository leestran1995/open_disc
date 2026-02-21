package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	opendisc "open_discord"
	"open_discord/internal/auth"
	"open_discord/internal/logic"
	postgresql2 "open_discord/internal/postgresql"
	"strings"

	"github.com/google/uuid"
)

type SseHandler struct {
	RoomService    *postgresql2.RoomService
	MessageService *postgresql2.MessageService
	Rooms          *map[uuid.UUID]*logic.Room
	TokenService   *auth.TokenService
	ClientRegistry *logic.ClientRegistry
}

func (s *SseHandler) CreateNewSseConnection(w http.ResponseWriter, r *http.Request) {

	// Auth checks
	authHeader := r.Header.Get("Authorization")
	startsWithBearer := strings.HasPrefix(authHeader, "Bearer ")
	if !startsWithBearer {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	bearerToken := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := s.TokenService.ValidateJWT(bearerToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	username := claims.Username

	sendChannel := make(chan opendisc.RoomEvent, 50)

	roomClient := logic.RoomClient{
		Username:    username,
		SendChannel: sendChannel,
	}

	s.ClientRegistry.Connect(&roomClient)

	// Set CORS headers to allow all origins. You may want to restrict this to specific origins in a production environment.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for {
		select {
		case <-r.Context().Done():
			s.ClientRegistry.Disconnect(roomClient)
			return

		case message := <-sendChannel:
			jsonBytes, err := json.Marshal(message)
			if err != nil {
				fmt.Printf("Error marshaling message: %v\n", err)
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", string(jsonBytes))
			w.(http.Flusher).Flush()
		}
	}
}
