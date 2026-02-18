ALTER TABLE data_sources
  ADD COLUMN is_builtin TINYINT(1) NOT NULL DEFAULT 0,
  ADD COLUMN last_fetch_at DATETIME NULL,
  ADD COLUMN last_fetch_status VARCHAR(32) NULL,
  ADD COLUMN last_fetch_error VARCHAR(1024) NULL,
  ADD COLUMN deleted_at DATETIME NULL;

CREATE INDEX idx_data_sources_deleted_at ON data_sources (deleted_at);
CREATE INDEX idx_data_sources_builtin_type_enabled ON data_sources (is_builtin, source_type, enabled);

CREATE TABLE IF NOT EXISTS ingest_runs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  source_id BIGINT NOT NULL,
  source_url VARCHAR(512) NOT NULL,
  parser_kind VARCHAR(64) NOT NULL DEFAULT 'generic',
  status ENUM('queued','running','succeeded','failed') NOT NULL DEFAULT 'queued',
  fetched_count INT NOT NULL DEFAULT 0,
  error_msg TEXT NULL,
  started_at DATETIME NULL,
  finished_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_ingest_runs_source_created (source_id, created_at),
  KEY idx_ingest_runs_status_created (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ingest_run_errors (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  run_id BIGINT NOT NULL,
  error_msg TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_ingest_run_errors_run (run_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO data_sources (name, source_type, platform, source_url, parser_kind, enabled, is_builtin, rights_display, rights_playback, rights_ai_summary)
SELECT 'UFC 官方赛程', 'schedule', 'ufc', 'https://www.ufc.com/events', 'ufc_schedule', 1, 1, 1, 0, 1
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM data_sources WHERE name = 'UFC 官方赛程');

INSERT INTO data_sources (name, source_type, platform, source_url, parser_kind, enabled, is_builtin, rights_display, rights_playback, rights_ai_summary)
SELECT 'ONE Championship 官方赛程', 'schedule', 'one', 'https://www.onefc.com/events/', 'one_schedule', 1, 1, 1, 0, 1
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM data_sources WHERE name = 'ONE Championship 官方赛程');

INSERT INTO data_sources (name, source_type, platform, source_url, parser_kind, enabled, is_builtin, rights_display, rights_playback, rights_ai_summary)
SELECT 'PFL 官方赛程', 'schedule', 'pfl', 'https://pflmma.com/events/', 'pfl_schedule', 1, 1, 1, 0, 1
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM data_sources WHERE name = 'PFL 官方赛程');

INSERT INTO data_sources (name, source_type, platform, source_url, parser_kind, enabled, is_builtin, rights_display, rights_playback, rights_ai_summary)
SELECT 'JCK 官方赛程', 'schedule', 'jck', 'https://jcksport.com/', 'jck_schedule', 1, 1, 1, 0, 1
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM data_sources WHERE name = 'JCK 官方赛程');

INSERT INTO data_sources (name, source_type, platform, source_url, parser_kind, enabled, is_builtin, rights_display, rights_playback, rights_ai_summary)
SELECT 'WBA 官方赛程', 'schedule', 'wba', 'https://www.wbaboxing.com/', 'wba_schedule', 1, 1, 1, 0, 1
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM data_sources WHERE name = 'WBA 官方赛程');

INSERT INTO data_sources (name, source_type, platform, source_url, parser_kind, enabled, is_builtin, rights_display, rights_playback, rights_ai_summary)
SELECT 'WBC 官方赛程', 'schedule', 'wbc', 'https://www.wbcboxing.com/', 'wbc_schedule', 1, 1, 1, 0, 1
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM data_sources WHERE name = 'WBC 官方赛程');

INSERT INTO data_sources (name, source_type, platform, source_url, parser_kind, enabled, is_builtin, rights_display, rights_playback, rights_ai_summary)
SELECT 'IBF 官方赛程', 'schedule', 'ibf', 'https://www.ibf-usba-boxing.com/', 'ibf_schedule', 1, 1, 1, 0, 1
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM data_sources WHERE name = 'IBF 官方赛程');

INSERT INTO data_sources (name, source_type, platform, source_url, parser_kind, enabled, is_builtin, rights_display, rights_playback, rights_ai_summary)
SELECT 'WBO 官方赛程', 'schedule', 'wbo', 'https://www.wboboxing.com/', 'wbo_schedule', 1, 1, 1, 0, 1
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM data_sources WHERE name = 'WBO 官方赛程');
