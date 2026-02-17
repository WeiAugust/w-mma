package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/bajiaozhi/w-mma/backend/internal/bootstrap"
	mysqlrepo "github.com/bajiaozhi/w-mma/backend/internal/repository/mysql"
	"github.com/bajiaozhi/w-mma/backend/internal/review"
	"github.com/bajiaozhi/w-mma/backend/internal/takedown"
)

func TestE2E_TakedownOfflinesArticleAndPublicAPIHidesIt(t *testing.T) {
	dsn := setupMySQLDSNForTest(t)
	cfg := bootstrap.Config{MySQLDSN: dsn}

	db, err := bootstrap.NewMySQL(cfg)
	if err != nil {
		t.Fatalf("open mysql failed: %v", err)
	}
	if err := bootstrap.RunMigrations(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("run migrations failed: %v", err)
	}

	articleRepo := mysqlrepo.NewArticleRepository(db)
	reviewSvc := review.NewService(articleRepo)

	pending, err := articleRepo.CreatePending(context.Background(), review.PendingArticle{
		SourceID:  1,
		Title:     "compliance-title",
		Summary:   "compliance-summary",
		SourceURL: fmt.Sprintf("https://example.com/%d", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("create pending failed: %v", err)
	}
	if err := reviewSvc.Approve(context.Background(), pending.ID, 9001); err != nil {
		t.Fatalf("approve failed: %v", err)
	}

	before, err := articleRepo.ListPublished(context.Background())
	if err != nil {
		t.Fatalf("list before takedown failed: %v", err)
	}
	if len(before) != 1 {
		t.Fatalf("expected 1 published article before takedown, got %d", len(before))
	}

	takedownRepo := mysqlrepo.NewTakedownRepository(db)
	takedownSvc := takedown.NewService(takedownRepo, articleRepo, nil)
	ticket, err := takedownSvc.Create(context.Background(), takedown.CreateInput{
		TargetType: "article",
		TargetID:   before[0].ID,
		Reason:     "copyright complaint",
	})
	if err != nil {
		t.Fatalf("create takedown failed: %v", err)
	}

	if err := takedownSvc.Resolve(context.Background(), ticket.ID, takedown.ActionOfflined); err != nil {
		t.Fatalf("resolve takedown failed: %v", err)
	}

	after, err := articleRepo.ListPublished(context.Background())
	if err != nil {
		t.Fatalf("list after takedown failed: %v", err)
	}
	if len(after) != 0 {
		t.Fatalf("expected 0 published article after takedown, got %d", len(after))
	}
}
