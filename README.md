# TrustLot

Explainable wealth data trust layer.

TrustLot ingests synthetic multi-custodian data feeds, normalizes them into a canonical model, reconciles positions, transactions, and tax lots, assigns reason-coded exceptions, computes a Trust Score for automation readiness, and preserves lineage and replay for debugging and auditability.

This is back-office infrastructure software. All data is synthetic.

## Prerequisites

- Go 1.22+
- Node.js 20+
- Docker and Docker Compose
- PostgreSQL client tools (optional, for direct DB access)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI (optional, for `make migrate`)

## Quickstart

```bash
# Start Postgres and Redis
make dev-up

# Apply database migrations (option A: via golang-migrate)
make migrate

# Apply database migrations (option B: via psql through Docker)
for f in db/migrations/0000*.up.sql; do
  docker exec -i trustlot-postgres-1 psql -U trustlot -d trustlot < "$f"
done

# Seed synthetic data
docker exec -i trustlot-postgres-1 psql -U trustlot -d trustlot < db/seed.sql

# Start the API server (terminal 1)
make run-api

# Start the frontend dev server (terminal 2)
make web-install  # first time only
make web-dev
```

The API serves on `http://localhost:8080`. The frontend serves on `http://localhost:3000`.

### Verify it works

```bash
curl http://localhost:8080/healthz
# {"status":"ok","database":"connected"}

curl http://localhost:8080/api/v1/exceptions
# returns seeded exceptions

curl http://localhost:8080/api/v1/recon-runs
# returns completed recon runs with match/break counts

curl http://localhost:8080/api/v1/exceptions/{id}/explanation
# returns deterministic break explanation with field diffs and suggested actions

curl http://localhost:8080/api/v1/replay-cases
# returns captured replay cases
```

## Repository Layout

```
cmd/
  api/              API server entrypoint
  worker/           Background worker entrypoint
internal/
  config/           Configuration loading from environment
  db/               Connection pool + Store (query layer)
  model/            Domain types, reason codes, match statuses
  http/             HTTP server, routes, handlers, CORS middleware
  ingest/           Ingestion pipeline + CSV parser
  recon/            Reconciliation engine v1 (rules, pairing, persistence)
  explain/          Deterministic break explanation layer
  replay/           Replay harness v1 (capture, execute, compare)
  trust/            Trust Score computation interface
  exception/        Exception management interface
  lineage/          Lineage tracking interface
  telemetry/        Structured logging (slog), future OTel
  platform/         Shared platform utilities
apps/
  web/              Next.js operator console
db/
  migrations/       SQL migration files (5 files)
  seed.sql          Deterministic synthetic seed data
tools/
  synth/            Synthetic data generator CLI
docs/               Architecture and domain documentation
testdata/           Test fixtures
scripts/            Automation scripts
configs/            Configuration templates
deployments/
  docker/           Production Docker files (future)
```

## Current State

### Working

- Postgres connection with pooling (pgx v5)
- 5 SQL migrations covering 17 tables
- Deterministic seed data: 3 custodians, 3 accounts, 4 securities, positions, transactions, tax lots, 3 recon runs with results, 3 exceptions with audit events
- **Reconciliation engine v1**: deterministic rule-based comparison of internal vs custodian positions and transactions, with configurable tolerances
- **Explanation layer**: deterministic "Explain This Break" producing field diffs, evidence, and suggested actions per reason code
- **Replay harness v1**: capture replay cases from exceptions, re-execute against current rules, detect improvements/regressions
- API endpoints backed by real database queries:
  - `GET /healthz` — health check with DB connectivity status
  - `GET /api/v1/exceptions` — list with `?status=` filter and `?limit=`
  - `GET /api/v1/exceptions/{id}` — detail view
  - `GET /api/v1/exceptions/{id}/explanation` — deterministic break explanation
  - `GET /api/v1/recon-runs` — list with match/break counts
  - `GET /api/v1/replay-cases` — list captured replay cases
  - `GET /api/v1/replay-cases/{id}` — detail with snapshots
  - `POST /api/v1/replay-cases/from-exception/{id}` — capture case from exception
  - `POST /api/v1/replay-cases/{id}/run` — execute replay
  - `GET /api/v1/trust-scores` — stub
  - `GET /api/v1/lineage/{type}/{id}` — stub
- CORS middleware for frontend-to-API communication
- Next.js operator console with:
  - Exception list and detail with live explanation rendering
  - Replay case list and detail with run button and result display
  - Recon runs list with match/break counts
- CSV parser for NorthRiver-style feeds
- Tests: config, ingestion, HTTP handlers, DB integration, recon rules, explanation builders, replay execution

### Not Yet Implemented

- Ingestion pipeline execution (parsing + persisting)
- Trust Score computation
- Exception workflows (assignment, resolution)
- Lineage graph traversal
- Synthetic data generation (beyond seed.sql)
- Authentication
- OpenTelemetry SDK integration (using slog for now)
- Worker job processing

## Domain Flow

```
Custodian Feed -> Ingest -> Normalize -> Reconcile -> Score -> Exception -> Lineage
                                              |                     |
                                              v                     v
                                           Explain              Replay
```

## Database Schema

17 tables across 5 migrations:

| Migration | Tables |
|-----------|--------|
| 000001 | custodians, ingestion_batches, raw_records |
| 000002 | accounts, securities, positions, transactions, tax_lots |
| 000003 | recon_runs, recon_results, exceptions, trust_scores, lineage_edges, replay_cases |
| 000004 | custodian_accounts, security_identifiers, exception_events |
| 000005 | replay_cases extended with snapshot columns (ALTERs existing table) |

## Reconciliation Engine

The engine (`internal/recon`) compares internal vs custodian records deterministically:

- **Pairing**: records matched by account_id + security_id (positions) or account_id + security_id + txn_type + trade_date (transactions)
- **Rules**: ordered rule chain, stops at first non-MATCH
- **Position rules**: quantity mismatch, price/market value mismatch (configurable tolerance)
- **Transaction rules**: duplicate detection, amount mismatch
- **Outputs**: recon_results, exceptions with reason codes, exception_events

## Explanation Layer

The explanation layer (`internal/explain`) generates deterministic "Explain This Break" output:

- Field-by-field diffs (internal vs custodian with delta)
- Structured evidence items
- Suggested operator actions keyed to reason code
- No AI — all logic is rule-driven and grounded in source data

## Replay Harness

The replay harness (`internal/replay`) supports regression testing of reconciliation logic:

- **Capture**: snapshot internal + custodian records and config from a live exception
- **Execute**: deserialize snapshots, run current rules (never touches live DB)
- **Compare**: detect unchanged, improved, regressed, or changed outcomes
- **Status**: cases tracked as pending/passing/failing

## Seed Data

`db/seed.sql` contains deterministic synthetic data with fixed UUIDs:

- **3 custodians**: NorthRiver (CSV), Atlas Trust (JSON), Pioneer Clearing (fixed-width)
- **3 accounts**: Wellington Growth, Meridian Income, Pinnacle Balanced
- **4 securities**: AAPL, MSFT, VBTLX, TSLA
- **6 positions**: custodian vs internal pairs with intentional mismatches
- **4 transactions**: including a duplicate from Atlas
- **3 tax lots**: including a cost basis mismatch from Pioneer
- **3 recon runs**: position, transaction, tax_lot — all completed
- **9 recon results**: 4 matches, 2 near-matches/breaks per entity type
- **3 exceptions**: POS_QUANTITY_MISMATCH, TXN_DUPLICATE, LOT_COST_BASIS_MISMATCH
- **4 exception events**: creation + acknowledgment audit trail

## Synthetic Custodians

| Custodian | Format | Characteristics |
|-----------|--------|-----------------|
| NorthRiver | CSV/SFTP | Nightly batches, header drift, CUSIP identifiers |
| Atlas Trust | JSON API | Intermittent failures, reversals, ISIN identifiers |
| Pioneer Clearing | Fixed-width | Rounding quirks, delayed corporate actions |

## Make Targets

Run `make help` for all available targets.

| Target | Description |
|--------|-------------|
| `make dev-up` | Start Postgres + Redis |
| `make dev-down` | Stop infrastructure |
| `make build` | Build all Go binaries |
| `make run-api` | Run the API server |
| `make run-worker` | Run the background worker |
| `make migrate` | Apply migrations (needs golang-migrate) |
| `make seed` | Seed the database |
| `make db-reset` | Drop, re-migrate, re-seed |
| `make web-install` | Install frontend deps |
| `make web-dev` | Start frontend dev server |
| `make test` | Run all Go tests |
| `make lint` | Run go vet |
| `make fmt` | Format Go code |
| `make clean` | Remove build artifacts |

## Testing

```bash
# Run all tests (DB integration tests skip without DATABASE_URL)
make test

# Run with DB integration tests
DATABASE_URL="postgres://trustlot:trustlot@localhost:5432/trustlot?sslmode=disable" make test
```

Test coverage:
- `internal/config` — default values, env var overrides
- `internal/ingest` — CSV parser: valid input, empty input, header-only
- `internal/http` — health, exceptions, recon-runs endpoints, CORS headers
- `internal/model` — reason code completeness, match statuses
- `internal/db` — ping, list exceptions, status filter, list recon runs (integration, requires DB)
- `internal/recon` — position quantity/price rules, transaction duplicate/amount rules, rule ordering, deterministic sorting
- `internal/explain` — explanation builders for each reason code, missing data handling, unknown codes
- `internal/replay` — replay position/transaction cases, improvement detection, regression detection, missing snapshots, determinism verification
