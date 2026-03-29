package model

// MatchStatus represents the outcome of a reconciliation comparison.
type MatchStatus string

const (
	StatusMatch     MatchStatus = "MATCH"
	StatusNearMatch MatchStatus = "NEAR_MATCH"
	StatusBreak     MatchStatus = "BREAK"
)

// ReasonCode identifies why a reconciliation break occurred.
type ReasonCode string

const (
	// Ingestion
	ReasonIngSchemaDrift ReasonCode = "ING_SCHEMA_DRIFT"
	ReasonIngLateFeed    ReasonCode = "ING_LATE_FEED"

	// Mapping
	ReasonMapSecurityAmbiguous ReasonCode = "MAP_SECURITY_AMBIGUOUS"
	ReasonMapAccountUnmapped   ReasonCode = "MAP_ACCOUNT_UNMAPPED"

	// Transaction
	ReasonTxnTradeSettleWindow  ReasonCode = "TXN_TRADE_SETTLE_WINDOW"
	ReasonTxnDuplicate          ReasonCode = "TXN_DUPLICATE"
	ReasonTxnCorrectionPosted   ReasonCode = "TXN_CORRECTION_POSTED"
	ReasonTxnAmountMismatch     ReasonCode = "TXN_AMOUNT_MISMATCH"

	// Position
	ReasonPosQuantityMismatch ReasonCode = "POS_QUANTITY_MISMATCH"
	ReasonPosPriceSourceDiff  ReasonCode = "POS_PRICE_SOURCE_DIFF"

	// Cash
	ReasonCashUnpostedFee ReasonCode = "CASH_UNPOSTED_FEE"

	// Tax Lot
	ReasonLotCostBasisMismatch  ReasonCode = "LOT_COST_BASIS_MISMATCH"
	ReasonLotReliefMethodDiff   ReasonCode = "LOT_RELIEF_METHOD_DIFF"
	ReasonLotShareAdjustmentCA  ReasonCode = "LOT_SHARE_ADJUSTMENT_CA"
	ReasonLotWashSaleAdj        ReasonCode = "LOT_WASH_SALE_ADJ"

	// Corporate Action
	ReasonCAEventMissing ReasonCode = "CA_EVENT_MISSING"

	// Operations
	ReasonOpsManualOverride ReasonCode = "OPS_MANUAL_OVERRIDE"
)

// AllReasonCodes returns every defined reason code.
func AllReasonCodes() []ReasonCode {
	return []ReasonCode{
		ReasonIngSchemaDrift,
		ReasonIngLateFeed,
		ReasonMapSecurityAmbiguous,
		ReasonMapAccountUnmapped,
		ReasonTxnTradeSettleWindow,
		ReasonTxnDuplicate,
		ReasonTxnCorrectionPosted,
		ReasonTxnAmountMismatch,
		ReasonPosQuantityMismatch,
		ReasonPosPriceSourceDiff,
		ReasonCashUnpostedFee,
		ReasonLotCostBasisMismatch,
		ReasonLotReliefMethodDiff,
		ReasonLotShareAdjustmentCA,
		ReasonLotWashSaleAdj,
		ReasonCAEventMissing,
		ReasonOpsManualOverride,
	}
}
