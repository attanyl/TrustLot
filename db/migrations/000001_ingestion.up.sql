CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE custodians (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name       TEXT NOT NULL UNIQUE,
  format     TEXT NOT NULL CHECK (format IN ('csv', 'json', 'fixed_width')),
  endpoint   TEXT NOT NULL DEFAULT '',
  active     BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ingestion_batches (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  custodian_id UUID NOT NULL REFERENCES custodians(id),
  status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
  file_ref     TEXT NOT NULL DEFAULT '',
  record_count INTEGER NOT NULL DEFAULT 0,
  started_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ingestion_batches_custodian ON ingestion_batches(custodian_id);
CREATE INDEX idx_ingestion_batches_status ON ingestion_batches(status);

CREATE TABLE raw_records (
  id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  batch_id  UUID NOT NULL REFERENCES ingestion_batches(id),
  seq_num   INTEGER NOT NULL,
  raw_data  BYTEA NOT NULL,
  parsed_at TIMESTAMPTZ,
  error     TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_raw_records_batch ON raw_records(batch_id);
