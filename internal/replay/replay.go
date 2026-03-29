package replay

import "context"

// Service is the replay harness interface.
type Service interface {
	CreateCaseFromException(ctx context.Context, exceptionID string) (*ReplayCase, error)
	RunCase(ctx context.Context, replayCaseID string) (*ReplayRunResult, error)
	ListCases(ctx context.Context) ([]ReplayCase, error)
	GetCase(ctx context.Context, replayCaseID string) (*ReplayCaseDetail, error)
}

// ReplayCase is the API-facing summary of a captured replay case.
type ReplayCase struct {
	ID                  string  `json:"id"`
	Description         string  `json:"description"`
	Status              string  `json:"status"`
	SourceExceptionID   *string `json:"source_exception_id,omitempty"`
	EntityType          string  `json:"entity_type"`
	ExpectedMatchStatus string  `json:"expected_match_status"`
	ExpectedReasonCode  string  `json:"expected_reason_code"`
	CreatedAt           string  `json:"created_at"`
}

// ReplayCaseDetail includes snapshot data for the detail view.
type ReplayCaseDetail struct {
	ReplayCase
	InternalSnapshot  any `json:"internal_snapshot"`
	CustodianSnapshot any `json:"custodian_snapshot"`
	ConfigSnapshot    any `json:"config_snapshot"`
}

// ReplayRunResult is the output of replaying a case against current rules.
type ReplayRunResult struct {
	CaseID              string `json:"case_id"`
	ActualMatchStatus   string `json:"actual_match_status"`
	ActualReasonCode    string `json:"actual_reason_code"`
	ExpectedMatchStatus string `json:"expected_match_status"`
	ExpectedReasonCode  string `json:"expected_reason_code"`
	Changed             bool   `json:"changed"`
	Regression          bool   `json:"regression"`
	Improvement         bool   `json:"improvement"`
	Summary             string `json:"summary"`
}
