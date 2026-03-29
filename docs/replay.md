# Replay

## Philosophy

Replay is a first-class feature. Every subtle reconciliation or parsing bug should become a replayable regression asset.

## Current State

The replay harness v1 is implemented in `internal/replay/`. It supports:

1. **Capture** a replay case from any exception (snapshots records + config)
2. **Execute** the case against current reconciliation rules
3. **Compare** actual vs expected outcome (unchanged / improved / regressed / changed)
4. **Track** case status as pending / passing / failing

## Workflow

When a reconciliation bug is discovered:

1. **Identify** the bad exception in the operator console
2. **Capture** the case via `POST /api/v1/replay-cases/from-exception/{id}`
3. **Fix** the reconciliation rule or parser
4. **Replay** the case via `POST /api/v1/replay-cases/{id}/run`
5. **Verify** the outcome improved (BREAK became MATCH) without regressions

## How It Works

### Capture

When a replay case is created from an exception:

- The exception, recon_result, and recon_run are loaded
- The internal and custodian records are found using the same pairing logic the engine uses
- Both records and the recon config are serialized as JSONB snapshots
- The expected match status and reason code are recorded

This snapshot approach means replay is independent of live database state. Records can be deleted or modified without affecting captured replay cases.

### Execution

When a replay case is run:

- Stored JSONB snapshots are deserialized into `model.Position` or `model.Transaction`
- The config snapshot is deserialized into `recon.ReconConfig`
- `recon.ApplyRules()` is called with `recon.PositionRules()` or `recon.TransactionRules()`
- The actual result is compared to the expected result

No live database queries are made during replay execution. The replay system reuses the exact same rule evaluation path as the reconciliation engine — it does not duplicate any rule logic.

### Outcome Classification

| Outcome | Meaning |
|---------|---------|
| Unchanged | Actual matches expected |
| Improved | Was BREAK, now MATCH (or less severe) |
| Regressed | Was MATCH, now BREAK |
| Changed | Different but same severity (e.g., different reason code, still BREAK) |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/replay-cases` | List all replay cases |
| `GET` | `/api/v1/replay-cases/{id}` | Case detail with snapshots |
| `POST` | `/api/v1/replay-cases/from-exception/{id}` | Capture case from exception |
| `POST` | `/api/v1/replay-cases/{id}/run` | Execute replay |

## Database Schema

The `replay_cases` table (migration 000005) stores:

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Primary key |
| `description` | TEXT | Auto-generated from exception context |
| `status` | TEXT | pending / passing / failing |
| `source_exception_id` | UUID | Reference to source exception |
| `entity_type` | TEXT | position / transaction |
| `expected_match_status` | TEXT | MATCH / NEAR_MATCH / BREAK |
| `expected_reason_code` | TEXT | Reason code at capture time |
| `internal_snapshot` | JSONB | Serialized internal record |
| `custodian_snapshot` | JSONB | Serialized custodian record |
| `config_snapshot` | JSONB | Serialized recon config |
| `created_at` | TIMESTAMPTZ | Capture timestamp |

## Determinism

Replay results are stable across runs given identical stored inputs and config. The replay path:

- Deserializes from JSONB (deterministic)
- Calls `recon.ApplyRules()` with fixed inputs (deterministic)
- Compares string values (deterministic)

No live DB state, no randomness, no time-dependent logic.

## Tests

`internal/replay/service_test.go` covers:

- Position replay matching original outcome
- Position replay detecting improvement (BREAK became MATCH)
- Transaction replay matching original outcome
- Transaction replay detecting regression (MATCH became BREAK)
- Graceful handling of missing snapshots
- Determinism verification (same inputs produce identical outputs across runs)
