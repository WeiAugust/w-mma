package http

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/bajiaozhi/w-mma/backend/internal/auth"
	"github.com/bajiaozhi/w-mma/backend/internal/event"
	"github.com/bajiaozhi/w-mma/backend/internal/fighter"
	"github.com/bajiaozhi/w-mma/backend/internal/ingest"
	"github.com/bajiaozhi/w-mma/backend/internal/review"
)

type Dependencies struct {
	ReviewService   *review.Service
	PublishedRepo   review.PublishedRepository
	EventService    *event.Service
	FighterService  *fighter.Service
	IngestPublisher ingest.FetchPublisher
	AuthService     *auth.Service
}

func NewServer() *gin.Engine {
	r := gin.New()
	RegisterRoutes(r)
	return r
}

func NewServerWithDependencies(deps Dependencies) *gin.Engine {
	r := gin.New()
	RegisterRoutesWithDependencies(r, deps)
	return r
}

type pendingCreator interface {
	CreatePending(ctx context.Context, item review.PendingArticle) (review.PendingArticle, error)
}

type reviewIngestAdapter struct {
	repo pendingCreator
}

func (a *reviewIngestAdapter) SavePending(ctx context.Context, rec ingest.PendingRecord) error {
	_, err := a.repo.CreatePending(ctx, review.PendingArticle{
		Title:     rec.Title,
		Summary:   rec.Summary,
		SourceURL: rec.SourceURL,
	})
	return err
}

type immediatePublisher struct {
	queue  ingest.Queue
	worker *ingest.Worker
}

func (p *immediatePublisher) Enqueue(ctx context.Context, job ingest.FetchJob) error {
	p.queue.Push(job)
	p.worker.RunOnce(ctx)
	return nil
}
