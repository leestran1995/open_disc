package logic

import (
	"encoding/json"
	opendisc "open_discord"

	"github.com/google/uuid"
)

// RoomClient represents a user that is actively connected to open_disc
// UserID is their unique user identifier
// SendChannel is the channel that their SSE connection will receive messages from
type RoomClient struct {
	Username    string
	Nickname    string
	SendChannel chan opendisc.RoomEvent
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

func (r *Room) Send(message opendisc.Message) error {
	asJson, err := json.Marshal(message)
	if err != nil {
		return err
	}

	roomEvent := opendisc.RoomEvent{
		RoomEventType: opendisc.NewMessage,
		Payload:       asJson,
	}

	for _, client := range *r.ClientRegistry.Clients {
		client.SendChannel <- roomEvent
	}

	return nil
}
