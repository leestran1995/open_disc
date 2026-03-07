package serverevent

import (
	"backend/logic"
	"backend/model"
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerEventStore struct {
	DB             *pgxpool.Pool
	ClientRegistry *logic.ClientRegistry
}

func (s ServerEventStore) CreateAndBroadcast(ctx context.Context, eventType model.ServerEventType, payload any) (*model.ServerEvent, error) {
	var serverEvent model.ServerEvent
	var payloadBytes []byte
	var payloadJson json.RawMessage

	err := s.DB.QueryRow(ctx,
		`INSERT INTO open_discord.server_events (event_type, payload)
		 VALUES ($1, $2)
		 RETURNING id, event_type, payload, timestamp, event_order`,
		string(eventType), payload,
	).Scan(&serverEvent.ServerEventID, &serverEvent.ServerEventType, &payloadBytes, &serverEvent.ServerEventTime, &serverEvent.ServerEventOrder)

	err = json.Unmarshal(payloadBytes, &payloadJson)
	serverEvent.Payload = payloadJson
	if err != nil {
		return nil, err
	}

	s.ClientRegistry.FanOutServerEvent(serverEvent)

	return &serverEvent, nil
}

// GetServerEventsByEventOrder retrieves server events by their integer event order, to a limit of 20. If neither bound is provided it will
// get the most recent events first.
func (s ServerEventStore) GetServerEventsByEventOrder(ctx context.Context, eventOrderStart *int, eventOrderEnd *int) ([]*model.ServerEvent, error) {
	var messages []*model.ServerEvent

	rows, err := s.DB.Query(ctx,
		`select id, event_type, payload, timestamp, event_order from open_discord.server_events
				where ($1::int is null or event_order >= $1::int)
				and ($2::int is null or event_order <= $2::int)
				order by event_order desc 
				limit 20`, eventOrderStart, eventOrderEnd)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var message model.ServerEvent
		rows.Scan(&message.ServerEventID, &message.ServerEventType, &message.Payload, &message.ServerEventTime, &message.ServerEventOrder)
		messages = append(messages, &message)
	}

	return messages, nil
}
