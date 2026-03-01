package auth

import (
	"context"
	"errors"
	"log/slog"
	"regexp"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	DB *pgxpool.Pool
}

// Signup performs validation checks and signs the user up if all the checks pass.
func (a *Service) Signup(username string, password string, otc uuid.UUID) error {

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

	tx, err := a.DB.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Insert the new user
	_, err = tx.Exec(context.Background(),
		`insert into open_discord.users(nickname, username, password) values ($1,$2,$3)`, username, username, passwordHash)

	if err != nil {
		return err
	}

	// Mark the OTC as used
	_, err = tx.Exec(context.Background(), `update open_discord.signup_otcs o set used = true where o.code = $1`, otc)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}
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
	return cpr.HasUppercase && cpr.HasLowercase && cpr.HasNumber && cpr.HasSpecial
}

func CheckPasswordStrength(password string) CheckPasswordResult {
	return CheckPasswordResult{
		HasUppercase:  regexp.MustCompile("[A-Z]").MatchString(password),
		HasLowercase:  regexp.MustCompile("[a-z]").MatchString(password),
		HasNumber:     regexp.MustCompile("[0-9]").MatchString(password),
		HasSpecial:    regexp.MustCompile("[!@#$%^&*()-+]").MatchString(password),
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
