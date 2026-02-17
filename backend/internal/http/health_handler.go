package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bajiaozhi/w-mma/backend/internal/auth"
	"github.com/bajiaozhi/w-mma/backend/internal/event"
	"github.com/bajiaozhi/w-mma/backend/internal/fighter"
	"github.com/bajiaozhi/w-mma/backend/internal/ingest"
	"github.com/bajiaozhi/w-mma/backend/internal/review"
	"github.com/bajiaozhi/w-mma/backend/internal/source"
	"golang.org/x/crypto/bcrypt"
)

func RegisterRoutes(r *gin.Engine) {
	reviewRepo := review.NewMemoryRepository()
	reviewSvc := review.NewService(reviewRepo)
	eventSvc := event.NewService(event.NewInMemoryRepository())
	fighterSvc := fighter.NewService(fighter.NewInMemoryRepository())

	queue := ingest.NewMemoryQueue()
	worker := ingest.NewWorker(queue, &reviewIngestAdapter{repo: reviewRepo}, ingest.NewHTTPParser(nil))
	publisher := &immediatePublisher{queue: queue, worker: worker}
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("admin123456"), bcrypt.DefaultCost)
	authSvc := auth.NewService(auth.NewStaticUserRepository("admin", string(passwordHash)), "test-secret")
	sourceSvc := source.NewService(source.NewInMemoryRepository())

	RegisterRoutesWithDependencies(r, Dependencies{
		ReviewService:   reviewSvc,
		PublishedRepo:   reviewRepo,
		EventService:    eventSvc,
		FighterService:  fighterSvc,
		IngestPublisher: publisher,
		AuthService:     authSvc,
		SourceService:   sourceSvc,
	})
}

func RegisterRoutesWithDependencies(r *gin.Engine, deps Dependencies) {
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	if deps.AuthService != nil {
		auth.RegisterAdminAuthRoutes(r, deps.AuthService)
	}
	if deps.SourceService != nil {
		source.RegisterAdminSourceRoutes(r, deps.SourceService)
	}

	review.RegisterAdminReviewRoutes(r, deps.ReviewService)
	review.RegisterPublicContentRoutes(r, deps.PublishedRepo)
	event.RegisterEventRoutes(r, deps.EventService)
	event.RegisterAdminEventRoutes(r, deps.EventService)
	fighter.RegisterFighterRoutes(r, deps.FighterService)
	ingest.RegisterAdminIngestRoutes(r, deps.IngestPublisher)
}
