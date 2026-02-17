package e2e

import (
	"path/filepath"
	"testing"

	"github.com/bajiaozhi/w-mma/backend/internal/bootstrap"
)

func TestE2E_RunMigrationsTwiceShouldSucceed(t *testing.T) {
	dsn := setupMySQLDSNForTest(t)
	cfg := bootstrap.Config{MySQLDSN: dsn}

	db, err := bootstrap.NewMySQL(cfg)
	if err != nil {
		t.Fatalf("open mysql failed: %v", err)
	}

	migrationDir := filepath.Join("..", "..", "migrations")
	if err := bootstrap.RunMigrations(db, migrationDir); err != nil {
		t.Fatalf("first run migrations failed: %v", err)
	}
	if err := bootstrap.RunMigrations(db, migrationDir); err != nil {
		t.Fatalf("second run migrations failed: %v", err)
	}
}
