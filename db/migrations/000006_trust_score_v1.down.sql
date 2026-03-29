ALTER TABLE trust_scores
  DROP CONSTRAINT IF EXISTS uq_trust_scores_entity;

CREATE INDEX IF NOT EXISTS idx_trust_scores_entity ON trust_scores(entity_type, entity_id);

ALTER TABLE trust_scores
  DROP COLUMN IF EXISTS rationale_summary;

ALTER TABLE trust_scores
  ALTER COLUMN total TYPE NUMERIC(5,4);
