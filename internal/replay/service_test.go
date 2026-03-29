package replay

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/trustlot/trustlot/internal/db"
	"github.com/trustlot/trustlot/internal/model"
	"github.com/trustlot/trustlot/internal/recon"
)

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

func TestReplayPosition_MatchesOriginal(t *testing.T) {
	asOf := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	internal := model.Position{
		ID: "pos-1", AccountID: "acct-1", SecurityID: "sec-1",
		Quantity: 100.0, MarketValue: 5000.0, AsOfDate: asOf, Source: "internal",
	}
	custodian := model.Position{
		ID: "pos-2", AccountID: "acct-1", SecurityID: "sec-1",
		Quantity: 98.0, MarketValue: 5000.0, AsOfDate: asOf, Source: "custodian",
	}

	row := db.ReplayCaseRow{
		ID:                  "case-1",
		EntityType:          "position",
		ExpectedMatchStatus: "BREAK",
		ExpectedReasonCode:  "POS_QUANTITY_MISMATCH",
		InternalSnapshot:    mustMarshal(t, internal),
		CustodianSnapshot:   mustMarshal(t, custodian),
		ConfigSnapshot:      mustMarshal(t, recon.DefaultConfig()),
	}

	status, rc, err := replayFromSnapshots(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != model.StatusBreak {
		t.Errorf("expected BREAK, got %s", status)
	}
	if rc == nil || *rc != model.ReasonPosQuantityMismatch {
		t.Errorf("expected POS_QUANTITY_MISMATCH, got %v", rc)
	}

	// buildSummary should say unchanged.
	summary := buildSummary(false, false, false, row, status, string(*rc))
	if summary != "Replay matched original outcome." {
		t.Errorf("unexpected summary: %s", summary)
	}
}

func TestReplayPosition_Improvement(t *testing.T) {
	// Simulate a fix: quantities now match after a rule change.
	asOf := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	internal := model.Position{
		ID: "pos-1", AccountID: "acct-1", SecurityID: "sec-1",
		Quantity: 100.0, MarketValue: 5000.0, AsOfDate: asOf, Source: "internal",
	}
	custodian := model.Position{
		ID: "pos-2", AccountID: "acct-1", SecurityID: "sec-1",
		Quantity: 100.0, MarketValue: 5000.0, AsOfDate: asOf, Source: "custodian",
	}

	row := db.ReplayCaseRow{
		ID:                  "case-2",
		EntityType:          "position",
		ExpectedMatchStatus: "BREAK",
		ExpectedReasonCode:  "POS_QUANTITY_MISMATCH",
		InternalSnapshot:    mustMarshal(t, internal),
		CustodianSnapshot:   mustMarshal(t, custodian),
		ConfigSnapshot:      mustMarshal(t, recon.DefaultConfig()),
	}

	status, rc, err := replayFromSnapshots(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != model.StatusMatch {
		t.Errorf("expected MATCH, got %s", status)
	}
	if rc != nil {
		t.Errorf("expected nil reason code, got %v", *rc)
	}

	// Should detect improvement.
	changed := true
	improvement := true
	summary := buildSummary(changed, improvement, false, row, status, "")
	if !strings.Contains(summary, "improved") {
		t.Errorf("expected improvement summary, got: %s", summary)
	}
}

func TestReplayTransaction_MatchesOriginal(t *testing.T) {
	tradeDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	settleDate := time.Date(2025, 1, 17, 0, 0, 0, 0, time.UTC)

	// Use different quantities so the duplicate rule does not fire.
	// The amount rule should trigger instead.
	internal := model.Transaction{
		ID: "txn-1", AccountID: "acct-1", SecurityID: "sec-1", TxnType: "buy",
		Quantity: 100, Price: 50, Amount: 5000, TradeDate: tradeDate,
		SettleDate: settleDate, Source: "internal",
	}
	custodian := model.Transaction{
		ID: "txn-2", AccountID: "acct-1", SecurityID: "sec-1", TxnType: "buy",
		Quantity: 95, Price: 50, Amount: 5100, TradeDate: tradeDate,
		SettleDate: settleDate, Source: "custodian",
	}

	row := db.ReplayCaseRow{
		ID:                  "case-3",
		EntityType:          "transaction",
		ExpectedMatchStatus: "BREAK",
		ExpectedReasonCode:  "TXN_AMOUNT_MISMATCH",
		InternalSnapshot:    mustMarshal(t, internal),
		CustodianSnapshot:   mustMarshal(t, custodian),
		ConfigSnapshot:      mustMarshal(t, recon.DefaultConfig()),
	}

	status, rc, err := replayFromSnapshots(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != model.StatusBreak {
		t.Errorf("expected BREAK, got %s", status)
	}
	if rc == nil || *rc != model.ReasonTxnAmountMismatch {
		t.Errorf("expected TXN_AMOUNT_MISMATCH, got %v", rc)
	}
}

func TestReplayTransaction_Regression(t *testing.T) {
	// Simulate regression: was matching, now breaks due to new rule.
	tradeDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	settleDate := time.Date(2025, 1, 17, 0, 0, 0, 0, time.UTC)

	internal := model.Transaction{
		ID: "txn-1", AccountID: "acct-1", SecurityID: "sec-1", TxnType: "buy",
		Quantity: 100, Price: 50, Amount: 5000.05, TradeDate: tradeDate,
		SettleDate: settleDate, Source: "internal",
	}
	custodian := model.Transaction{
		ID: "txn-2", AccountID: "acct-1", SecurityID: "sec-1", TxnType: "buy",
		Quantity: 100, Price: 50, Amount: 5000.00, TradeDate: tradeDate,
		SettleDate: settleDate, Source: "custodian",
	}

	// Expected was MATCH (tolerance was higher), but current rules break at 0.01.
	row := db.ReplayCaseRow{
		ID:                  "case-4",
		EntityType:          "transaction",
		ExpectedMatchStatus: "MATCH",
		ExpectedReasonCode:  "",
		InternalSnapshot:    mustMarshal(t, internal),
		CustodianSnapshot:   mustMarshal(t, custodian),
		ConfigSnapshot:      mustMarshal(t, recon.DefaultConfig()),
	}

	status, rc, err := replayFromSnapshots(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With default tolerance 0.01, delta of 0.05 should break.
	if status != model.StatusBreak {
		t.Errorf("expected BREAK, got %s", status)
	}

	actualRC := ""
	if rc != nil {
		actualRC = string(*rc)
	}
	summary := buildSummary(true, false, true, row, status, actualRC)
	if !strings.Contains(summary, "regressed") {
		t.Errorf("expected regression summary, got: %s", summary)
	}
}

func TestReplay_MissingSnapshots(t *testing.T) {
	row := db.ReplayCaseRow{
		ID:                "case-5",
		EntityType:        "position",
		InternalSnapshot:  nil,
		CustodianSnapshot: nil,
	}

	_, _, err := replayFromSnapshots(row)
	if err == nil {
		t.Error("expected error for missing snapshots")
	}
	if !strings.Contains(err.Error(), "incomplete snapshots") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReplay_Determinism(t *testing.T) {
	// Run the same replay twice and verify identical output.
	asOf := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	internal := model.Position{
		ID: "pos-1", AccountID: "acct-1", SecurityID: "sec-1",
		Quantity: 100.0, MarketValue: 5000.02, AsOfDate: asOf, Source: "internal",
	}
	custodian := model.Position{
		ID: "pos-2", AccountID: "acct-1", SecurityID: "sec-1",
		Quantity: 100.0, MarketValue: 5000.0, AsOfDate: asOf, Source: "custodian",
	}

	row := db.ReplayCaseRow{
		ID:                "case-6",
		EntityType:        "position",
		InternalSnapshot:  mustMarshal(t, internal),
		CustodianSnapshot: mustMarshal(t, custodian),
		ConfigSnapshot:    mustMarshal(t, recon.DefaultConfig()),
	}

	status1, rc1, err1 := replayFromSnapshots(row)
	status2, rc2, err2 := replayFromSnapshots(row)

	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v, %v", err1, err2)
	}
	if status1 != status2 {
		t.Errorf("non-deterministic status: %s vs %s", status1, status2)
	}
	if (rc1 == nil) != (rc2 == nil) {
		t.Errorf("non-deterministic reason code presence")
	}
	if rc1 != nil && rc2 != nil && *rc1 != *rc2 {
		t.Errorf("non-deterministic reason code: %s vs %s", *rc1, *rc2)
	}
}
