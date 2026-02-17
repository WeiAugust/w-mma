CREATE TABLE IF NOT EXISTS pending_articles (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  summary TEXT NOT NULL,
  source_url VARCHAR(512) NOT NULL,
  raw_payload JSON NULL,
  status ENUM('pending','approved','rejected') NOT NULL DEFAULT 'pending',
  reviewer_id BIGINT NULL,
  reviewed_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_pending_articles_source_url (source_url),
  KEY idx_pending_articles_status_created_at (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS fighter_updates (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  fighter_id BIGINT NOT NULL,
  content TEXT NOT NULL,
  published_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_fighter_updates_fighter FOREIGN KEY (fighter_id) REFERENCES fighters(id),
  KEY idx_fighter_updates_fighter_published (fighter_id, published_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
