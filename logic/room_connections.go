package logic

import (
	"fmt"
	opendisc "open_discord"

	"github.com/google/uuid"
)

// RoomClient represents a user that is actively connected to open_disc
// UserID is their unique user identifier
// SendChannel is the channel that their SSE connection will receive messages from
type RoomClient struct {
	UserID      uuid.UUID
	SendChannel chan opendisc.Message
}

// Room represents a single room active on the server
// ConnectedClients is a map of UserIDs to RoomClient structs
// RoomID is the ID of the room as it exists in the DB
// Name is the name of the room
type Room struct {
	ConnectedClients map[uuid.UUID]*RoomClient
	RoomID           uuid.UUID
	Name             string
}

func (r *Room) ConnectToRoom(roomClient RoomClient) {
	fmt.Printf("Connecting user %s to room %s\n", r.RoomID, roomClient.UserID)
	r.ConnectedClients[roomClient.UserID] = &roomClient
}

func (r *Room) DisconnectFromRoom(roomClient RoomClient) {
	fmt.Printf("Disconnecting user %s from room %s\n", r.RoomID, roomClient.UserID)
	delete(r.ConnectedClients, roomClient.UserID)
}

func (r *Room) Send(message opendisc.Message) error {
	for _, client := range r.ConnectedClients {
		client.SendChannel <- message
	}
	return nil
}
