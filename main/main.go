package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	opendisc "open_discord"
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
	roomHandler := myHttp.RoomHandler{RoomService: roomService}

	messageService := postgresql.MessageService{DB: pool, Rooms: rooms}
	messageHandler := myHttp.MessageHandler{MessageService: messageService}

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
	http.HandleFunc("/connect/{userId}", wireEventHandler(roomService))
	http.ListenAndServe("localhost:8081", nil)
}

func wireEventHandler(roomService postgresql.RoomService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// Grab the user id from the path
		userId, err := uuid.Parse(r.PathValue("userId"))
		if err != nil {
			log.Fatalf("Unable to parse user id: %v\n", err)
			return
		}

		// Create a send channel for the newly-connected user
		sendChannel := make(chan opendisc.Message)

		defer close(sendChannel)

		roomClient := logic.RoomClient{
			UserID:      userId,
			SendChannel: sendChannel,
		}

		// Add this user's send channel to all of the rooms they belong to
		userRooms, err := roomService.GetRoomsForUser(context.Background(), userId)

		if err != nil {
			log.Fatalf("Unable to get all rooms: %v\n", err)
			return
		}

		for _, ur := range userRooms {
			matchingRoom := rooms[ur.ID]
			matchingRoom.ConnectToRoom(roomClient)
		}

		// Set CORS headers to allow all origins. You may want to restrict this to specific origins in a production environment.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		for {
			select {
			case <-r.Context().Done():
				fmt.Println("Client disconnected")

				for _, ur := range userRooms {
					matchingRoom := rooms[ur.ID]
					matchingRoom.DisconnectFromRoom(roomClient)
				}

				return

			case message := <-sendChannel:
				fmt.Println("Received message in SSE handler")
				asJson, err := json.Marshal(message)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf("Event %s", string(asJson)))
				w.(http.Flusher).Flush()
			}
		}
	}
}
