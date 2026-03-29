package explain

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/trustlot/trustlot/internal/db"
	"github.com/trustlot/trustlot/internal/model"
)

// Service generates deterministic explanations for exceptions.
type Service interface {
	ExplainException(ctx context.Context, exceptionID string) (*Explanation, error)
}

type service struct {
	store *db.Store
}

// NewService creates an explanation service backed by the given store.
func NewService(store *db.Store) Service {
	return &service{store: store}
}

func (s *service) ExplainException(ctx context.Context, exceptionID string) (*Explanation, error) {
	exc, err := s.store.GetException(ctx, exceptionID)
	if err != nil {
		return nil, fmt.Errorf("load exception: %w", err)
	}

	result, err := s.store.GetReconResultForException(ctx, exc.ReconRunID, exc.EntityType, exc.EntityID)
	if err != nil {
		slog.Warn("explain: recon result not found, generating partial explanation",
			"exception_id", exceptionID, "error", err)
		expl := BuildExplanation(exc, model.ReconResult{}, nil, nil)
		return &expl, nil
	}

	run, err := s.store.GetReconRun(ctx, exc.ReconRunID)
	if err != nil {
		slog.Warn("explain: recon run not found", "exception_id", exceptionID, "error", err)
		expl := BuildExplanation(exc, result, nil, nil)
		return &expl, nil
	}

	internal, custodian := s.loadRecordPair(ctx, exc, run)

	expl := BuildExplanation(exc, result, internal, custodian)
	return &expl, nil
}

// loadRecordPair loads the internal entity by ID, then finds its custodian counterpart
// using the same deterministic pairing logic the engine uses.
func (s *service) loadRecordPair(ctx context.Context, exc model.Exception, run model.ReconRun) (internal, custodian any) {
	switch exc.EntityType {
	case "position":
		return s.loadPositionPair(ctx, exc.EntityID, run)
	case "transaction":
		return s.loadTransactionPair(ctx, exc.EntityID, run)
	default:
		return nil, nil
	}
}

func (s *service) loadPositionPair(ctx context.Context, entityID string, run model.ReconRun) (any, any) {
	intPos, err := s.store.GetPosition(ctx, entityID)
	if err != nil {
		slog.Debug("explain: internal position not found", "entity_id", entityID, "error", err)
		// Try loading as custodian-side record (unmatched custodian).
		custPos, err2 := s.store.GetPosition(ctx, entityID)
		if err2 != nil {
			return nil, nil
		}
		return nil, custPos
	}

	// Find the paired custodian position.
	oppositeSource := "custodian"
	if intPos.Source == "custodian" {
		oppositeSource = "internal"
	}

	paired, err := s.store.FindPairedPosition(ctx, intPos.AccountID, intPos.SecurityID, run.AsOfDate, oppositeSource)
	if err != nil {
		slog.Debug("explain: paired position not found", "account_id", intPos.AccountID, "security_id", intPos.SecurityID, "error", err)
		if intPos.Source == "internal" {
			return intPos, nil
		}
		return nil, intPos
	}

	if intPos.Source == "internal" {
		return intPos, paired
	}
	return paired, intPos
}

func (s *service) loadTransactionPair(ctx context.Context, entityID string, run model.ReconRun) (any, any) {
	intTxn, err := s.store.GetTransaction(ctx, entityID)
	if err != nil {
		slog.Debug("explain: internal transaction not found", "entity_id", entityID, "error", err)
		return nil, nil
	}

	oppositeSource := "custodian"
	if intTxn.Source == "custodian" {
		oppositeSource = "internal"
	}

	paired, err := s.store.FindPairedTransaction(ctx, intTxn.AccountID, intTxn.SecurityID, intTxn.TxnType, intTxn.TradeDate, oppositeSource)
	if err != nil {
		slog.Debug("explain: paired transaction not found", "account_id", intTxn.AccountID, "security_id", intTxn.SecurityID, "error", err)
		if intTxn.Source == "internal" {
			return intTxn, nil
		}
		return nil, intTxn
	}

	if intTxn.Source == "internal" {
		return intTxn, paired
	}
	return paired, intTxn
}
