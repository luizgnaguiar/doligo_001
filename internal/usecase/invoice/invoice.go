package invoice

import (
	"context"

	"doligo_001/internal/api/dto"
	"github.com/google/uuid"
	"doligo_001/internal/domain/invoice"
)

type Usecase interface {
	Create(ctx context.Context, req *dto.CreateInvoiceRequest) (*invoice.Invoice, error)
	GetByID(ctx context.Context, id uuid.UUID) (*invoice.Invoice, error)
	QueueInvoicePDFGeneration(ctx context.Context, invoiceID uuid.UUID) error
}

type Repository interface {
	Create(ctx context.Context, invoice *invoice.Invoice) error
	Update(ctx context.Context, invoice *invoice.Invoice) error
	FindByID(ctx context.Context, id uuid.UUID) (*invoice.Invoice, error)
	FindByIDWithDetails(ctx context.Context, id uuid.UUID) (*invoice.Invoice, error)
}
