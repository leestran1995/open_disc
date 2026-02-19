package opendisc

import (
	"context"

	"github.com/google/uuid"
)

type User struct {
	UserID   uuid.UUID `json:"user_id"`
	Nickname string    `json:"nickname"`
}

type CreateUserRequest struct {
	Nickname string `json:"nickname"`
}

// RoomJoinRequest In the future the UserID should come as a header fromm the API Gateway
type RoomJoinRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

type UserService interface {
	CreateUser(ctx context.Context, request CreateUserRequest) (*User, error)
	GetUserByID(ctx context.Context, userId uuid.UUID) (*User, error)
	GetUserByNickname(ctx context.Context, nickname string) (*User, error)
	DeleteUser(ctx context.Context, userId uuid.UUID) error
}
