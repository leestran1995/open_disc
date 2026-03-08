package room

import (
	"backend/logic"
	"backend/model"
	"backend/serverevent"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomHandler struct {
	RoomService      *RoomService
	Rooms            *map[uuid.UUID]*logic.Room
	ClientRegistry   *logic.ClientRegistry
	ServerEventStore *serverevent.ServerEventStore
}

func NewRoomHandler(
	roomService *RoomService,
	Rooms *map[uuid.UUID]*logic.Room,
	ClientRegistry *logic.ClientRegistry,
	serverEventStore *serverevent.ServerEventStore,
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
	router.GET("/rooms", RoomHandler.HandleGetAllRooms)
	router.PUT("/rooms/order", RoomHandler.HandleSwapRoomOrder)
	router.PUT("/rooms/:roomId/star", RoomHandler.HandleStarRoom)
	router.DELETE("/rooms/:roomId/star", RoomHandler.HandleStarRoom)
}

func (h *RoomHandler) HandleCreateRoom(c *gin.Context) {
	var request CreateRoomRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newRoom, err := h.RoomService.Create(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	(*h.Rooms)[newRoom.ID] = &logic.Room{
		ClientRegistry: h.ClientRegistry,
		RoomID:         newRoom.ID,
		Name:           newRoom.Name,
	}

	roomCreatedEvent := model.ServerEvent{
		ServerEventType: model.RoomCreated,
		Payload:         newRoom.Name,
	}

	h.ServerEventStore.Create(c, model.RoomCreated, newRoom, nil)
	// This should be in the service layer, alas
	h.ClientRegistry.FanOutMessage(roomCreatedEvent, nil)

	c.JSON(http.StatusCreated, newRoom)
}

func (h *RoomHandler) HandleSwapRoomOrder(c *gin.Context) {
	var req SwapRoomOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.RoomService.Reorder(c.Request.Context(), req)
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

	res, err := h.RoomService.GetAll(c, &asUuid)

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
		err := h.RoomService.Star(c, userUuid.(uuid.UUID), roomUuid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	case "DELETE":
		err := h.RoomService.Unstar(c, userUuid.(uuid.UUID), roomUuid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid method"})
	}
}
