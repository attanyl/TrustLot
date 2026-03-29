-- Widen total from NUMERIC(5,4) to NUMERIC(5,2) to hold 0-100 range.
-- Add rationale_summary for explainable trust output.
-- Add unique constraint on (entity_type, entity_id) for upsert support.
ALTER TABLE trust_scores
  ALTER COLUMN total TYPE NUMERIC(5,2);

ALTER TABLE trust_scores
  ADD COLUMN IF NOT EXISTS rationale_summary TEXT NOT NULL DEFAULT '';

-- Replace the non-unique index with a unique constraint for upsert.
DROP INDEX IF EXISTS idx_trust_scores_entity;
ALTER TABLE trust_scores
  ADD CONSTRAINT uq_trust_scores_entity UNIQUE (entity_type, entity_id);
