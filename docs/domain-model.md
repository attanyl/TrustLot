# Domain Model

## Ingestion Layer

### Custodian
A data source that delivers portfolio feeds.
- Fields: ID (UUID), Name (unique), Format (csv/json/fixed_width), Endpoint, Active
- One custodian produces many ingestion batches
- Three synthetic custodians seeded: NorthRiver, Atlas Trust, Pioneer Clearing

### IngestionBatch
A single delivery of data from a custodian.
- Fields: ID, CustodianID, Status (pending/processing/completed/failed), FileRef, RecordCount, StartedAt, CompletedAt, CreatedAt
- One batch contains many raw records
- Status tracks the ingestion lifecycle

### RawRecord
An unparsed record from an ingestion batch.
- Fields: ID, BatchID, SeqNum, RawData (BYTEA), ParsedAt, Error
- Preserves the original byte-level data for replay and debugging
- Error field captures parse failures per-row

## Canonical Model

### Account
A portfolio account.
- Fields: ID, ExternalID, CustodianID, Name, AccountType, CreatedAt
- Uniquely identified by (ExternalID, CustodianID)
- Linked to custodians via foreign key

### CustodianAccount
Maps custodian-specific account references to canonical accounts.
- Fields: ID, AccountID, CustodianID, ExternalRef, Status (active/inactive/pending), CreatedAt
- Uniquely identified by (CustodianID, ExternalRef)
- Supports the case where one canonical account has representations at multiple custodians

### Security
A financial instrument.
- Fields: ID, Symbol, CUSIP, ISIN, SEDOL, Name, SecurityType, CreatedAt
- Identifier columns on this table are for convenience; the normalized identifier table is `security_identifiers`

### SecurityIdentifier
Normalized security identifier for cross-custodian matching.
- Fields: ID, SecurityID, IDType (cusip/isin/sedol/ticker/internal), IDValue, Source, CreatedAt
- Uniquely identified by (IDType, IDValue)
- Separating identifiers handles the common case where a security has a CUSIP from one custodian and an ISIN from another

### Position
A holding at a point in time.
- Fields: ID, AccountID, SecurityID, Quantity (NUMERIC 20,8), MarketValue (NUMERIC 20,4), AsOfDate, Source (custodian/internal), BatchID (nullable), CreatedAt
- Source indicates whether the position came from a custodian feed or was internally computed
- Reconciliation compares custodian positions against internal positions

### Transaction
A trade, dividend, transfer, or fee.
- Fields: ID, AccountID, SecurityID, TxnType, Quantity, Price, Amount, TradeDate, SettleDate, Source, BatchID, CreatedAt
- TradeDate vs SettleDate distinction matters for reconciliation (T+1/T+2 mismatches)

### TaxLot
A cost basis record for a specific purchase lot.
- Fields: ID, AccountID, SecurityID, OpenDate, Quantity, CostBasis, ReliefMethod (FIFO/LIFO/HIFO/SpecID), Source, BatchID, CreatedAt
- Cost basis mismatches between custodian and internal records are a common exception source

## Reconciliation Layer

### ReconRun
A single execution of the reconciliation engine.
- Fields: ID, EntityType (position/transaction/tax_lot), AsOfDate, Status (running/completed/failed), StartedAt, CompletedAt
- Each run targets one entity type on one as-of date
- Engine loads internal and custodian records for that date, pairs them, applies rules

### ReconResult
The outcome of reconciling a single entity.
- Fields: ID, RunID, EntityType, EntityID, Status (MATCH/NEAR_MATCH/BREAK), ReasonCode (nullable), Details, CreatedAt
- Every BREAK should have a reason code; MATCHes have null reason codes
- Details field stores the rule name that triggered (used by explanation layer)
- Partial index on reason_code for efficient break queries

### Exception
A reconciliation break requiring operator attention.
- Fields: ID, ReconRunID, EntityType, EntityID, ReasonCode, Status (open/acknowledged/resolved/suppressed), AssignedTo, CreatedAt, UpdatedAt
- Created automatically from BREAK recon results
- Status lifecycle: open -> acknowledged -> resolved (or suppressed)

### ExceptionEvent
Audit trail for exception lifecycle changes.
- Fields: ID, ExceptionID, EventType (created/acknowledged/assigned/resolved/suppressed/reopened/commented), Actor, Detail, CreatedAt
- Every state change is recorded for auditability

## Explanation Layer

### Explanation (in-memory, not persisted)
Generated on-demand by `internal/explain` from exception + recon_result + source records.
- Summary: deterministic natural-language sentence describing the break
- FieldDiffs: side-by-side internal vs custodian values with delta
- Evidence: labeled key-value pairs (account, security, date, rule, source)
- SuggestedActions: coded next steps keyed to reason code

## Trust Layer

### TrustScore
Measures automation readiness for an entity.
- Fields: ID, EntityType, EntityID, Total (NUMERIC 5,4), plus 7 component scores (each NUMERIC 5,4), ComputedAt
- Components: SourceReliability, MappingConfidence, ReconStatus, Freshness, Completeness, HumanReviewState, AnomalyPenalty
- All components stored as individual columns (queryable via SQL, not JSONB)
- Not yet computed — schema and model are in place

### LineageEdge
Connects two entities in the data lineage graph.
- Fields: ID, SourceType, SourceID, TargetType, TargetID, Relation, CreatedAt
- Generic graph structure: types are strings like "raw_record", "position", "recon_result"
- Indexed on both (source_type, source_id) and (target_type, target_id)

## Replay Layer

### ReplayCase
A captured scenario for regression testing.
- Fields: ID, Description, Status (pending/passing/failing), SourceExceptionID, EntityType, ExpectedMatchStatus, ExpectedReasonCode, InternalSnapshot (JSONB), CustodianSnapshot (JSONB), ConfigSnapshot (JSONB), CreatedAt
- Snapshots freeze the exact inputs at capture time so replay is independent of live DB state
- Status updated to passing/failing after each replay execution
- See docs/replay.md for the replay workflow

## Relationships

```
Custodian --1:N--> IngestionBatch --1:N--> RawRecord
Custodian --1:N--> Account
Custodian --1:N--> CustodianAccount --N:1--> Account
Security  --1:N--> SecurityIdentifier
Account   --1:N--> Position
Account   --1:N--> Transaction
Account   --1:N--> TaxLot
ReconRun  --1:N--> ReconResult
ReconRun  --1:N--> Exception --1:N--> ExceptionEvent
Exception --0:N--> ReplayCase (via source_exception_id)
```
