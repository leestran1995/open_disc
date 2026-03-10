package serverevent

import (
	"backend/logic"
	"backend/model"
	"context"
	"time"
)

type ServerEventStore struct {
	ClientRegistry *logic.ClientRegistry
}

func NewServerEventStore(clientRegistry *logic.ClientRegistry) *ServerEventStore {
	return &ServerEventStore{
		ClientRegistry: clientRegistry,
	}
}

func (s ServerEventStore) Create(
	ctx context.Context,
	eventType model.ServerEventType,
	payload any,
	roles *[]string,
) (*model.ServerEvent, error) {
	var serverEvent model.ServerEvent
	serverEvent = model.ServerEvent{
		ServerEventType: eventType,
		Payload:         payload,
		ServerEventTime: time.Now(),
	}
	s.ClientRegistry.FanOutMessage(serverEvent, roles)
	return &serverEvent, nil
}
