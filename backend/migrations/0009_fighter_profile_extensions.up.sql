ALTER TABLE fighters
  ADD COLUMN name_zh VARCHAR(128) NULL,
  ADD COLUMN stats_json JSON NULL,
  ADD COLUMN records_json JSON NULL;

CREATE INDEX idx_fighters_name_zh ON fighters (name_zh);

INSERT INTO data_sources (name, source_type, platform, source_url, parser_kind, enabled, is_builtin, rights_display, rights_playback, rights_ai_summary)
SELECT 'UFC 官方选手库', 'fighter', 'ufc', 'https://www.ufc.com/athletes', 'ufc_athletes', 0, 1, 1, 0, 1
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM data_sources WHERE name = 'UFC 官方选手库');
