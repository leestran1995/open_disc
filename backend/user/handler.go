package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	UserService *UserService
}

func BindUserRoutes(router *gin.Engine, handler *UserHandler) {
	router.GET("/users", handler.GetAllUsers)
	router.GET("/users/:id", handler.GetUserByID)
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	result, err := h.UserService.GetAllUsers(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	asUuid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	result, err := h.UserService.GetUserByID(c, asUuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
