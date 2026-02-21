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
		 RETURNING id, name, sort_order`,
		request.Name,
	).Scan(&room.ID, &room.Name, &room.SortOrder)

	if err != nil {
		return nil, err
	}

	return &room, nil
}

func (s RoomService) GetByID(ctx context.Context, serverId uuid.UUID) (*opendisc.Room, error) {
	var room opendisc.Room
	row := s.DB.QueryRow(ctx, "select * from open_discord.rooms where id = $1", serverId)

	err := row.Scan(&room.ID, &room.Name, &room.SortOrder)
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
	rows, err := s.DB.Query(ctx, "select * from open_discord.rooms order by sort_order desc")
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
		err := rows.Scan(&room.ID, &room.Name, &room.SortOrder)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
		hasNext = rows.Next()
	}

	return rooms, nil
}

func (s RoomService) SwapRoomOrder(ctx context.Context, req opendisc.SwapRoomOrderRequest) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	firstRoom, err := s.GetByID(ctx, req.RoomIDOne)
	if err != nil {
		return err
	}

	secondRoom, err := s.GetByID(ctx, req.RoomIDTwo)
	if err != nil {
		return err
	}

	// Swap rooms' sort orders
	_, err = tx.Exec(ctx, `UPDATE open_discord.rooms
								SET sort_order = CASE 
									WHEN id = $1 THEN $3::int
									WHEN id = $2 THEN $4::int
								END
								WHERE id in ($1, $2)`, firstRoom.ID, secondRoom.ID, secondRoom.SortOrder, firstRoom.SortOrder)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
