package opendisc

import "github.com/google/uuid"

type Room struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type CreateRoomRequest struct {
	Name string `json:"name"`
}
