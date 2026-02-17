CREATE TABLE IF NOT EXISTS articles (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  content TEXT NOT NULL,
  source_url VARCHAR(512) NOT NULL,
  status ENUM('pending','published') NOT NULL DEFAULT 'pending',
  published_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_articles_source_url (source_url)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS events (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  org VARCHAR(32) NOT NULL,
  name VARCHAR(255) NOT NULL,
  status ENUM('scheduled','live','completed') NOT NULL DEFAULT 'scheduled',
  starts_at DATETIME NOT NULL,
  venue VARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_events_starts_at (starts_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS fighters (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(128) NOT NULL,
  nickname VARCHAR(128) NULL,
  country VARCHAR(64) NULL,
  weight_class VARCHAR(64) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_fighters_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS bouts (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  event_id BIGINT NOT NULL,
  red_fighter_id BIGINT NOT NULL,
  blue_fighter_id BIGINT NOT NULL,
  sequence_no INT NOT NULL,
  result VARCHAR(128) NULL,
  winner_fighter_id BIGINT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_bouts_event FOREIGN KEY (event_id) REFERENCES events(id),
  CONSTRAINT fk_bouts_red_fighter FOREIGN KEY (red_fighter_id) REFERENCES fighters(id),
  CONSTRAINT fk_bouts_blue_fighter FOREIGN KEY (blue_fighter_id) REFERENCES fighters(id),
  CONSTRAINT fk_bouts_winner_fighter FOREIGN KEY (winner_fighter_id) REFERENCES fighters(id),
  UNIQUE KEY uk_bouts_event_sequence (event_id, sequence_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS pending_reviews (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  payload_json JSON NOT NULL,
  status ENUM('pending','approved','rejected') NOT NULL DEFAULT 'pending',
  reviewer_id BIGINT NULL,
  reviewed_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_pending_reviews_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
