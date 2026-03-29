package recon

import (
	"math"

	"github.com/trustlot/trustlot/internal/model"
)

// Rule evaluates a single reconciliation condition against a paired record set.
type Rule interface {
	Name() string
	Evaluate(ctx RuleContext) (model.MatchStatus, *model.ReasonCode)
}

// RuleContext holds the data a rule needs to make a determination.
type RuleContext struct {
	Internal  interface{} // model.Position or model.Transaction
	Custodian interface{} // model.Position or model.Transaction
	Config    ReconConfig
}

// --- Position rules ---

// positionQuantityRule checks whether internal and custodian quantities agree.
type positionQuantityRule struct{}

func (r positionQuantityRule) Name() string { return "position_quantity" }

func (r positionQuantityRule) Evaluate(ctx RuleContext) (model.MatchStatus, *model.ReasonCode) {
	internal := ctx.Internal.(model.Position)
	custodian := ctx.Custodian.(model.Position)
	diff := math.Abs(internal.Quantity - custodian.Quantity)
	if diff > ctx.Config.QuantityTolerance {
		rc := model.ReasonPosQuantityMismatch
		return model.StatusBreak, &rc
	}
	return model.StatusMatch, nil
}

// positionPriceRule checks whether internal and custodian market values agree
// within the configured price tolerance.
type positionPriceRule struct{}

func (r positionPriceRule) Name() string { return "position_price" }

func (r positionPriceRule) Evaluate(ctx RuleContext) (model.MatchStatus, *model.ReasonCode) {
	internal := ctx.Internal.(model.Position)
	custodian := ctx.Custodian.(model.Position)
	diff := math.Abs(internal.MarketValue - custodian.MarketValue)
	if diff > ctx.Config.PriceTolerance {
		rc := model.ReasonPosPriceSourceDiff
		return model.StatusBreak, &rc
	}
	return model.StatusMatch, nil
}

// PositionRules returns the ordered rule set for position reconciliation.
func PositionRules() []Rule {
	return []Rule{
		positionQuantityRule{},
		positionPriceRule{},
	}
}

// --- Transaction rules ---

// transactionDuplicateRule flags duplicate transactions based on identical
// symbol, quantity, and trade date within the custodian feed.
type transactionDuplicateRule struct{}

func (r transactionDuplicateRule) Name() string { return "transaction_duplicate" }

func (r transactionDuplicateRule) Evaluate(ctx RuleContext) (model.MatchStatus, *model.ReasonCode) {
	internal := ctx.Internal.(model.Transaction)
	custodian := ctx.Custodian.(model.Transaction)

	// If the pair has identical security, quantity, and trade date but different IDs,
	// and the amounts also match, it may be a duplicate posting. The engine handles
	// true duplicates at the pairing level; this rule catches when a custodian sends
	// the same trade twice with different amounts.
	if internal.SecurityID == custodian.SecurityID &&
		internal.Quantity == custodian.Quantity &&
		internal.TradeDate.Equal(custodian.TradeDate) &&
		internal.ID != custodian.ID &&
		internal.Amount != custodian.Amount {
		rc := model.ReasonTxnDuplicate
		return model.StatusBreak, &rc
	}
	return model.StatusMatch, nil
}

// transactionAmountRule checks whether internal and custodian transaction
// amounts agree.
type transactionAmountRule struct{}

func (r transactionAmountRule) Name() string { return "transaction_amount" }

func (r transactionAmountRule) Evaluate(ctx RuleContext) (model.MatchStatus, *model.ReasonCode) {
	internal := ctx.Internal.(model.Transaction)
	custodian := ctx.Custodian.(model.Transaction)
	diff := math.Abs(internal.Amount - custodian.Amount)
	if diff > ctx.Config.PriceTolerance {
		rc := model.ReasonTxnAmountMismatch
		return model.StatusBreak, &rc
	}
	return model.StatusMatch, nil
}

// TransactionRules returns the ordered rule set for transaction reconciliation.
func TransactionRules() []Rule {
	return []Rule{
		transactionDuplicateRule{},
		transactionAmountRule{},
	}
}
