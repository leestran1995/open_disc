package logic

import "backend/model"

type ClientRegistry struct {
	Clients *map[string]*RoomClient
}

func (c *ClientRegistry) Connect(rc *RoomClient) {
	(*c.Clients)[rc.Username] = rc
	connectEvent := model.ServerEvent{
		ServerEventType: model.UserJoined,
		Payload:         rc.Username,
	}

	c.FanOutMessage(connectEvent)
}

func (c *ClientRegistry) Disconnect(rc RoomClient) {
	delete(*c.Clients, rc.Username)

	disconnectEvent := model.ServerEvent{
		ServerEventType: model.UserLeft,
		Payload:         rc.Username,
	}

	c.FanOutMessage(disconnectEvent)
}

func (c *ClientRegistry) IsOnline(username string) bool {
	return (*c.Clients)[username] != nil
}

func (c *ClientRegistry) FanOutMessage(message model.ServerEvent) {
	for _, rc := range *c.Clients {
		rc.SendChannel <- message
	}
}
