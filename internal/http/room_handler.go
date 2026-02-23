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
	RoomService    postgresql.RoomService
	Rooms          map[uuid.UUID]*logic.Room
	ClientRegistry *logic.ClientRegistry
}

func BindRoomRoutes(router *gin.Engine, RoomHandler *RoomHandler) {
	router.POST("/rooms", RoomHandler.HandleCreateRoom)
	router.GET("/rooms/:id", RoomHandler.HandleGetRoomByID)
	router.GET("/rooms", RoomHandler.HandleGetAllRooms)
	router.POST("/rooms/:id/join", RoomHandler.HandleJoinRoom)
	router.PUT("/rooms/order", RoomHandler.HandleSwapRoomOrder)
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

	h.Rooms[u.ID] = &logic.Room{
		ClientRegistry: h.ClientRegistry,
		RoomID:         u.ID,
		Name:           u.Name,
	}

	roomCreatedEvent := opendisc.RoomEvent{
		RoomEventType: opendisc.RoomCreated,
		Payload:       u.Name,
	}

	// This should be in the service layer, alas
	h.ClientRegistry.FanOutMessage(roomCreatedEvent)

	c.JSON(http.StatusCreated, u)
}

func (h *RoomHandler) HandleGetRoomByID(c *gin.Context) {
	roomId := c.Param("id")
	asUuid, err := uuid.Parse(roomId)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	user, err := h.RoomService.GetByID(c.Request.Context(), asUuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *RoomHandler) HandleJoinRoom(c *gin.Context) {
	var joinRequest opendisc.RoomJoinRequest

	if err := c.ShouldBindJSON(&joinRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roomId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	h.RoomService.JoinRoom(c.Request.Context(), joinRequest, roomId)
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
	res, err := h.RoomService.GetAllRooms(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
