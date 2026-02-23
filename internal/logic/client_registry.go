package logic

import opendisc "open_discord"

type ClientRegistry struct {
	Clients *map[string]*RoomClient
}

func (c *ClientRegistry) Connect(rc *RoomClient) {
	(*c.Clients)[rc.Username] = rc
	connectEvent := opendisc.ServerEvent{
		ServerEventType: opendisc.UserJoined,
		Payload:         rc.Username,
	}

	c.FanOutMessage(connectEvent)
}

func (c *ClientRegistry) Disconnect(rc RoomClient) {
	delete(*c.Clients, rc.Username)

	disconnectEvent := opendisc.ServerEvent{
		ServerEventType: opendisc.UserLeft,
		Payload:         rc.Username,
	}

	c.FanOutMessage(disconnectEvent)
}

func (c *ClientRegistry) IsOnline(username string) bool {
	return (*c.Clients)[username] != nil
}

func (c *ClientRegistry) FanOutMessage(message opendisc.ServerEvent) {
	for _, rc := range *c.Clients {
		rc.SendChannel <- message
	}
}
