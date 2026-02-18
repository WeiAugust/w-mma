DROP INDEX idx_bouts_event_segment ON bouts;

ALTER TABLE bouts
  DROP COLUMN card_segment,
  DROP COLUMN weight_class,
  DROP COLUMN red_ranking,
  DROP COLUMN blue_ranking;
