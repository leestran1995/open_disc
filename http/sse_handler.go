package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	opendisc "open_discord"
	"open_discord/logic"
	"open_discord/postgresql"
	"time"

	"github.com/google/uuid"
)

type SseHandler struct {
	RoomService    *postgresql.RoomService
	MessageService *postgresql.MessageService
	Rooms          map[uuid.UUID]*logic.Room
}

func (s *SseHandler) CreateNewSseConnection(w http.ResponseWriter, r *http.Request) {

	// Grab the user ID from the path, eventually this will come from an auth token header
	userId, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		log.Fatalf("Unable to parse user id: %v\n", err)
		return
	}

	sendChannel := make(chan opendisc.RoomEvent, 50)

	roomClient := logic.RoomClient{
		UserID:      userId,
		SendChannel: sendChannel,
	}

	userRooms, err := s.RoomService.GetRoomsForUser(context.Background(), userId)

	for _, userRoom := range userRooms {
		roomMessages, err := s.MessageService.GetMessagesByTimestamp(context.Background(), userRoom.ID, time.Now())
		if err != nil {
			log.Fatalf("Unable to get messages by timestamp: %v\n", err)
		}

		toJson, err := json.Marshal(roomMessages)
		if err != nil {
			log.Fatalf("Unable to marshal room messages: %v\n", err)
		}

		roomEvent := opendisc.RoomEvent{
			RoomEventType: opendisc.HistoricalMessages,
			Payload:       toJson,
		}

		sendChannel <- roomEvent
	}

	if err != nil {
		log.Fatalf("Unable to get all rooms: %v\n", err)
		return
	}

	for _, ur := range userRooms {
		matchingRoom := s.Rooms[ur.ID]
		matchingRoom.ConnectToRoom(roomClient)
	}

	// Set CORS headers to allow all origins. You may want to restrict this to specific origins in a production environment.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for {
		select {
		case <-r.Context().Done():
			fmt.Println("Client disconnected")

			for _, ur := range userRooms {
				matchingRoom := s.Rooms[ur.ID]
				matchingRoom.DisconnectFromRoom(roomClient)
			}

			return

		case message := <-sendChannel:
			fmt.Println("Received message in SSE handler")
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
