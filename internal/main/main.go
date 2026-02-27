package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	http2 "open_discord/internal/http"
	"open_discord/internal/logic"
	"open_discord/internal/util"
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
var clients map[string]*logic.RoomClient

func main() {
	fmt.Println("Starting application")

	rooms = make(map[uuid.UUID]*logic.Room)
	clients = make(map[string]*logic.RoomClient)

	clientRegistry := logic.ClientRegistry{Clients: &clients}

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

	services := util.CreateServices(pool, jwtSecret, &rooms, &clientRegistry)
	handlers := util.CreateHandlers(services, &rooms, &clientRegistry)

	// Add all existing rooms to memory
	allRooms, err := services.RoomsService.GetAllRooms(context.Background(), nil)
	if err != nil {
		/clog.Fatalf("Unable to get all rooms: %v\n", err)
	}

	for _, room := range allRooms {
		connectionRoom := logic.Room{
			ClientRegistry: &clientRegistry,
			RoomID:         room.ID,
			Name:           room.Name,
		}
		rooms[room.ID] = &connectionRoom
	}

	// Router setup
	router := setupRouter()
	// router.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"https://chat.lee.fail"},
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Authorization", "Content-Type"},
	// 	AllowCredentials: true,
	// }))
	router.Use(http2.AuthMiddleware(&services.TokenService))

	http2.BindUserRoutes(router, &handlers.UserHandler)
	http2.BindRoomRoutes(router, &handlers.RoomHandler)
	http2.BindMessageRoutes(router, &handlers.MessagesHandler)
	http2.BindAuthRoutes(router, &handlers.AuthHandler)
	http2.BindServerEventRoutes(router, &handlers.ServerEventHandler)
	router.GET(
		"/connect",
		handlers.SseHandler.HandleGinSseConnection,
	)
	router.Run(":8080")
}
