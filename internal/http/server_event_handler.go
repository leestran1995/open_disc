package http

import (
	"net/http"
	"open_discord/internal/postgresql"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ServerEventHandler struct {
	ServerEventStore postgresql.ServerEventStore
}

func BindServerEventRoutes(router *gin.Engine, serverEventHandler *ServerEventHandler) {
	router.GET("/events", serverEventHandler.HandleGetServerEvents)
}

func (ses *ServerEventHandler) HandleGetServerEvents(c *gin.Context) {
	var eventOrderStart *int
	eventOrderStartString := c.Query("event_order_start")

	if eventOrderStartString != "" {
		result, err := strconv.Atoi(eventOrderStartString)
		eventOrderStart = &result
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		eventOrderStart = nil
	}

	var eventOrderEnd *int
	eventOrderEndString := c.Query("event_order_end")
	if eventOrderEndString != "" {
		result, err := strconv.Atoi(eventOrderEndString)
		eventOrderEnd = &result
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		eventOrderEnd = nil
	}

	result, err := ses.ServerEventStore.GetServerEventsByEventOrder(c, eventOrderStart, eventOrderEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"server_events": result})
}
