package explain

import (
	"fmt"
	"math"

	"github.com/trustlot/trustlot/internal/model"
)

// builder produces an Explanation from an internal/custodian record pair.
// Each reason code has its own builder function.
type builder func(exc model.Exception, result model.ReconResult, internal, custodian any) Explanation

var builders = map[model.ReasonCode]builder{
	model.ReasonPosQuantityMismatch: buildPosQuantityMismatch,
	model.ReasonPosPriceSourceDiff:  buildPosPriceMismatch,
	model.ReasonTxnDuplicate:        buildTxnDuplicate,
	model.ReasonTxnAmountMismatch:   buildTxnAmountMismatch,
}

// BuildExplanation dispatches to the correct builder for the reason code.
// If no builder exists or records are nil, it returns a generic explanation.
func BuildExplanation(exc model.Exception, result model.ReconResult, internal, custodian any) Explanation {
	base := Explanation{
		ExceptionID:   exc.ID,
		ReconResultID: result.ID,
		EntityType:    exc.EntityType,
		MatchStatus:   string(result.Status),
		ReasonCode:    string(exc.ReasonCode),
	}

	if internal == nil || custodian == nil {
		base.Summary = fmt.Sprintf("Unmatched %s record: no paired %s record found.",
			exc.EntityType, pairedSource(internal == nil))
		base.Evidence = []EvidenceItem{
			{Label: "Reason Code", Value: string(exc.ReasonCode)},
			{Label: "Entity ID", Value: exc.EntityID},
		}
		base.SuggestedActions = []SuggestedAction{
			{Code: "REVIEW_PENDING_ACTIVITY", Label: "Review pending ingestion activity for missing record"},
		}
		return base
	}

	fn, ok := builders[exc.ReasonCode]
	if !ok {
		base.Summary = fmt.Sprintf("Break detected: %s.", exc.ReasonCode)
		base.Evidence = []EvidenceItem{
			{Label: "Reason Code", Value: string(exc.ReasonCode)},
			{Label: "Rule", Value: result.Details},
		}
		base.SuggestedActions = []SuggestedAction{
			{Code: "REVIEW_PENDING_ACTIVITY", Label: "Review pending activity and investigate manually"},
		}
		return base
	}

	return fn(exc, result, internal, custodian)
}

func pairedSource(internalMissing bool) string {
	if internalMissing {
		return "internal"
	}
	return "custodian"
}

// --- Position builders ---

func buildPosQuantityMismatch(exc model.Exception, result model.ReconResult, intRaw, custRaw any) Explanation {
	intPos := intRaw.(model.Position)
	custPos := custRaw.(model.Position)
	delta := intPos.Quantity - custPos.Quantity

	return Explanation{
		ExceptionID:   exc.ID,
		ReconResultID: result.ID,
		EntityType:    exc.EntityType,
		MatchStatus:   string(result.Status),
		ReasonCode:    string(exc.ReasonCode),
		Summary: fmt.Sprintf("Position quantity mismatch: internal quantity %.4f vs custodian quantity %.4f (delta %.4f).",
			intPos.Quantity, custPos.Quantity, delta),
		Evidence: []EvidenceItem{
			{Label: "Account", Value: intPos.AccountID},
			{Label: "Security", Value: intPos.SecurityID},
			{Label: "As-Of Date", Value: intPos.AsOfDate.Format("2006-01-02")},
			{Label: "Rule", Value: result.Details},
			{Label: "Internal Source", Value: intPos.Source},
			{Label: "Custodian Source", Value: custPos.Source},
		},
		FieldDiffs: []FieldDiff{
			{
				Field:          "quantity",
				InternalValue:  fmt.Sprintf("%.4f", intPos.Quantity),
				CustodianValue: fmt.Sprintf("%.4f", custPos.Quantity),
				Delta:          fmt.Sprintf("%.4f", delta),
			},
		},
		SuggestedActions: []SuggestedAction{
			{Code: "REVIEW_PENDING_ACTIVITY", Label: "Check for pending transactions that may explain the difference"},
			{Code: "CHECK_LOT_RELIEF", Label: "Verify tax-lot relief has been applied consistently"},
		},
	}
}

func buildPosPriceMismatch(exc model.Exception, result model.ReconResult, intRaw, custRaw any) Explanation {
	intPos := intRaw.(model.Position)
	custPos := custRaw.(model.Position)
	delta := intPos.MarketValue - custPos.MarketValue

	return Explanation{
		ExceptionID:   exc.ID,
		ReconResultID: result.ID,
		EntityType:    exc.EntityType,
		MatchStatus:   string(result.Status),
		ReasonCode:    string(exc.ReasonCode),
		Summary: fmt.Sprintf("Position market value mismatch: internal %.4f vs custodian %.4f (delta %.4f, tolerance exceeded).",
			intPos.MarketValue, custPos.MarketValue, delta),
		Evidence: []EvidenceItem{
			{Label: "Account", Value: intPos.AccountID},
			{Label: "Security", Value: intPos.SecurityID},
			{Label: "As-Of Date", Value: intPos.AsOfDate.Format("2006-01-02")},
			{Label: "Rule", Value: result.Details},
		},
		FieldDiffs: []FieldDiff{
			{
				Field:          "market_value",
				InternalValue:  fmt.Sprintf("%.4f", intPos.MarketValue),
				CustodianValue: fmt.Sprintf("%.4f", custPos.MarketValue),
				Delta:          fmt.Sprintf("%.4f", delta),
			},
		},
		SuggestedActions: []SuggestedAction{
			{Code: "CHECK_SOURCE_PRICE", Label: "Compare pricing sources for this security"},
			{Code: "REVIEW_PENDING_ACTIVITY", Label: "Check for intra-day price movements or stale pricing"},
		},
	}
}

// --- Transaction builders ---

func buildTxnDuplicate(exc model.Exception, result model.ReconResult, intRaw, custRaw any) Explanation {
	intTxn := intRaw.(model.Transaction)
	custTxn := custRaw.(model.Transaction)

	return Explanation{
		ExceptionID:   exc.ID,
		ReconResultID: result.ID,
		EntityType:    exc.EntityType,
		MatchStatus:   string(result.Status),
		ReasonCode:    string(exc.ReasonCode),
		Summary: fmt.Sprintf("Duplicate transaction detected: same security %s, quantity %.4f, trade date %s with differing amounts.",
			intTxn.SecurityID, intTxn.Quantity, intTxn.TradeDate.Format("2006-01-02")),
		Evidence: []EvidenceItem{
			{Label: "Account", Value: intTxn.AccountID},
			{Label: "Security", Value: intTxn.SecurityID},
			{Label: "Trade Date", Value: intTxn.TradeDate.Format("2006-01-02")},
			{Label: "Type", Value: intTxn.TxnType},
			{Label: "Rule", Value: result.Details},
		},
		FieldDiffs: []FieldDiff{
			{
				Field:          "amount",
				InternalValue:  fmt.Sprintf("%.4f", intTxn.Amount),
				CustodianValue: fmt.Sprintf("%.4f", custTxn.Amount),
				Delta:          fmt.Sprintf("%.4f", intTxn.Amount-custTxn.Amount),
			},
		},
		SuggestedActions: []SuggestedAction{
			{Code: "CONFIRM_DUPLICATE", Label: "Confirm whether custodian posted the same trade twice"},
			{Code: "LINK_CORRECTION_CHAIN", Label: "Check if this is part of a cancel/correct sequence"},
		},
	}
}

func buildTxnAmountMismatch(exc model.Exception, result model.ReconResult, intRaw, custRaw any) Explanation {
	intTxn := intRaw.(model.Transaction)
	custTxn := custRaw.(model.Transaction)
	delta := intTxn.Amount - custTxn.Amount

	diffs := []FieldDiff{
		{
			Field:          "amount",
			InternalValue:  fmt.Sprintf("%.4f", intTxn.Amount),
			CustodianValue: fmt.Sprintf("%.4f", custTxn.Amount),
			Delta:          fmt.Sprintf("%.4f", delta),
		},
	}

	// Include quantity diff if also mismatched.
	if math.Abs(intTxn.Quantity-custTxn.Quantity) > 0 {
		diffs = append(diffs, FieldDiff{
			Field:          "quantity",
			InternalValue:  fmt.Sprintf("%.4f", intTxn.Quantity),
			CustodianValue: fmt.Sprintf("%.4f", custTxn.Quantity),
			Delta:          fmt.Sprintf("%.4f", intTxn.Quantity-custTxn.Quantity),
		})
	}

	return Explanation{
		ExceptionID:   exc.ID,
		ReconResultID: result.ID,
		EntityType:    exc.EntityType,
		MatchStatus:   string(result.Status),
		ReasonCode:    string(exc.ReasonCode),
		Summary: fmt.Sprintf("Transaction amount mismatch: internal %.4f vs custodian %.4f (delta %.4f).",
			intTxn.Amount, custTxn.Amount, delta),
		Evidence: []EvidenceItem{
			{Label: "Account", Value: intTxn.AccountID},
			{Label: "Security", Value: intTxn.SecurityID},
			{Label: "Trade Date", Value: intTxn.TradeDate.Format("2006-01-02")},
			{Label: "Type", Value: intTxn.TxnType},
			{Label: "Rule", Value: result.Details},
		},
		FieldDiffs:     diffs,
		SuggestedActions: []SuggestedAction{
			{Code: "REVIEW_PENDING_ACTIVITY", Label: "Check for pending corrections or fee adjustments"},
			{Code: "LINK_CORRECTION_CHAIN", Label: "Look for a correction chain that may resolve this difference"},
		},
	}
}
