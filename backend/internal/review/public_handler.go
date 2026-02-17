package review

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PublishedRepository interface {
	ListPublished(ctx context.Context) ([]PendingArticle, error)
}

func RegisterPublicContentRoutes(r *gin.Engine, repo PublishedRepository) {
	r.GET("/api/articles", func(c *gin.Context) {
		items, err := repo.ListPublished(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	})
}
