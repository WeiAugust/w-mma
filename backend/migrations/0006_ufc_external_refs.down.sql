DROP INDEX idx_events_external_url ON events;
ALTER TABLE events DROP COLUMN external_url;

DROP INDEX idx_fighters_external_url ON fighters;
ALTER TABLE fighters DROP COLUMN external_url;
