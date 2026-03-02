package user

import (
	"backend/logic"
	"context"


	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	DB             *pgxpool.Pool
	ClientRegistry *logic.ClientRegistry
}

type User struct {
	UserID   uuid.UUID `json:"user_id"`
	Nickname string    `json:"nickname"`
	Username string    `json:"username"`
	IsOnline bool      `json:"is_online"`
}

func (u UserService) GetUserByID(ctx context.Context, userId uuid.UUID) (*User, error) {
	var user User
	row := u.DB.QueryRow(context.Background(), "select * from open_discord.users where id = $1", userId)

	err := row.Scan(&user.UserID, &user.Nickname)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u UserService) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	row := u.DB.QueryRow(context.Background(), "select id, nickname from open_discord.users where username = $1", username)

	err := row.Scan(&user.UserID, &user.Nickname)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u UserService) GetAllUsers(ctx context.Context) ([]User, error) {
	var users []User
	rows, err := u.DB.Query(ctx, "select id, nickname, username from open_discord.users")

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var user User
		err = rows.Scan(&user.UserID, &user.Nickname, &user.Username)
		if err != nil {
			return nil, err
		}
		if u.ClientRegistry.IsOnline(user.Username) {
			user.IsOnline = true
		} else {
			user.IsOnline = false
		}
		users = append(users, user)
	}
	return users, nil
}
