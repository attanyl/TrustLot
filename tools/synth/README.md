# Synthetic Data Generator

Generates test data for TrustLot's three fictional custodians.

## Current State

The generator CLI (`tools/synth/main.go`) accepts a custodian name argument and prints a stub message. For Milestone 1, synthetic data is provided via `db/seed.sql` which inserts deterministic records with fixed UUIDs directly into the database.

The seed data covers:
- 3 custodians, 3 accounts, 4 securities
- 6 positions (custodian vs internal pairs with mismatches)
- 4 transactions (including a duplicate)
- 3 tax lots (including a cost basis mismatch)
- 3 recon runs with 9 results
- 3 exceptions with audit events

## Custodians

### NorthRiver
- **Format**: CSV / SFTP-style delivery
- **Parser**: Implemented in `internal/ingest/csv_parser.go`
- **Characteristics**: Nightly batch files, occasional header changes, pending transactions, CUSIP-heavy identifiers
- **Failure modes**: Schema drift, late files, duplicate transactions

### Atlas Trust
- **Format**: JSON API-style delivery
- **Characteristics**: Intermittent API failures, reversal/replacement corrections, ISIN-heavy identifiers
- **Failure modes**: Missing identifiers, correction chains, API timeouts

### Pioneer Clearing
- **Format**: Fixed-width file delivery
- **Characteristics**: Rounding quirks, delayed corporate action files, truncation issues
- **Failure modes**: Tax-lot mismatches, corporate-action-induced basis drift, trade/settle date mismatches

## Usage

```bash
# Build
make build-synth

# Run (currently prints stub message)
./bin/synth northriver
./bin/synth atlas
./bin/synth pioneer

# For actual test data, use the seed file
docker exec -i trustlot-postgres-1 psql -U trustlot -d trustlot < db/seed.sql
```

## Scenario Types (planned for Milestone 2)

The generator will be able to intentionally produce:
- Schema drift (column changes, field type changes)
- Missing identifiers (no CUSIP, ambiguous ISIN)
- Late files (delivery outside window)
- Duplicate transactions
- Correction chains (reversal + replacement)
- Tax-lot mismatches (cost basis disagreements)
- Corporate-action-induced basis drift
- Trade-date vs settlement-date mismatches

## Seed Data Conventions

- All UUIDs follow a pattern: `a0000000-...-00000000000N` for custodians, `b0...` for accounts, `c0...` for securities, etc.
- The seed is idempotent (`ON CONFLICT DO NOTHING`)
- Timestamps use relative offsets from `now()` so the data always looks recent
