package opendisc

import "github.com/google/uuid"

type Room struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	Starred   bool      `json:"starred"`
}

type CreateRoomRequest struct {
	Name string `json:"name"`
}

type SwapRoomOrderRequest struct {
	RoomIDs []uuid.UUID `json:"room_ids"`
}
