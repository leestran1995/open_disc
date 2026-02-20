package auth

import (
	"context"
	"errors"

	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	DB *pgxpool.Pool
}

// Signup
// Signs a user up, returns an error if signup failed. Sets initial nickname to be same as username
// /*
func (a *Service) Signup(username string, password string) error {
	if a.UsernameExists(username) {
		return errors.New("username exists")
	}

	passwordHash, err := a.ValidateAndHashPassword(password)
	if err != nil {
		return err
	}

	_, err = a.DB.Query(context.Background(),
		`insert into open_discord.users(nickname, username, password) values ($1,$2,$3)`, username, username, passwordHash)

	if err != nil {
		return err
	}

	return nil
}

func (a *Service) UsernameExists(username string) bool {
	var exists bool

	a.DB.QueryRow(context.Background(),
		`select exists(
    select * from open_discord.users u
             where u.username = $1)`, username).Scan(&exists)

	return exists
}

// ValidatePassword TODO: Implement better password verification
func (a *Service) ValidateAndHashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", errors.New("password must be at least 8 characters")
	}
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func (a *Service) CheckPassword(username, password string) (bool, error) {
	var existingPassword string

	row := a.DB.QueryRow(context.Background(),
		`select u.password from open_discord.users u where u.username = $1`, username)

	err := row.Scan(&existingPassword)
	if err != nil {
		return false, err
	}

	result, err := argon2id.ComparePasswordAndHash(password, existingPassword)
	if err != nil {
		return false, err
	}

	return result, nil
}
