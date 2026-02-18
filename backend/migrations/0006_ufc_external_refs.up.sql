ALTER TABLE events
  ADD COLUMN external_url VARCHAR(512) NULL;

CREATE INDEX idx_events_external_url ON events (external_url(191));

ALTER TABLE fighters
  ADD COLUMN external_url VARCHAR(512) NULL;

CREATE INDEX idx_fighters_external_url ON fighters (external_url(191));
