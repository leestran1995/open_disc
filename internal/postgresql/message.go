package postgresql

import (
	"context"
	"fmt"
	"open_discord/internal/logic"
	"time"

	"open_discord"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageService struct {
	DB             *pgxpool.Pool
	ClientRegistry *logic.ClientRegistry
}

func (s MessageService) Create(ctx context.Context, request opendisc.MessageCreateRequest, username string) (*opendisc.Message, error) {
	var message opendisc.Message

	err := s.DB.QueryRow(ctx,
		`INSERT INTO open_discord.messages (room_id, message, username)
		 VALUES ($1, $2, $3)
		 RETURNING id, message, username, timestamp, room_id`,
		request.RoomID, request.Message, username,
	).Scan(&message.ID, &message.Message, &message.Username, &message.TimeStamp, &message.RoomID)

	if err != nil {
		return nil, err
	}

	roomEvent := opendisc.RoomEvent{
		RoomEventType: opendisc.NewMessage,
		Payload:       message,
	}
	s.ClientRegistry.FanOutMessage(roomEvent)

	return &message, nil
}

// GetMessagesByTimestamp treat timestamp like a cursor to allow for infinite scrolling
func (s MessageService) GetMessagesByTimestamp(ctx context.Context, roomId uuid.UUID, timestamp time.Time) ([]*opendisc.Message, error) {
	var messages []*opendisc.Message
	fmt.Println(roomId.String(), timestamp)

	rows, err := s.DB.Query(ctx,
		`select id, timestamp, room_id, message, username from open_discord.messages m
			where m.room_id = $1
			and m.timestamp < $2
			order by m.timestamp desc
			limit 10`, roomId, timestamp)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var message opendisc.Message
		rows.Scan(&message.ID, &message.TimeStamp, &message.RoomID, &message.Message, &message.Username)
		messages = append(messages, &message)
	}

	return messages, nil
}
