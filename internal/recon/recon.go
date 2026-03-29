package recon

import "context"

// Engine runs reconciliation for a specific entity type.
type Engine interface {
	Run(ctx context.Context, runID string) error
}
