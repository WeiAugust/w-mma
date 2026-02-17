package main

import (
	"context"
	"log"
	"time"

	"github.com/bajiaozhi/w-mma/backend/internal/auth"
	"github.com/bajiaozhi/w-mma/backend/internal/bootstrap"
	"github.com/bajiaozhi/w-mma/backend/internal/event"
	"github.com/bajiaozhi/w-mma/backend/internal/fighter"
	apihttp "github.com/bajiaozhi/w-mma/backend/internal/http"
	"github.com/bajiaozhi/w-mma/backend/internal/ingest"
	"github.com/bajiaozhi/w-mma/backend/internal/media"
	"github.com/bajiaozhi/w-mma/backend/internal/queue"
	"github.com/bajiaozhi/w-mma/backend/internal/repository/cache"
	mysqlrepo "github.com/bajiaozhi/w-mma/backend/internal/repository/mysql"
	"github.com/bajiaozhi/w-mma/backend/internal/review"
	"github.com/bajiaozhi/w-mma/backend/internal/source"
	"github.com/bajiaozhi/w-mma/backend/internal/summary"
	"github.com/bajiaozhi/w-mma/backend/internal/takedown"
)

func main() {
	cfg, err := bootstrap.LoadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	db, err := bootstrap.NewMySQL(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := bootstrap.RunMigrations(db, "migrations"); err != nil {
		log.Fatal(err)
	}

	redisClient, err := bootstrap.NewRedisClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	articleRepo := mysqlrepo.NewArticleRepository(db)
	articleCache := cache.NewArticleCache(redisClient, 120*time.Second)
	reviewSvc := review.NewService(articleRepo, articleCache)

	eventRepo := mysqlrepo.NewEventRepository(db)
	eventCache := cache.NewEventCache(redisClient)
	eventSvc := event.NewService(eventRepo, eventCache)

	fighterRepo := mysqlrepo.NewFighterRepository(db)
	fighterCache := cache.NewFighterCache(redisClient)
	fighterSvc := fighter.NewService(fighterRepo, fighterCache)
	sourceRepo := mysqlrepo.NewSourceRepository(db)
	sourceSvc := source.NewService(sourceRepo)
	mediaRepo := mysqlrepo.NewMediaRepository(db)
	mediaSvc := media.NewService(mediaRepo)
	summaryRepo := mysqlrepo.NewSummaryJobRepository(db)
	summarySvc := summary.NewService(summaryRepo, summary.Config{
		Provider: cfg.SummaryProvider,
		APIBase:  cfg.SummaryAPIBase,
		APIKey:   cfg.SummaryAPIKey,
	})
	takedownRepo := mysqlrepo.NewTakedownRepository(db)
	takedownSvc := takedown.NewService(takedownRepo, articleRepo, articleCache)

	stream := queue.NewStreamQueue(redisClient, ingest.FetchStreamName, "worker", "api")
	if err := stream.EnsureGroup(context.Background()); err != nil {
		log.Fatal(err)
	}
	publisher := ingest.NewStreamPublisher(stream)
	authRepo := auth.NewStaticUserRepository(cfg.AdminUsername, cfg.AdminPasswordHash)
	authSvc := auth.NewService(authRepo, cfg.AdminJWTSecret)

	srv := apihttp.NewServerWithDependencies(apihttp.Dependencies{
		ReviewService:   reviewSvc,
		PendingCreator:  articleRepo,
		PublishedRepo:   articleRepo,
		EventService:    eventSvc,
		FighterService:  fighterSvc,
		IngestPublisher: publisher,
		AuthService:     authSvc,
		SourceService:   sourceSvc,
		MediaService:    mediaSvc,
		SummaryService:  summarySvc,
		TakedownService: takedownSvc,
	})

	if err := srv.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
