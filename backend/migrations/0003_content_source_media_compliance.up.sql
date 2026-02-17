CREATE TABLE IF NOT EXISTS admin_users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  status ENUM('active','disabled') NOT NULL DEFAULT 'active',
  last_login_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_admin_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS data_sources (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(128) NOT NULL,
  source_type ENUM('news','schedule','fighter') NOT NULL,
  platform VARCHAR(64) NOT NULL,
  account_id VARCHAR(128) NULL,
  source_url VARCHAR(512) NOT NULL,
  parser_kind VARCHAR(64) NOT NULL DEFAULT 'generic',
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  rights_display TINYINT(1) NOT NULL DEFAULT 0,
  rights_playback TINYINT(1) NOT NULL DEFAULT 0,
  rights_ai_summary TINYINT(1) NOT NULL DEFAULT 0,
  rights_expires_at DATETIME NULL,
  rights_proof_url VARCHAR(512) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_data_sources_type_enabled (source_type, enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS media_assets (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  owner_type ENUM('article','event','fighter') NOT NULL,
  owner_id BIGINT NOT NULL,
  media_type ENUM('image','video') NOT NULL,
  url VARCHAR(512) NOT NULL,
  cover_url VARCHAR(512) NULL,
  title VARCHAR(255) NULL,
  sort_no INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_media_owner_sort (owner_type, owner_id, sort_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS summary_jobs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  source_id BIGINT NOT NULL,
  target_type ENUM('article') NOT NULL,
  target_id BIGINT NOT NULL,
  status ENUM('pending','running','done','failed','manual_required') NOT NULL,
  provider VARCHAR(64) NULL,
  error_msg TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_summary_jobs_status_created (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS rights_takedowns (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  target_type ENUM('article','event','fighter') NOT NULL,
  target_id BIGINT NOT NULL,
  reason TEXT NOT NULL,
  complainant VARCHAR(128) NULL,
  evidence_url VARCHAR(512) NULL,
  status ENUM('open','resolved') NOT NULL DEFAULT 'open',
  action ENUM('offlined','rejected') NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  resolved_at DATETIME NULL,
  KEY idx_takedowns_target_status (target_type, target_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

ALTER TABLE articles
  ADD COLUMN source_id BIGINT NULL;

ALTER TABLE articles
  ADD COLUMN cover_url VARCHAR(512) NULL;

ALTER TABLE articles
  ADD COLUMN video_url VARCHAR(512) NULL;

ALTER TABLE articles
  ADD COLUMN published_mode ENUM('manual','ai') NOT NULL DEFAULT 'manual';

ALTER TABLE articles
  MODIFY COLUMN status ENUM('pending','published','offline') NOT NULL DEFAULT 'pending';

ALTER TABLE pending_articles
  ADD COLUMN source_id BIGINT NULL;

ALTER TABLE events
  ADD COLUMN source_id BIGINT NULL;

ALTER TABLE events
  ADD COLUMN poster_url VARCHAR(512) NULL;

ALTER TABLE events
  ADD COLUMN promo_video_url VARCHAR(512) NULL;

ALTER TABLE fighters
  ADD COLUMN source_id BIGINT NULL;

ALTER TABLE fighters
  ADD COLUMN avatar_url VARCHAR(512) NULL;

ALTER TABLE fighters
  ADD COLUMN intro_video_url VARCHAR(512) NULL;

ALTER TABLE fighters
  ADD COLUMN is_manual TINYINT(1) NOT NULL DEFAULT 0;
