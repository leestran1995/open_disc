package logic

import (
	"encoding/json"
	"fmt"
	opendisc "open_discord"

	"github.com/google/uuid"
)

// RoomClient represents a user that is actively connected to open_disc
// UserID is their unique user identifier
// SendChannel is the channel that their SSE connection will receive messages from
type RoomClient struct {
	UserID      uuid.UUID
	SendChannel chan opendisc.RoomEvent
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

	connectedEvent := opendisc.UserEvent{
		RoomID: r.RoomID,
		UserID: roomClient.UserID,
	}

	asJson, err := json.Marshal(connectedEvent)
	if err != nil {
		panic(err)
	}

	roomEvent := opendisc.RoomEvent{
		RoomEventType: opendisc.UserJoined,
		Payload:       asJson,
	}
	for _, client := range r.ConnectedClients {
		client.SendChannel <- roomEvent // Why is this blocking
	}
}

func (r *Room) DisconnectFromRoom(roomClient RoomClient) {
	fmt.Printf("Disconnecting user %s from room %s\n", r.RoomID, roomClient.UserID)
	delete(r.ConnectedClients, roomClient.UserID)

	r.ConnectedClients[roomClient.UserID] = &roomClient

	connectedEvent := opendisc.UserEvent{
		RoomID: r.RoomID,
		UserID: roomClient.UserID,
	}

	asJson, err := json.Marshal(connectedEvent)
	if err != nil {
		panic(err)
	}

	roomEvent := opendisc.RoomEvent{
		RoomEventType: opendisc.UserLeft,
		Payload:       asJson,
	}
	for _, client := range r.ConnectedClients {
		client.SendChannel <- roomEvent // Why is this blocking
	}
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

	for _, client := range r.ConnectedClients {
		client.SendChannel <- roomEvent
	}
	return nil
}
