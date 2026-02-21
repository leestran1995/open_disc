package http

import (
	"net/http"
	opendisc "open_discord"
	"open_discord/logic"
	"open_discord/postgresql"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomHandler struct {
	RoomService postgresql.RoomService
	Rooms       map[uuid.UUID]*logic.Room
}

func BindRoomRoutes(router *gin.Engine, RoomHandler *RoomHandler) {
	router.POST("/rooms", RoomHandler.HandleCreateRoom)
	router.GET("/rooms/:id", RoomHandler.HandleGetRoomByID)
	router.POST("/rooms/:id/join", RoomHandler.HandleJoinRoom)
}

func (h *RoomHandler) HandleCreateRoom(c *gin.Context) {
	var request opendisc.CreateRoomRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	u, err := h.RoomService.Create(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.Rooms[u.ID] = &logic.Room{
		ConnectedClients: make(map[string]*logic.RoomClient),
		RoomID:           u.ID,
		Name:             u.Name,
	}

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
	}

	c.JSON(http.StatusOK, user)
}

func (h *RoomHandler) HandleJoinRoom(c *gin.Context) {
	var joinRequest opendisc.RoomJoinRequest

	if err := c.ShouldBindJSON(&joinRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	roomId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	h.RoomService.JoinRoom(c.Request.Context(), joinRequest, roomId)
}
