package invoice

import (
	"context"

	"github.com/google/uuid"
	"doligo_001/internal/domain/invoice"
)

type Repository interface {
	Create(ctx context.Context, invoice *invoice.Invoice) error
	FindByID(ctx context.Context, id uuid.UUID) (*invoice.Invoice, error)
}
