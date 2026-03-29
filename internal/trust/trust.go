package trust

import (
	"context"
	"time"
)

// ScoreComponents holds the individual factors that compose a Trust Score.
// Each component is a raw subscore in the range [0, 1] before weighting.
type ScoreComponents struct {
	SourceReliability float64 `json:"source_reliability"`
	Reconciliation    float64 `json:"reconciliation"`
	Freshness         float64 `json:"freshness"`
	Completeness      float64 `json:"completeness"`
	HumanReview       float64 `json:"human_review"`
	AnomalyPenalty    float64 `json:"anomaly_penalty"`
}

// TrustScoreResult is the full output of a trust score computation.
type TrustScoreResult struct {
	ID               string          `json:"id"`
	EntityType       string          `json:"entity_type"`
	EntityID         string          `json:"entity_id"`
	Score            float64         `json:"score"`
	Components       ScoreComponents `json:"components"`
	RationaleSummary string          `json:"rationale_summary"`
	ComputedAt       time.Time       `json:"computed_at"`
}

// Service computes, persists, and retrieves trust scores.
type Service interface {
	ComputeForReconResult(ctx context.Context, reconResultID string) (*TrustScoreResult, error)
	ComputeForException(ctx context.Context, exceptionID string) (*TrustScoreResult, error)
	ListScores(ctx context.Context, limit int) ([]TrustScoreResult, error)
}
