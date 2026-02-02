package item

import (
	"context"

	"github.com/google/uuid"
	"doligo_001/internal/domain/item"
)

type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*item.Item, error)
}
