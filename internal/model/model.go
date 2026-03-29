package model

import "time"

// Custodian represents a data source that delivers portfolio feeds.
type Custodian struct {
	ID       string
	Name     string
	Format   string // csv, json, fixed_width
	Endpoint string
	Active   bool
}

// IngestionBatch represents a single delivery of data from a custodian.
type IngestionBatch struct {
	ID          string
	CustodianID string
	Status      string // pending, processing, completed, failed
	FileRef     string
	RecordCount int
	StartedAt   time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
}

// RawRecord is an unparsed record from an ingestion batch.
type RawRecord struct {
	ID       string
	BatchID  string
	SeqNum   int
	RawData  []byte
	ParsedAt *time.Time
	Error    string
}

// Account is a canonical account in the wealth data model.
type Account struct {
	ID          string
	ExternalID  string
	CustodianID string
	Name        string
	AccountType string
	CreatedAt   time.Time
}

// Security is a canonical security.
type Security struct {
	ID           string
	Symbol       string
	CUSIP        string
	ISIN         string
	SEDOL        string
	Name         string
	SecurityType string
}

// Position is a canonical holding at a point in time.
type Position struct {
	ID          string
	AccountID   string
	SecurityID  string
	Quantity    float64
	MarketValue float64
	AsOfDate    time.Time
	Source      string // custodian, internal
	BatchID     *string
}

// Transaction is a canonical trade or movement.
type Transaction struct {
	ID         string
	AccountID  string
	SecurityID string
	TxnType    string // buy, sell, dividend, transfer, fee, etc.
	Quantity   float64
	Price      float64
	Amount     float64
	TradeDate  time.Time
	SettleDate time.Time
	Source     string
	BatchID    *string
}

// TaxLot tracks cost basis for a specific purchase lot.
type TaxLot struct {
	ID           string
	AccountID    string
	SecurityID   string
	OpenDate     time.Time
	Quantity     float64
	CostBasis    float64
	ReliefMethod string // FIFO, LIFO, HIFO, SpecID
	Source       string
	BatchID      *string
}

// ReconRun represents a single execution of the reconciliation engine.
type ReconRun struct {
	ID         string
	EntityType string // position, transaction, tax_lot
	AsOfDate   time.Time
	Status     string // running, completed, failed
	StartedAt  time.Time
	CompletedAt *time.Time
}

// ReconResult is the outcome of reconciling a single entity.
type ReconResult struct {
	ID         string
	RunID      string
	EntityType string
	EntityID   string
	Status     MatchStatus
	ReasonCode ReasonCode
	Details    string
}

// Exception is a reconciliation break requiring attention.
type Exception struct {
	ID         string
	ReconRunID string
	EntityType string
	EntityID   string
	ReasonCode ReasonCode
	Status     string // open, acknowledged, resolved, suppressed
	AssignedTo string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// TrustScore measures automation readiness for an entity.
type TrustScore struct {
	ID                string
	EntityType        string
	EntityID          string
	Total             float64
	SourceReliability float64
	MappingConfidence float64
	ReconStatus       float64
	Freshness         float64
	Completeness      float64
	HumanReviewState  float64
	AnomalyPenalty    float64
	ComputedAt        time.Time
}

// LineageEdge connects two entities in the data lineage graph.
type LineageEdge struct {
	ID         string
	SourceType string
	SourceID   string
	TargetType string
	TargetID   string
	Relation   string
	CreatedAt  time.Time
}

// ReplayCase is a captured scenario for regression testing.
type ReplayCase struct {
	ID          string
	Description string
	FixtureRef  string
	Status      string // pending, passing, failing
	CreatedAt   time.Time
}
