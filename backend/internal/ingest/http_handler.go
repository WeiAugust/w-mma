package ingest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type fetchRequest struct {
	SourceID int64  `json:"source_id"`
	URL      string `json:"url"`
}

func RegisterAdminIngestRoutes(r *gin.Engine, publisher FetchPublisher) {
	r.POST("/admin/ingest/fetch", func(c *gin.Context) {
		var req fetchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
			return
		}

		if err := publisher.Enqueue(c.Request.Context(), FetchJob{SourceID: req.SourceID, URL: req.URL}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}
