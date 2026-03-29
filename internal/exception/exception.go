package exception

import (
	"context"

	"github.com/trustlot/trustlot/internal/model"
)

// Manager handles creation and retrieval of exceptions.
type Manager interface {
	Create(ctx context.Context, exc model.Exception) error
	List(ctx context.Context, status string, limit int) ([]model.Exception, error)
	Get(ctx context.Context, id string) (model.Exception, error)
}
