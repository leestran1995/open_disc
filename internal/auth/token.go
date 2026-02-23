package auth

import (
	"context"
	"errors"
	"fmt"
	"open_discord/internal/postgresql"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService struct {
	Secret      []byte
	UserService *postgresql.UserService
}

type Claims struct {
	UserID   uuid.UUID `json:"id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

func (t *TokenService) GenerateJWT(username string) (string, error) {
	user, err := t.UserService.GetUserByUsername(context.Background(), username)
	claims := Claims{
		Username: username,
		UserID:   user.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "open_disc",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(t.Secret)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (t *TokenService) ValidateJWT(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return t.Secret, nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
