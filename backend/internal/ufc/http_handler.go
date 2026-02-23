package ufc

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(r *gin.Engine, svc *Service) {
	r.POST("/admin/sources/:id/sync", func(c *gin.Context) {
		sourceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source id"})
			return
		}

		result, err := svc.SyncSource(c.Request.Context(), sourceID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "result": result})
	})

	r.POST("/admin/sources/:id/sync-fighter", func(c *gin.Context) {
		sourceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source id"})
			return
		}
		var req struct {
			URL string `json:"url"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := svc.SyncSingleAthlete(c.Request.Context(), sourceID, req.URL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "result": result})
	})

	r.POST("/admin/sources/sync-enabled", func(c *gin.Context) {
		result, err := svc.SyncEnabledSources(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "result": result})
	})
}
