package review

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PendingCreator interface {
	CreatePending(ctx context.Context, item PendingArticle) (PendingArticle, error)
}

type manualArticleRequest struct {
	SourceID  int64  `json:"source_id"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	SourceURL string `json:"source_url"`
	CoverURL  string `json:"cover_url"`
	VideoURL  string `json:"video_url"`
}

func RegisterAdminManualArticleRoutes(r *gin.Engine, creator PendingCreator) {
	r.POST("/admin/articles/manual", func(c *gin.Context) {
		var req manualArticleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.SourceID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "source_id is required"})
			return
		}
		if req.Title == "" || req.Summary == "" || req.SourceURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "title, summary and source_url are required"})
			return
		}

		item, err := creator.CreatePending(c.Request.Context(), PendingArticle{
			SourceID:  req.SourceID,
			Title:     req.Title,
			Summary:   req.Summary,
			SourceURL: req.SourceURL,
			CoverURL:  req.CoverURL,
			VideoURL:  req.VideoURL,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, item)
	})
}
