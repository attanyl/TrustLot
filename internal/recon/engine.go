package recon

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/trustlot/trustlot/internal/db"
	"github.com/trustlot/trustlot/internal/model"
)

type engine struct {
	store  *db.Store
	config ReconConfig
}

// NewEngine creates a reconciliation engine backed by the given store.
func NewEngine(store *db.Store, config ReconConfig) Engine {
	return &engine{store: store, config: config}
}

func (e *engine) Run(ctx context.Context, runID string) error {
	run, err := e.store.GetReconRun(ctx, runID)
	if err != nil {
		return fmt.Errorf("load recon run: %w", err)
	}

	slog.Info("recon run started", "run_id", runID, "entity_type", run.EntityType, "as_of_date", run.AsOfDate.Format("2006-01-02"))

	var runErr error
	switch run.EntityType {
	case "position":
		runErr = e.reconcilePositions(ctx, run)
	case "transaction":
		runErr = e.reconcileTransactions(ctx, run)
	default:
		runErr = fmt.Errorf("unsupported entity type: %s", run.EntityType)
	}

	now := time.Now()
	if runErr != nil {
		_ = e.store.UpdateReconRunStatus(ctx, runID, "failed", &now)
		return runErr
	}

	if err := e.store.UpdateReconRunStatus(ctx, runID, "completed", &now); err != nil {
		return fmt.Errorf("mark run completed: %w", err)
	}

	slog.Info("recon run completed", "run_id", runID)
	return nil
}

// reconcilePositions pairs internal and custodian positions, applies rules, and persists results.
func (e *engine) reconcilePositions(ctx context.Context, run model.ReconRun) error {
	internal, err := e.store.ListPositionsBySource(ctx, "internal", run.AsOfDate)
	if err != nil {
		return fmt.Errorf("load internal positions: %w", err)
	}
	custodian, err := e.store.ListPositionsBySource(ctx, "custodian", run.AsOfDate)
	if err != nil {
		return fmt.Errorf("load custodian positions: %w", err)
	}

	sortPositions(internal)
	sortPositions(custodian)

	custodianMap := make(map[string]model.Position, len(custodian))
	for _, p := range custodian {
		key := positionKey(p)
		custodianMap[key] = p
	}

	matched := make(map[string]bool)
	var matchCount, breakCount int

	rules := PositionRules()

	for _, intPos := range internal {
		key := positionKey(intPos)
		custPos, found := custodianMap[key]
		if !found {
			// Unmatched internal position: BREAK with no specific rule hit.
			if err := e.persistBreak(ctx, run, intPos.ID, "position", model.ReasonMapAccountUnmapped, "no custodian position found"); err != nil {
				return err
			}
			breakCount++
			continue
		}
		matched[key] = true

		status, reasonCode, ruleName := ApplyRules(rules, RuleContext{
			Internal:  intPos,
			Custodian: custPos,
			Config:    e.config,
		})

		if err := e.persistResult(ctx, run, intPos.ID, "position", status, reasonCode, ruleName); err != nil {
			return err
		}

		if status == model.StatusBreak {
			breakCount++
		} else {
			matchCount++
		}
	}

	// Unmatched custodian positions.
	for _, custPos := range custodian {
		key := positionKey(custPos)
		if matched[key] {
			continue
		}
		if err := e.persistBreak(ctx, run, custPos.ID, "position", model.ReasonMapAccountUnmapped, "no internal position found"); err != nil {
			return err
		}
		breakCount++
	}

	slog.Info("position recon summary", "run_id", run.ID, "matches", matchCount, "breaks", breakCount)
	return nil
}

// reconcileTransactions pairs internal and custodian transactions, applies rules, and persists results.
func (e *engine) reconcileTransactions(ctx context.Context, run model.ReconRun) error {
	internal, err := e.store.ListTransactionsBySource(ctx, "internal", run.AsOfDate)
	if err != nil {
		return fmt.Errorf("load internal transactions: %w", err)
	}
	custodian, err := e.store.ListTransactionsBySource(ctx, "custodian", run.AsOfDate)
	if err != nil {
		return fmt.Errorf("load custodian transactions: %w", err)
	}

	sortTransactions(internal)
	sortTransactions(custodian)

	custodianMap := make(map[string]model.Transaction, len(custodian))
	for _, t := range custodian {
		key := transactionKey(t)
		custodianMap[key] = t
	}

	matched := make(map[string]bool)
	var matchCount, breakCount int

	rules := TransactionRules()

	for _, intTxn := range internal {
		key := transactionKey(intTxn)
		custTxn, found := custodianMap[key]
		if !found {
			if err := e.persistBreak(ctx, run, intTxn.ID, "transaction", model.ReasonMapAccountUnmapped, "no custodian transaction found"); err != nil {
				return err
			}
			breakCount++
			continue
		}
		matched[key] = true

		status, reasonCode, ruleName := ApplyRules(rules, RuleContext{
			Internal:  intTxn,
			Custodian: custTxn,
			Config:    e.config,
		})

		if err := e.persistResult(ctx, run, intTxn.ID, "transaction", status, reasonCode, ruleName); err != nil {
			return err
		}

		if status == model.StatusBreak {
			breakCount++
		} else {
			matchCount++
		}
	}

	for _, custTxn := range custodian {
		key := transactionKey(custTxn)
		if matched[key] {
			continue
		}
		if err := e.persistBreak(ctx, run, custTxn.ID, "transaction", model.ReasonMapAccountUnmapped, "no internal transaction found"); err != nil {
			return err
		}
		breakCount++
	}

	slog.Info("transaction recon summary", "run_id", run.ID, "matches", matchCount, "breaks", breakCount)
	return nil
}

// ApplyRules runs rules in order, stopping at the first non-MATCH.
// Returns the resulting status, an optional reason code, and the name of the
// rule that triggered (empty string if all rules matched).
func ApplyRules(rules []Rule, ctx RuleContext) (model.MatchStatus, *model.ReasonCode, string) {
	for _, rule := range rules {
		status, rc := rule.Evaluate(ctx)
		if status != model.StatusMatch {
			slog.Debug("rule triggered", "rule", rule.Name(), "status", status)
			return status, rc, rule.Name()
		}
	}
	return model.StatusMatch, nil, ""
}

// persistResult writes a recon_result and, if it's a break, an exception + event.
func (e *engine) persistResult(ctx context.Context, run model.ReconRun, entityID, entityType string, status model.MatchStatus, reasonCode *model.ReasonCode, details string) error {
	resultID := generateID()
	rc := model.ReasonCode("")
	if reasonCode != nil {
		rc = *reasonCode
	}

	result := model.ReconResult{
		ID:         resultID,
		RunID:      run.ID,
		EntityType: entityType,
		EntityID:   entityID,
		Status:     status,
		ReasonCode: rc,
		Details:    details,
	}
	if err := e.store.InsertReconResult(ctx, result); err != nil {
		return fmt.Errorf("persist recon result: %w", err)
	}

	if status == model.StatusBreak && reasonCode != nil {
		return e.createException(ctx, run, entityID, entityType, *reasonCode)
	}
	return nil
}

// persistBreak is a convenience for unmatched records.
func (e *engine) persistBreak(ctx context.Context, run model.ReconRun, entityID, entityType string, rc model.ReasonCode, details string) error {
	return e.persistResult(ctx, run, entityID, entityType, model.StatusBreak, &rc, details)
}

func (e *engine) createException(ctx context.Context, run model.ReconRun, entityID, entityType string, rc model.ReasonCode) error {
	now := time.Now()
	excID := generateID()

	exc := model.Exception{
		ID:         excID,
		ReconRunID: run.ID,
		EntityType: entityType,
		EntityID:   entityID,
		ReasonCode: rc,
		Status:     "open",
		AssignedTo: "",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := e.store.InsertException(ctx, exc); err != nil {
		return fmt.Errorf("create exception: %w", err)
	}

	if err := e.store.InsertExceptionEvent(ctx, excID, "created", "system", fmt.Sprintf(`{"reason_code":"%s"}`, rc)); err != nil {
		return fmt.Errorf("create exception event: %w", err)
	}
	return nil
}

// --- Pairing keys ---

func positionKey(p model.Position) string {
	return p.AccountID + "|" + p.SecurityID
}

func transactionKey(t model.Transaction) string {
	return t.AccountID + "|" + t.SecurityID + "|" + t.TxnType + "|" + t.TradeDate.Format("2006-01-02")
}

// --- Deterministic sorting ---

func sortPositions(ps []model.Position) {
	sort.Slice(ps, func(i, j int) bool {
		if ps[i].AccountID != ps[j].AccountID {
			return ps[i].AccountID < ps[j].AccountID
		}
		return ps[i].SecurityID < ps[j].SecurityID
	})
}

func sortTransactions(ts []model.Transaction) {
	sort.Slice(ts, func(i, j int) bool {
		if ts[i].AccountID != ts[j].AccountID {
			return ts[i].AccountID < ts[j].AccountID
		}
		if ts[i].SecurityID != ts[j].SecurityID {
			return ts[i].SecurityID < ts[j].SecurityID
		}
		if ts[i].TxnType != ts[j].TxnType {
			return ts[i].TxnType < ts[j].TxnType
		}
		return ts[i].TradeDate.Before(ts[j].TradeDate)
	})
}

// generateID produces a random hex ID suitable for use as a UUID-format string.
// In production this would use a proper UUID library; for MVP this avoids adding a dependency.
func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	// Format as UUID-style string.
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	)
}
