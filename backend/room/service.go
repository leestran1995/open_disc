package room

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type RoomService struct {
	DB          *pgxpool.Pool
	RedisClient *redis.Client
}

func NewRoomService(db *pgxpool.Pool, redisClient *redis.Client) *RoomService {
	return &RoomService{
		DB:          db,
		RedisClient: redisClient,
	}
}

func (s RoomService) Create(ctx context.Context, request CreateRoomRequest) (*Room, error) {
	slog.Info("Creating new room",
		slog.String("room name", request.Name),
	)

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		slog.Warn("Failed to begin transaction for creating room",
			slog.String("roomName", request.Name),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	defer tx.Rollback(ctx)
	var room Room

	// Insert room
	err = tx.QueryRow(ctx,
		`INSERT INTO open_discord.rooms (name)
		 VALUES ($1)
		 RETURNING id, name, sort_order`,
		request.Name,
	).Scan(&room.ID, &room.Name, &room.SortOrder)

	// Fetch and assign default role to room
	slog.Info("Assigning default role to room",
		slog.String("room name", request.Name),
	)
	var defaultRoleId uuid.UUID
	err = tx.QueryRow(ctx,
		`select id from open_discord.roles where name = 'default'`).Scan(&defaultRoleId)
	if err != nil {
		slog.Warn("Failed to find default role for new room",
			slog.String("roomName", request.Name),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	_, err = tx.Exec(ctx, `insert into open_discord.room_roles (room_id, role_id) values ($1, $2)`, room.ID, defaultRoleId)
	if err != nil {
		return nil, err
	}
	tx.Commit(ctx)
	return &room, nil
}

// GetAll returns all rooms along with whether the calling user has starred them. If there is no calling user,
// then userId will be null and we will mark all rooms as false (for the purposes of system calls)
func (s RoomService) GetAll(ctx context.Context, userId *uuid.UUID) ([]Room, error) {
	var rooms []Room
	var sql string
	var rows pgx.Rows
	var err error

	if userId == nil {
		sql = `select id, name, sort_order, false as starred from open_discord.rooms`
		rows, err = s.DB.Query(ctx, sql)
	} else {
		sql = `SELECT DISTINCT r.id, r.name, r.sort_order,
                urs.user_id IS NOT NULL AS starred
				FROM open_discord.rooms r
					LEFT JOIN open_discord.room_roles rr ON rr.room_id = r.id
					LEFT JOIN open_discord.user_roles ur ON ur.role_id = rr.role_id
														AND ur.user_id = $1
					LEFT JOIN open_discord.user_room_stars urs ON urs.room_id = r.id
															AND urs.user_id = $1
				WHERE ur.user_id IS NOT NULL  -- user has access via a role
				OR rr.room_id IS NULL      -- room has no roles attached (public)
				ORDER BY r.sort_order`
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
		var room Room
		err := rows.Scan(&room.ID, &room.Name, &room.SortOrder, &room.Starred)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
		hasNext = rows.Next()
	}

	return rooms, nil
}

func (s RoomService) Reorder(ctx context.Context, req SwapRoomOrderRequest) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, id := range req.RoomIDs {
		_, err := tx.Exec(ctx,
			`update open_discord.rooms set sort_order = $1 where id = $2`, i+1, id)
		if err != nil {
			slog.Warn("Failed to reorder rooms",
				slog.String("error", err.Error()),
				slog.Int("index", i),
				slog.String("roomId", id.String()),
			)
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s RoomService) Star(ctx context.Context, userUuid uuid.UUID, roomUuid uuid.UUID) error {
	_, err := s.DB.Exec(ctx,
		`insert into open_discord.user_room_stars(user_id, room_id) values ($1, $2)`,
		userUuid, roomUuid)
	if err != nil {
		slog.Warn("Failed to star room",
			slog.String("userUuid", userUuid.String()),
			slog.String("roomUuid", roomUuid.String()),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func (s RoomService) Unstar(ctx context.Context, userUuid uuid.UUID, roomUuid uuid.UUID) error {
	_, err := s.DB.Exec(ctx,
		`delete from open_discord.user_room_stars where user_id = $1 and room_id = $2`, userUuid, roomUuid)
	if err != nil {
		slog.Warn("Failed to unstar room",
			slog.String("userUuid", userUuid.String()),
			slog.String("roomUuid", roomUuid.String()),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func roomRoleRedisKey(roomId uuid.UUID) string {
	return "room_roles:" + roomId.String()
}

func (s RoomService) GetRolesForRoom(ctx context.Context, roomId uuid.UUID) ([]string, error) {
	// Check Redis first
	redisKey := roomRoleRedisKey(roomId)
	cachedRoles, err := s.RedisClient.Get(ctx, redisKey).Result()
	if err == nil {
		slog.Info("Cache hit for room roles", slog.String("room_id", roomId.String()))
		return strings.Split(cachedRoles, ","), nil
	} else {
		slog.Info("Cache miss for room roles", slog.String("room_id", roomId.String()))
		slog.Error(err.Error())
	}

	// Fetch from DB
	var roles []string
	rows, err := s.DB.Query(ctx, `select r.name from open_discord.roles r join open_discord.room_roles rr on r.id = rr.role_id where rr.room_id = $1`, roomId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var roleName string
		err = rows.Scan(&roleName)
		if err != nil {
			return nil, err
		}
		roles = append(roles, roleName)
	}
	return roles, nil
}

func (s RoomService) AssignRoomRole(ctx context.Context, roomName, roleName string) error {
	var roomId uuid.UUID
	err := s.DB.QueryRow(ctx, `select id from open_discord.rooms r where r.name = $1`, roomName).Scan(&roomId)

	if err != nil {
		slog.Warn("Failed to find room for assigning room role",
			slog.String("roomName", roomName),
			slog.String("roleName", roleName),
			slog.String("error", err.Error()),
		)
		return err
	}

	var roldId uuid.UUID
	err = s.DB.QueryRow(ctx, `select id from open_discord.roles r where r.name = $1`, roleName).Scan(&roldId)

	if err != nil {
		slog.Warn("Failed to find role for assigning room role",
			slog.String("roomName", roomName),
			slog.String("roleName", roleName),
			slog.String("error", err.Error()),
		)
		return err
	}

	_, err = s.DB.Exec(ctx, `insert into open_discord.room_roles (room_id, role_id) values ($1, $2)`, roomId, roldId)
	if err != nil {
		return err
	}

	// Evict the cache for the room's roles since we've made a change
	redisKey := roomRoleRedisKey(roomId)
	err = s.RedisClient.Del(ctx, redisKey).Err()
	if err != nil {
		slog.Error(
			"Error invalidating room roles cache",
			slog.String("room_id", roomId.String()),
		)
	}
	return nil
}

func (s RoomService) RemoveRoomRole(ctx context.Context, roomName, roleName string) error {
	var roomId uuid.UUID
	err := s.DB.QueryRow(ctx, `select id from open_discord.rooms r where r.name = $1`, roomName).Scan(&roomId)
	if err != nil {
		slog.Warn("Failed to find room for removing room role",
			slog.String("roomName", roomName),
			slog.String("roleName", roleName),
			slog.String("error", err.Error()),
		)
		return err
	}

	var roldId uuid.UUID
	err = s.DB.QueryRow(ctx, `select id from open_discord.roles r where r.name = $1`, roleName).Scan(&roldId)
	if err != nil {
		slog.Warn("Failed to find role for removing room role",
			slog.String("roomName", roomName),
			slog.String("roleName", roleName),
			slog.String("error", err.Error()),
		)
		return err
	}
	_, err = s.DB.Exec(ctx, `delete from open_discord.room_roles where room_id = $1 and role_id = $2`, roomId, roldId)
	if err != nil {
		slog.Warn("Failed to remove room role",
			slog.String("roomName", roomName),
			slog.String("roleName", roleName),
			slog.String("error", err.Error()),
		)
		return err
	}

	// Evict the cache for the room's roles since we've made a change
	redisKey := roomRoleRedisKey(roomId)
	err = s.RedisClient.Del(ctx, redisKey).Err()
	if err != nil {
		slog.Error(
			"Error invalidating room roles cache",
			slog.String("room_id", roomId.String()),
		)
	}
	return nil
}
