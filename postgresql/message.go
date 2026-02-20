package postgresql

import (
	"context"
	"open_discord/logic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"open_discord"
)

type MessageService struct {
	DB    *pgxpool.Pool
	Rooms map[uuid.UUID]*logic.Room
}

func (s MessageService) Create(ctx context.Context, request opendisc.MessageCreateRequest) (*opendisc.Message, error) {
	var message opendisc.Message

	err := s.DB.QueryRow(ctx,
		`INSERT INTO open_discord.messages (room_id, message, user_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, message, user_id, timestamp, room_id`,
		request.RoomID, request.Message, request.UserID,
	).Scan(&message.ID, &message.Message, &message.UserID, &message.TimeStamp, &message.RoomID)

	if err != nil {
		return nil, err
	}

	err = s.Rooms[message.RoomID].Send(message)

	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (s MessageService) GetMessagesByTimestamp(ctx context.Context, roomId uuid.UUID, timestamp time.Time) ([]*opendisc.Message, error) {
	var messages []*opendisc.Message

	rows, err := s.DB.Query(ctx,
		`select id, timestamp, room_id, message, user_id from open_discord.messages m
			where m.room_id = $1
			and m.timestamp < $2
			limit 10`, roomId, timestamp)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var message opendisc.Message
		rows.Scan(&message.ID, &message.TimeStamp, &message.RoomID, &message.Message, &message.UserID)
		messages = append(messages, &message)
	}

	return messages, nil
}
