package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bajiaozhi/w-mma/backend/internal/bootstrap"
	"github.com/bajiaozhi/w-mma/backend/internal/ingest"
	"github.com/bajiaozhi/w-mma/backend/internal/live"
	"github.com/bajiaozhi/w-mma/backend/internal/queue"
	"github.com/bajiaozhi/w-mma/backend/internal/repository/cache"
	mysqlrepo "github.com/bajiaozhi/w-mma/backend/internal/repository/mysql"
	"github.com/bajiaozhi/w-mma/backend/internal/review"
	"github.com/bajiaozhi/w-mma/backend/internal/source"
	"github.com/bajiaozhi/w-mma/backend/internal/ufc"
)

type reviewPendingCreator interface {
	CreatePending(ctx context.Context, item review.PendingArticle) (review.PendingArticle, error)
}

type reviewPendingAdapter struct {
	repo reviewPendingCreator
}

func (a *reviewPendingAdapter) SavePending(ctx context.Context, rec ingest.PendingRecord) error {
	_, err := a.repo.CreatePending(ctx, review.PendingArticle{
		Title:     rec.Title,
		Summary:   rec.Summary,
		SourceURL: rec.SourceURL,
	})
	return err
}

type ufcLiveRepoAdapter struct {
	repo *mysqlrepo.EventRepository
}

func (a *ufcLiveRepoAdapter) ListTrackableEvents(ctx context.Context) ([]live.UFCTrackableEvent, error) {
	rows, err := a.repo.ListUFCLiveTrackableEvents(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]live.UFCTrackableEvent, 0, len(rows))
	for _, row := range rows {
		items = append(items, live.UFCTrackableEvent{
			ID:          row.ID,
			Status:      row.Status,
			StartsAt:    row.StartsAt,
			ExternalURL: row.ExternalURL,
		})
	}
	return items, nil
}

func (a *ufcLiveRepoAdapter) ListBoutSnapshots(ctx context.Context, eventID int64) ([]live.UFCBoutSnapshot, error) {
	rows, err := a.repo.ListUFCLiveBoutSnapshots(ctx, eventID)
	if err != nil {
		return nil, err
	}
	items := make([]live.UFCBoutSnapshot, 0, len(rows))
	for _, row := range rows {
		items = append(items, live.UFCBoutSnapshot{
			BoutID:        row.BoutID,
			SequenceNo:    row.SequenceNo,
			RedFighterID:  row.RedFighterID,
			BlueFighterID: row.BlueFighterID,
			WinnerID:      row.WinnerID,
			Method:        row.Method,
			Round:         row.Round,
			TimeSec:       row.TimeSec,
			Result:        row.Result,
		})
	}
	return items, nil
}

func (a *ufcLiveRepoAdapter) UpdateEventStatus(ctx context.Context, eventID int64, status string) error {
	return a.repo.UpdateEventStatus(ctx, eventID, status)
}

func (a *ufcLiveRepoAdapter) UpsertBoutResult(ctx context.Context, eventID int64, boutID int64, winnerID int64, method string, round int, timeSec int, result string) error {
	return a.repo.UpsertUFCLiveBoutResult(ctx, eventID, boutID, winnerID, method, round, timeSec, result)
}

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
	eventRepo := mysqlrepo.NewEventRepository(db)
	eventCache := cache.NewEventCache(redisClient)
	sourceRepo := mysqlrepo.NewSourceRepository(db)
	sourceSvc := source.NewService(sourceRepo)
	ufcSyncRepo := mysqlrepo.NewUFCSyncRepository(db)
	imageMirror := ufc.NewLocalImageMirror(ufc.LocalImageMirrorConfig{
		StorageDir: cfg.MediaCacheDir,
		PublicBase: cfg.PublicBaseURL,
	})
	ufcSyncSvc := ufc.NewService(sourceSvc, ufcSyncRepo, ufc.NewHTTPClient(nil), ufc.WithImageMirror(imageMirror))
	worker := ingest.NewQueuelessWorker(&reviewPendingAdapter{repo: articleRepo}, ingest.NewDefaultParserRegistry(nil))

	stream := queue.NewStreamQueue(redisClient, ingest.FetchStreamName, "worker", "worker-1")
	if err := stream.EnsureGroup(context.Background()); err != nil {
		log.Fatal(err)
	}
	consumer := ingest.NewStreamConsumer(stream)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go ufc.StartScheduler(ctx, ufcSyncSvc, 12*time.Hour)
	ufcLiveMonitor := live.NewUFCLiveMonitor(
		&ufcLiveRepoAdapter{repo: eventRepo},
		ufc.NewHTTPClient(nil),
		eventCache,
		live.UFCLiveMonitorConfig{},
	)
	go ufcLiveMonitor.Run(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := consumer.ConsumeOnce(ctx, func(job ingest.FetchJob) error {
				return worker.HandleJob(ctx, job)
			}); err != nil {
				log.Printf("consume ingest stream failed: %v", err)
			}
		}
	}
}
