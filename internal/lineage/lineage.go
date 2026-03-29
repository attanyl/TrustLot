package lineage

import "context"

// LineageGraph is the full traversal result for an entity.
type LineageGraph struct {
	RootType string       `json:"root_type"`
	RootID   string       `json:"root_id"`
	Nodes    []LineageNode `json:"nodes"`
	Edges    []LineageEdge `json:"edges"`
}

// LineageNode is a single entity in the lineage graph.
type LineageNode struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Label    string         `json:"label"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// LineageEdge connects two nodes in the lineage graph.
type LineageEdge struct {
	FromID   string `json:"from_id"`
	ToID     string `json:"to_id"`
	EdgeType string `json:"edge_type"`
}

// Service assembles lineage graphs for entities.
type Service interface {
	GetLineageForEntity(ctx context.Context, entityType string, entityID string) (*LineageGraph, error)
	GetLineageForException(ctx context.Context, exceptionID string) (*LineageGraph, error)
}
