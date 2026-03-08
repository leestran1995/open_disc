package message

import (
	"backend/model"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	DB *pgxpool.Pool
}

func NewMessageService(db *pgxpool.Pool) *Service {
	return &Service{
		DB: db,
	}
}

func (s *Service) GetMessagesForRoom(c *gin.Context, roomId uuid.UUID, cursorTimestamp *time.Time) (*[]model.Message, error) {

	var messages []model.Message
	rows, err := s.DB.Query(
		c,
		`SELECT id, room_id, user_id, message, timestamp FROM open_discord.messages WHERE room_id = $1 AND ($2 is null or timestamp < $2) ORDER BY timestamp DESC limit 25`,
		roomId,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var message model.Message
		err := rows.Scan(&message.ID, &message.RoomID, &message.UserID, &message.Message, &message.TimeStamp)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return &messages, nil
}

func (s *Service) CreateMessage(request *model.MessageCreateRequest) (*model.Message, error) {
	var message model.Message
	err := s.DB.QueryRow(
		context.Background(),
		`INSERT INTO open_discord.messages (room_id, user_id, message) VALUES ($1, $2, $3) RETURNING id, room_id, user_id, message, timestamp`,
		request.RoomID, request.UserID, request.Message,
	).Scan(&message.ID, &message.RoomID, &message.UserID, &message.Message, &message.TimeStamp)
	if err != nil {
		return nil, err
	}

	return &message, nil
}
