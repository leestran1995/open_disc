package util

import (
	auth2 "backend/auth"
	http2 "backend/http"
	"backend/logic"
	postgresql2 "backend/postgresql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Services struct {
	UsersService     postgresql2.UserService
	RoomsService     postgresql2.RoomService
	MessagesService  postgresql2.MessageService
	AuthService      auth2.Service
	TokenService     auth2.TokenService
	ServerEventStore postgresql2.ServerEventStore
}

func CreateServices(
	db *pgxpool.Pool,
	secret string,
	rooms *map[uuid.UUID]*logic.Room,
	clientRegistry *logic.ClientRegistry,
) *Services {
	usersService := postgresql2.UserService{DB: db, ClientRegistry: clientRegistry}
	return &Services{
		UsersService:     postgresql2.UserService{DB: db, ClientRegistry: clientRegistry},
		RoomsService:     postgresql2.RoomService{DB: db},
		MessagesService:  postgresql2.MessageService{DB: db, ClientRegistry: clientRegistry},
		AuthService:      auth2.Service{DB: db},
		TokenService:     auth2.TokenService{Secret: []byte(secret), UserService: &usersService},
		ServerEventStore: postgresql2.ServerEventStore{DB: db, ClientRegistry: clientRegistry},
	}
}

type Handlers struct {
	AuthHandler        http2.AuthHandler
	UserHandler        http2.UserHandler
	RoomHandler        http2.RoomHandler
	MessagesHandler    http2.MessageHandler
	SseHandler         http2.SseHandler
	ServerEventHandler http2.ServerEventHandler
}

func CreateHandlers(services *Services, rooms *map[uuid.UUID]*logic.Room, clientRegistry *logic.ClientRegistry) *Handlers {
	return &Handlers{
		AuthHandler: http2.AuthHandler{
			Auth:  &services.AuthService,
			Token: &services.TokenService,
		},
		UserHandler: http2.UserHandler{
			UserService: &services.UsersService,
		},
		RoomHandler: *http2.NewRoomHandler(
			&services.RoomsService,
			rooms,
			clientRegistry,
			&services.ServerEventStore,
		),
		MessagesHandler: http2.MessageHandler{
			MessageService:   services.MessagesService,
			ServerEventStore: services.ServerEventStore,
		},
		SseHandler: http2.SseHandler{
			RoomService:    &services.RoomsService,
			MessageService: &services.MessagesService,
			Rooms:          rooms,
			TokenService:   &services.TokenService,
			ClientRegistry: clientRegistry,
		},
		ServerEventHandler: http2.ServerEventHandler{
			ServerEventStore: services.ServerEventStore,
		},
	}
}
