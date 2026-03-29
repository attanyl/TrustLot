-- TrustLot seed data — deterministic synthetic records for local development.
-- Run after migrations: psql $DATABASE_URL -f db/seed.sql
--
-- Uses fixed UUIDs so the seed is idempotent (re-runnable after truncation).

BEGIN;

-- Custodians
INSERT INTO custodians (id, name, format, endpoint) VALUES
  ('a0000000-0000-0000-0000-000000000001', 'NorthRiver',       'csv',         '/feeds/northriver'),
  ('a0000000-0000-0000-0000-000000000002', 'Atlas Trust',      'json',        'https://api.atlastrust.synth/v1'),
  ('a0000000-0000-0000-0000-000000000003', 'Pioneer Clearing', 'fixed_width', '/feeds/pioneer')
ON CONFLICT (name) DO NOTHING;

-- Accounts
INSERT INTO accounts (id, external_id, custodian_id, name, account_type) VALUES
  ('b0000000-0000-0000-0000-000000000001', 'NR-100201', 'a0000000-0000-0000-0000-000000000001', 'Wellington Growth Fund',   'investment'),
  ('b0000000-0000-0000-0000-000000000002', 'AT-55432',  'a0000000-0000-0000-0000-000000000002', 'Meridian Income Strategy', 'investment'),
  ('b0000000-0000-0000-0000-000000000003', 'PC-8810',   'a0000000-0000-0000-0000-000000000003', 'Pinnacle Balanced Fund',   'investment')
ON CONFLICT (external_id, custodian_id) DO NOTHING;

-- Securities
INSERT INTO securities (id, symbol, cusip, isin, name, security_type) VALUES
  ('c0000000-0000-0000-0000-000000000001', 'AAPL',  '037833100', 'US0378331005', 'Apple Inc.',         'equity'),
  ('c0000000-0000-0000-0000-000000000002', 'MSFT',  '594918104', 'US5949181045', 'Microsoft Corp.',    'equity'),
  ('c0000000-0000-0000-0000-000000000003', 'VBTLX', '922908769', '',             'Vanguard Total Bond','mutual_fund'),
  ('c0000000-0000-0000-0000-000000000004', 'TSLA',  '88160R101', 'US88160R1014', 'Tesla Inc.',         'equity')
ON CONFLICT DO NOTHING;

-- Ingestion batches (one per custodian, completed)
INSERT INTO ingestion_batches (id, custodian_id, status, file_ref, record_count, completed_at) VALUES
  ('d0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', 'completed', 'northriver_20260328.csv',  4, now() - interval '2 hours'),
  ('d0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000002', 'completed', 'atlas_positions_20260328',  3, now() - interval '1 hour'),
  ('d0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000003', 'completed', 'pioneer_20260328.dat',      3, now() - interval '30 minutes')
ON CONFLICT DO NOTHING;

-- Positions (mix of custodian sources — these are what reconciliation compares)
INSERT INTO positions (id, account_id, security_id, quantity, market_value, as_of_date, source, batch_id) VALUES
  -- NorthRiver says 150 AAPL
  ('e0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', 150.00000000, 26250.0000, '2026-03-28', 'custodian', 'd0000000-0000-0000-0000-000000000001'),
  -- Internal book says 152 AAPL (mismatch!)
  ('e0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', 152.00000000, 26600.0000, '2026-03-28', 'internal', NULL),
  -- Atlas says 500 MSFT
  ('e0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000002', 'c0000000-0000-0000-0000-000000000002', 500.00000000, 210000.0000, '2026-03-28', 'custodian', 'd0000000-0000-0000-0000-000000000002'),
  -- Internal book agrees on 500 MSFT
  ('e0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000002', 'c0000000-0000-0000-0000-000000000002', 500.00000000, 210500.0000, '2026-03-28', 'internal', NULL),
  -- Pioneer says 1000 VBTLX
  ('e0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000003', 'c0000000-0000-0000-0000-000000000003', 1000.00000000, 10500.0000, '2026-03-28', 'custodian', 'd0000000-0000-0000-0000-000000000003'),
  -- Internal book says 1000 VBTLX
  ('e0000000-0000-0000-0000-000000000006', 'b0000000-0000-0000-0000-000000000003', 'c0000000-0000-0000-0000-000000000003', 1000.00000000, 10500.0000, '2026-03-28', 'internal', NULL)
ON CONFLICT DO NOTHING;

-- Transactions
INSERT INTO transactions (id, account_id, security_id, txn_type, quantity, price, amount, trade_date, settle_date, source, batch_id) VALUES
  ('f0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', 'buy',      50.00000000, 175.00000000, 8750.0000, '2026-03-25', '2026-03-27', 'custodian', 'd0000000-0000-0000-0000-000000000001'),
  ('f0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000004', 'sell',     25.00000000, 180.00000000, 4500.0000, '2026-03-26', '2026-03-28', 'custodian', 'd0000000-0000-0000-0000-000000000001'),
  -- Duplicate transaction from Atlas (same trade, posted twice)
  ('f0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000002', 'c0000000-0000-0000-0000-000000000002', 'buy',     100.00000000, 420.00000000, 42000.0000, '2026-03-27', '2026-03-28', 'custodian', 'd0000000-0000-0000-0000-000000000002'),
  ('f0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000002', 'c0000000-0000-0000-0000-000000000002', 'buy',     100.00000000, 420.00000000, 42000.0000, '2026-03-27', '2026-03-28', 'custodian', 'd0000000-0000-0000-0000-000000000002')
ON CONFLICT DO NOTHING;

-- Tax lots
INSERT INTO tax_lots (id, account_id, security_id, open_date, quantity, cost_basis, relief_method, source, batch_id) VALUES
  ('10000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', '2025-06-15', 100.00000000, 15000.0000, 'FIFO', 'custodian', 'd0000000-0000-0000-0000-000000000001'),
  ('10000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', '2026-03-25',  50.00000000,  8750.0000, 'FIFO', 'custodian', 'd0000000-0000-0000-0000-000000000001'),
  -- Pioneer reports cost basis as 10400 but internal says 10500 (mismatch)
  ('10000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000003', 'c0000000-0000-0000-0000-000000000003', '2025-01-10', 1000.00000000, 10400.0000, 'FIFO', 'custodian', 'd0000000-0000-0000-0000-000000000003')
ON CONFLICT DO NOTHING;

-- Recon runs
INSERT INTO recon_runs (id, entity_type, as_of_date, status, started_at, completed_at) VALUES
  ('20000000-0000-0000-0000-000000000001', 'position',    '2026-03-28', 'completed', now() - interval '25 minutes', now() - interval '24 minutes'),
  ('20000000-0000-0000-0000-000000000002', 'transaction', '2026-03-28', 'completed', now() - interval '20 minutes', now() - interval '19 minutes'),
  ('20000000-0000-0000-0000-000000000003', 'tax_lot',     '2026-03-28', 'completed', now() - interval '15 minutes', now() - interval '14 minutes')
ON CONFLICT DO NOTHING;

-- Recon results: positions
INSERT INTO recon_results (id, run_id, entity_type, entity_id, status, reason_code, details) VALUES
  -- AAPL: qty mismatch 150 vs 152
  ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'position', 'e0000000-0000-0000-0000-000000000001', 'BREAK', 'POS_QUANTITY_MISMATCH', 'custodian=150, internal=152, delta=2'),
  -- MSFT: price diff 210000 vs 210500 (quantities match)
  ('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000001', 'position', 'e0000000-0000-0000-0000-000000000003', 'NEAR_MATCH', 'POS_PRICE_SOURCE_DIFF', 'custodian_mv=210000, internal_mv=210500, delta=500'),
  -- VBTLX: perfect match
  ('30000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000001', 'position', 'e0000000-0000-0000-0000-000000000005', 'MATCH', NULL, '')
ON CONFLICT DO NOTHING;

-- Recon results: transactions
INSERT INTO recon_results (id, run_id, entity_type, entity_id, status, reason_code, details) VALUES
  -- Buy AAPL: matched
  ('30000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000002', 'transaction', 'f0000000-0000-0000-0000-000000000001', 'MATCH', NULL, ''),
  -- Sell TSLA: matched
  ('30000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000002', 'transaction', 'f0000000-0000-0000-0000-000000000002', 'MATCH', NULL, ''),
  -- Duplicate MSFT buy
  ('30000000-0000-0000-0000-000000000006', '20000000-0000-0000-0000-000000000002', 'transaction', 'f0000000-0000-0000-0000-000000000004', 'BREAK', 'TXN_DUPLICATE', 'duplicate of f0000000-0000-0000-0000-000000000003')
ON CONFLICT DO NOTHING;

-- Recon results: tax lots
INSERT INTO recon_results (id, run_id, entity_type, entity_id, status, reason_code, details) VALUES
  ('30000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000003', 'tax_lot', '10000000-0000-0000-0000-000000000001', 'MATCH', NULL, ''),
  ('30000000-0000-0000-0000-000000000008', '20000000-0000-0000-0000-000000000003', 'tax_lot', '10000000-0000-0000-0000-000000000002', 'MATCH', NULL, ''),
  -- Pioneer cost basis mismatch
  ('30000000-0000-0000-0000-000000000009', '20000000-0000-0000-0000-000000000003', 'tax_lot', '10000000-0000-0000-0000-000000000003', 'BREAK', 'LOT_COST_BASIS_MISMATCH', 'custodian=10400, internal=10500, delta=100')
ON CONFLICT DO NOTHING;

-- Exceptions (created from BREAKs)
INSERT INTO exceptions (id, recon_run_id, entity_type, entity_id, reason_code, status) VALUES
  ('40000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'position',    'e0000000-0000-0000-0000-000000000001', 'POS_QUANTITY_MISMATCH',   'open'),
  ('40000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', 'transaction', 'f0000000-0000-0000-0000-000000000004', 'TXN_DUPLICATE',           'open'),
  ('40000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000003', 'tax_lot',     '10000000-0000-0000-0000-000000000003', 'LOT_COST_BASIS_MISMATCH', 'acknowledged')
ON CONFLICT DO NOTHING;

-- Exception events (audit trail)
INSERT INTO exception_events (exception_id, event_type, actor, detail) VALUES
  ('40000000-0000-0000-0000-000000000001', 'created', 'system', 'generated from recon run 20000000-...-000000000001'),
  ('40000000-0000-0000-0000-000000000002', 'created', 'system', 'generated from recon run 20000000-...-000000000002'),
  ('40000000-0000-0000-0000-000000000003', 'created', 'system', 'generated from recon run 20000000-...-000000000003'),
  ('40000000-0000-0000-0000-000000000003', 'acknowledged', 'ops-analyst', 'reviewing Pioneer cost basis discrepancy')
ON CONFLICT DO NOTHING;

-- Lineage edges (source → target relationships for auditability)
-- Positions: entity → recon_result
INSERT INTO lineage_edges (source_type, source_id, target_type, target_id, relation) VALUES
  ('position', 'e0000000-0000-0000-0000-000000000001', 'recon_result', '30000000-0000-0000-0000-000000000001', 'reconciled_as'),
  ('position', 'e0000000-0000-0000-0000-000000000003', 'recon_result', '30000000-0000-0000-0000-000000000002', 'reconciled_as'),
  ('position', 'e0000000-0000-0000-0000-000000000005', 'recon_result', '30000000-0000-0000-0000-000000000003', 'reconciled_as')
ON CONFLICT DO NOTHING;

-- Transactions: entity → recon_result
INSERT INTO lineage_edges (source_type, source_id, target_type, target_id, relation) VALUES
  ('transaction', 'f0000000-0000-0000-0000-000000000001', 'recon_result', '30000000-0000-0000-0000-000000000004', 'reconciled_as'),
  ('transaction', 'f0000000-0000-0000-0000-000000000002', 'recon_result', '30000000-0000-0000-0000-000000000005', 'reconciled_as'),
  ('transaction', 'f0000000-0000-0000-0000-000000000004', 'recon_result', '30000000-0000-0000-0000-000000000006', 'reconciled_as')
ON CONFLICT DO NOTHING;

-- Recon result → exception (for BREAKs)
INSERT INTO lineage_edges (source_type, source_id, target_type, target_id, relation) VALUES
  ('recon_result', '30000000-0000-0000-0000-000000000001', 'exception', '40000000-0000-0000-0000-000000000001', 'raised_exception'),
  ('recon_result', '30000000-0000-0000-0000-000000000006', 'exception', '40000000-0000-0000-0000-000000000002', 'raised_exception'),
  ('recon_result', '30000000-0000-0000-0000-000000000009', 'exception', '40000000-0000-0000-0000-000000000003', 'raised_exception')
ON CONFLICT DO NOTHING;

COMMIT;
