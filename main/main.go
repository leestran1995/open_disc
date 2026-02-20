package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	myHttp "open_discord/http"
	"open_discord/logic"
	"open_discord/postgresql"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	return router
}

var rooms map[uuid.UUID]*logic.Room

func main() {
	fmt.Println("Starting application")

	rooms = make(map[uuid.UUID]*logic.Room)
	ctx := context.Background()

	// Load configs into our env
	err := godotenv.Load("local.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbURL := os.Getenv("DATABASE_URL")

	// Create DB Pool
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	defer pool.Close()

	// Wire up our services and handlers, this boilerplate should eventually get moved somewhere else
	userService := postgresql.UserService{DB: pool}
	userHandler := myHttp.UserHandler{UserService: userService}

	roomService := postgresql.RoomService{DB: pool}
	roomHandler := myHttp.RoomHandler{RoomService: roomService, Rooms: rooms}

	messageService := postgresql.MessageService{DB: pool, Rooms: rooms}
	messageHandler := myHttp.MessageHandler{MessageService: messageService}

	sseHandler := myHttp.SseHandler{
		RoomService:    &roomService,
		Rooms:          rooms,
		MessageService: &messageService,
	}

	allRooms, err := roomService.GetAllRooms(context.Background())
	if err != nil {
		log.Fatalf("Unable to get all rooms: %v\n", err)
	}

	for _, room := range allRooms {
		connectionRoom := logic.Room{
			ConnectedClients: make(map[uuid.UUID]*logic.RoomClient),
			RoomID:           room.ID,
			Name:             room.Name,
		}
		rooms[room.ID] = &connectionRoom
	}

	// Router setup
	router := setupRouter()

	myHttp.BindUserRoutes(router, &userHandler)
	myHttp.BindRoomRoutes(router, &roomHandler)
	myHttp.BindMessageRoutes(router, &messageHandler)

	go router.Run("localhost:8080")

	// Start SSE listener
	http.HandleFunc("/connect/{userId}", sseHandler.CreateNewSseConnection)
	http.ListenAndServe("localhost:8081", nil)
}
