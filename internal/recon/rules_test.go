package recon

import (
	"testing"
	"time"

	"github.com/trustlot/trustlot/internal/model"
)

func TestPositionQuantityMismatch(t *testing.T) {
	cfg := DefaultConfig()
	rule := positionQuantityRule{}

	t.Run("exact match", func(t *testing.T) {
		ctx := RuleContext{
			Internal:  model.Position{Quantity: 100.0, MarketValue: 5000.0},
			Custodian: model.Position{Quantity: 100.0, MarketValue: 5000.0},
			Config:    cfg,
		}
		status, rc := rule.Evaluate(ctx)
		if status != model.StatusMatch {
			t.Errorf("expected MATCH, got %s", status)
		}
		if rc != nil {
			t.Errorf("expected nil reason code, got %s", *rc)
		}
	})

	t.Run("quantity break", func(t *testing.T) {
		ctx := RuleContext{
			Internal:  model.Position{Quantity: 100.0, MarketValue: 5000.0},
			Custodian: model.Position{Quantity: 99.0, MarketValue: 5000.0},
			Config:    cfg,
		}
		status, rc := rule.Evaluate(ctx)
		if status != model.StatusBreak {
			t.Errorf("expected BREAK, got %s", status)
		}
		if rc == nil || *rc != model.ReasonPosQuantityMismatch {
			t.Errorf("expected POS_QUANTITY_MISMATCH reason code")
		}
	})
}

func TestPositionPriceMismatch(t *testing.T) {
	cfg := DefaultConfig()
	rule := positionPriceRule{}

	t.Run("within tolerance", func(t *testing.T) {
		ctx := RuleContext{
			Internal:  model.Position{Quantity: 100.0, MarketValue: 5000.00},
			Custodian: model.Position{Quantity: 100.0, MarketValue: 5000.005},
			Config:    cfg,
		}
		status, rc := rule.Evaluate(ctx)
		if status != model.StatusMatch {
			t.Errorf("expected MATCH, got %s", status)
		}
		if rc != nil {
			t.Errorf("expected nil reason code")
		}
	})

	t.Run("exceeds tolerance", func(t *testing.T) {
		ctx := RuleContext{
			Internal:  model.Position{Quantity: 100.0, MarketValue: 5000.00},
			Custodian: model.Position{Quantity: 100.0, MarketValue: 5000.02},
			Config:    cfg,
		}
		status, rc := rule.Evaluate(ctx)
		if status != model.StatusBreak {
			t.Errorf("expected BREAK, got %s", status)
		}
		if rc == nil || *rc != model.ReasonPosPriceSourceDiff {
			t.Errorf("expected POS_PRICE_SOURCE_DIFF reason code")
		}
	})
}

func TestTransactionAmountMismatch(t *testing.T) {
	cfg := DefaultConfig()
	rule := transactionAmountRule{}
	tradeDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("amounts match", func(t *testing.T) {
		ctx := RuleContext{
			Internal:  model.Transaction{Amount: 10000.00, TradeDate: tradeDate},
			Custodian: model.Transaction{Amount: 10000.00, TradeDate: tradeDate},
			Config:    cfg,
		}
		status, rc := rule.Evaluate(ctx)
		if status != model.StatusMatch {
			t.Errorf("expected MATCH, got %s", status)
		}
		if rc != nil {
			t.Errorf("expected nil reason code")
		}
	})

	t.Run("amounts differ", func(t *testing.T) {
		ctx := RuleContext{
			Internal:  model.Transaction{Amount: 10000.00, TradeDate: tradeDate},
			Custodian: model.Transaction{Amount: 10050.00, TradeDate: tradeDate},
			Config:    cfg,
		}
		status, rc := rule.Evaluate(ctx)
		if status != model.StatusBreak {
			t.Errorf("expected BREAK, got %s", status)
		}
		if rc == nil || *rc != model.ReasonTxnAmountMismatch {
			t.Errorf("expected TXN_AMOUNT_MISMATCH reason code")
		}
	})
}

func TestTransactionDuplicateDetection(t *testing.T) {
	cfg := DefaultConfig()
	rule := transactionDuplicateRule{}
	tradeDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("not duplicate", func(t *testing.T) {
		ctx := RuleContext{
			Internal: model.Transaction{
				ID: "a", SecurityID: "sec1", Quantity: 100, TradeDate: tradeDate, Amount: 5000,
			},
			Custodian: model.Transaction{
				ID: "b", SecurityID: "sec1", Quantity: 100, TradeDate: tradeDate, Amount: 5000,
			},
			Config: cfg,
		}
		status, rc := rule.Evaluate(ctx)
		if status != model.StatusMatch {
			t.Errorf("expected MATCH, got %s", status)
		}
		if rc != nil {
			t.Errorf("expected nil reason code")
		}
	})

	t.Run("duplicate with different amount", func(t *testing.T) {
		ctx := RuleContext{
			Internal: model.Transaction{
				ID: "a", SecurityID: "sec1", Quantity: 100, TradeDate: tradeDate, Amount: 5000,
			},
			Custodian: model.Transaction{
				ID: "b", SecurityID: "sec1", Quantity: 100, TradeDate: tradeDate, Amount: 5100,
			},
			Config: cfg,
		}
		status, rc := rule.Evaluate(ctx)
		if status != model.StatusBreak {
			t.Errorf("expected BREAK, got %s", status)
		}
		if rc == nil || *rc != model.ReasonTxnDuplicate {
			t.Errorf("expected TXN_DUPLICATE reason code")
		}
	})
}

func TestApplyRulesStopsAtFirstBreak(t *testing.T) {
	cfg := DefaultConfig()
	rules := PositionRules()

	// Quantity mismatch should fire before price rule is checked.
	ctx := RuleContext{
		Internal:  model.Position{Quantity: 100.0, MarketValue: 5000.00},
		Custodian: model.Position{Quantity: 50.0, MarketValue: 9999.99},
		Config:    cfg,
	}
	status, rc, ruleName := ApplyRules(rules, ctx)
	if status != model.StatusBreak {
		t.Fatalf("expected BREAK, got %s", status)
	}
	if rc == nil || *rc != model.ReasonPosQuantityMismatch {
		t.Errorf("expected POS_QUANTITY_MISMATCH, got %v", rc)
	}
	if ruleName != "position_quantity" {
		t.Errorf("expected position_quantity rule, got %s", ruleName)
	}
}

func TestDeterministicSortPositions(t *testing.T) {
	positions := []model.Position{
		{AccountID: "b", SecurityID: "2"},
		{AccountID: "a", SecurityID: "2"},
		{AccountID: "a", SecurityID: "1"},
		{AccountID: "b", SecurityID: "1"},
	}
	sortPositions(positions)

	expected := []struct{ acct, sec string }{
		{"a", "1"}, {"a", "2"}, {"b", "1"}, {"b", "2"},
	}
	for i, e := range expected {
		if positions[i].AccountID != e.acct || positions[i].SecurityID != e.sec {
			t.Errorf("index %d: expected %s|%s, got %s|%s", i, e.acct, e.sec, positions[i].AccountID, positions[i].SecurityID)
		}
	}
}

func TestDeterministicSortTransactions(t *testing.T) {
	d1 := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	txns := []model.Transaction{
		{AccountID: "a", SecurityID: "1", TxnType: "sell", TradeDate: d1},
		{AccountID: "a", SecurityID: "1", TxnType: "buy", TradeDate: d2},
		{AccountID: "a", SecurityID: "1", TxnType: "buy", TradeDate: d1},
	}
	sortTransactions(txns)

	if txns[0].TxnType != "buy" || !txns[0].TradeDate.Equal(d1) {
		t.Errorf("index 0: expected buy@d1, got %s@%v", txns[0].TxnType, txns[0].TradeDate)
	}
	if txns[1].TxnType != "buy" || !txns[1].TradeDate.Equal(d2) {
		t.Errorf("index 1: expected buy@d2, got %s@%v", txns[1].TxnType, txns[1].TradeDate)
	}
	if txns[2].TxnType != "sell" {
		t.Errorf("index 2: expected sell, got %s", txns[2].TxnType)
	}
}
