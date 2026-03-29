package model

import "testing"

func TestAllReasonCodesReturnsExpectedCount(t *testing.T) {
	codes := AllReasonCodes()
	if len(codes) != 17 {
		t.Errorf("expected 17 reason codes, got %d", len(codes))
	}
}

func TestReasonCodesAreNonEmpty(t *testing.T) {
	for _, code := range AllReasonCodes() {
		if code == "" {
			t.Error("found empty reason code")
		}
	}
}

func TestMatchStatusValues(t *testing.T) {
	statuses := []MatchStatus{StatusMatch, StatusNearMatch, StatusBreak}
	for _, s := range statuses {
		if s == "" {
			t.Error("found empty match status")
		}
	}
}
