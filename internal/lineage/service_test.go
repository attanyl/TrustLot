package lineage

import (
	"sort"
	"testing"
)

func TestGraphBuilderDeduplicatesNodes(t *testing.T) {
	b := newBuilder("position", "pos-1")

	b.addNode(LineageNode{ID: "a", Type: "position", Label: "Position A"})
	b.addNode(LineageNode{ID: "a", Type: "position", Label: "Position A updated"})
	b.addNode(LineageNode{ID: "b", Type: "recon_result", Label: "Recon B"})

	g := b.build()
	if len(g.Nodes) != 2 {
		t.Errorf("expected 2 unique nodes, got %d", len(g.Nodes))
	}
	// Second add should overwrite label.
	for _, n := range g.Nodes {
		if n.ID == "a" && n.Label != "Position A updated" {
			t.Errorf("expected updated label, got %q", n.Label)
		}
	}
}

func TestGraphBuilderDeduplicatesEdges(t *testing.T) {
	b := newBuilder("position", "pos-1")

	b.addEdge(LineageEdge{FromID: "a", ToID: "b", EdgeType: "reconciled_as"})
	b.addEdge(LineageEdge{FromID: "a", ToID: "b", EdgeType: "reconciled_as"})
	b.addEdge(LineageEdge{FromID: "a", ToID: "c", EdgeType: "raised_exception"})

	g := b.build()
	if len(g.Edges) != 2 {
		t.Errorf("expected 2 unique edges, got %d", len(g.Edges))
	}
}

func TestGraphBuilderDeterministicNodeOrder(t *testing.T) {
	b := newBuilder("position", "pos-1")

	// Add in random order.
	b.addNode(LineageNode{ID: "exc-1", Type: "exception", Label: "Exc"})
	b.addNode(LineageNode{ID: "pos-1", Type: "position", Label: "Position"})
	b.addNode(LineageNode{ID: "rr-1", Type: "recon_result", Label: "Recon"})
	b.addNode(LineageNode{ID: "batch-1", Type: "ingestion_batch", Label: "Batch"})
	b.addNode(LineageNode{ID: "ts-1", Type: "trust_score", Label: "Trust"})
	b.addNode(LineageNode{ID: "raw-1", Type: "raw_record", Label: "Raw"})

	g := b.build()

	expectedOrder := []string{
		"ingestion_batch",
		"raw_record",
		"position",
		"recon_result",
		"exception",
		"trust_score",
	}

	for i, n := range g.Nodes {
		if n.Type != expectedOrder[i] {
			t.Errorf("node[%d] type = %q, want %q", i, n.Type, expectedOrder[i])
		}
	}
}

func TestGraphBuilderDeterministicEdgeOrder(t *testing.T) {
	b := newBuilder("position", "pos-1")

	b.addEdge(LineageEdge{FromID: "z", ToID: "a", EdgeType: "raised_exception"})
	b.addEdge(LineageEdge{FromID: "a", ToID: "c", EdgeType: "scored_as"})
	b.addEdge(LineageEdge{FromID: "a", ToID: "b", EdgeType: "reconciled_as"})

	g := b.build()

	// Should be sorted by from_id, then to_id, then edge_type.
	if g.Edges[0].FromID != "a" || g.Edges[0].ToID != "b" {
		t.Errorf("edges not sorted: first edge = %+v", g.Edges[0])
	}
	if g.Edges[1].FromID != "a" || g.Edges[1].ToID != "c" {
		t.Errorf("edges not sorted: second edge = %+v", g.Edges[1])
	}
	if g.Edges[2].FromID != "z" || g.Edges[2].ToID != "a" {
		t.Errorf("edges not sorted: third edge = %+v", g.Edges[2])
	}
}

func TestGraphBuilderStableOutput(t *testing.T) {
	// Same input should always produce identical output.
	for run := 0; run < 10; run++ {
		b := newBuilder("transaction", "txn-1")

		b.addNode(LineageNode{ID: "exc-1", Type: "exception", Label: "Exc"})
		b.addNode(LineageNode{ID: "txn-1", Type: "transaction", Label: "Txn"})
		b.addNode(LineageNode{ID: "rr-1", Type: "recon_result", Label: "Recon"})
		b.addNode(LineageNode{ID: "ts-1", Type: "trust_score", Label: "Trust"})

		b.addEdge(LineageEdge{FromID: "txn-1", ToID: "rr-1", EdgeType: "reconciled_as"})
		b.addEdge(LineageEdge{FromID: "txn-1", ToID: "exc-1", EdgeType: "raised_exception"})
		b.addEdge(LineageEdge{FromID: "txn-1", ToID: "ts-1", EdgeType: "scored_as"})

		g := b.build()

		if g.RootType != "transaction" || g.RootID != "txn-1" {
			t.Fatalf("root mismatch")
		}
		if len(g.Nodes) != 4 {
			t.Fatalf("node count %d, want 4", len(g.Nodes))
		}
		if len(g.Edges) != 3 {
			t.Fatalf("edge count %d, want 3", len(g.Edges))
		}

		// Verify node order is the same every time.
		expectedTypes := []string{"transaction", "recon_result", "exception", "trust_score"}
		for i, n := range g.Nodes {
			if n.Type != expectedTypes[i] {
				t.Errorf("run %d: node[%d] type = %q, want %q", run, i, n.Type, expectedTypes[i])
			}
		}
	}
}

func TestNodeTypeOrder(t *testing.T) {
	types := []string{"trust_score", "exception", "position", "ingestion_batch", "raw_record", "recon_result", "transaction"}
	sort.Slice(types, func(i, j int) bool {
		return nodeTypeOrder(types[i]) < nodeTypeOrder(types[j])
	})

	expected := []string{"ingestion_batch", "raw_record", "position", "transaction", "recon_result", "exception", "trust_score"}
	for i, typ := range types {
		if typ != expected[i] {
			t.Errorf("sorted[%d] = %q, want %q", i, typ, expected[i])
		}
	}
}

func TestPositionFullChain(t *testing.T) {
	// Simulate what the service builds for a position with full chain.
	b := newBuilder("position", "pos-1")

	b.addNode(LineageNode{ID: "batch-1", Type: "ingestion_batch", Label: "Batch: northriver.csv"})
	b.addNode(LineageNode{ID: "raw-1", Type: "raw_record", Label: "Raw Record #1"})
	b.addNode(LineageNode{ID: "pos-1", Type: "position", Label: "Position: custodian qty=150"})
	b.addNode(LineageNode{ID: "rr-1", Type: "recon_result", Label: "Recon: BREAK"})
	b.addNode(LineageNode{ID: "exc-1", Type: "exception", Label: "Exception: POS_QUANTITY_MISMATCH"})
	b.addNode(LineageNode{ID: "ts-1", Type: "trust_score", Label: "Trust Score: 25.3"})

	b.addEdge(LineageEdge{FromID: "batch-1", ToID: "raw-1", EdgeType: "contains"})
	b.addEdge(LineageEdge{FromID: "raw-1", ToID: "pos-1", EdgeType: "normalized_to"})
	b.addEdge(LineageEdge{FromID: "pos-1", ToID: "rr-1", EdgeType: "reconciled_as"})
	b.addEdge(LineageEdge{FromID: "pos-1", ToID: "exc-1", EdgeType: "raised_exception"})
	b.addEdge(LineageEdge{FromID: "pos-1", ToID: "ts-1", EdgeType: "scored_as"})

	g := b.build()

	if len(g.Nodes) != 6 {
		t.Fatalf("expected 6 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 5 {
		t.Fatalf("expected 5 edges, got %d", len(g.Edges))
	}

	// First node should be batch (lowest order).
	if g.Nodes[0].Type != "ingestion_batch" {
		t.Errorf("first node type = %q, want ingestion_batch", g.Nodes[0].Type)
	}
	// Last node should be trust_score (highest order).
	if g.Nodes[5].Type != "trust_score" {
		t.Errorf("last node type = %q, want trust_score", g.Nodes[5].Type)
	}
}

func TestTransactionFullChain(t *testing.T) {
	b := newBuilder("transaction", "txn-4")

	b.addNode(LineageNode{ID: "batch-2", Type: "ingestion_batch", Label: "Batch: atlas_positions"})
	b.addNode(LineageNode{ID: "txn-4", Type: "transaction", Label: "Transaction: buy custodian qty=100"})
	b.addNode(LineageNode{ID: "rr-6", Type: "recon_result", Label: "Recon: BREAK"})
	b.addNode(LineageNode{ID: "exc-2", Type: "exception", Label: "Exception: TXN_DUPLICATE"})

	b.addEdge(LineageEdge{FromID: "txn-4", ToID: "rr-6", EdgeType: "reconciled_as"})
	b.addEdge(LineageEdge{FromID: "txn-4", ToID: "exc-2", EdgeType: "raised_exception"})

	g := b.build()

	if g.RootType != "transaction" {
		t.Errorf("root_type = %q, want transaction", g.RootType)
	}
	if len(g.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(g.Nodes))
	}
}

func TestPartialLineageMissingTrustScore(t *testing.T) {
	// Simulate lineage where trust score hasn't been computed yet.
	b := newBuilder("position", "pos-5")

	b.addNode(LineageNode{ID: "pos-5", Type: "position", Label: "Position: internal qty=1000"})
	b.addNode(LineageNode{ID: "rr-3", Type: "recon_result", Label: "Recon: MATCH"})

	b.addEdge(LineageEdge{FromID: "pos-5", ToID: "rr-3", EdgeType: "reconciled_as"})

	g := b.build()

	// No trust_score or exception nodes — partial lineage is valid.
	if len(g.Nodes) != 2 {
		t.Fatalf("expected 2 nodes (partial), got %d", len(g.Nodes))
	}
	if len(g.Edges) != 1 {
		t.Fatalf("expected 1 edge (partial), got %d", len(g.Edges))
	}

	// Verify no trust_score type in nodes.
	for _, n := range g.Nodes {
		if n.Type == "trust_score" {
			t.Error("should not contain trust_score in partial lineage")
		}
	}
}

func TestExceptionLineageChain(t *testing.T) {
	b := newBuilder("exception", "exc-1")

	b.addNode(LineageNode{ID: "exc-1", Type: "exception", Label: "Exception: POS_QUANTITY_MISMATCH"})
	b.addNode(LineageNode{ID: "pos-1", Type: "position", Label: "Position: custodian qty=150"})
	b.addNode(LineageNode{ID: "rr-1", Type: "recon_result", Label: "Recon: BREAK"})
	b.addNode(LineageNode{ID: "ts-1", Type: "trust_score", Label: "Trust Score: 25.3"})

	b.addEdge(LineageEdge{FromID: "pos-1", ToID: "rr-1", EdgeType: "reconciled_as"})
	b.addEdge(LineageEdge{FromID: "pos-1", ToID: "exc-1", EdgeType: "raised_exception"})
	b.addEdge(LineageEdge{FromID: "rr-1", ToID: "exc-1", EdgeType: "raised_exception"})
	b.addEdge(LineageEdge{FromID: "pos-1", ToID: "ts-1", EdgeType: "scored_as"})

	g := b.build()

	if g.RootType != "exception" {
		t.Errorf("root_type = %q, want exception", g.RootType)
	}
	if len(g.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 4 {
		t.Fatalf("expected 4 edges, got %d", len(g.Edges))
	}
}
