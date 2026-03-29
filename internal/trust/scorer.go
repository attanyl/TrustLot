package trust

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/trustlot/trustlot/internal/model"
)

// Weights for the v1 scoring model. These are explicit and deterministic.
// They sum to 1.0 (anomaly penalty is subtracted, not added).
var weights = struct {
	SourceReliability float64
	Reconciliation    float64
	Freshness         float64
	Completeness      float64
	HumanReview       float64
	AnomalyPenalty    float64
}{
	SourceReliability: 0.15,
	Reconciliation:    0.35,
	Freshness:         0.15,
	Completeness:      0.10,
	HumanReview:       0.15,
	AnomalyPenalty:    0.10,
}

// reconInput holds the data needed to compute a trust score.
type reconInput struct {
	MatchStatus model.MatchStatus
	ReasonCode  model.ReasonCode
	Source      string // "internal", "custodian"
	AsOfDate    time.Time
	Now         time.Time

	// Exception state (empty string if no exception exists).
	ExceptionStatus string

	// Entity completeness signals.
	HasAccountID  bool
	HasSecurityID bool
	HasQuantity   bool
	HasAmount     bool
}

// computeComponents produces raw subscores from structured input.
func computeComponents(in reconInput) ScoreComponents {
	return ScoreComponents{
		SourceReliability: sourceReliabilityScore(in.Source),
		Reconciliation:    reconciliationScore(in.MatchStatus),
		Freshness:         freshnessScore(in.AsOfDate, in.Now),
		Completeness:      completenessScore(in),
		HumanReview:       humanReviewScore(in.ExceptionStatus),
		AnomalyPenalty:    anomalyPenaltyScore(in.MatchStatus, in.ReasonCode),
	}
}

// computeTotal applies weights to components and returns a score in [0, 100].
func computeTotal(c ScoreComponents) float64 {
	raw := c.SourceReliability*weights.SourceReliability +
		c.Reconciliation*weights.Reconciliation +
		c.Freshness*weights.Freshness +
		c.Completeness*weights.Completeness +
		c.HumanReview*weights.HumanReview -
		c.AnomalyPenalty*weights.AnomalyPenalty

	score := raw * 100.0
	return math.Round(clamp(score, 0, 100)*100) / 100
}

// sourceReliabilityScore: internal data is more trusted than custodian data.
//
//	internal  → 0.95
//	custodian → 0.85
//	unknown   → 0.70
func sourceReliabilityScore(source string) float64 {
	switch source {
	case "internal":
		return 0.95
	case "custodian":
		return 0.85
	default:
		return 0.70
	}
}

// reconciliationScore: direct mapping from match status.
//
//	MATCH      → 1.0
//	NEAR_MATCH → 0.7
//	BREAK      → 0.2
func reconciliationScore(status model.MatchStatus) float64 {
	switch status {
	case model.StatusMatch:
		return 1.0
	case model.StatusNearMatch:
		return 0.7
	case model.StatusBreak:
		return 0.2
	default:
		return 0.0
	}
}

// freshnessScore: how recent the data is relative to now.
//
//	same day            → 1.0
//	1 day stale         → 0.7
//	2-7 days stale      → 0.3
//	older than 7 days   → 0.1
func freshnessScore(asOfDate time.Time, now time.Time) float64 {
	days := int(now.Sub(asOfDate).Hours() / 24)
	switch {
	case days <= 0:
		return 1.0
	case days == 1:
		return 0.7
	case days <= 7:
		return 0.3
	default:
		return 0.1
	}
}

// completenessScore: proportion of required fields present.
func completenessScore(in reconInput) float64 {
	fields := []bool{in.HasAccountID, in.HasSecurityID, in.HasQuantity, in.HasAmount}
	present := 0
	for _, f := range fields {
		if f {
			present++
		}
	}
	return float64(present) / float64(len(fields))
}

// humanReviewScore: based on exception lifecycle state.
//
//	no exception / no review needed → 0.5
//	open exception                  → 0.0
//	acknowledged                    → 0.4
//	resolved / suppressed           → 1.0
func humanReviewScore(exceptionStatus string) float64 {
	switch exceptionStatus {
	case "":
		return 0.5
	case "open":
		return 0.0
	case "acknowledged":
		return 0.4
	case "resolved", "suppressed":
		return 1.0
	default:
		return 0.5
	}
}

// anomalyPenaltyScore: simple deterministic penalty for known anomaly patterns.
//
//	no break                → 0.0
//	BREAK + duplicate       → 0.4
//	BREAK + quantity/amount → 0.3
//	BREAK (other)           → 0.2
func anomalyPenaltyScore(status model.MatchStatus, reason model.ReasonCode) float64 {
	if status != model.StatusBreak {
		return 0.0
	}
	switch reason {
	case model.ReasonTxnDuplicate:
		return 0.4
	case model.ReasonPosQuantityMismatch, model.ReasonTxnAmountMismatch:
		return 0.3
	default:
		return 0.2
	}
}

// buildRationale generates a deterministic human-readable summary from components.
func buildRationale(score float64, c ScoreComponents, matchStatus model.MatchStatus, exceptionStatus string) string {
	var level string
	switch {
	case score >= 75:
		level = "High trust"
	case score >= 40:
		level = "Medium trust"
	default:
		level = "Low trust"
	}

	var parts []string

	switch matchStatus {
	case model.StatusMatch:
		parts = append(parts, "matched reconciliation result")
	case model.StatusNearMatch:
		parts = append(parts, "near-match reconciliation result")
	case model.StatusBreak:
		parts = append(parts, "reconciliation break")
	}

	if c.Freshness >= 0.9 {
		parts = append(parts, "same-day data")
	} else if c.Freshness <= 0.3 {
		parts = append(parts, "stale data")
	}

	if c.Completeness >= 1.0 {
		parts = append(parts, "complete record fields")
	} else if c.Completeness < 0.75 {
		parts = append(parts, "missing record fields")
	}

	switch exceptionStatus {
	case "open":
		parts = append(parts, "open exception pending review")
	case "acknowledged":
		parts = append(parts, "exception acknowledged but not resolved")
	case "resolved":
		parts = append(parts, "exception resolved")
	}

	if c.AnomalyPenalty >= 0.3 {
		parts = append(parts, "significant anomaly penalty applied")
	}

	detail := strings.Join(parts, " and ")
	if detail == "" {
		detail = "standard conditions"
	}

	return fmt.Sprintf("%s: %s.", level, detail)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
