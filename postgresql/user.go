package postgresql

import (
	"context"

	"open_discord"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	DB *pgxpool.Pool
}

func (u UserService) CreateUser(ctx context.Context, request opendisc.CreateUserRequest) (*opendisc.User, error) {
	var user opendisc.User

	err := u.DB.QueryRow(ctx,
		`INSERT INTO open_discord.users (nickname)
		 VALUES ($1)
		 RETURNING id, nickname`,
		request.Nickname,
	).Scan(&user.UserID, &user.Nickname)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u UserService) GetUserByID(ctx context.Context, userId uuid.UUID) (*opendisc.User, error) {
	var user opendisc.User
	row := u.DB.QueryRow(context.Background(), "select * from open_discord.users where id = $1", userId)

	err := row.Scan(&user.UserID, &user.Nickname)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u UserService) GetUserByNickname(ctx context.Context, nickname string) (*opendisc.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserService) DeleteUser(ctx context.Context, userId uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}
