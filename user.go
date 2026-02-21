package opendisc

import (
	"github.com/google/uuid"
)

type User struct {
	UserID   uuid.UUID `json:"user_id"`
	Nickname string    `json:"nickname"`
	Username string    `json:"username"`
}

type CreateUserRequest struct {
	Nickname string `json:"nickname"`
}

// RoomJoinRequest In the future the UserID should come as a header fromm the API Gateway
type RoomJoinRequest struct {
	UserID uuid.UUID `json:"user_id"`
}
