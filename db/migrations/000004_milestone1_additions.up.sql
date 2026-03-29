-- Custodian-specific account mappings.
-- A single canonical account may have representations at multiple custodians.
CREATE TABLE custodian_accounts (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id      UUID NOT NULL REFERENCES accounts(id),
  custodian_id    UUID NOT NULL REFERENCES custodians(id),
  external_ref    TEXT NOT NULL,
  status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'pending')),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (custodian_id, external_ref)
);

CREATE INDEX idx_custodian_accounts_account ON custodian_accounts(account_id);

-- Normalized security identifiers.
-- Separating identifiers lets us handle the common case where a security
-- has a CUSIP from one custodian and an ISIN from another.
CREATE TABLE security_identifiers (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  security_id UUID NOT NULL REFERENCES securities(id),
  id_type     TEXT NOT NULL CHECK (id_type IN ('cusip', 'isin', 'sedol', 'ticker', 'internal')),
  id_value    TEXT NOT NULL,
  source      TEXT NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (id_type, id_value)
);

CREATE INDEX idx_security_identifiers_security ON security_identifiers(security_id);

-- Exception lifecycle events for audit trail.
CREATE TABLE exception_events (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  exception_id UUID NOT NULL REFERENCES exceptions(id),
  event_type   TEXT NOT NULL CHECK (event_type IN ('created', 'acknowledged', 'assigned', 'resolved', 'suppressed', 'reopened', 'commented')),
  actor        TEXT NOT NULL DEFAULT 'system',
  detail       TEXT NOT NULL DEFAULT '',
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_exception_events_exception ON exception_events(exception_id);
