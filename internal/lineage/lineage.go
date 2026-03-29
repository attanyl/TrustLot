package lineage

import (
	"context"

	"github.com/trustlot/trustlot/internal/model"
)

// Tracker records and queries lineage edges.
type Tracker interface {
	Record(ctx context.Context, edge model.LineageEdge) error
	Trace(ctx context.Context, entityType, entityID string) ([]model.LineageEdge, error)
}
