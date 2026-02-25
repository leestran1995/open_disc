package http

import (
	"net/http"
	opendisc "open_discord"
	"open_discord/internal/logic"
	"open_discord/internal/postgresql"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomHandler struct {
	RoomService      *postgresql.RoomService
	Rooms            *map[uuid.UUID]*logic.Room
	ClientRegistry   *logic.ClientRegistry
	ServerEventStore *postgresql.ServerEventStore
}

func NewRoomHandler(
	roomService *postgresql.RoomService,
	Rooms *map[uuid.UUID]*logic.Room,
	ClientRegistry *logic.ClientRegistry,
	serverEventStore *postgresql.ServerEventStore,
) *RoomHandler {
	return &RoomHandler{
		RoomService:      roomService,
		Rooms:            Rooms,
		ClientRegistry:   ClientRegistry,
		ServerEventStore: serverEventStore,
	}
}

func BindRoomRoutes(router *gin.Engine, RoomHandler *RoomHandler) {
	router.POST("/rooms", RoomHandler.HandleCreateRoom)
	router.GET("/rooms/:id", RoomHandler.HandleGetRoomByID)
	router.GET("/rooms", RoomHandler.HandleGetAllRooms)
	router.PUT("/rooms/order", RoomHandler.HandleSwapRoomOrder)
	router.PUT("/rooms/:roomId/star", RoomHandler.HandleStarRoom)
	router.DELETE("/rooms/:roomId/star", RoomHandler.HandleStarRoom)
}

func (h *RoomHandler) HandleCreateRoom(c *gin.Context) {
	var request opendisc.CreateRoomRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.RoomService.Create(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	(*h.Rooms)[u.ID] = &logic.Room{
		ClientRegistry: h.ClientRegistry,
		RoomID:         u.ID,
		Name:           u.Name,
	}

	roomCreatedEvent := opendisc.ServerEvent{
		ServerEventType: opendisc.RoomCreated,
		Payload:         u.Name,
	}

	h.ServerEventStore.Create(c, opendisc.RoomCreated, u)
	// This should be in the service layer, alas
	h.ClientRegistry.FanOutMessage(roomCreatedEvent)

	c.JSON(http.StatusCreated, u)
}

func (h *RoomHandler) HandleGetRoomByID(c *gin.Context) {
	roomId := c.Param("id")
	asUuid, err := uuid.Parse(roomId)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.RoomService.GetByID(c.Request.Context(), asUuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *RoomHandler) HandleSwapRoomOrder(c *gin.Context) {
	var req opendisc.SwapRoomOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.RoomService.ReorderRooms(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (h *RoomHandler) HandleGetAllRooms(c *gin.Context) {
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	asUuid := userId.(uuid.UUID)

	res, err := h.RoomService.GetAllRooms(c, &asUuid)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *RoomHandler) HandleStarRoom(c *gin.Context) {
	roomId := c.Param("roomId")
	if roomId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roomId is empty"})
		return
	}

	var roomUuid, err = uuid.Parse(roomId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userUuid, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found"})
		return
	}

	methodType := c.Request.Method

	switch methodType {
	case "PUT":
		err := h.RoomService.StarRoom(c, userUuid.(uuid.UUID), roomUuid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	case "DELETE":
		err := h.RoomService.UnstarRoom(c, userUuid.(uuid.UUID), roomUuid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid method"})
		return
	}
}
