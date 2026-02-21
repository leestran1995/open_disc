package postgresql

import (
	"context"

	"open_discord"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomService struct {
	DB *pgxpool.Pool
}

func (s RoomService) Create(ctx context.Context, request opendisc.CreateRoomRequest) (*opendisc.Room, error) {
	var room opendisc.Room

	err := s.DB.QueryRow(ctx,
		`INSERT INTO open_discord.rooms (name)
		 VALUES ($1)
		 RETURNING id, name`,
		request.Name,
	).Scan(&room.ID, &room.Name)

	if err != nil {
		return nil, err
	}

	return &room, nil
}

func (s RoomService) GetByID(ctx context.Context, serverId uuid.UUID) (*opendisc.Room, error) {
	var room opendisc.Room
	row := s.DB.QueryRow(ctx, "select * from open_discord.rooms where id = $1", serverId)

	err := row.Scan(&room.ID, &room.Name)
	if err != nil {
		return nil, err
	}

	return &room, nil
}

func (s RoomService) JoinRoom(ctx context.Context, request opendisc.RoomJoinRequest, roomId uuid.UUID) error {
	err := s.DB.QueryRow(ctx,
		`insert into open_discord.user_room_pivot (user_id, room_id) values ($1, $2)`, request.UserID, roomId).Scan()

	if err != nil {
		return err
	}

	return nil
}

func (s RoomService) GetAllRooms(ctx context.Context) ([]opendisc.Room, error) {
	var rooms []opendisc.Room
	rows, err := s.DB.Query(ctx, "select * from open_discord.rooms")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hasNext = rows.Next()
	if !hasNext {
		return rooms, nil
	}

	for hasNext {
		var room opendisc.Room
		err := rows.Scan(&room.ID, &room.Name)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
		hasNext = rows.Next()
	}

	return rooms, nil
}
