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

func RegisterAdminEventRoutes(r *gin.Engine, svc *Service) {
	r.GET("/admin/events", func(c *gin.Context) {
		events, err := svc.ListEvents(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": events})
	})

	r.PUT("/admin/events/:id", func(c *gin.Context) {
		eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
			return
		}

		var req struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := svc.UpdateEvent(c.Request.Context(), eventID, UpdateEventInput{
			Name:   req.Name,
			Status: req.Status,
		}); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}
