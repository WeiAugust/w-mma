package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bajiaozhi/w-mma/backend/internal/event"
	"github.com/bajiaozhi/w-mma/backend/internal/fighter"
	"github.com/bajiaozhi/w-mma/backend/internal/ingest"
	"github.com/bajiaozhi/w-mma/backend/internal/review"
)

func RegisterRoutes(r *gin.Engine) {
	reviewRepo := review.NewMemoryRepository()
	reviewSvc := review.NewService(reviewRepo)
	eventSvc := event.NewService(event.NewInMemoryRepository())
	fighterSvc := fighter.NewService(fighter.NewInMemoryRepository())

	queue := ingest.NewMemoryQueue()
	worker := ingest.NewWorker(queue, &reviewIngestAdapter{repo: reviewRepo}, ingest.NewHTTPParser(nil))
	publisher := &immediatePublisher{queue: queue, worker: worker}

	RegisterRoutesWithDependencies(r, Dependencies{
		ReviewService:   reviewSvc,
		PublishedRepo:   reviewRepo,
		EventService:    eventSvc,
		FighterService:  fighterSvc,
		IngestPublisher: publisher,
	})
}

func RegisterRoutesWithDependencies(r *gin.Engine, deps Dependencies) {
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	review.RegisterAdminReviewRoutes(r, deps.ReviewService)
	review.RegisterPublicContentRoutes(r, deps.PublishedRepo)
	event.RegisterEventRoutes(r, deps.EventService)
	event.RegisterAdminEventRoutes(r, deps.EventService)
	fighter.RegisterFighterRoutes(r, deps.FighterService)
	ingest.RegisterAdminIngestRoutes(r, deps.IngestPublisher)
}
