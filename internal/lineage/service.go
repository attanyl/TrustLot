package lineage

import (
	"context"
	"fmt"
	"sort"

	"github.com/trustlot/trustlot/internal/db"
)

type service struct {
	store *db.Store
}

// NewService creates a lineage service backed by the given store.
func NewService(store *db.Store) Service {
	return &service{store: store}
}

func (s *service) GetLineageForEntity(ctx context.Context, entityType string, entityID string) (*LineageGraph, error) {
	b := newBuilder(entityType, entityID)

	switch entityType {
	case "position":
		if err := s.addPositionLineage(ctx, b, entityID); err != nil {
			return nil, err
		}
	case "transaction":
		if err := s.addTransactionLineage(ctx, b, entityID); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported entity type for lineage: %s", entityType)
	}

	return b.build(), nil
}

func (s *service) GetLineageForException(ctx context.Context, exceptionID string) (*LineageGraph, error) {
	exc, err := s.store.GetException(ctx, exceptionID)
	if err != nil {
		return nil, fmt.Errorf("lineage: load exception: %w", err)
	}

	b := newBuilder("exception", exceptionID)

	// Add the exception node.
	b.addNode(LineageNode{
		ID:   exc.ID,
		Type: "exception",
		Label: fmt.Sprintf("Exception: %s", exc.ReasonCode),
		Metadata: map[string]any{
			"status":      exc.Status,
			"reason_code": string(exc.ReasonCode),
			"assigned_to": exc.AssignedTo,
		},
	})

	// Add the canonical entity and its upstream/downstream.
	switch exc.EntityType {
	case "position":
		if err := s.addPositionLineage(ctx, b, exc.EntityID); err != nil {
			return nil, err
		}
	case "transaction":
		if err := s.addTransactionLineage(ctx, b, exc.EntityID); err != nil {
			return nil, err
		}
	}

	// Link entity → exception.
	b.addEdge(LineageEdge{FromID: exc.EntityID, ToID: exc.ID, EdgeType: "raised_exception"})

	// Link recon_result → exception if we can find it.
	rr, err := s.store.GetReconResultForException(ctx, exc.ReconRunID, exc.EntityType, exc.EntityID)
	if err == nil {
		b.addEdge(LineageEdge{FromID: rr.ID, ToID: exc.ID, EdgeType: "raised_exception"})
	}

	// Trust score for the entity.
	s.addTrustScore(ctx, b, exc.EntityType, exc.EntityID)

	return b.build(), nil
}

// addPositionLineage adds the full chain for a position.
func (s *service) addPositionLineage(ctx context.Context, b *graphBuilder, positionID string) error {
	p, err := s.store.GetPosition(ctx, positionID)
	if err != nil {
		return fmt.Errorf("lineage: load position: %w", err)
	}

	b.addNode(LineageNode{
		ID:   p.ID,
		Type: "position",
		Label: fmt.Sprintf("Position: %s qty=%.2f", p.Source, p.Quantity),
		Metadata: map[string]any{
			"account_id":   p.AccountID,
			"security_id":  p.SecurityID,
			"quantity":     p.Quantity,
			"market_value": p.MarketValue,
			"as_of_date":   p.AsOfDate.Format("2006-01-02"),
			"source":       p.Source,
		},
	})

	// Upstream: raw_records via batch_id.
	if p.BatchID != nil {
		s.addRawRecords(ctx, b, *p.BatchID, p.ID)
	}

	// Downstream: recon_results, exceptions, trust_score.
	s.addReconResultsForEntity(ctx, b, "position", p.ID)
	s.addExceptionsForEntity(ctx, b, "position", p.ID)
	s.addTrustScore(ctx, b, "position", p.ID)

	return nil
}

// addTransactionLineage adds the full chain for a transaction.
func (s *service) addTransactionLineage(ctx context.Context, b *graphBuilder, txnID string) error {
	t, err := s.store.GetTransaction(ctx, txnID)
	if err != nil {
		return fmt.Errorf("lineage: load transaction: %w", err)
	}

	b.addNode(LineageNode{
		ID:   t.ID,
		Type: "transaction",
		Label: fmt.Sprintf("Transaction: %s %s qty=%.2f", t.TxnType, t.Source, t.Quantity),
		Metadata: map[string]any{
			"account_id":  t.AccountID,
			"security_id": t.SecurityID,
			"txn_type":    t.TxnType,
			"quantity":    t.Quantity,
			"price":       t.Price,
			"amount":      t.Amount,
			"trade_date":  t.TradeDate.Format("2006-01-02"),
			"settle_date": t.SettleDate.Format("2006-01-02"),
			"source":      t.Source,
		},
	})

	// Upstream: raw_records via batch_id.
	if t.BatchID != nil {
		s.addRawRecords(ctx, b, *t.BatchID, t.ID)
	}

	// Downstream: recon_results, exceptions, trust_score.
	s.addReconResultsForEntity(ctx, b, "transaction", t.ID)
	s.addExceptionsForEntity(ctx, b, "transaction", t.ID)
	s.addTrustScore(ctx, b, "transaction", t.ID)

	return nil
}

// addRawRecords finds raw records from the same batch and links them upstream.
func (s *service) addRawRecords(ctx context.Context, b *graphBuilder, batchID, entityID string) {
	records, err := s.store.ListRawRecordsByBatch(ctx, batchID)
	if err != nil {
		return // partial lineage is acceptable
	}
	for _, r := range records {
		b.addNode(LineageNode{
			ID:   r.ID,
			Type: "raw_record",
			Label: fmt.Sprintf("Raw Record #%d", r.SeqNum),
			Metadata: map[string]any{
				"batch_id": r.BatchID,
				"seq_num":  r.SeqNum,
			},
		})
		b.addEdge(LineageEdge{FromID: r.ID, ToID: entityID, EdgeType: "normalized_to"})
	}

	// Also add the batch node.
	batch, err := s.store.GetIngestionBatch(ctx, batchID)
	if err == nil {
		b.addNode(LineageNode{
			ID:   batch.ID,
			Type: "ingestion_batch",
			Label: fmt.Sprintf("Batch: %s", batch.FileRef),
			Metadata: map[string]any{
				"custodian_id": batch.CustodianID,
				"status":       batch.Status,
				"record_count": batch.RecordCount,
				"file_ref":     batch.FileRef,
			},
		})
		for _, r := range records {
			b.addEdge(LineageEdge{FromID: batch.ID, ToID: r.ID, EdgeType: "contains"})
		}
	}
}

// addReconResultsForEntity finds recon results that reference the entity.
func (s *service) addReconResultsForEntity(ctx context.Context, b *graphBuilder, entityType, entityID string) {
	results, err := s.store.ListReconResultsForEntity(ctx, entityType, entityID)
	if err != nil {
		return
	}
	for _, r := range results {
		b.addNode(LineageNode{
			ID:   r.ID,
			Type: "recon_result",
			Label: fmt.Sprintf("Recon: %s", r.Status),
			Metadata: map[string]any{
				"run_id":      r.RunID,
				"status":      string(r.Status),
				"reason_code": string(r.ReasonCode),
				"details":     r.Details,
			},
		})
		b.addEdge(LineageEdge{FromID: entityID, ToID: r.ID, EdgeType: "reconciled_as"})
	}
}

// addExceptionsForEntity finds exceptions that reference the entity.
func (s *service) addExceptionsForEntity(ctx context.Context, b *graphBuilder, entityType, entityID string) {
	exceptions, err := s.store.ListExceptionsForEntity(ctx, entityType, entityID)
	if err != nil {
		return
	}
	for _, e := range exceptions {
		b.addNode(LineageNode{
			ID:   e.ID,
			Type: "exception",
			Label: fmt.Sprintf("Exception: %s", e.ReasonCode),
			Metadata: map[string]any{
				"status":      e.Status,
				"reason_code": string(e.ReasonCode),
				"assigned_to": e.AssignedTo,
			},
		})
		b.addEdge(LineageEdge{FromID: entityID, ToID: e.ID, EdgeType: "raised_exception"})
	}
}

// addTrustScore adds the trust score for an entity if one exists.
func (s *service) addTrustScore(ctx context.Context, b *graphBuilder, entityType, entityID string) {
	ts, err := s.store.GetTrustScoreForEntity(ctx, entityType, entityID)
	if err != nil {
		return
	}
	b.addNode(LineageNode{
		ID:   ts.ID,
		Type: "trust_score",
		Label: fmt.Sprintf("Trust Score: %.1f", ts.Total),
		Metadata: map[string]any{
			"total":             ts.Total,
			"rationale_summary": ts.RationaleSummary,
			"computed_at":       ts.ComputedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
	b.addEdge(LineageEdge{FromID: entityID, ToID: ts.ID, EdgeType: "scored_as"})
}

// graphBuilder accumulates nodes and edges, deduplicating as it goes.
type graphBuilder struct {
	rootType string
	rootID   string
	nodeMap  map[string]LineageNode
	edgeSet  map[string]LineageEdge
}

func newBuilder(rootType, rootID string) *graphBuilder {
	return &graphBuilder{
		rootType: rootType,
		rootID:   rootID,
		nodeMap:  make(map[string]LineageNode),
		edgeSet:  make(map[string]LineageEdge),
	}
}

func (b *graphBuilder) addNode(n LineageNode) {
	b.nodeMap[n.ID] = n
}

func (b *graphBuilder) addEdge(e LineageEdge) {
	key := e.FromID + "|" + e.ToID + "|" + e.EdgeType
	b.edgeSet[key] = e
}

func (b *graphBuilder) build() *LineageGraph {
	nodes := make([]LineageNode, 0, len(b.nodeMap))
	for _, n := range b.nodeMap {
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Type != nodes[j].Type {
			return nodeTypeOrder(nodes[i].Type) < nodeTypeOrder(nodes[j].Type)
		}
		return nodes[i].ID < nodes[j].ID
	})

	edges := make([]LineageEdge, 0, len(b.edgeSet))
	for _, e := range b.edgeSet {
		edges = append(edges, e)
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].FromID != edges[j].FromID {
			return edges[i].FromID < edges[j].FromID
		}
		if edges[i].ToID != edges[j].ToID {
			return edges[i].ToID < edges[j].ToID
		}
		return edges[i].EdgeType < edges[j].EdgeType
	})

	return &LineageGraph{
		RootType: b.rootType,
		RootID:   b.rootID,
		Nodes:    nodes,
		Edges:    edges,
	}
}

// nodeTypeOrder defines a deterministic display order for node types.
func nodeTypeOrder(t string) int {
	switch t {
	case "ingestion_batch":
		return 0
	case "raw_record":
		return 1
	case "position":
		return 2
	case "transaction":
		return 3
	case "recon_result":
		return 4
	case "exception":
		return 5
	case "trust_score":
		return 6
	default:
		return 99
	}
}
