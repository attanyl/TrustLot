CREATE TABLE recon_runs (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  entity_type  TEXT NOT NULL CHECK (entity_type IN ('position', 'transaction', 'tax_lot')),
  as_of_date   DATE NOT NULL,
  status       TEXT NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'completed', 'failed')),
  started_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ
);

CREATE INDEX idx_recon_runs_date ON recon_runs(as_of_date);

CREATE TABLE recon_results (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  run_id      UUID NOT NULL REFERENCES recon_runs(id),
  entity_type TEXT NOT NULL,
  entity_id   UUID NOT NULL,
  status      TEXT NOT NULL CHECK (status IN ('MATCH', 'NEAR_MATCH', 'BREAK')),
  reason_code TEXT,
  details     TEXT NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_recon_results_run ON recon_results(run_id);
CREATE INDEX idx_recon_results_breaks ON recon_results(reason_code) WHERE reason_code IS NOT NULL;

CREATE TABLE exceptions (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  recon_run_id UUID NOT NULL REFERENCES recon_runs(id),
  entity_type  TEXT NOT NULL,
  entity_id    UUID NOT NULL,
  reason_code  TEXT NOT NULL,
  status       TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'acknowledged', 'resolved', 'suppressed')),
  assigned_to  TEXT NOT NULL DEFAULT '',
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_exceptions_status ON exceptions(status);
CREATE INDEX idx_exceptions_reason ON exceptions(reason_code);

CREATE TABLE trust_scores (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  entity_type        TEXT NOT NULL,
  entity_id          UUID NOT NULL,
  total              NUMERIC(5,4) NOT NULL,
  source_reliability NUMERIC(5,4) NOT NULL DEFAULT 0,
  mapping_confidence NUMERIC(5,4) NOT NULL DEFAULT 0,
  recon_status       NUMERIC(5,4) NOT NULL DEFAULT 0,
  freshness          NUMERIC(5,4) NOT NULL DEFAULT 0,
  completeness       NUMERIC(5,4) NOT NULL DEFAULT 0,
  human_review_state NUMERIC(5,4) NOT NULL DEFAULT 0,
  anomaly_penalty    NUMERIC(5,4) NOT NULL DEFAULT 0,
  computed_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_trust_scores_entity ON trust_scores(entity_type, entity_id);

CREATE TABLE lineage_edges (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_type TEXT NOT NULL,
  source_id   UUID NOT NULL,
  target_type TEXT NOT NULL,
  target_id   UUID NOT NULL,
  relation    TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_lineage_source ON lineage_edges(source_type, source_id);
CREATE INDEX idx_lineage_target ON lineage_edges(target_type, target_id);

CREATE TABLE replay_cases (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  description TEXT NOT NULL,
  fixture_ref TEXT NOT NULL DEFAULT '',
  status      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'passing', 'failing')),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
