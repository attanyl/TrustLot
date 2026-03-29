# Architecture

## System Context

TrustLot sits between synthetic custodian data feeds and operator workflows. It does not trade, move money, or interact with real custodians.

```
Custodian Feeds (synthetic)
        |
    [ Ingest ]
        |
    [ Normalize ]
        |
    [ Reconcile ] ---> [ Explain ]
       / \
      /   \
[ Trust ] [ Exception ]
      \   /
       \ /
    [ Lineage ]
        |
    [ Replay ]
```

## Components

### API Server (`cmd/api`)
HTTP service using chi router. Connects to Postgres on startup via pgx connection pool. Degrades gracefully if the database is unavailable — endpoints return empty data instead of crashing.

Current endpoints:
- `GET /healthz` — returns status + database connectivity
- `GET /api/v1/exceptions` — list exceptions with optional `?status=` filter and `?limit=`
- `GET /api/v1/exceptions/{id}` — single exception detail
- `GET /api/v1/exceptions/{id}/explanation` — deterministic break explanation with field diffs, evidence, and suggested actions
- `GET /api/v1/recon-runs` — list runs with match/break counts (joined from recon_results)
- `GET /api/v1/replay-cases` — list captured replay cases
- `GET /api/v1/replay-cases/{id}` — replay case detail with snapshots
- `POST /api/v1/replay-cases/from-exception/{id}` — capture replay case from exception
- `POST /api/v1/replay-cases/{id}/run` — execute replay and return result
- `GET /api/v1/trust-scores` — stub (returns empty)
- `GET /api/v1/lineage/{entityType}/{entityID}` — stub (returns empty)

### Worker (`cmd/worker`)
Background process that will execute ingestion batches, reconciliation runs, trust score computation, and replay cases. Currently a placeholder that blocks on SIGINT/SIGTERM.

### Reconciliation Engine (`internal/recon`)
Deterministic rule-based engine that compares internal vs custodian records.

- **Pairing**: records matched by account_id + security_id (positions) or account_id + security_id + txn_type + trade_date (transactions)
- **Rules**: ordered rule chain with `Rule` interface. Engine stops at first non-MATCH. Rules are composable and testable in isolation.
- **Position rules**: quantity mismatch (`POS_QUANTITY_MISMATCH`), market value mismatch with configurable tolerance (`POS_PRICE_SOURCE_DIFF`)
- **Transaction rules**: duplicate detection (`TXN_DUPLICATE`), amount mismatch (`TXN_AMOUNT_MISMATCH`)
- **Persistence**: writes recon_results, exceptions, and exception_events
- **Determinism**: inputs sorted before processing, no randomness in rule evaluation

### Explanation Layer (`internal/explain`)
Generates deterministic "Explain This Break" output for any exception. No AI — all logic is reason-code-driven.

- Loads exception, recon_result, and reconstructs the internal/custodian record pair
- Dispatches to reason-code-specific builders that produce: summary, field diffs, evidence items, suggested actions
- Graceful degradation when records are partially missing

### Replay Harness (`internal/replay`)
Supports regression testing of reconciliation logic by capturing and replaying cases.

- **Capture**: snapshots internal + custodian records and config from a live exception into JSONB columns
- **Execute**: deserializes stored snapshots, calls the same `recon.ApplyRules()` path (never touches live DB during replay)
- **Compare**: classifies outcome as unchanged, improved, regressed, or changed
- **Status tracking**: cases marked as pending/passing/failing after execution

### Operator Console (`apps/web`)
Next.js 14 App Router application. Fetches data from the API server via `src/lib/api.ts`. Pages:

- Dashboard — open exceptions, total exceptions, recon run count
- Health — pings `/healthz`, shows API and DB status
- Exceptions — table with reason codes, status badges, links to detail
- Exception detail — all fields plus explanation (summary, field diffs, evidence, suggested actions)
- Recon Runs — table with match/break counts
- Replay Cases — list with status badges, detail with snapshots and run button
- Trust Scores, Lineage — stub pages

### Synthetic Generator (`tools/synth`)
CLI tool for generating test data from three fictional custodians. Currently prints stub messages. Actual data generation uses `db/seed.sql` patterns.

## Data Layer

### Store Pattern
`internal/db/store.go` wraps `pgxpool.Pool` with domain-specific query methods. Handlers receive a `*Store` (which may be nil for testing without a database). No ORM — all queries are hand-written SQL.

### Migration Strategy
5 sequential migration files, split by dependency chain:
1. Ingestion layer (custodians, batches, raw_records)
2. Canonical model (accounts, securities, positions, transactions, tax_lots)
3. Reconciliation and trust (recon_runs, recon_results, exceptions, trust_scores, lineage_edges, replay_cases)
4. Milestone 1 additions (custodian_accounts, security_identifiers, exception_events)
5. Replay harness (ALTERs replay_cases with snapshot JSONB columns)

Each migration has an up and down file. Schema uses UUID primary keys, NUMERIC for financial precision, and CHECK constraints for enum-like columns.

### Seed Data
`db/seed.sql` provides deterministic synthetic data with fixed UUIDs. It is idempotent (uses `ON CONFLICT DO NOTHING`) and can be re-run after truncation. The seed covers all three custodians with intentional data quality issues for reconciliation to catch.

## Data Flow

1. **Ingest**: Raw custodian files arrive (CSV, JSON, fixed-width). The CSV parser (`internal/ingest/csv_parser.go`) is implemented for NorthRiver-style feeds. Parsed into `raw_records` with batch tracking.
2. **Normalize**: Raw records are mapped to canonical `positions`, `transactions`, and `tax_lots`. Security and account identifiers are resolved via `security_identifiers` and `custodian_accounts` tables.
3. **Reconcile**: Canonical records are compared across sources via ordered rule chains. Each comparison yields a MATCH, NEAR_MATCH, or BREAK with a reason code. Results stored in `recon_results`.
4. **Explain**: BREAKs get deterministic explanations with field diffs, evidence, and suggested operator actions.
5. **Score**: Trust Scores are computed from seven weighted components. Every score includes a component breakdown. (Not yet implemented.)
6. **Exception**: BREAKs generate exceptions routed to an operator inbox. Exception lifecycle is tracked via `exception_events` audit trail.
7. **Lineage**: Every transformation is recorded as a directed edge in the lineage graph via `lineage_edges`.
8. **Replay**: Captured failure cases can be re-run against updated logic for regression testing. Snapshots ensure replay is independent of live DB state.

## Technology Choices

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Backend language | Go | Performance, type safety, simple deployment |
| HTTP router | chi v5 | Lightweight, idiomatic, stdlib-compatible |
| Database | PostgreSQL 16 | Relational integrity, NUMERIC precision, UUID support |
| DB driver | pgx v5 | Native Go driver with connection pooling |
| Frontend | Next.js 14 + TypeScript | App Router, server components, Tailwind CSS |
| Queue (future) | Redis or NATS | TBD — Redis container already in docker-compose |
| Observability | slog (now), OpenTelemetry (future) | Structured JSON logging, trace/metric providers later |
| Containerization | Docker Compose | Local development only — Postgres and Redis |

## Key Constraints

- All data is synthetic. No real custodian integrations.
- Single go.mod monorepo. No microservices.
- REST API. No gRPC until genuinely needed.
- No ORM. SQL queries and migrations are explicit.
- No auth in MVP. Stubs only.
- API gracefully degrades without database — returns empty responses, not errors.
- Reconciliation and replay are deterministic — same inputs always produce same outputs.
