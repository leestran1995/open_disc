package logic

import (
	"backend/domain"
)

type ClientRegistry struct {
	Clients *map[string]*RoomClient
}

func (c *ClientRegistry) Connect(rc *RoomClient) {
	(*c.Clients)[rc.Username] = rc
	connectEvent := domain.ServerEvent{
		ServerEventType: domain.UserJoined,
		Payload:         rc.Username,
	}

	c.FanOutMessage(connectEvent)
}

func (c *ClientRegistry) Disconnect(rc RoomClient) {
	delete(*c.Clients, rc.Username)

	disconnectEvent := domain.ServerEvent{
		ServerEventType: domain.UserLeft,
		Payload:         rc.Username,
	}

	c.FanOutMessage(disconnectEvent)
}

func (c *ClientRegistry) IsOnline(username string) bool {
	return (*c.Clients)[username] != nil
}

func (c *ClientRegistry) FanOutMessage(message domain.ServerEvent) {
	for _, rc := range *c.Clients {
		rc.SendChannel <- message
	}
}
