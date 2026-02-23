DELETE FROM data_sources
WHERE name = 'UFC 官方选手库'
  AND source_type = 'fighter'
  AND platform = 'ufc'
  AND parser_kind = 'ufc_athletes'
  AND is_builtin = 1;

DROP INDEX idx_fighters_name_zh ON fighters;

ALTER TABLE fighters
  DROP COLUMN name_zh,
  DROP COLUMN stats_json,
  DROP COLUMN records_json;
