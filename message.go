package opendisc

import (
	"encoding/json"
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

type UserEvent struct {
	RoomID uuid.UUID `json:"room_id"`
	UserID uuid.UUID `json:"user_id"`
}

type RoomEventType string

const (
	NewMessage         RoomEventType = "new_message"
	UserJoined         RoomEventType = "user_joined"
	UserLeft           RoomEventType = "user_left"
	HistoricalMessages RoomEventType = "historical_messages"
)

type RoomEvent struct {
	RoomEventType RoomEventType   `json:"room_event_type"`
	Payload       json.RawMessage `json:"payload"`
}
