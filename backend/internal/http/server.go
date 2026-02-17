package http

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/bajiaozhi/w-mma/backend/internal/ingest"
	"github.com/bajiaozhi/w-mma/backend/internal/review"
)

func NewServer() *gin.Engine {
	r := gin.New()
	RegisterRoutes(r)
	return r
}

type reviewIngestAdapter struct {
	repo *review.MemoryRepository
}

func (a *reviewIngestAdapter) SavePending(ctx context.Context, rec ingest.PendingRecord) error {
	_, err := a.repo.CreatePending(ctx, review.PendingArticle{
		Title:     rec.Title,
		Summary:   rec.Summary,
		SourceURL: rec.SourceURL,
	})
	return err
}
