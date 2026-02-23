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
	applyMigration(t, db, filepath.Join("..", "..", "migrations", "0004_fighter_record_column.up.sql"))
	applyMigration(t, db, filepath.Join("..", "..", "migrations", "0005_source_ingest_ops.up.sql"))
	applyMigration(t, db, filepath.Join("..", "..", "migrations", "0006_ufc_external_refs.up.sql"))
	applyMigration(t, db, filepath.Join("..", "..", "migrations", "0007_event_bout_display_fields.up.sql"))
	applyMigration(t, db, filepath.Join("..", "..", "migrations", "0008_bout_result_fields.up.sql"))
	applyMigration(t, db, filepath.Join("..", "..", "migrations", "0009_fighter_profile_extensions.up.sql"))

	mustHaveTable(t, db, "admin_users")
	mustHaveTable(t, db, "data_sources")
	mustHaveTable(t, db, "media_assets")
	mustHaveTable(t, db, "summary_jobs")
	mustHaveTable(t, db, "rights_takedowns")
	mustHaveTable(t, db, "ingest_runs")
	mustHaveTable(t, db, "ingest_run_errors")

	mustHaveColumn(t, db, "articles", "source_id")
	mustHaveColumn(t, db, "articles", "cover_url")
	mustHaveColumn(t, db, "articles", "video_url")

	mustHaveColumn(t, db, "events", "source_id")
	mustHaveColumn(t, db, "events", "poster_url")
	mustHaveColumn(t, db, "events", "promo_video_url")
	mustHaveColumn(t, db, "bouts", "card_segment")
	mustHaveColumn(t, db, "bouts", "weight_class")
	mustHaveColumn(t, db, "bouts", "red_ranking")
	mustHaveColumn(t, db, "bouts", "blue_ranking")
	mustHaveColumn(t, db, "bouts", "method")
	mustHaveColumn(t, db, "bouts", "round")
	mustHaveColumn(t, db, "bouts", "time_sec")

	mustHaveColumn(t, db, "fighters", "source_id")
	mustHaveColumn(t, db, "fighters", "avatar_url")
	mustHaveColumn(t, db, "fighters", "intro_video_url")
	mustHaveColumn(t, db, "fighters", "record")
	mustHaveColumn(t, db, "fighters", "is_manual")
	mustHaveColumn(t, db, "fighters", "name_zh")
	mustHaveColumn(t, db, "fighters", "stats_json")
	mustHaveColumn(t, db, "fighters", "records_json")
	mustHaveColumn(t, db, "pending_articles", "source_id")
	mustHaveColumn(t, db, "events", "external_url")
	mustHaveColumn(t, db, "fighters", "external_url")
	mustHaveColumn(t, db, "data_sources", "is_builtin")
	mustHaveColumn(t, db, "data_sources", "last_fetch_at")
	mustHaveColumn(t, db, "data_sources", "last_fetch_status")
	mustHaveColumn(t, db, "data_sources", "last_fetch_error")
	mustHaveColumn(t, db, "data_sources", "deleted_at")
	mustHaveBuiltInSource(t, db, "UFC 官方赛程")
	mustHaveBuiltInSource(t, db, "ONE Championship 官方赛程")
	mustHaveBuiltInSource(t, db, "PFL 官方赛程")
	mustHaveBuiltInSource(t, db, "JCK 官方赛程")
	mustHaveBuiltInSource(t, db, "WBA 官方赛程")
	mustHaveBuiltInSource(t, db, "WBC 官方赛程")
	mustHaveBuiltInSource(t, db, "IBF 官方赛程")
	mustHaveBuiltInSource(t, db, "WBO 官方赛程")
	mustHaveBuiltInSource(t, db, "UFC 官方选手库")
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

func mustHaveBuiltInSource(t *testing.T, db *sql.DB, name string) {
	t.Helper()

	var count int
	if err := db.QueryRow(`
		SELECT COUNT(1)
		FROM data_sources
		WHERE name = ?
		  AND is_builtin = 1
	`, name).Scan(&count); err != nil {
		t.Fatalf("query built-in source failed: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected built-in source %q to exist", name)
	}
}
