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
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	reviewRepo := review.NewMemoryRepository()
	reviewSvc := review.NewService(reviewRepo)
	review.RegisterAdminReviewRoutes(r, reviewSvc)
	review.RegisterPublicContentRoutes(r, reviewRepo)

	eventSvc := event.NewService(event.NewInMemoryRepository())
	event.RegisterEventRoutes(r, eventSvc)
	event.RegisterAdminEventRoutes(r, eventSvc)

	fighterSvc := fighter.NewService(fighter.NewInMemoryRepository())
	fighter.RegisterFighterRoutes(r, fighterSvc)

	queue := ingest.NewMemoryQueue()
	worker := ingest.NewWorker(queue, &reviewIngestAdapter{repo: reviewRepo}, ingest.NewHTTPParser(nil))
	ingest.RegisterAdminIngestRoutes(r, queue, worker)
}
