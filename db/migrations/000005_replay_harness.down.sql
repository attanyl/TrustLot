DROP INDEX IF EXISTS idx_replay_cases_exception;

ALTER TABLE replay_cases
  DROP COLUMN IF EXISTS source_exception_id,
  DROP COLUMN IF EXISTS entity_type,
  DROP COLUMN IF EXISTS expected_match_status,
  DROP COLUMN IF EXISTS expected_reason_code,
  DROP COLUMN IF EXISTS internal_snapshot,
  DROP COLUMN IF EXISTS custodian_snapshot,
  DROP COLUMN IF EXISTS config_snapshot;
