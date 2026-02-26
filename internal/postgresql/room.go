package postgresql

import (
	"context"

	"open_discord"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

// GetAllRooms returns all rooms along with whether the calling user has starred them. If there is no calling user,
// then userId will be null and we will mark all rooms as false (for the purposes of system calls)
func (s RoomService) GetAllRooms(ctx context.Context, userId *uuid.UUID) ([]opendisc.Room, error) {
	var rooms []opendisc.Room
	var sql string
	var rows pgx.Rows
	var err error

	if userId == nil {
		sql = `select id, name, sort_order, false as starred from open_discord.rooms`
		rows, err = s.DB.Query(ctx, sql)
	} else {
		sql = `select r.id, r.name, r.sort_order, urs.user_id is not null as starred
				from open_discord.rooms r
						 left join open_discord.user_room_stars urs on r.id = urs.room_id and urs.user_id = $1
				order by r.sort_order`
		rows, err = s.DB.Query(ctx, sql, *userId)
	}

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
		err := rows.Scan(&room.ID, &room.Name, &room.SortOrder, &room.Starred)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
		hasNext = rows.Next()
	}

	return rooms, nil
}

func (s RoomService) ReorderRooms(ctx context.Context, req opendisc.SwapRoomOrderRequest) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, id := range req.RoomIDs {
		_, err := tx.Exec(ctx,
			`update open_discord.rooms set sort_order = $1 where id = $2`, i+1, id)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s RoomService) StarRoom(ctx context.Context, userUuid uuid.UUID, roomUuid uuid.UUID) error {
	_, err := s.DB.Exec(ctx,
		`insert into open_discord.user_room_stars(user_id, room_id) values ($1, $2)`,
		userUuid, roomUuid)
	if err != nil {
		return err
	}
	return nil
}

func (s RoomService) UnstarRoom(ctx context.Context, userUuid uuid.UUID, roomUuid uuid.UUID) error {
	_, err := s.DB.Exec(ctx,
		`delete from open_discord.user_room_stars where user_id = $1 and room_id = $2`, userUuid, roomUuid)
	if err != nil {
		return err
	}
	return nil
}
