package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	DB *pgxpool.Pool
}

var (
	reUppercase = regexp.MustCompile(`[A-Z]`)
	reLowercase = regexp.MustCompile(`[a-z]`)
	reNumber    = regexp.MustCompile(`[0-9]`)
	reSpecial   = regexp.MustCompile(`[!@#$%^&*()\-+]`)
)

// Signup performs validation checks and signs the user up if all the checks pass.
func (a *Service) Signup(username string, password string, otc uuid.UUID) error {

	slog.Info("Signing up user",
		slog.String("username", username),
	)
	// Username validation
	if a.UsernameExists(username) {
		return errors.New("username exists")
	}

	err := a.ValidateUsername(username)

	if err != nil {
		return err
	}

	// Password validation
	result := CheckPasswordStrength(password)

	if !result.IsValid() {
		return errors.New("password does not meet strength requirements")
	}

	// OTC validation
	passwordHash, err := a.HashPassword(password)
	if err != nil {
		return err
	}

	var otcExists bool
	a.DB.QueryRow(context.Background(),
		`select exists(
		select * from open_discord.signup_otcs o
			 where o.code = $1
			 and o.used = false
			 and o.time_expires > now())
			 `, otc).Scan(&otcExists)

	if !otcExists {
		return errors.New("invalid otc")
	}

	slog.Info("User passed validation checks, inserting them into database",
		slog.String("username", username),
	)

	tx, err := a.DB.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	var userId uuid.UUID
	// Insert the new user
	err = tx.QueryRow(context.Background(),
		`insert into open_discord.users(nickname, username, password) values ($1,$2,$3) returning id`, username, username, passwordHash).Scan(&userId)

	if err != nil {
		return err
	}

	// Mark the OTC as used
	_, err = tx.Exec(context.Background(), `update open_discord.signup_otcs o set used = true where o.code = $1`, otc)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	slog.Info("Assigning user to default role",
		slog.String("username", username),
	)
	// Assign the user the default role
	var defaultRoleId uuid.UUID
	err = tx.QueryRow(context.Background(),
		`select id from open_discord.roles where name = 'default'`).Scan(&defaultRoleId)
	if err != nil {
		fmt.Print(err.Error())
		return err
	}
	_, err = tx.Exec(context.Background(),
		`insert into open_discord.user_roles(user_id, role_id) values ($1, $2)`, userId, defaultRoleId)
	if err != nil {
		fmt.Print(err.Error())
		return err
	}
	tx.Commit(context.Background())
	return nil
}

// Username functions

// ValidateUsername TODO: Implement actual username validation
func (a *Service) ValidateUsername(username string) error {
	return nil
}

// UsernameExists helper function to check if a username exists already.
// In the future, we could try and suggest a new username if it's already taken.
// Maybe using Redis and a bloom filter.
func (a *Service) UsernameExists(username string) bool {
	var exists bool

	a.DB.QueryRow(context.Background(),
		`select exists(
    select * from open_discord.users u
             where u.username = $1)`, username).Scan(&exists)

	return exists
}

// Password functions

// HashPassword helper function to hash a password using argon
func (a *Service) HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

// A CheckPasswordResult represents the result of checking a password's strength against certain criteria.
// we include boolean fields for specific criteria so the FE can better communicate to the user which criteria their password does or does not meet.
type CheckPasswordResult struct {
	HasUppercase  bool `json:"has_uppercase"`
	HasLowercase  bool `json:"has_lowercase"`
	HasNumber     bool `json:"has_number"`
	HasSpecial    bool `json:"has_special"`
	HasEightChars bool `json:"has_eight_chars"`
}

func (cpr *CheckPasswordResult) IsValid() bool {
	return cpr.HasUppercase && cpr.HasLowercase && cpr.HasNumber && cpr.HasSpecial && cpr.HasEightChars
}

func CheckPasswordStrength(password string) CheckPasswordResult {
	return CheckPasswordResult{
		HasUppercase:  reUppercase.MatchString(password),
		HasLowercase:  reLowercase.MatchString(password),
		HasNumber:     reNumber.MatchString(password),
		HasSpecial:    reSpecial.MatchString(password),
		HasEightChars: len(password) >= 8,
	}
}

// CheckPassword get the existing password from the DB and use its salt to hash the provided password
// and check if they match.
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

func (s *Service) ChangePassword(username, oldPassword, newPassword string) error {
	if oldPassword == newPassword {
		return errors.New("new password cannot be the same as the old password")
	}

	// First, verify the old password is correct
	isValid, err := s.CheckPassword(username, oldPassword)
	if err != nil {
		return err
	}
	if !isValid {
		return errors.New("old password is incorrect")
	}

	// Validate the new password strength
	result := CheckPasswordStrength(newPassword)
	if !result.IsValid() {
		return errors.New("new password does not meet strength requirements")
	}

	// Hash the new password
	newPasswordHash, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update the user's password in the database
	_, err = s.DB.Exec(context.Background(),
		`update open_discord.users set password = $1 where username = $2`, newPasswordHash, username)

	return err
}
