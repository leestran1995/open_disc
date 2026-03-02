package opendisc

import (
	"github.com/google/uuid"
)

type User struct {
	UserID   uuid.UUID `json:"user_id"`
	Nickname string    `json:"nickname"`
	Username string    `json:"username"`
	IsOnline bool      `json:"is_online"`
}
