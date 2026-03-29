package trust

import (
	"math"
	"testing"
	"time"

	"github.com/trustlot/trustlot/internal/model"
)

func TestReconciliationScore(t *testing.T) {
	tests := []struct {
		status model.MatchStatus
		want   float64
	}{
		{model.StatusMatch, 1.0},
		{model.StatusNearMatch, 0.7},
		{model.StatusBreak, 0.2},
		{"", 0.0},
	}
	for _, tt := range tests {
		got := reconciliationScore(tt.status)
		if got != tt.want {
			t.Errorf("reconciliationScore(%q) = %v, want %v", tt.status, got, tt.want)
		}
	}
}

func TestSourceReliabilityScore(t *testing.T) {
	tests := []struct {
		source string
		want   float64
	}{
		{"internal", 0.95},
		{"custodian", 0.85},
		{"unknown", 0.70},
		{"", 0.70},
	}
	for _, tt := range tests {
		got := sourceReliabilityScore(tt.source)
		if got != tt.want {
			t.Errorf("sourceReliabilityScore(%q) = %v, want %v", tt.source, got, tt.want)
		}
	}
}

func TestFreshnessScore(t *testing.T) {
	now := time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		asOf    time.Time
		want    float64
	}{
		{"same day", now, 1.0},
		{"1 day stale", now.AddDate(0, 0, -1), 0.7},
		{"3 days stale", now.AddDate(0, 0, -3), 0.3},
		{"7 days stale", now.AddDate(0, 0, -7), 0.3},
		{"8 days stale", now.AddDate(0, 0, -8), 0.1},
		{"30 days stale", now.AddDate(0, 0, -30), 0.1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := freshnessScore(tt.asOf, now)
			if got != tt.want {
				t.Errorf("freshnessScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompletenessScore(t *testing.T) {
	tests := []struct {
		name string
		in   reconInput
		want float64
	}{
		{
			"all present",
			reconInput{HasAccountID: true, HasSecurityID: true, HasQuantity: true, HasAmount: true},
			1.0,
		},
		{
			"missing amount",
			reconInput{HasAccountID: true, HasSecurityID: true, HasQuantity: true, HasAmount: false},
			0.75,
		},
		{
			"missing two",
			reconInput{HasAccountID: true, HasSecurityID: false, HasQuantity: true, HasAmount: false},
			0.5,
		},
		{
			"all missing",
			reconInput{},
			0.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := completenessScore(tt.in)
			if got != tt.want {
				t.Errorf("completenessScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHumanReviewScore(t *testing.T) {
	tests := []struct {
		status string
		want   float64
	}{
		{"", 0.5},
		{"open", 0.0},
		{"acknowledged", 0.4},
		{"resolved", 1.0},
		{"suppressed", 1.0},
		{"unknown", 0.5},
	}
	for _, tt := range tests {
		got := humanReviewScore(tt.status)
		if got != tt.want {
			t.Errorf("humanReviewScore(%q) = %v, want %v", tt.status, got, tt.want)
		}
	}
}

func TestAnomalyPenaltyScore(t *testing.T) {
	tests := []struct {
		status model.MatchStatus
		reason model.ReasonCode
		want   float64
	}{
		{model.StatusMatch, "", 0.0},
		{model.StatusNearMatch, "", 0.0},
		{model.StatusBreak, model.ReasonTxnDuplicate, 0.4},
		{model.StatusBreak, model.ReasonPosQuantityMismatch, 0.3},
		{model.StatusBreak, model.ReasonTxnAmountMismatch, 0.3},
		{model.StatusBreak, model.ReasonPosPriceSourceDiff, 0.2},
		{model.StatusBreak, model.ReasonMapAccountUnmapped, 0.2},
	}
	for _, tt := range tests {
		got := anomalyPenaltyScore(tt.status, tt.reason)
		if got != tt.want {
			t.Errorf("anomalyPenaltyScore(%q, %q) = %v, want %v", tt.status, tt.reason, got, tt.want)
		}
	}
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) < tolerance
}

func TestMatchHighScore(t *testing.T) {
	// MATCH + same-day + complete + no exception = high trust
	in := reconInput{
		MatchStatus:   model.StatusMatch,
		Source:        "internal",
		AsOfDate:      time.Now(),
		Now:           time.Now(),
		HasAccountID:  true,
		HasSecurityID: true,
		HasQuantity:   true,
		HasAmount:     true,
	}
	c := computeComponents(in)
	total := computeTotal(c)

	if total < 75 {
		t.Errorf("MATCH with complete same-day data should score >= 75, got %v", total)
	}
	if c.Reconciliation != 1.0 {
		t.Errorf("MATCH recon component should be 1.0, got %v", c.Reconciliation)
	}
	if c.AnomalyPenalty != 0.0 {
		t.Errorf("MATCH should have no anomaly penalty, got %v", c.AnomalyPenalty)
	}
}

func TestBreakOpenExceptionLowScore(t *testing.T) {
	// BREAK + open exception + stale = low trust
	in := reconInput{
		MatchStatus:     model.StatusBreak,
		ReasonCode:      model.ReasonPosQuantityMismatch,
		Source:          "custodian",
		AsOfDate:        time.Now().AddDate(0, 0, -10),
		Now:             time.Now(),
		ExceptionStatus: "open",
		HasAccountID:    true,
		HasSecurityID:   true,
		HasQuantity:     true,
		HasAmount:       true,
	}
	c := computeComponents(in)
	total := computeTotal(c)

	if total >= 40 {
		t.Errorf("BREAK with open exception and stale data should score < 40, got %v", total)
	}
	if c.Reconciliation != 0.2 {
		t.Errorf("BREAK recon component should be 0.2, got %v", c.Reconciliation)
	}
	if c.HumanReview != 0.0 {
		t.Errorf("open exception review component should be 0.0, got %v", c.HumanReview)
	}
}

func TestResolvedExceptionImprovesScore(t *testing.T) {
	base := reconInput{
		MatchStatus:   model.StatusBreak,
		ReasonCode:    model.ReasonPosQuantityMismatch,
		Source:        "custodian",
		AsOfDate:      time.Now(),
		Now:           time.Now(),
		HasAccountID:  true,
		HasSecurityID: true,
		HasQuantity:   true,
		HasAmount:     true,
	}

	openInput := base
	openInput.ExceptionStatus = "open"
	openTotal := computeTotal(computeComponents(openInput))

	resolvedInput := base
	resolvedInput.ExceptionStatus = "resolved"
	resolvedTotal := computeTotal(computeComponents(resolvedInput))

	if resolvedTotal <= openTotal {
		t.Errorf("resolved exception (%v) should score higher than open (%v)", resolvedTotal, openTotal)
	}
}

func TestStaleDataLowersFreshness(t *testing.T) {
	now := time.Now()
	fresh := reconInput{
		MatchStatus:   model.StatusMatch,
		Source:        "internal",
		AsOfDate:      now,
		Now:           now,
		HasAccountID:  true,
		HasSecurityID: true,
		HasQuantity:   true,
		HasAmount:     true,
	}

	stale := fresh
	stale.AsOfDate = now.AddDate(0, 0, -14)

	freshTotal := computeTotal(computeComponents(fresh))
	staleTotal := computeTotal(computeComponents(stale))

	if staleTotal >= freshTotal {
		t.Errorf("stale data (%v) should score lower than fresh (%v)", staleTotal, freshTotal)
	}
}

func TestMissingFieldsLowerCompleteness(t *testing.T) {
	complete := reconInput{
		MatchStatus:   model.StatusMatch,
		Source:        "internal",
		AsOfDate:      time.Now(),
		Now:           time.Now(),
		HasAccountID:  true,
		HasSecurityID: true,
		HasQuantity:   true,
		HasAmount:     true,
	}

	incomplete := complete
	incomplete.HasAmount = false
	incomplete.HasSecurityID = false

	completeTotal := computeTotal(computeComponents(complete))
	incompleteTotal := computeTotal(computeComponents(incomplete))

	if incompleteTotal >= completeTotal {
		t.Errorf("incomplete data (%v) should score lower than complete (%v)", incompleteTotal, completeTotal)
	}
}

func TestDeterministicOutput(t *testing.T) {
	// Same input must always produce the same output.
	fixedTime := time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC)
	in := reconInput{
		MatchStatus:     model.StatusBreak,
		ReasonCode:      model.ReasonTxnDuplicate,
		Source:          "custodian",
		AsOfDate:        fixedTime.AddDate(0, 0, -2),
		Now:             fixedTime,
		ExceptionStatus: "acknowledged",
		HasAccountID:    true,
		HasSecurityID:   true,
		HasQuantity:     true,
		HasAmount:       false,
	}

	c1 := computeComponents(in)
	t1 := computeTotal(c1)

	c2 := computeComponents(in)
	t2 := computeTotal(c2)

	if t1 != t2 {
		t.Errorf("determinism violated: %v != %v", t1, t2)
	}
	if c1 != c2 {
		t.Errorf("components determinism violated: %+v != %+v", c1, c2)
	}
}

func TestRationaleGeneration(t *testing.T) {
	tests := []struct {
		name       string
		score      float64
		match      model.MatchStatus
		excStatus  string
		wantPrefix string
	}{
		{"high match", 85.0, model.StatusMatch, "", "High trust"},
		{"medium near-match", 55.0, model.StatusNearMatch, "acknowledged", "Medium trust"},
		{"low break", 20.0, model.StatusBreak, "open", "Low trust"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := computeComponents(reconInput{MatchStatus: tt.match, ExceptionStatus: tt.excStatus, Now: time.Now(), AsOfDate: time.Now()})
			r := buildRationale(tt.score, c, tt.match, tt.excStatus)
			if len(r) == 0 {
				t.Error("rationale should not be empty")
			}
			if r[:len(tt.wantPrefix)] != tt.wantPrefix {
				t.Errorf("rationale should start with %q, got %q", tt.wantPrefix, r)
			}
		})
	}
}

func TestScoreRange(t *testing.T) {
	// Exhaustive check: all combinations stay within [0, 100].
	statuses := []model.MatchStatus{model.StatusMatch, model.StatusNearMatch, model.StatusBreak}
	sources := []string{"internal", "custodian", "unknown"}
	excStatuses := []string{"", "open", "acknowledged", "resolved"}
	now := time.Now()
	dates := []time.Time{now, now.AddDate(0, 0, -1), now.AddDate(0, 0, -30)}

	for _, st := range statuses {
		for _, src := range sources {
			for _, exc := range excStatuses {
				for _, d := range dates {
					in := reconInput{
						MatchStatus:     st,
						Source:          src,
						AsOfDate:        d,
						Now:             now,
						ExceptionStatus: exc,
						HasAccountID:    true,
						HasSecurityID:   true,
						HasQuantity:     true,
						HasAmount:       true,
					}
					total := computeTotal(computeComponents(in))
					if total < 0 || total > 100 {
						t.Errorf("score out of range: %v for input %+v", total, in)
					}
				}
			}
		}
	}
}
