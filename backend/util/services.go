package util

import (
	auth "backend/auth"
	http2 "backend/http"
	"backend/logic"
	"backend/serverevent"
	"backend/sse"

	"backend/room"
	"backend/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Services struct {
	UsersService     user.UserService
	RoomsService     room.RoomService
	AuthService      auth.Service
	TokenService     auth.TokenService
	ServerEventStore serverevent.ServerEventStore
}

func CreateServices(
	db *pgxpool.Pool,
	secret string,
	rooms *map[uuid.UUID]*logic.Room,
	clientRegistry *logic.ClientRegistry,
	redisClient *redis.Client,
) *Services {
	usersService := user.NewUserService(db, clientRegistry, redisClient)
	return &Services{
		UsersService:     *usersService,
		RoomsService:     *room.NewRoomService(db, redisClient),
		AuthService:      auth.Service{DB: db},
		TokenService:     auth.TokenService{Secret: []byte(secret), UserService: usersService},
		ServerEventStore: serverevent.ServerEventStore{DB: db, ClientRegistry: clientRegistry},
	}
}

type Handlers struct {
	AuthHandler        auth.AuthHandler
	UserHandler        user.UserHandler
	RoomHandler        room.RoomHandler
	MessagesHandler    http2.MessageHandler
	SseHandler         sse.SseHandler
	ServerEventHandler serverevent.ServerEventHandler
}

func CreateHandlers(services *Services, rooms *map[uuid.UUID]*logic.Room, clientRegistry *logic.ClientRegistry) *Handlers {
	return &Handlers{
		AuthHandler: *auth.NewAuthHandler(&services.AuthService, &services.TokenService, &services.UsersService),
		UserHandler: user.UserHandler{
			UserService: &services.UsersService,
		},
		RoomHandler: *room.NewRoomHandler(
			&services.RoomsService,
			rooms,
			clientRegistry,
			&services.ServerEventStore,
		),
		MessagesHandler: http2.MessageHandler{
			ServerEventStore: services.ServerEventStore,
		},
		SseHandler: sse.SseHandler{
			RoomService:    &services.RoomsService,
			Rooms:          rooms,
			TokenService:   &services.TokenService,
			ClientRegistry: clientRegistry,
		},
		ServerEventHandler: serverevent.ServerEventHandler{
			ServerEventStore: services.ServerEventStore,
		},
	}
}
