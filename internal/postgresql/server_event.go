package postgresql

import (
	"context"
	"encoding/json"
	"open_discord/internal/logic"

	"open_discord"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerEventStore struct {
	DB             *pgxpool.Pool
	ClientRegistry *logic.ClientRegistry
}

func (s ServerEventStore) Create(ctx context.Context, eventType opendisc.ServerEventType, payload any) (*opendisc.ServerEvent, error) {
	var serverEvent opendisc.ServerEvent
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

	s.ClientRegistry.FanOutMessage(serverEvent)

	return &serverEvent, nil
}

func (s ServerEventStore) GetServerEventsByEventOrder(ctx context.Context, eventOrderStart *int, eventOrderEnd *int) ([]*opendisc.ServerEvent, error) {
	var messages []*opendisc.ServerEvent

	rows, err := s.DB.Query(ctx,
		`select id, event_type, payload, timestamp, event_order from open_discord.server_events
				where ($1::int is null or event_order >= $1::int)
				and ($2::int is null or event_order <= $2::int)
				order by event_order desc 
				limit 10`, eventOrderStart, eventOrderEnd)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var message opendisc.ServerEvent
		rows.Scan(&message.ServerEventID, &message.ServerEventType, &message.Payload, &message.ServerEventTime, &message.ServerEventOrder)
		messages = append(messages, &message)
	}

	return messages, nil
}
