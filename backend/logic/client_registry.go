package logic

import (
	"backend/model"

	"github.com/google/uuid"
)

type ClientRegistry struct {
	Clients *map[uuid.UUID]*RoomClient
}

func (c *ClientRegistry) Connect(rc *RoomClient) {
	(*c.Clients)[rc.UserID] = rc
	connectEvent := model.ServerEvent{
		ServerEventType: model.UserJoined,
		Payload:         rc.UserID,
	}

	c.FanOutMessage(connectEvent, nil)
}

func (c *ClientRegistry) Disconnect(rc RoomClient) {
	delete(*c.Clients, rc.UserID)

	disconnectEvent := model.ServerEvent{
		ServerEventType: model.UserLeft,
		Payload:         rc.UserID,
	}

	c.FanOutMessage(disconnectEvent, nil)
}

func (c *ClientRegistry) IsOnline(userID uuid.UUID) bool {
	return (*c.Clients)[userID] != nil
}

func (c *ClientRegistry) FanOutMessage(message model.ServerEvent, roles *[]string) {
	for _, rc := range *c.Clients {
		message.Roles = roles
		rc.SendChannel <- message
	}
}
