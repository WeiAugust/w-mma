package event

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterEventRoutes(r *gin.Engine, svc *Service) {
	r.GET("/api/events", func(c *gin.Context) {
		events, err := svc.ListEvents(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": events})
	})

	r.GET("/api/events/:id", func(c *gin.Context) {
		eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
			return
		}

		card, err := svc.GetEventCard(c.Request.Context(), eventID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, card)
	})
}
