package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSchema_HasPendingArticlesAndFighterUpdates(t *testing.T) {
	db := setupMySQLForTest(t)
	applyMigration(t, db, filepath.Join("..", "..", "migrations", "0001_init_schema.up.sql"))
	applyMigration(t, db, filepath.Join("..", "..", "migrations", "0002_persistence_schema.up.sql"))
	mustHaveTable(t, db, "pending_articles")
	mustHaveTable(t, db, "fighter_updates")
}

func applyMigration(t *testing.T, db *sql.DB, migrationPath string) {
	t.Helper()

	content, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration file: %v", err)
	}

	statements := strings.Split(string(content), ";")
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" {
			continue
		}
		if _, err := db.Exec(trimmed); err != nil {
			t.Fatalf("apply migration failed: %v", err)
		}
	}
}
