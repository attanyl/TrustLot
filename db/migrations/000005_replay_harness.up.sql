-- Extend replay_cases with snapshot data for deterministic replay.
-- The original table only had description/fixture_ref/status which is
-- insufficient to replay a reconciliation comparison without hitting live data.

ALTER TABLE replay_cases
  ADD COLUMN source_exception_id UUID REFERENCES exceptions(id),
  ADD COLUMN entity_type         TEXT NOT NULL DEFAULT '',
  ADD COLUMN expected_match_status TEXT NOT NULL DEFAULT '',
  ADD COLUMN expected_reason_code  TEXT NOT NULL DEFAULT '',
  ADD COLUMN internal_snapshot    JSONB,
  ADD COLUMN custodian_snapshot   JSONB,
  ADD COLUMN config_snapshot      JSONB;

CREATE INDEX idx_replay_cases_exception ON replay_cases(source_exception_id)
  WHERE source_exception_id IS NOT NULL;
