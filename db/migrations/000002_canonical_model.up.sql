CREATE TABLE accounts (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  external_id  TEXT NOT NULL,
  custodian_id UUID NOT NULL REFERENCES custodians(id),
  name         TEXT NOT NULL DEFAULT '',
  account_type TEXT NOT NULL DEFAULT '',
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (external_id, custodian_id)
);

CREATE TABLE securities (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  symbol        TEXT NOT NULL DEFAULT '',
  cusip         TEXT NOT NULL DEFAULT '',
  isin          TEXT NOT NULL DEFAULT '',
  sedol         TEXT NOT NULL DEFAULT '',
  name          TEXT NOT NULL DEFAULT '',
  security_type TEXT NOT NULL DEFAULT '',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_securities_cusip ON securities(cusip) WHERE cusip != '';
CREATE INDEX idx_securities_isin ON securities(isin) WHERE isin != '';

CREATE TABLE positions (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id   UUID NOT NULL REFERENCES accounts(id),
  security_id  UUID NOT NULL REFERENCES securities(id),
  quantity     NUMERIC(20,8) NOT NULL,
  market_value NUMERIC(20,4) NOT NULL DEFAULT 0,
  as_of_date   DATE NOT NULL,
  source       TEXT NOT NULL DEFAULT 'custodian',
  batch_id     UUID REFERENCES ingestion_batches(id),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_positions_account ON positions(account_id);
CREATE INDEX idx_positions_as_of ON positions(as_of_date);

CREATE TABLE transactions (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id  UUID NOT NULL REFERENCES accounts(id),
  security_id UUID NOT NULL REFERENCES securities(id),
  txn_type    TEXT NOT NULL,
  quantity    NUMERIC(20,8) NOT NULL,
  price       NUMERIC(20,8) NOT NULL DEFAULT 0,
  amount      NUMERIC(20,4) NOT NULL DEFAULT 0,
  trade_date  DATE NOT NULL,
  settle_date DATE NOT NULL,
  source      TEXT NOT NULL DEFAULT 'custodian',
  batch_id    UUID REFERENCES ingestion_batches(id),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_transactions_account ON transactions(account_id);
CREATE INDEX idx_transactions_trade_date ON transactions(trade_date);

CREATE TABLE tax_lots (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id    UUID NOT NULL REFERENCES accounts(id),
  security_id   UUID NOT NULL REFERENCES securities(id),
  open_date     DATE NOT NULL,
  quantity      NUMERIC(20,8) NOT NULL,
  cost_basis    NUMERIC(20,4) NOT NULL,
  relief_method TEXT NOT NULL DEFAULT 'FIFO',
  source        TEXT NOT NULL DEFAULT 'custodian',
  batch_id      UUID REFERENCES ingestion_batches(id),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tax_lots_account ON tax_lots(account_id);
