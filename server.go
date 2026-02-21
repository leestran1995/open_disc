package opendisc

import "github.com/google/uuid"

type Room struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
}

type CreateRoomRequest struct {
	Name string `json:"name"`
}

type SwapRoomOrderRequest struct {
	RoomIDOne uuid.UUID `json:"room_id_one"`
	RoomIDTwo uuid.UUID `json:"room_id_two"`
}
