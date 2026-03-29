package explain

import (
	"strings"
	"testing"
	"time"

	"github.com/trustlot/trustlot/internal/model"
)

func TestBuildPosQuantityMismatch(t *testing.T) {
	exc := model.Exception{
		ID:         "exc-1",
		ReconRunID: "run-1",
		EntityType: "position",
		EntityID:   "pos-1",
		ReasonCode: model.ReasonPosQuantityMismatch,
		Status:     "open",
	}
	result := model.ReconResult{
		ID:         "rr-1",
		RunID:      "run-1",
		EntityType: "position",
		EntityID:   "pos-1",
		Status:     model.StatusBreak,
		ReasonCode: model.ReasonPosQuantityMismatch,
		Details:    "position_quantity",
	}
	asOf := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	internal := model.Position{
		ID: "pos-1", AccountID: "acct-1", SecurityID: "sec-1",
		Quantity: 100.0, MarketValue: 5000.0, AsOfDate: asOf, Source: "internal",
	}
	custodian := model.Position{
		ID: "pos-2", AccountID: "acct-1", SecurityID: "sec-1",
		Quantity: 98.0, MarketValue: 5000.0, AsOfDate: asOf, Source: "custodian",
	}

	expl := BuildExplanation(exc, result, internal, custodian)

	if expl.ExceptionID != "exc-1" {
		t.Errorf("expected exception_id exc-1, got %s", expl.ExceptionID)
	}
	if expl.ReasonCode != "POS_QUANTITY_MISMATCH" {
		t.Errorf("expected reason code POS_QUANTITY_MISMATCH, got %s", expl.ReasonCode)
	}
	if !strings.Contains(expl.Summary, "100.0000") || !strings.Contains(expl.Summary, "98.0000") {
		t.Errorf("summary should contain both quantities, got: %s", expl.Summary)
	}
	if !strings.Contains(expl.Summary, "2.0000") {
		t.Errorf("summary should contain delta, got: %s", expl.Summary)
	}
	if len(expl.FieldDiffs) != 1 {
		t.Fatalf("expected 1 field diff, got %d", len(expl.FieldDiffs))
	}
	if expl.FieldDiffs[0].Field != "quantity" {
		t.Errorf("expected field diff on quantity, got %s", expl.FieldDiffs[0].Field)
	}
	if expl.FieldDiffs[0].InternalValue != "100.0000" {
		t.Errorf("expected internal value 100.0000, got %s", expl.FieldDiffs[0].InternalValue)
	}
	if expl.FieldDiffs[0].CustodianValue != "98.0000" {
		t.Errorf("expected custodian value 98.0000, got %s", expl.FieldDiffs[0].CustodianValue)
	}
	if len(expl.SuggestedActions) == 0 {
		t.Error("expected at least one suggested action")
	}
	if len(expl.Evidence) == 0 {
		t.Error("expected at least one evidence item")
	}
}

func TestBuildPosPriceMismatch(t *testing.T) {
	exc := model.Exception{
		ID: "exc-2", ReconRunID: "run-1", EntityType: "position",
		EntityID: "pos-1", ReasonCode: model.ReasonPosPriceSourceDiff, Status: "open",
	}
	result := model.ReconResult{
		ID: "rr-2", RunID: "run-1", EntityType: "position", EntityID: "pos-1",
		Status: model.StatusBreak, ReasonCode: model.ReasonPosPriceSourceDiff, Details: "position_price",
	}
	asOf := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	internal := model.Position{
		ID: "pos-1", AccountID: "acct-1", SecurityID: "sec-1",
		Quantity: 100.0, MarketValue: 5000.00, AsOfDate: asOf, Source: "internal",
	}
	custodian := model.Position{
		ID: "pos-2", AccountID: "acct-1", SecurityID: "sec-1",
		Quantity: 100.0, MarketValue: 5000.05, AsOfDate: asOf, Source: "custodian",
	}

	expl := BuildExplanation(exc, result, internal, custodian)

	if !strings.Contains(expl.Summary, "market value mismatch") {
		t.Errorf("summary should mention market value mismatch, got: %s", expl.Summary)
	}
	if len(expl.FieldDiffs) != 1 || expl.FieldDiffs[0].Field != "market_value" {
		t.Errorf("expected field diff on market_value, got %v", expl.FieldDiffs)
	}
	hasCheckPrice := false
	for _, a := range expl.SuggestedActions {
		if a.Code == "CHECK_SOURCE_PRICE" {
			hasCheckPrice = true
		}
	}
	if !hasCheckPrice {
		t.Error("expected CHECK_SOURCE_PRICE in suggested actions")
	}
}

func TestBuildTxnDuplicate(t *testing.T) {
	tradeDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	exc := model.Exception{
		ID: "exc-3", ReconRunID: "run-2", EntityType: "transaction",
		EntityID: "txn-1", ReasonCode: model.ReasonTxnDuplicate, Status: "open",
	}
	result := model.ReconResult{
		ID: "rr-3", RunID: "run-2", EntityType: "transaction", EntityID: "txn-1",
		Status: model.StatusBreak, ReasonCode: model.ReasonTxnDuplicate, Details: "transaction_duplicate",
	}
	internal := model.Transaction{
		ID: "txn-1", AccountID: "acct-1", SecurityID: "sec-1", TxnType: "buy",
		Quantity: 100, Amount: 5000, TradeDate: tradeDate, Source: "internal",
	}
	custodian := model.Transaction{
		ID: "txn-2", AccountID: "acct-1", SecurityID: "sec-1", TxnType: "buy",
		Quantity: 100, Amount: 5100, TradeDate: tradeDate, Source: "custodian",
	}

	expl := BuildExplanation(exc, result, internal, custodian)

	if !strings.Contains(expl.Summary, "Duplicate transaction") {
		t.Errorf("summary should mention duplicate, got: %s", expl.Summary)
	}
	hasConfirm := false
	for _, a := range expl.SuggestedActions {
		if a.Code == "CONFIRM_DUPLICATE" {
			hasConfirm = true
		}
	}
	if !hasConfirm {
		t.Error("expected CONFIRM_DUPLICATE in suggested actions")
	}
}

func TestBuildTxnAmountMismatch(t *testing.T) {
	tradeDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	exc := model.Exception{
		ID: "exc-4", ReconRunID: "run-2", EntityType: "transaction",
		EntityID: "txn-1", ReasonCode: model.ReasonTxnAmountMismatch, Status: "open",
	}
	result := model.ReconResult{
		ID: "rr-4", RunID: "run-2", EntityType: "transaction", EntityID: "txn-1",
		Status: model.StatusBreak, ReasonCode: model.ReasonTxnAmountMismatch, Details: "transaction_amount",
	}
	internal := model.Transaction{
		ID: "txn-1", AccountID: "acct-1", SecurityID: "sec-1", TxnType: "buy",
		Quantity: 100, Amount: 10000.00, TradeDate: tradeDate, Source: "internal",
	}
	custodian := model.Transaction{
		ID: "txn-2", AccountID: "acct-1", SecurityID: "sec-1", TxnType: "buy",
		Quantity: 100, Amount: 10050.00, TradeDate: tradeDate, Source: "custodian",
	}

	expl := BuildExplanation(exc, result, internal, custodian)

	if !strings.Contains(expl.Summary, "amount mismatch") {
		t.Errorf("summary should mention amount mismatch, got: %s", expl.Summary)
	}
	if !strings.Contains(expl.Summary, "10000.0000") || !strings.Contains(expl.Summary, "10050.0000") {
		t.Errorf("summary should contain both amounts, got: %s", expl.Summary)
	}
	// Quantity is equal so only amount diff should appear.
	if len(expl.FieldDiffs) != 1 {
		t.Errorf("expected 1 field diff (amount only since qty matches), got %d", len(expl.FieldDiffs))
	}
}

func TestBuildExplanation_MissingCustodianRecord(t *testing.T) {
	exc := model.Exception{
		ID: "exc-5", ReconRunID: "run-1", EntityType: "position",
		EntityID: "pos-1", ReasonCode: model.ReasonMapAccountUnmapped, Status: "open",
	}
	result := model.ReconResult{
		ID: "rr-5", RunID: "run-1", EntityType: "position", EntityID: "pos-1",
		Status: model.StatusBreak, ReasonCode: model.ReasonMapAccountUnmapped, Details: "",
	}

	expl := BuildExplanation(exc, result, model.Position{Source: "internal"}, nil)

	if !strings.Contains(expl.Summary, "custodian") {
		t.Errorf("summary should mention missing custodian, got: %s", expl.Summary)
	}
	if len(expl.SuggestedActions) == 0 {
		t.Error("expected at least one suggested action for missing record")
	}
}

func TestBuildExplanation_BothRecordsMissing(t *testing.T) {
	exc := model.Exception{
		ID: "exc-6", ReconRunID: "run-1", EntityType: "position",
		EntityID: "pos-1", ReasonCode: model.ReasonMapAccountUnmapped, Status: "open",
	}
	result := model.ReconResult{
		ID: "rr-6", RunID: "run-1", EntityType: "position", EntityID: "pos-1",
		Status: model.StatusBreak, ReasonCode: model.ReasonMapAccountUnmapped,
	}

	expl := BuildExplanation(exc, result, nil, nil)

	if expl.Summary == "" {
		t.Error("summary should not be empty even with missing records")
	}
	if expl.ExceptionID != "exc-6" {
		t.Errorf("exception ID should be preserved, got %s", expl.ExceptionID)
	}
}

func TestBuildExplanation_UnknownReasonCode(t *testing.T) {
	exc := model.Exception{
		ID: "exc-7", EntityType: "position", EntityID: "pos-1",
		ReasonCode: model.ReasonCode("UNKNOWN_FUTURE_CODE"), Status: "open",
	}
	result := model.ReconResult{
		ID: "rr-7", Status: model.StatusBreak,
		ReasonCode: model.ReasonCode("UNKNOWN_FUTURE_CODE"), Details: "some_rule",
	}

	expl := BuildExplanation(exc, result, model.Position{}, model.Position{})

	if !strings.Contains(expl.Summary, "UNKNOWN_FUTURE_CODE") {
		t.Errorf("summary should include the unknown reason code, got: %s", expl.Summary)
	}
}
