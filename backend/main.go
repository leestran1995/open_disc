package main

import (
	"backend/auth"
	"backend/cli"
	http2 "backend/http"
	"backend/logic"
	"backend/role"
	"backend/room"
	"backend/serverevent"
	"backend/util"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"backend/user"

	"github.com/gin-contrib/cors"
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
	allRooms, err := services.RoomsService.GetAll(context.Background(), nil)
	if err != nil {
		log.Fatalf("Unable to get all rooms: %v\n", err)
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
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://chat.lee.fail"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
	router.Use(auth.AuthMiddleware(&services.TokenService))

	user.BindUserRoutes(router, &handlers.UserHandler)
	room.BindRoomRoutes(router, &handlers.RoomHandler)
	http2.BindMessageRoutes(router, &handlers.MessagesHandler)
	auth.BindAuthRoutes(router, &handlers.AuthHandler)
	serverevent.BindServerEventRoutes(router, &handlers.ServerEventHandler)
	router.GET(
		"/connect",
		handlers.SseHandler.EstablishSSEConnection,
	)
	fmt.Println("Starting CLI")
	otc := auth.Otc{DB: pool}
	roleService := role.Service{DB: pool}
	cli := cli.Cli{Otc: &otc, RoleService: &roleService}
	go cli.Run()
	router.Run(":8080")
}
