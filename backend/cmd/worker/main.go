package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bajiaozhi/w-mma/backend/internal/bootstrap"
	"github.com/bajiaozhi/w-mma/backend/internal/ingest"
	"github.com/bajiaozhi/w-mma/backend/internal/queue"
	mysqlrepo "github.com/bajiaozhi/w-mma/backend/internal/repository/mysql"
	"github.com/bajiaozhi/w-mma/backend/internal/review"
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
	worker := ingest.NewQueuelessWorker(&reviewPendingAdapter{repo: articleRepo}, ingest.NewHTTPParser(nil))

	stream := queue.NewStreamQueue(redisClient, ingest.FetchStreamName, "worker", "worker-1")
	if err := stream.EnsureGroup(context.Background()); err != nil {
		log.Fatal(err)
	}
	consumer := ingest.NewStreamConsumer(stream)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
