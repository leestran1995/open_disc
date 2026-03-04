package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Otc struct {
	DB *pgxpool.Pool
}

func (o *Otc) GenerateUuid() (uuid.UUID, error) {
	newId := uuid.New()

	_, err := o.DB.Exec(context.Background(), `insert into open_discord.signup_otcs(code) values ($1)`, newId)
	if err != nil {
		return uuid.Nil, err
	}
	return newId, nil
}
