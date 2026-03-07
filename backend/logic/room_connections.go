package logic

import (
	"backend/model"
	"encoding/json"

	"github.com/google/uuid"
)

// RoomClient represents a user that is actively connected to open_disc
// UserID is their unique user identifier
// SendChannel is the channel that their SSE connection will receive messages from
type RoomClient struct {
	Username    string
	Nickname    string
	SendChannel chan model.ServerEvent
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

func (r *Room) Send(message model.Message) error {
	asJson, err := json.Marshal(message)
	if err != nil {
		return err
	}

	roomEvent := model.ServerEvent{
		ServerEventType: model.NewMessage,
		Payload:         asJson,
	}

	for _, client := range *r.ClientRegistry.Clients {
		client.SendChannel <- roomEvent
	}

	return nil
}
