package util

import (
	"open_discord/auth"
	"open_discord/http"
	"open_discord/logic"
	"open_discord/postgresql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Services struct {
	UsersService    postgresql.UserService
	RoomsService    postgresql.RoomService
	MessagesService postgresql.MessageService
	AuthService     auth.Service
	TokenService    auth.TokenService
}

func CreateServices(db *pgxpool.Pool, secret string) *Services {
	return &Services{
		UsersService:    postgresql.UserService{DB: db},
		RoomsService:    postgresql.RoomService{DB: db},
		MessagesService: postgresql.MessageService{DB: db},
		AuthService:     auth.Service{DB: db},
		TokenService:    auth.TokenService{Secret: []byte(secret)},
	}
}

type Handlers struct {
	AuthHandler     http.AuthHandler
	UserHandler     http.UserHandler
	RoomHandler     http.RoomHandler
	MessagesHandler http.MessageHandler
	SseHandler      http.SseHandler
}

func CreateHandlers(services *Services, rooms map[uuid.UUID]*logic.Room) *Handlers {
	return &Handlers{
		AuthHandler: http.AuthHandler{
			Auth:  &services.AuthService,
			Token: &services.TokenService,
		},
		UserHandler: http.UserHandler{
			UserService: &services.UsersService,
		},
		RoomHandler: http.RoomHandler{
			RoomService: services.RoomsService,
		},
		MessagesHandler: http.MessageHandler{
			MessageService: services.MessagesService,
		},
		SseHandler: http.SseHandler{
			RoomService:    &services.RoomsService,
			MessageService: &services.MessagesService,
			Rooms:          rooms,
		},
	}
}
