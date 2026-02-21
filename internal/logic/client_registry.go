package logic

type ClientRegistry struct {
	Clients *map[string]*RoomClient
}

func (c *ClientRegistry) Connect(rc *RoomClient) {
	(*c.Clients)[rc.Username] = rc
}

func (c *ClientRegistry) Disconnect(rc RoomClient) {
	delete(*c.Clients, rc.Username)
}
