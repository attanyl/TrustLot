# CLAUDE.md

## Project: TrustLot

TrustLot is an explainable wealth data trust layer.

It ingests synthetic multi-custodian data feeds, normalizes them into a canonical model, reconciles positions / transactions / tax lots, assigns reason-coded exceptions, computes a Trust Score for automation readiness, and preserves lineage and replay for debugging and auditability.

This is not a consumer app and not a flashy dashboard project. It is deliberately deep back-office / platform work.

---

## Core product thesis

The hard problem is not "can we parse custodian files?"
The hard problem is:

> Can we determine, explain, and prove whether a portfolio record is safe to automate on?

TrustLot exists to answer that question.

A record is only useful if we can:
- ingest it reliably
- normalize it consistently
- reconcile it deterministically
- explain discrepancies clearly
- replay historical failures
- measure trust explicitly

---

## What this repo should optimize for

1. Deterministic correctness over cleverness
2. Explainability over magic
3. Replayability over one-off fixes
4. Clear architecture over premature abstraction
5. Dirty operational realism over demo polish
6. Synthetic data only — never real client or custodian data
7. Strong local development experience
8. Testability at every layer
9. Lineage and auditability as first-class concerns
10. Narrow AI usage only at the edges

---

## Product scope

### MVP scope
Build a working local platform that can:

- ingest 3 synthetic custodians:
  - CSV / SFTP-style
  - JSON API-style
  - fixed-width file-style
- normalize data into a canonical model
- reconcile:
  - positions
  - transactions
  - tax lots
- generate exceptions with reason codes
- compute a Trust Score
- show operator workflows in a web UI
- preserve lineage from raw record to canonical record to recon result
- replay historical ingestion/reconciliation runs against newer parser / rule versions

### Explicit non-goals
Do not build:
- real custodian integrations
- real client onboarding
- real trading automation
- real money movement
- production auth / enterprise SSO
- a full portfolio accounting platform
- a generic chatbot
- investor-facing UX

---

## Architecture constraints

### High-level architecture
Use a monorepo.

Recommended top-level components:
- backend API/service layer
- ingestion workers
- reconciliation engine
- synthetic data generator
- frontend operator console
- shared schemas/types
- infrastructure/dev tooling

### Preferred stack
- Backend: Go
- Frontend: Next.js + TypeScript
- Database: PostgreSQL
- Caching / queue: Redis or NATS (choose one and stay consistent)
- Observability: OpenTelemetry + Prometheus-compatible metrics
- Containerization: Docker Compose
- Migrations: SQL migration files
- API style: REST first for speed, optionally structured for future gRPC

If there is a conflict between speed and architectural purity, choose the path that preserves correctness and maintainability.

---

## Domain model expectations

The codebase should model, at minimum, these concepts:

### Source / ingestion
- Custodian
- IngestionBatch
- RawRecord
- SchemaVersion
- DriftEvent

### Core canonical model
- Account
- CustodianAccount
- Security
- SecurityIdentifier
- Transaction
- Position
- TaxLot
- CorporateAction (at least stubbed if not fully used in MVP)

### Reconciliation
- ReconRun
- ReconResult
- MatchStatus
- ReasonCode
- Exception
- ExceptionEvent

### Trust / explainability
- TrustScore
- LineageEdge
- ReplayCase
- ReplayRun

---

## Trust Score principles

Trust Score is not a vague AI confidence metric.

It should represent:

> automation readiness for a record given current evidence

At minimum, the score should be composed from:
- source reliability
- mapping confidence
- reconciliation status
- freshness
- completeness
- human review state
- anomaly penalty

Every Trust Score must be explainable with component breakdowns.

Do not produce opaque scalar scores without reasons.

---

## Reconciliation principles

Reconciliation must be deterministic and rule-driven.

Support these match categories:
- MATCH
- NEAR_MATCH
- BREAK

Every BREAK must try to map to a reason code.

Starter reason codes include:
- ING_SCHEMA_DRIFT
- ING_LATE_FEED
- MAP_SECURITY_AMBIGUOUS
- MAP_ACCOUNT_UNMAPPED
- TXN_TRADE_SETTLE_WINDOW
- TXN_DUPLICATE
- TXN_CORRECTION_POSTED
- POS_QUANTITY_MISMATCH
- POS_PRICE_SOURCE_DIFF
- CASH_UNPOSTED_FEE
- LOT_COST_BASIS_MISMATCH
- LOT_RELIEF_METHOD_DIFF
- LOT_SHARE_ADJUSTMENT_CA
- LOT_WASH_SALE_ADJ
- CA_EVENT_MISSING
- OPS_MANUAL_OVERRIDE

Rules should be implemented in a way that is easy to extend, test, and reason about.

---

## Synthetic custodians

The project should include 3 fictional custodians with different failure modes.

### 1. NorthRiver
CSV / SFTP-style delivery.
Characteristics:
- nightly batch files
- occasional header changes
- some pending transactions
- CUSIP-heavy identifiers

### 2. Atlas Trust
JSON API-style delivery.
Characteristics:
- intermittent API failures
- reversal/replacement corrections
- ISIN-heavy identifiers

### 3. Pioneer Clearing
Fixed-width file delivery.
Characteristics:
- rounding quirks
- delayed corporate action files
- truncation issues

The synthetic generator should be able to intentionally produce:
- schema drift
- missing identifiers
- late files
- duplicate transactions
- correction chains
- tax-lot mismatches
- corporate-action-induced basis drift
- trade-date vs settlement-date mismatches

---

## Frontend expectations

The UI is an operator console, not a marketing website.

Must include:
- daily health view
- exception inbox
- exception detail view
- trust score breakdown
- lineage view
- replay view
- synthetic scenario explorer

UI priorities:
- dense, useful tables
- filters and drilldowns
- crisp diffs
- evidence visibility
- low visual fluff

---

## Replay philosophy

Replay is a first-class feature.

When a parsing or reconciliation bug is discovered:
1. capture the synthetic failing case
2. persist it as a replay fixture
3. write a failing test
4. fix the bug
5. replay old runs against the new version
6. ensure regression coverage exists

Never fix a subtle issue without turning it into a replayable asset.

---

## Observability requirements

Every major subsystem must emit:
- logs
- metrics
- traces

At minimum, instrument:
- ingestion batch lifecycle
- parsing failures
- schema drift detection
- normalization writes
- recon runs
- exception creation
- trust score computation
- replay runs

Useful metrics include:
- batch success rate
- parse failure counts
- match rate by entity type
- exceptions by reason code
- exception aging
- trust score distribution
- replay pass/fail rate

---

## Testing standards

Tests matter a lot in this repo.

Required testing layers:
- unit tests for parsing and reconciliation rules
- integration tests for DB-backed flows
- golden fixture tests for synthetic custodian files
- replay regression tests
- deterministic seed-based scenario tests

Important invariants:
- matched position qty should align with lot sum unless intentionally broken
- canonical entities must preserve source lineage
- Trust Score output must include explainable components
- every BREAK should have a reason code or explicit fallback reason

---

## Coding standards

### General
- Prefer clear names over short names
- Keep functions small and single-purpose
- Avoid framework overengineering
- Avoid hidden global state
- Make errors explicit and structured
- Keep core logic deterministic

### Go
- Use idiomatic Go
- Prefer small interfaces
- Keep domain logic separate from transport and storage layers
- Avoid premature generics unless genuinely helpful
- Return rich errors with machine-readable context where useful

### TypeScript / frontend
- Prefer strict typing
- Keep API contracts explicit
- Separate data fetching from presentation
- Build useful operator components, not generic design-system theater

### SQL / schema
- Favor explicit migrations
- Index for reconciliation query paths
- Preserve raw + normalized + derived lineage relationships
- Use JSONB only where flexibility is genuinely useful

---

## AI usage rules

AI is allowed only in narrow, grounded places.

Good uses:
- incident summarization from structured evidence
- synthetic document extraction experiments
- developer productivity and scaffolding assistance

Bad uses:
- deciding reconciliation truth
- deciding final reason codes without deterministic rules
- replacing explainable logic with LLM guesses
- hiding missing logic behind “AI”

Any AI-generated output shown in product features must be grounded in structured evidence.

---

## Compliance / professionalism rules

Even though this is a side project with synthetic data:
- never imply affiliation with real custodians
- never use production data
- never copy internal Vestmark code, schemas, naming, or logic
- never represent this as a real trading or accounting system
- clearly label all sample data as synthetic

This project should signal professionalism, not recklessness.

---

## Naming and tone

This is infrastructure software.

Write docs and code comments like you are building:
- a control plane
- a reliability layer
- a reconciliation engine
- an exception management system

Do not write in startup cliché language.
Do not oversell.
Be concrete.

---

## Directory design expectations

The initial repo should likely include directories resembling:

- `/apps/web`
- `/cmd/api`
- `/cmd/worker`
- `/internal/...`
- `/pkg/...` only if genuinely needed
- `/db/migrations`
- `/configs`
- `/deployments/docker`
- `/tools/synth`
- `/testdata`
- `/docs`
- `/scripts`

Do not create meaningless folders just to look enterprise.

---

## Initial milestones

### Milestone 1
Repo scaffolding, local dev, database, migrations, health endpoint, web shell

### Milestone 2
Synthetic generator + one custodian ingestion path

### Milestone 3
Canonical model persistence + lineage

### Milestone 4
Reconciliation engine v1 for positions and transactions

### Milestone 5
Tax-lot reconciliation + reason codes

### Milestone 6
Exception inbox + detail UI

### Milestone 7
Trust Score + replay harness

### Milestone 8
Observability + demo polish

---

## What good output from Claude looks like

When modifying this repo, Claude should:
- preserve domain clarity
- keep the architecture coherent
- write real code, not pseudocode
- create migrations and tests alongside features
- explain tradeoffs briefly and concretely
- avoid giant speculative rewrites
- prefer incremental, reviewable changes

When uncertain, Claude should choose the more deterministic and testable design.

---

## Immediate build priority

The first job is not to fully implement TrustLot.

The first job is to generate a clean, serious skeleton that supports:
- local development
- clear bounded contexts
- migrations
- synthetic data tooling
- future reconciliation engine work
- frontend operator console scaffolding

That foundation matters more than surface polish.