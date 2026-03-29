package trust

import (
	"context"
	"fmt"
	"time"

	"github.com/trustlot/trustlot/internal/db"
)

type service struct {
	store *db.Store
}

// NewService creates a trust score service backed by the given store.
func NewService(store *db.Store) Service {
	return &service{store: store}
}

func (s *service) ComputeForReconResult(ctx context.Context, reconResultID string) (*TrustScoreResult, error) {
	rr, err := s.store.GetReconResult(ctx, reconResultID)
	if err != nil {
		return nil, fmt.Errorf("trust: load recon result: %w", err)
	}

	run, err := s.store.GetReconRun(ctx, rr.RunID)
	if err != nil {
		return nil, fmt.Errorf("trust: load recon run: %w", err)
	}

	exceptionStatus := s.lookupExceptionStatus(ctx, rr.RunID, rr.EntityType, rr.EntityID)
	source, asOfDate, complete := s.entitySignals(ctx, rr.EntityType, rr.EntityID, run.AsOfDate)

	in := reconInput{
		MatchStatus:     rr.Status,
		ReasonCode:      rr.ReasonCode,
		Source:          source,
		AsOfDate:        asOfDate,
		Now:             time.Now(),
		ExceptionStatus: exceptionStatus,
		HasAccountID:    complete.hasAccountID,
		HasSecurityID:   complete.hasSecurityID,
		HasQuantity:     complete.hasQuantity,
		HasAmount:       complete.hasAmount,
	}

	return s.computeAndPersist(ctx, rr.EntityType, rr.EntityID, in)
}

func (s *service) ComputeForException(ctx context.Context, exceptionID string) (*TrustScoreResult, error) {
	exc, err := s.store.GetException(ctx, exceptionID)
	if err != nil {
		return nil, fmt.Errorf("trust: load exception: %w", err)
	}

	run, err := s.store.GetReconRun(ctx, exc.ReconRunID)
	if err != nil {
		return nil, fmt.Errorf("trust: load recon run: %w", err)
	}

	rr, err := s.store.GetReconResultForException(ctx, exc.ReconRunID, exc.EntityType, exc.EntityID)
	if err != nil {
		return nil, fmt.Errorf("trust: load recon result for exception: %w", err)
	}

	source, asOfDate, complete := s.entitySignals(ctx, exc.EntityType, exc.EntityID, run.AsOfDate)

	in := reconInput{
		MatchStatus:     rr.Status,
		ReasonCode:      rr.ReasonCode,
		Source:          source,
		AsOfDate:        asOfDate,
		Now:             time.Now(),
		ExceptionStatus: exc.Status,
		HasAccountID:    complete.hasAccountID,
		HasSecurityID:   complete.hasSecurityID,
		HasQuantity:     complete.hasQuantity,
		HasAmount:       complete.hasAmount,
	}

	return s.computeAndPersist(ctx, exc.EntityType, exc.EntityID, in)
}

func (s *service) ListScores(ctx context.Context, limit int) ([]TrustScoreResult, error) {
	rows, err := s.store.ListTrustScores(ctx, limit)
	if err != nil {
		return nil, err
	}

	out := make([]TrustScoreResult, len(rows))
	for i, r := range rows {
		out[i] = rowToResult(r)
	}
	return out, nil
}

func (s *service) computeAndPersist(ctx context.Context, entityType, entityID string, in reconInput) (*TrustScoreResult, error) {
	components := computeComponents(in)
	total := computeTotal(components)
	rationale := buildRationale(total, components, in.MatchStatus, in.ExceptionStatus)
	now := time.Now()

	row := db.TrustScoreRow{
		EntityType:        entityType,
		EntityID:          entityID,
		Total:             total,
		SourceReliability: components.SourceReliability,
		MappingConfidence: 0.0, // not used in v1
		ReconStatus:       components.Reconciliation,
		Freshness:         components.Freshness,
		Completeness:      components.Completeness,
		HumanReviewState:  components.HumanReview,
		AnomalyPenalty:    components.AnomalyPenalty,
		RationaleSummary:  rationale,
		ComputedAt:        now,
	}

	id, err := s.store.UpsertTrustScore(ctx, row)
	if err != nil {
		return nil, fmt.Errorf("trust: persist score: %w", err)
	}

	result := &TrustScoreResult{
		ID:               id,
		EntityType:       entityType,
		EntityID:         entityID,
		Score:            total,
		Components:       components,
		RationaleSummary: rationale,
		ComputedAt:       now,
	}
	return result, nil
}

func (s *service) lookupExceptionStatus(ctx context.Context, runID, entityType, entityID string) string {
	exc, err := s.store.FindExceptionForEntity(ctx, runID, entityType, entityID)
	if err != nil {
		return ""
	}
	return exc.Status
}

type completenessSignals struct {
	hasAccountID  bool
	hasSecurityID bool
	hasQuantity   bool
	hasAmount     bool
}

func (s *service) entitySignals(ctx context.Context, entityType, entityID string, fallbackDate time.Time) (string, time.Time, completenessSignals) {
	switch entityType {
	case "position":
		p, err := s.store.GetPosition(ctx, entityID)
		if err == nil {
			return p.Source, p.AsOfDate, completenessSignals{
				hasAccountID:  p.AccountID != "",
				hasSecurityID: p.SecurityID != "",
				hasQuantity:   true,
				hasAmount:     p.MarketValue != 0,
			}
		}
	case "transaction":
		t, err := s.store.GetTransaction(ctx, entityID)
		if err == nil {
			return t.Source, t.TradeDate, completenessSignals{
				hasAccountID:  t.AccountID != "",
				hasSecurityID: t.SecurityID != "",
				hasQuantity:   true,
				hasAmount:     t.Amount != 0,
			}
		}
	}

	return "custodian", fallbackDate, completenessSignals{
		hasAccountID:  true,
		hasSecurityID: true,
		hasQuantity:   true,
		hasAmount:     true,
	}
}

func rowToResult(r db.TrustScoreRow) TrustScoreResult {
	return TrustScoreResult{
		ID:         r.ID,
		EntityType: r.EntityType,
		EntityID:   r.EntityID,
		Score:      r.Total,
		Components: ScoreComponents{
			SourceReliability: r.SourceReliability,
			Reconciliation:    r.ReconStatus,
			Freshness:         r.Freshness,
			Completeness:      r.Completeness,
			HumanReview:       r.HumanReviewState,
			AnomalyPenalty:    r.AnomalyPenalty,
		},
		RationaleSummary: r.RationaleSummary,
		ComputedAt:       r.ComputedAt,
	}
}
