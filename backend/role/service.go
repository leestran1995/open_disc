package role

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	DB *pgxpool.Pool
}

func (s Service) CreateRole(name string) (*Role, error) {
	var role Role
	err := s.DB.QueryRow(context.Background(), "insert into open_discord.roles(name) values ($1) returning id, name", name).Scan(&role.ID, &role.Name)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (s Service) DeleteRole(name string) error {
	_, err := s.DB.Exec(context.Background(), "delete from open_discord.roles where name = $1", name)
	return err
}
