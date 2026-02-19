package postgresql

import (
	"context"
	"open_discord/logic"

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
		`INSERT INTO open_discord.messages (server_id, message, user_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, message, user_id, timestamp, server_id`,
		request.ServerID, request.Message, request.UserID,
	).Scan(&message.ID, &message.Message, &message.UserID, &message.TimeStamp, &message.ServerID)

	if err != nil {
		return nil, err
	}

	err = s.Rooms[message.ServerID].Send(message)

	if err != nil {
		return nil, err
	}

	return &message, nil
}
