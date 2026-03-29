package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/trustlot/trustlot/internal/model"
)

// Store wraps a connection pool and provides domain query methods.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates a Store from an existing pool.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Ping verifies database connectivity.
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// ListExceptions returns exceptions ordered by creation time descending.
func (s *Store) ListExceptions(ctx context.Context, status string, limit int) ([]model.Exception, error) {
	query := `
		SELECT id, recon_run_id, entity_type, entity_id, reason_code,
		       status, assigned_to, created_at, updated_at
		FROM exceptions`

	args := []any{}
	argIdx := 1

	if status != "" {
		query += fmt.Sprintf(" WHERE status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list exceptions: %w", err)
	}
	defer rows.Close()

	var out []model.Exception
	for rows.Next() {
		var e model.Exception
		if err := rows.Scan(
			&e.ID, &e.ReconRunID, &e.EntityType, &e.EntityID, &e.ReasonCode,
			&e.Status, &e.AssignedTo, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan exception: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetException returns a single exception by ID.
func (s *Store) GetException(ctx context.Context, id string) (model.Exception, error) {
	var e model.Exception
	err := s.pool.QueryRow(ctx, `
		SELECT id, recon_run_id, entity_type, entity_id, reason_code,
		       status, assigned_to, created_at, updated_at
		FROM exceptions WHERE id = $1`, id,
	).Scan(
		&e.ID, &e.ReconRunID, &e.EntityType, &e.EntityID, &e.ReasonCode,
		&e.Status, &e.AssignedTo, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return e, fmt.Errorf("exception not found: %s", id)
	}
	if err != nil {
		return e, fmt.Errorf("get exception: %w", err)
	}
	return e, nil
}

// ReconRunRow includes summary counts alongside the run metadata.
type ReconRunRow struct {
	model.ReconRun
	MatchCount int
	BreakCount int
}

// GetReconRun returns a single recon run by ID.
func (s *Store) GetReconRun(ctx context.Context, id string) (model.ReconRun, error) {
	var r model.ReconRun
	err := s.pool.QueryRow(ctx, `
		SELECT id, entity_type, as_of_date, status, started_at, completed_at
		FROM recon_runs WHERE id = $1`, id,
	).Scan(&r.ID, &r.EntityType, &r.AsOfDate, &r.Status, &r.StartedAt, &r.CompletedAt)
	if err != nil {
		return r, fmt.Errorf("get recon run: %w", err)
	}
	return r, nil
}

// UpdateReconRunStatus sets the status and completed_at timestamp of a run.
func (s *Store) UpdateReconRunStatus(ctx context.Context, id, status string, completedAt *time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE recon_runs SET status = $2, completed_at = $3 WHERE id = $1`,
		id, status, completedAt,
	)
	if err != nil {
		return fmt.Errorf("update recon run status: %w", err)
	}
	return nil
}

// ListPositionsBySource returns positions for a given source and date.
func (s *Store) ListPositionsBySource(ctx context.Context, source string, asOfDate time.Time) ([]model.Position, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, account_id, security_id, quantity, market_value, as_of_date, source, batch_id
		FROM positions
		WHERE source = $1 AND as_of_date = $2
		ORDER BY account_id, security_id`, source, asOfDate,
	)
	if err != nil {
		return nil, fmt.Errorf("list positions by source: %w", err)
	}
	defer rows.Close()

	var out []model.Position
	for rows.Next() {
		var p model.Position
		if err := rows.Scan(&p.ID, &p.AccountID, &p.SecurityID, &p.Quantity, &p.MarketValue, &p.AsOfDate, &p.Source, &p.BatchID); err != nil {
			return nil, fmt.Errorf("scan position: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListTransactionsBySource returns transactions for a given source and trade date.
func (s *Store) ListTransactionsBySource(ctx context.Context, source string, tradeDate time.Time) ([]model.Transaction, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, account_id, security_id, txn_type, quantity, price, amount, trade_date, settle_date, source, batch_id
		FROM transactions
		WHERE source = $1 AND trade_date = $2
		ORDER BY account_id, security_id, txn_type`, source, tradeDate,
	)
	if err != nil {
		return nil, fmt.Errorf("list transactions by source: %w", err)
	}
	defer rows.Close()

	var out []model.Transaction
	for rows.Next() {
		var t model.Transaction
		if err := rows.Scan(&t.ID, &t.AccountID, &t.SecurityID, &t.TxnType, &t.Quantity, &t.Price, &t.Amount, &t.TradeDate, &t.SettleDate, &t.Source, &t.BatchID); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// InsertReconResult persists a single reconciliation result.
func (s *Store) InsertReconResult(ctx context.Context, r model.ReconResult) error {
	var reasonCode *string
	if r.ReasonCode != "" {
		rc := string(r.ReasonCode)
		reasonCode = &rc
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO recon_results (id, run_id, entity_type, entity_id, status, reason_code, details)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		r.ID, r.RunID, r.EntityType, r.EntityID, string(r.Status), reasonCode, r.Details,
	)
	if err != nil {
		return fmt.Errorf("insert recon result: %w", err)
	}
	return nil
}

// InsertException persists a new exception.
func (s *Store) InsertException(ctx context.Context, e model.Exception) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO exceptions (id, recon_run_id, entity_type, entity_id, reason_code, status, assigned_to, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		e.ID, e.ReconRunID, e.EntityType, e.EntityID, string(e.ReasonCode), e.Status, e.AssignedTo, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert exception: %w", err)
	}
	return nil
}

// InsertExceptionEvent persists an exception lifecycle event.
func (s *Store) InsertExceptionEvent(ctx context.Context, exceptionID, eventType, actor, detail string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO exception_events (exception_id, event_type, actor, detail)
		VALUES ($1, $2, $3, $4)`,
		exceptionID, eventType, actor, detail,
	)
	if err != nil {
		return fmt.Errorf("insert exception event: %w", err)
	}
	return nil
}

// GetReconResultForException finds the recon_result that corresponds to an exception.
func (s *Store) GetReconResultForException(ctx context.Context, runID, entityType, entityID string) (model.ReconResult, error) {
	var r model.ReconResult
	var reasonCode *string
	err := s.pool.QueryRow(ctx, `
		SELECT id, run_id, entity_type, entity_id, status, reason_code, details
		FROM recon_results
		WHERE run_id = $1 AND entity_type = $2 AND entity_id = $3
		LIMIT 1`, runID, entityType, entityID,
	).Scan(&r.ID, &r.RunID, &r.EntityType, &r.EntityID, &r.Status, &reasonCode, &r.Details)
	if err != nil {
		return r, fmt.Errorf("get recon result for exception: %w", err)
	}
	if reasonCode != nil {
		r.ReasonCode = model.ReasonCode(*reasonCode)
	}
	return r, nil
}

// GetPosition returns a single position by ID.
func (s *Store) GetPosition(ctx context.Context, id string) (model.Position, error) {
	var p model.Position
	err := s.pool.QueryRow(ctx, `
		SELECT id, account_id, security_id, quantity, market_value, as_of_date, source, batch_id
		FROM positions WHERE id = $1`, id,
	).Scan(&p.ID, &p.AccountID, &p.SecurityID, &p.Quantity, &p.MarketValue, &p.AsOfDate, &p.Source, &p.BatchID)
	if err != nil {
		return p, fmt.Errorf("get position: %w", err)
	}
	return p, nil
}

// GetTransaction returns a single transaction by ID.
func (s *Store) GetTransaction(ctx context.Context, id string) (model.Transaction, error) {
	var t model.Transaction
	err := s.pool.QueryRow(ctx, `
		SELECT id, account_id, security_id, txn_type, quantity, price, amount, trade_date, settle_date, source, batch_id
		FROM transactions WHERE id = $1`, id,
	).Scan(&t.ID, &t.AccountID, &t.SecurityID, &t.TxnType, &t.Quantity, &t.Price, &t.Amount, &t.TradeDate, &t.SettleDate, &t.Source, &t.BatchID)
	if err != nil {
		return t, fmt.Errorf("get transaction: %w", err)
	}
	return t, nil
}

// FindPairedPosition finds a position by account, security, date, and source.
func (s *Store) FindPairedPosition(ctx context.Context, accountID, securityID string, asOfDate time.Time, source string) (model.Position, error) {
	var p model.Position
	err := s.pool.QueryRow(ctx, `
		SELECT id, account_id, security_id, quantity, market_value, as_of_date, source, batch_id
		FROM positions
		WHERE account_id = $1 AND security_id = $2 AND as_of_date = $3 AND source = $4
		LIMIT 1`, accountID, securityID, asOfDate, source,
	).Scan(&p.ID, &p.AccountID, &p.SecurityID, &p.Quantity, &p.MarketValue, &p.AsOfDate, &p.Source, &p.BatchID)
	if err != nil {
		return p, fmt.Errorf("find paired position: %w", err)
	}
	return p, nil
}

// FindPairedTransaction finds a transaction by account, security, type, trade date, and source.
func (s *Store) FindPairedTransaction(ctx context.Context, accountID, securityID, txnType string, tradeDate time.Time, source string) (model.Transaction, error) {
	var t model.Transaction
	err := s.pool.QueryRow(ctx, `
		SELECT id, account_id, security_id, txn_type, quantity, price, amount, trade_date, settle_date, source, batch_id
		FROM transactions
		WHERE account_id = $1 AND security_id = $2 AND txn_type = $3 AND trade_date = $4 AND source = $5
		LIMIT 1`, accountID, securityID, txnType, tradeDate, source,
	).Scan(&t.ID, &t.AccountID, &t.SecurityID, &t.TxnType, &t.Quantity, &t.Price, &t.Amount, &t.TradeDate, &t.SettleDate, &t.Source, &t.BatchID)
	if err != nil {
		return t, fmt.Errorf("find paired transaction: %w", err)
	}
	return t, nil
}

// ListReconRuns returns recon runs with match/break counts.
func (s *Store) ListReconRuns(ctx context.Context, limit int) ([]ReconRunRow, error) {
	query := `
		SELECT
			r.id, r.entity_type, r.as_of_date, r.status, r.started_at, r.completed_at,
			COUNT(*) FILTER (WHERE rr.status = 'MATCH')    AS match_count,
			COUNT(*) FILTER (WHERE rr.status = 'BREAK')    AS break_count
		FROM recon_runs r
		LEFT JOIN recon_results rr ON rr.run_id = r.id
		GROUP BY r.id
		ORDER BY r.started_at DESC`

	args := []any{}
	if limit > 0 {
		query += " LIMIT $1"
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list recon runs: %w", err)
	}
	defer rows.Close()

	var out []ReconRunRow
	for rows.Next() {
		var row ReconRunRow
		var completedAt *time.Time
		if err := rows.Scan(
			&row.ID, &row.EntityType, &row.AsOfDate, &row.Status,
			&row.StartedAt, &completedAt,
			&row.MatchCount, &row.BreakCount,
		); err != nil {
			return nil, fmt.Errorf("scan recon run: %w", err)
		}
		row.CompletedAt = completedAt
		out = append(out, row)
	}
	return out, rows.Err()
}

// ReplayCaseRow holds all columns for the extended replay_cases table.
type ReplayCaseRow struct {
	ID                  string
	Description         string
	Status              string
	SourceExceptionID   *string
	EntityType          string
	ExpectedMatchStatus string
	ExpectedReasonCode  string
	InternalSnapshot    []byte
	CustodianSnapshot   []byte
	ConfigSnapshot      []byte
	CreatedAt           time.Time
}

// InsertReplayCase persists a new replay case with snapshot data.
func (s *Store) InsertReplayCase(ctx context.Context, rc ReplayCaseRow) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO replay_cases (id, description, status, source_exception_id, entity_type,
			expected_match_status, expected_reason_code, internal_snapshot, custodian_snapshot,
			config_snapshot, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		rc.ID, rc.Description, rc.Status, rc.SourceExceptionID, rc.EntityType,
		rc.ExpectedMatchStatus, rc.ExpectedReasonCode, rc.InternalSnapshot,
		rc.CustodianSnapshot, rc.ConfigSnapshot, rc.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert replay case: %w", err)
	}
	return nil
}

// GetReplayCase returns a single replay case by ID.
func (s *Store) GetReplayCase(ctx context.Context, id string) (ReplayCaseRow, error) {
	var rc ReplayCaseRow
	err := s.pool.QueryRow(ctx, `
		SELECT id, description, status, source_exception_id, entity_type,
		       expected_match_status, expected_reason_code, internal_snapshot,
		       custodian_snapshot, config_snapshot, created_at
		FROM replay_cases WHERE id = $1`, id,
	).Scan(&rc.ID, &rc.Description, &rc.Status, &rc.SourceExceptionID, &rc.EntityType,
		&rc.ExpectedMatchStatus, &rc.ExpectedReasonCode, &rc.InternalSnapshot,
		&rc.CustodianSnapshot, &rc.ConfigSnapshot, &rc.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return rc, fmt.Errorf("replay case not found: %s", id)
	}
	if err != nil {
		return rc, fmt.Errorf("get replay case: %w", err)
	}
	return rc, nil
}

// ListReplayCases returns all replay cases ordered by creation time descending.
func (s *Store) ListReplayCases(ctx context.Context) ([]ReplayCaseRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, description, status, source_exception_id, entity_type,
		       expected_match_status, expected_reason_code, internal_snapshot,
		       custodian_snapshot, config_snapshot, created_at
		FROM replay_cases
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list replay cases: %w", err)
	}
	defer rows.Close()

	var out []ReplayCaseRow
	for rows.Next() {
		var rc ReplayCaseRow
		if err := rows.Scan(&rc.ID, &rc.Description, &rc.Status, &rc.SourceExceptionID,
			&rc.EntityType, &rc.ExpectedMatchStatus, &rc.ExpectedReasonCode,
			&rc.InternalSnapshot, &rc.CustodianSnapshot, &rc.ConfigSnapshot,
			&rc.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan replay case: %w", err)
		}
		out = append(out, rc)
	}
	return out, rows.Err()
}

// UpdateReplayCaseStatus sets the status of a replay case.
func (s *Store) UpdateReplayCaseStatus(ctx context.Context, id, status string) error {
	_, err := s.pool.Exec(ctx, `UPDATE replay_cases SET status = $2 WHERE id = $1`, id, status)
	if err != nil {
		return fmt.Errorf("update replay case status: %w", err)
	}
	return nil
}

// GetReconResult returns a single recon result by ID.
func (s *Store) GetReconResult(ctx context.Context, id string) (model.ReconResult, error) {
	var r model.ReconResult
	var reasonCode *string
	err := s.pool.QueryRow(ctx, `
		SELECT id, run_id, entity_type, entity_id, status, reason_code, details
		FROM recon_results WHERE id = $1`, id,
	).Scan(&r.ID, &r.RunID, &r.EntityType, &r.EntityID, &r.Status, &reasonCode, &r.Details)
	if err == pgx.ErrNoRows {
		return r, fmt.Errorf("recon result not found: %s", id)
	}
	if err != nil {
		return r, fmt.Errorf("get recon result: %w", err)
	}
	if reasonCode != nil {
		r.ReasonCode = model.ReasonCode(*reasonCode)
	}
	return r, nil
}

// FindExceptionForEntity finds an exception by recon run, entity type, and entity ID.
func (s *Store) FindExceptionForEntity(ctx context.Context, runID, entityType, entityID string) (model.Exception, error) {
	var e model.Exception
	err := s.pool.QueryRow(ctx, `
		SELECT id, recon_run_id, entity_type, entity_id, reason_code,
		       status, assigned_to, created_at, updated_at
		FROM exceptions
		WHERE recon_run_id = $1 AND entity_type = $2 AND entity_id = $3
		ORDER BY updated_at DESC
		LIMIT 1`, runID, entityType, entityID,
	).Scan(
		&e.ID, &e.ReconRunID, &e.EntityType, &e.EntityID, &e.ReasonCode,
		&e.Status, &e.AssignedTo, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return e, fmt.Errorf("find exception for entity: %w", err)
	}
	return e, nil
}

// TrustScoreRow holds all columns for a persisted trust score.
type TrustScoreRow struct {
	ID                string
	EntityType        string
	EntityID          string
	Total             float64
	SourceReliability float64
	MappingConfidence float64
	ReconStatus       float64
	Freshness         float64
	Completeness      float64
	HumanReviewState  float64
	AnomalyPenalty    float64
	RationaleSummary  string
	ComputedAt        time.Time
}

// UpsertTrustScore inserts or replaces a trust score for a given entity.
// Returns the ID of the persisted row.
func (s *Store) UpsertTrustScore(ctx context.Context, ts TrustScoreRow) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `
		INSERT INTO trust_scores (entity_type, entity_id, total,
			source_reliability, mapping_confidence, recon_status,
			freshness, completeness, human_review_state, anomaly_penalty,
			rationale_summary, computed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT ON CONSTRAINT uq_trust_scores_entity
		DO UPDATE SET
			total = EXCLUDED.total,
			source_reliability = EXCLUDED.source_reliability,
			mapping_confidence = EXCLUDED.mapping_confidence,
			recon_status = EXCLUDED.recon_status,
			freshness = EXCLUDED.freshness,
			completeness = EXCLUDED.completeness,
			human_review_state = EXCLUDED.human_review_state,
			anomaly_penalty = EXCLUDED.anomaly_penalty,
			rationale_summary = EXCLUDED.rationale_summary,
			computed_at = EXCLUDED.computed_at
		RETURNING id`,
		ts.EntityType, ts.EntityID, ts.Total,
		ts.SourceReliability, ts.MappingConfidence, ts.ReconStatus,
		ts.Freshness, ts.Completeness, ts.HumanReviewState, ts.AnomalyPenalty,
		ts.RationaleSummary, ts.ComputedAt,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("upsert trust score: %w", err)
	}
	return id, nil
}

// GetTrustScoreForEntity returns the trust score for a specific entity.
func (s *Store) GetTrustScoreForEntity(ctx context.Context, entityType, entityID string) (TrustScoreRow, error) {
	var ts TrustScoreRow
	err := s.pool.QueryRow(ctx, `
		SELECT id, entity_type, entity_id, total,
			source_reliability, mapping_confidence, recon_status,
			freshness, completeness, human_review_state, anomaly_penalty,
			rationale_summary, computed_at
		FROM trust_scores
		WHERE entity_type = $1 AND entity_id = $2`, entityType, entityID,
	).Scan(
		&ts.ID, &ts.EntityType, &ts.EntityID, &ts.Total,
		&ts.SourceReliability, &ts.MappingConfidence, &ts.ReconStatus,
		&ts.Freshness, &ts.Completeness, &ts.HumanReviewState, &ts.AnomalyPenalty,
		&ts.RationaleSummary, &ts.ComputedAt,
	)
	if err != nil {
		return ts, fmt.Errorf("get trust score for entity: %w", err)
	}
	return ts, nil
}

// ListRawRecordsByBatch returns raw records belonging to a batch.
func (s *Store) ListRawRecordsByBatch(ctx context.Context, batchID string) ([]model.RawRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, batch_id, seq_num, raw_data, parsed_at, error
		FROM raw_records
		WHERE batch_id = $1
		ORDER BY seq_num`, batchID,
	)
	if err != nil {
		return nil, fmt.Errorf("list raw records by batch: %w", err)
	}
	defer rows.Close()

	var out []model.RawRecord
	for rows.Next() {
		var r model.RawRecord
		if err := rows.Scan(&r.ID, &r.BatchID, &r.SeqNum, &r.RawData, &r.ParsedAt, &r.Error); err != nil {
			return nil, fmt.Errorf("scan raw record: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetIngestionBatch returns a single ingestion batch by ID.
func (s *Store) GetIngestionBatch(ctx context.Context, id string) (model.IngestionBatch, error) {
	var b model.IngestionBatch
	err := s.pool.QueryRow(ctx, `
		SELECT id, custodian_id, status, file_ref, record_count, started_at, completed_at, created_at
		FROM ingestion_batches WHERE id = $1`, id,
	).Scan(&b.ID, &b.CustodianID, &b.Status, &b.FileRef, &b.RecordCount, &b.StartedAt, &b.CompletedAt, &b.CreatedAt)
	if err != nil {
		return b, fmt.Errorf("get ingestion batch: %w", err)
	}
	return b, nil
}

// ListReconResultsForEntity returns all recon results for a given entity.
func (s *Store) ListReconResultsForEntity(ctx context.Context, entityType, entityID string) ([]model.ReconResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, run_id, entity_type, entity_id, status, reason_code, details
		FROM recon_results
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC`, entityType, entityID,
	)
	if err != nil {
		return nil, fmt.Errorf("list recon results for entity: %w", err)
	}
	defer rows.Close()

	var out []model.ReconResult
	for rows.Next() {
		var r model.ReconResult
		var reasonCode *string
		if err := rows.Scan(&r.ID, &r.RunID, &r.EntityType, &r.EntityID, &r.Status, &reasonCode, &r.Details); err != nil {
			return nil, fmt.Errorf("scan recon result: %w", err)
		}
		if reasonCode != nil {
			r.ReasonCode = model.ReasonCode(*reasonCode)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListExceptionsForEntity returns all exceptions for a given entity.
func (s *Store) ListExceptionsForEntity(ctx context.Context, entityType, entityID string) ([]model.Exception, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, recon_run_id, entity_type, entity_id, reason_code,
		       status, assigned_to, created_at, updated_at
		FROM exceptions
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC`, entityType, entityID,
	)
	if err != nil {
		return nil, fmt.Errorf("list exceptions for entity: %w", err)
	}
	defer rows.Close()

	var out []model.Exception
	for rows.Next() {
		var e model.Exception
		if err := rows.Scan(
			&e.ID, &e.ReconRunID, &e.EntityType, &e.EntityID, &e.ReasonCode,
			&e.Status, &e.AssignedTo, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan exception: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// InsertLineageEdge persists a lineage edge.
func (s *Store) InsertLineageEdge(ctx context.Context, sourceType, sourceID, targetType, targetID, relation string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO lineage_edges (source_type, source_id, target_type, target_id, relation)
		VALUES ($1, $2, $3, $4, $5)`,
		sourceType, sourceID, targetType, targetID, relation,
	)
	if err != nil {
		return fmt.Errorf("insert lineage edge: %w", err)
	}
	return nil
}

// ListTrustScores returns trust scores ordered by computed_at descending.
func (s *Store) ListTrustScores(ctx context.Context, limit int) ([]TrustScoreRow, error) {
	query := `
		SELECT id, entity_type, entity_id, total,
			source_reliability, mapping_confidence, recon_status,
			freshness, completeness, human_review_state, anomaly_penalty,
			rationale_summary, computed_at
		FROM trust_scores
		ORDER BY computed_at DESC`

	args := []any{}
	if limit > 0 {
		query += " LIMIT $1"
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list trust scores: %w", err)
	}
	defer rows.Close()

	var out []TrustScoreRow
	for rows.Next() {
		var ts TrustScoreRow
		if err := rows.Scan(
			&ts.ID, &ts.EntityType, &ts.EntityID, &ts.Total,
			&ts.SourceReliability, &ts.MappingConfidence, &ts.ReconStatus,
			&ts.Freshness, &ts.Completeness, &ts.HumanReviewState, &ts.AnomalyPenalty,
			&ts.RationaleSummary, &ts.ComputedAt,
		); err != nil {
			return nil, fmt.Errorf("scan trust score: %w", err)
		}
		out = append(out, ts)
	}
	return out, rows.Err()
}
