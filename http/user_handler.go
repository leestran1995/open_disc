package http

import (
	"net/http"
	opendisc "open_discord"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	UserService opendisc.UserService
}

func BindUserRoutes(router *gin.Engine, handler *UserHandler) {
	router.POST("/users", handler.HandleCreateUser)
	router.GET("/users/:id", handler.handleGetUserByID)
}

func (h *UserHandler) HandleCreateUser(c *gin.Context) {
	var request opendisc.CreateUserRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	u, err := h.UserService.CreateUser(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusCreated, u)
}

func (h *UserHandler) handleGetUserByID(c *gin.Context) {
	userId := c.Param("id")
	asUuid, err := uuid.Parse(userId)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	user, err := h.UserService.GetUserByID(c.Request.Context(), asUuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, user)
}
