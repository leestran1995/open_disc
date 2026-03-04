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

func (u UserService) GetUserByID(ctx context.Context, userId uuid.UUID) (*User, error) {
	var user User
	row := u.DB.QueryRow(context.Background(), "select * from open_discord.users where id = $1", userId)

	err := row.Scan(&user.UserID, &user.Nickname)
	if err != nil {
		return nil, err
	}

	userRoles, err := u.GetUserRoles(ctx, user.UserID)
	if err != nil {
		return nil, err
	}

	user.Roles = userRoles

	return &user, nil
}

func (u UserService) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	row := u.DB.QueryRow(context.Background(), "select id, nickname from open_discord.users where username = $1", username)

	err := row.Scan(&user.UserID, &user.Nickname)
	if err != nil {
		return nil, err
	}

	userRoles, err := u.GetUserRoles(ctx, user.UserID)
	if err != nil {
		return nil, err
	}

	user.Roles = userRoles
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
		userRoles, err := u.GetUserRoles(ctx, user.UserID)
		if err != nil {
			return nil, err
		}
		user.Roles = userRoles
		users = append(users, user)
	}
	return users, nil
}

func (u UserService) GetUserRoles(ctx context.Context, userId uuid.UUID) ([]string, error) {
	var roles []string
	rows, err := u.DB.Query(ctx, "select r.name from open_discord.roles r join open_discord.user_roles ur on r.id = ur.role_id where ur.user_id = $1", userId)
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

// AssignUserToRole uses username and rolename as parameters since this will usually be used by a human and we don't want to make them
// look up the user and role ids themselves
func (u UserService) AssignUserToRole(ctx context.Context, username string, rolename string) error {
	var userId uuid.UUID
	var roleId uuid.UUID

	err := u.DB.QueryRow(ctx, "select id from open_discord.users where username = $1", username).Scan(&userId)
	if err != nil {
		return err
	}

	err = u.DB.QueryRow(ctx, "select id from open_discord.roles where name = $1", rolename).Scan(&roleId)
	if err != nil {
		return err
	}

	_, err = u.DB.Exec(ctx, "insert into open_discord.user_roles(user_id, role_id) values ($1, $2)", userId, roleId)
	return err
}

func (u UserService) RemoveUserFromRole(ctx context.Context, username string, rolename string) error {
	var userId uuid.UUID
	var roleId uuid.UUID

	err := u.DB.QueryRow(ctx, "select id from open_discord.users where username = $1", username).Scan(&userId)
	if err != nil {
		return err
	}

	err = u.DB.QueryRow(ctx, "select id from open_discord.roles where name = $1", rolename).Scan(&roleId)
	if err != nil {
		return err
	}

	_, err = u.DB.Exec(ctx, "delete from open_discord.user_roles where user_id = $1 and role_id = $2", userId, roleId)
	return err
}
