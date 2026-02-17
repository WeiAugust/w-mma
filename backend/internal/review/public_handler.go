package review

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterPublicContentRoutes(r *gin.Engine, repo *MemoryRepository) {
	r.GET("/api/articles", func(c *gin.Context) {
		items, err := repo.ListPublished(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	})
}
