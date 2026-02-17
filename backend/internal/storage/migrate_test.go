package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrations_CreateCoreTables(t *testing.T) {
	db := setupMySQLForTest(t)
	applyMigrations(t, db)
	mustHaveTable(t, db, "articles")
	mustHaveTable(t, db, "events")
	mustHaveTable(t, db, "bouts")
	mustHaveTable(t, db, "fighters")
}

func applyMigrations(t *testing.T, db *sql.DB) {
	t.Helper()

	migrationPath := filepath.Join("..", "..", "migrations", "0001_init_schema.up.sql")
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

func mustHaveTable(t *testing.T, db *sql.DB, table string) {
	t.Helper()

	var count int
	if err := db.QueryRow(`
		SELECT COUNT(1)
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
		  AND table_name = ?
	`, table).Scan(&count); err != nil {
		t.Fatalf("query table metadata failed: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected table %q to exist", table)
	}
}
