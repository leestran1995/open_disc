package http

import (
	"net/http"
	"open_discord/postgresql"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserService *postgresql.UserService
}

func BindUserRoutes(router *gin.Engine, handler *UserHandler) {
	router.GET("/users", handler.GetAllUsers)
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	result, err := h.UserService.GetAllUsers(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, result)
}
