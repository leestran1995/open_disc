package logic

import (
	"encoding/json"
	"backend/domain"

	"github.com/google/uuid"
)

// RoomClient represents a user that is actively connected to open_disc
// UserID is their unique user identifier
// SendChannel is the channel that their SSE connection will receive messages from
type RoomClient struct {
	Username    string
	Nickname    string
	SendChannel chan domain.ServerEvent
}

// Room represents a single room active on the server
// ConnectedClients is a map of UserIDs to RoomClient structs
// RoomID is the ID of the room as it exists in the DB
// Name is the name of the room
type Room struct {
	ClientRegistry *ClientRegistry
	RoomID         uuid.UUID
	Name           string
}

func (r *Room) Send(message domain.Message) error {
	asJson, err := json.Marshal(message)
	if err != nil {
		return err
	}

	roomEvent := domain.ServerEvent{
		ServerEventType: domain.NewMessage,
		Payload:         asJson,
	}

	for _, client := range *r.ClientRegistry.Clients {
		client.SendChannel <- roomEvent
	}

	return nil
}
