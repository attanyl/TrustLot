# Reason Codes

Every reconciliation BREAK must map to a reason code. These codes categorize the root cause of the discrepancy and drive operator workflows.

17 reason codes are defined in `internal/model/reason_codes.go` as typed constants.

## Active in Reconciliation Engine

These reason codes are produced by the reconciliation engine v1 rules:

| Code | Category | Rule | Trigger |
|------|----------|------|---------|
| `POS_QUANTITY_MISMATCH` | Position | `positionQuantityRule` | Internal vs custodian quantity differs beyond tolerance |
| `POS_PRICE_SOURCE_DIFF` | Position | `positionPriceRule` | Market value differs beyond price tolerance (default 0.01) |
| `TXN_DUPLICATE` | Transaction | `transactionDuplicateRule` | Same security, quantity, trade date with different amounts |
| `TXN_AMOUNT_MISMATCH` | Transaction | `transactionAmountRule` | Transaction amount differs beyond tolerance |
| `MAP_ACCOUNT_UNMAPPED` | Mapping | (engine pairing) | No paired record found for internal or custodian record |

## Active in Seed Data

These reason codes appear in the current seed data:

| Code | Category | Seed Scenario |
|------|----------|---------------|
| `POS_QUANTITY_MISMATCH` | Position | NorthRiver AAPL: custodian=150, internal=152 |
| `POS_PRICE_SOURCE_DIFF` | Position | Atlas MSFT: market value differs by $500 (near-match) |
| `TXN_DUPLICATE` | Transaction | Atlas MSFT buy posted twice on same date |
| `LOT_COST_BASIS_MISMATCH` | Tax Lot | Pioneer VBTLX: custodian=10400, internal=10500 |

## Full Reference

### Ingestion

| Code | Description | Common Causes | Typical Resolution |
|------|-------------|---------------|-------------------|
| `ING_SCHEMA_DRIFT` | Custodian file schema changed unexpectedly | Header column added/removed, field type change | Update parser, add schema version |
| `ING_LATE_FEED` | Data arrived outside expected delivery window | Custodian delay, network issue | Wait for delivery, escalate if SLA breached |

### Mapping

| Code | Description | Common Causes | Typical Resolution |
|------|-------------|---------------|-------------------|
| `MAP_SECURITY_AMBIGUOUS` | Security identifier could not be uniquely resolved | Missing CUSIP/ISIN, conflicting identifiers across custodians | Manual mapping, add identifier rule |
| `MAP_ACCOUNT_UNMAPPED` | Account ID not found in account mapping table | New account, custodian-side ID change | Add account mapping entry |

### Transaction

| Code | Description | Common Causes | Typical Resolution |
|------|-------------|---------------|-------------------|
| `TXN_TRADE_SETTLE_WINDOW` | Trade date vs settlement date mismatch between sources | T+1 vs T+2 conventions, holiday calendars | Verify settlement convention, apply tolerance |
| `TXN_DUPLICATE` | Duplicate transaction detected | Custodian resend, batch overlap | Deduplicate, mark as suppressed |
| `TXN_CORRECTION_POSTED` | Corrective transaction posted by custodian | Reversal/replacement pair | Link correction chain, resolve original |
| `TXN_AMOUNT_MISMATCH` | Transaction amount differs between sources | Fee adjustment, rounding, partial fill | Review fee schedule, check lot math |

### Position

| Code | Description | Common Causes | Typical Resolution |
|------|-------------|---------------|-------------------|
| `POS_QUANTITY_MISMATCH` | Position quantity differs between sources | Pending trades, timing difference | Compare trade blotter, check settlement |
| `POS_PRICE_SOURCE_DIFF` | Market value differs due to pricing source | Different pricing vendors, stale prices | Verify pricing source hierarchy |

### Cash

| Code | Description | Common Causes | Typical Resolution |
|------|-------------|---------------|-------------------|
| `CASH_UNPOSTED_FEE` | Cash balance off due to unposted fee | Advisory fee, custody fee pending | Check fee schedule, wait for posting |

### Tax Lot

| Code | Description | Common Causes | Typical Resolution |
|------|-------------|---------------|-------------------|
| `LOT_COST_BASIS_MISMATCH` | Cost basis differs between sources | Wash sale adjustment, corporate action | Verify adjustment history |
| `LOT_RELIEF_METHOD_DIFF` | Tax lot relief method differs | FIFO vs HIFO disagreement | Confirm client election |
| `LOT_SHARE_ADJUSTMENT_CA` | Share count adjusted by corporate action | Stock split, merger | Process corporate action, recalculate |
| `LOT_WASH_SALE_ADJ` | Wash sale disallowed loss adjustment | IRS wash sale rule applied inconsistently | Review 30-day window, verify adjustment |

### Corporate Action

| Code | Description | Common Causes | Typical Resolution |
|------|-------------|---------------|-------------------|
| `CA_EVENT_MISSING` | Expected corporate action event not received | Delayed announcement, custodian processing lag | Check corporate action calendar |

### Operations

| Code | Description | Common Causes | Typical Resolution |
|------|-------------|---------------|-------------------|
| `OPS_MANUAL_OVERRIDE` | Operator manually overrode reconciliation result | Business judgment, known acceptable break | Document justification |

## Match Statuses

| Status | Meaning | Creates Exception? |
|--------|---------|-------------------|
| `MATCH` | Records agree within tolerance | No |
| `NEAR_MATCH` | Records are close but have a notable difference | Sometimes (configurable) |
| `BREAK` | Records disagree — requires investigation | Yes |

## Code

Reason codes are defined in `internal/model/reason_codes.go` as the `ReasonCode` string type with 17 constants. The `AllReasonCodes()` function returns the complete list and is covered by unit tests.
