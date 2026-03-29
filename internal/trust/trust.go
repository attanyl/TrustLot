package trust

import "context"

// ScoreComponents holds the individual factors that compose a Trust Score.
type ScoreComponents struct {
	SourceReliability float64
	MappingConfidence float64
	ReconStatus       float64
	Freshness         float64
	Completeness      float64
	HumanReviewState  float64
	AnomalyPenalty    float64
}

// Score is a computed Trust Score for an entity.
type Score struct {
	EntityType string
	EntityID   string
	Total      float64
	Components ScoreComponents
}

// Computer calculates Trust Scores for entities.
type Computer interface {
	Compute(ctx context.Context, entityType, entityID string) (Score, error)
}
