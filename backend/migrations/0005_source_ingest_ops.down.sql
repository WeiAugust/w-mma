DROP TABLE IF EXISTS ingest_run_errors;
DROP TABLE IF EXISTS ingest_runs;

DROP INDEX idx_data_sources_builtin_type_enabled ON data_sources;
DROP INDEX idx_data_sources_deleted_at ON data_sources;

ALTER TABLE data_sources
  DROP COLUMN deleted_at,
  DROP COLUMN last_fetch_error,
  DROP COLUMN last_fetch_status,
  DROP COLUMN last_fetch_at,
  DROP COLUMN is_builtin;
