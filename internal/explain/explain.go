package explain

// Explanation is the structured output for "Explain This Break."
type Explanation struct {
	ExceptionID    string            `json:"exception_id"`
	ReconResultID  string            `json:"recon_result_id"`
	EntityType     string            `json:"entity_type"`
	MatchStatus    string            `json:"match_status"`
	ReasonCode     string            `json:"reason_code"`
	Summary        string            `json:"summary"`
	Evidence       []EvidenceItem    `json:"evidence"`
	FieldDiffs     []FieldDiff       `json:"field_diffs"`
	SuggestedActions []SuggestedAction `json:"suggested_actions"`
}

// EvidenceItem is a labeled piece of evidence supporting the classification.
type EvidenceItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// FieldDiff shows a side-by-side comparison of internal vs custodian values.
type FieldDiff struct {
	Field          string `json:"field"`
	InternalValue  string `json:"internal_value"`
	CustodianValue string `json:"custodian_value"`
	Delta          string `json:"delta,omitempty"`
}

// SuggestedAction is a deterministic next-step recommendation.
type SuggestedAction struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}
