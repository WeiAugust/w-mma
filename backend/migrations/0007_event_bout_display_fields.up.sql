ALTER TABLE bouts
  ADD COLUMN card_segment VARCHAR(24) NULL,
  ADD COLUMN weight_class VARCHAR(64) NULL,
  ADD COLUMN red_ranking VARCHAR(32) NULL,
  ADD COLUMN blue_ranking VARCHAR(32) NULL;

CREATE INDEX idx_bouts_event_segment ON bouts (event_id, card_segment, sequence_no);
