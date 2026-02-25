package opendisc

import (
	"time"

	"github.com/google/uuid"
)

type MessageCreateRequest struct {
	UserID  uuid.UUID `json:"user_id"`
	RoomID  uuid.UUID `json:"room_id"`
	Message string    `json:"message"`
}

type ServerEventType string

const (
	NewMessage  ServerEventType = "new_message"
	UserJoined  ServerEventType = "user_joined"
	UserLeft    ServerEventType = "user_left"
	RoomCreated ServerEventType = "room_created"
	RoomDeleted ServerEventType = "room_deleted"
)

type ServerEvent struct {
	ServerEventType  ServerEventType `json:"server_event_type"`
	ServerEventID    uuid.UUID       `json:"server_event_id"`
	ServerEventOrder int             `json:"server_event_order"`
	ServerEventTime  time.Time       `json:"server_event_time"`
	Payload          any             `json:"payload"`
}

type Message struct {
	UserID    uuid.UUID `json:"user_id"`
	RoomID    uuid.UUID `json:"room_id"`
	Message   string    `json:"message"`
	TimeStamp time.Time `json:"timestamp"`

	// The following fields are deprecated and should be removed once everything is migrated to ServerEvents
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

// UserConnectionEvent Applicable to either UserJoined or UserLeft event types
type UserConnectionEvent struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Nickname string    `json:"nickname"`
}

// RoomExistenceEvent Applicable to either RoomCreated or RoomDeleted event types
type RoomExistenceEvent struct {
	RoomID   uuid.UUID `json:"room_id"`
	RoomName string    `json:"room_name"`
}
