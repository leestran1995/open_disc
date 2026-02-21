package logic

import "sync"

// ClientRegistry tracks all active SSE clients so new rooms
// can connect them on creation.
type ClientRegistry struct {
	mu      sync.RWMutex
	clients map[string]*RoomClient // keyed by username
}

func NewClientRegistry() *ClientRegistry {
	return &ClientRegistry{
		clients: make(map[string]*RoomClient),
	}
}

func (cr *ClientRegistry) Register(client *RoomClient) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.clients[client.Username] = client
}

func (cr *ClientRegistry) Unregister(username string) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	delete(cr.clients, username)
}

func (cr *ClientRegistry) GetAll() []*RoomClient {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	result := make([]*RoomClient, 0, len(cr.clients))
	for _, c := range cr.clients {
		result = append(result, c)
	}
	return result
}
