ALTER TABLE fighters
  DROP COLUMN is_manual,
  DROP COLUMN intro_video_url,
  DROP COLUMN avatar_url,
  DROP COLUMN source_id;

ALTER TABLE events
  DROP COLUMN promo_video_url,
  DROP COLUMN poster_url,
  DROP COLUMN source_id;

ALTER TABLE articles
  MODIFY COLUMN status ENUM('pending','published') NOT NULL DEFAULT 'pending',
  DROP COLUMN published_mode,
  DROP COLUMN video_url,
  DROP COLUMN cover_url,
  DROP COLUMN source_id;

ALTER TABLE pending_articles
  DROP COLUMN source_id;

DROP TABLE IF EXISTS rights_takedowns;
DROP TABLE IF EXISTS summary_jobs;
DROP TABLE IF EXISTS media_assets;
DROP TABLE IF EXISTS data_sources;
DROP TABLE IF EXISTS admin_users;
