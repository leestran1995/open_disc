package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	myHttp "open_discord/http"
	"open_discord/logic"
	"open_discord/util"
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
	jwtSecret := os.Getenv("JWT_SECRET")

	// Create DB Pool
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	defer pool.Close()

	services := util.CreateServices(pool, jwtSecret)
	handlers := util.CreateHandlers(services, rooms)

	allRooms, err := services.RoomsService.GetAllRooms(context.Background())
	if err != nil {
		log.Fatalf("Unable to get all rooms: %v\n", err)
	}

	for _, room := range allRooms {
		connectionRoom := logic.Room{
			ConnectedClients: make(map[string]*logic.RoomClient),
			RoomID:           room.ID,
			Name:             room.Name,
		}
		rooms[room.ID] = &connectionRoom
	}

	// Router setup
	router := setupRouter()
	router.Use(myHttp.AuthMiddleware(&services.TokenService))

	myHttp.BindUserRoutes(router, &handlers.UserHandler)
	myHttp.BindRoomRoutes(router, &handlers.RoomHandler)
	myHttp.BindMessageRoutes(router, &handlers.MessagesHandler)
	myHttp.BindAuthRoutes(router, &handlers.AuthHandler)

	go router.Run("localhost:8080")

	// Start SSE listener
	http.HandleFunc("/connect/{userId}", handlers.SseHandler.CreateNewSseConnection)
	http.ListenAndServe("localhost:8081", nil)
}
