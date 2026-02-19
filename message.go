package opendisc

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID `json:"id"`
	ServerID  uuid.UUID `json:"server_id"`
	Message   string    `json:"message"`
	TimeStamp time.Time `json:"timestamp"`
	UserID    uuid.UUID `json:"user_id"`
}

type MessageCreateRequest struct {
	ServerID uuid.UUID `json:"server_id"`
	Message  string    `json:"message"`
	UserID   uuid.UUID `json:"user_id"`
}
