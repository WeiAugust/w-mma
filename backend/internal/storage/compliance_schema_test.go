package storage

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestSchema_HasAuthSourceMediaAndTakedownTables(t *testing.T) {
	db := setupMySQLForTest(t)
	applyMigration(t, db, filepath.Join("..", "..", "migrations", "0001_init_schema.up.sql"))
	applyMigration(t, db, filepath.Join("..", "..", "migrations", "0002_persistence_schema.up.sql"))
	applyMigration(t, db, filepath.Join("..", "..", "migrations", "0003_content_source_media_compliance.up.sql"))

	mustHaveTable(t, db, "admin_users")
	mustHaveTable(t, db, "data_sources")
	mustHaveTable(t, db, "media_assets")
	mustHaveTable(t, db, "summary_jobs")
	mustHaveTable(t, db, "rights_takedowns")

	mustHaveColumn(t, db, "articles", "source_id")
	mustHaveColumn(t, db, "articles", "cover_url")
	mustHaveColumn(t, db, "articles", "video_url")

	mustHaveColumn(t, db, "events", "source_id")
	mustHaveColumn(t, db, "events", "poster_url")
	mustHaveColumn(t, db, "events", "promo_video_url")

	mustHaveColumn(t, db, "fighters", "source_id")
	mustHaveColumn(t, db, "fighters", "avatar_url")
	mustHaveColumn(t, db, "fighters", "intro_video_url")
	mustHaveColumn(t, db, "fighters", "is_manual")
}

func mustHaveColumn(t *testing.T, db *sql.DB, table string, column string) {
	t.Helper()

	var count int
	if err := db.QueryRow(`
		SELECT COUNT(1)
		FROM information_schema.columns
		WHERE table_schema = DATABASE()
		  AND table_name = ?
		  AND column_name = ?
	`, table, column).Scan(&count); err != nil {
		t.Fatalf("query column metadata failed: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected column %q to exist in table %q", column, table)
	}
}
