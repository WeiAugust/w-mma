package review

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PublishedRepository interface {
	ListPublished(ctx context.Context) ([]PendingArticle, error)
}

type PlaybackPolicy interface {
	CanPlay(ctx context.Context, sourceID int64) bool
}

func RegisterPublicContentRoutes(r *gin.Engine, repo PublishedRepository, policy ...PlaybackPolicy) {
	var playbackPolicy PlaybackPolicy
	if len(policy) > 0 {
		playbackPolicy = policy[0]
	}

	r.GET("/api/articles", func(c *gin.Context) {
		items, err := repo.ListPublished(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for idx := range items {
			if items[idx].VideoURL == "" {
				items[idx].CanPlay = false
				continue
			}

			canPlay := false
			if playbackPolicy != nil {
				canPlay = playbackPolicy.CanPlay(c.Request.Context(), items[idx].SourceID)
			}
			items[idx].CanPlay = canPlay
			if !canPlay {
				items[idx].VideoURL = ""
			}
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	})
}
