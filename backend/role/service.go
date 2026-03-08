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

func (s Service) GetAllRoles() ([]Role, error) {
	rows, err := s.DB.Query(context.Background(), "select id, name from open_discord.roles")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []Role
	for rows.Next() {
		var role Role
		err := rows.Scan(&role.ID, &role.Name)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func IsAdmin(roles []string) bool {
	for _, role := range roles {
		if role == "admin" {
			return true
		}
	}
	return false
}

func HasCommonRole(userRoles, roomRoles *[]string) bool {
	if userRoles == nil || len(*userRoles) == 0 {
		return false
	}
	if roomRoles == nil || len(*roomRoles) == 0 {
		return true
	}
	roleSet := make(map[string]struct{})
	for _, role := range *userRoles {
		roleSet[role] = struct{}{}
	}
	for _, role := range *roomRoles {
		if _, exists := roleSet[role]; exists {
			return true
		}
	}
	return false
}
