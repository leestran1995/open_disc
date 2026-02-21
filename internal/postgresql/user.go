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

func (u UserService) GetAllUsers(ctx context.Context) ([]opendisc.User, error) {
	var users []opendisc.User
	rows, err := u.DB.Query(ctx, "select id, nickname, username from open_discord.users")

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var user opendisc.User
		err = rows.Scan(&user.UserID, &user.Nickname, &user.Username)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
