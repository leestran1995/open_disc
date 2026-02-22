package opendisc

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID `json:"id"`
	RoomID    uuid.UUID `json:"room_id"`
	Message   string    `json:"message"`
	TimeStamp time.Time `json:"timestamp"`
	Username  string    `json:"username"`
}

type MessageCreateRequest struct {
	RoomID  uuid.UUID `json:"room_id"`
	Message string    `json:"message"`
}

type UserEvent struct {
	RoomID   uuid.UUID `json:"room_id"`
	Username string    `json:"username"`
}

type RoomEventType string

const (
	NewMessage         RoomEventType = "new_message"
	UserJoined         RoomEventType = "user_joined"
	UserLeft           RoomEventType = "user_left"
	HistoricalMessages RoomEventType = "historical_messages"
)

type RoomEvent struct {
	RoomEventType RoomEventType `json:"room_event_type"`
	Payload       any           `json:"payload"`
}
