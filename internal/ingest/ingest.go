package ingest

import (
	"context"

	"github.com/trustlot/trustlot/internal/model"
)

// Parser converts raw custodian data into normalized records.
type Parser interface {
	Parse(ctx context.Context, custodian model.Custodian, data []byte) ([]model.RawRecord, error)
}

// Pipeline orchestrates the ingestion of custodian feeds.
type Pipeline struct{}

// NewPipeline creates an ingestion pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{}
}
