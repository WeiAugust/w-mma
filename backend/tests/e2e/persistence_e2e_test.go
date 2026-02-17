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
)

func TestE2E_PersistedArticleSurvivesRepositoryRecreate(t *testing.T) {
	dsn := setupMySQLDSNForTest(t)
	cfg := bootstrap.Config{MySQLDSN: dsn}

	db, err := bootstrap.NewMySQL(cfg)
	if err != nil {
		t.Fatalf("open mysql failed: %v", err)
	}
	if err := bootstrap.RunMigrations(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("run migrations failed: %v", err)
	}

	repo := mysqlrepo.NewArticleRepository(db)
	svc := review.NewService(repo)

	pending, err := repo.CreatePending(context.Background(), review.PendingArticle{
		Title:     "persist-title",
		Summary:   "persist-summary",
		SourceURL: fmt.Sprintf("https://example.com/%d", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("create pending failed: %v", err)
	}

	if err := svc.Approve(context.Background(), pending.ID, 9001); err != nil {
		t.Fatalf("approve failed: %v", err)
	}

	db2, err := bootstrap.NewMySQL(cfg)
	if err != nil {
		t.Fatalf("re-open mysql failed: %v", err)
	}
	repo2 := mysqlrepo.NewArticleRepository(db2)
	items, err := repo2.ListPublished(context.Background())
	if err != nil {
		t.Fatalf("list published failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 published article, got %d", len(items))
	}
}
