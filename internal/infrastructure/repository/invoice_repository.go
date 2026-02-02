package repository

import (
	"context"

	"doligo_001/internal/domain/thirdparty"
	"github.com/google/uuid"
	"doligo_001/internal/domain/invoice"
	"doligo_001/internal/infrastructure/db/models"
	"gorm.io/gorm"
)

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) *invoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) Create(ctx context.Context, domainInvoice *invoice.Invoice) error {
	modelInvoice := toInvoiceModel(domainInvoice)
	return r.db.WithContext(ctx).Create(modelInvoice).Error
}

func (r *invoiceRepository) FindByID(ctx context.Context, id uuid.UUID) (*invoice.Invoice, error) {
	var modelInvoice models.Invoice
	err := r.db.WithContext(ctx).Preload("Lines").First(&modelInvoice, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return toInvoiceDomain(&modelInvoice), nil
}

func (r *invoiceRepository) FindByIDWithDetails(ctx context.Context, id uuid.UUID) (*invoice.Invoice, error) {
	var modelInvoice models.Invoice
	err := r.db.WithContext(ctx).Preload("Lines").Preload("ThirdParty").First(&modelInvoice, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return toInvoiceDomain(&modelInvoice), nil
}

func toInvoiceModel(d *invoice.Invoice) *models.Invoice {
	lines := make([]models.InvoiceLine, len(d.Lines))
	for i, line := range d.Lines {
		lines[i] = *toInvoiceLineModel(&line)
	}

	return &models.Invoice{
		BaseModel: models.BaseModel{
			ID:        d.ID,
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
		ThirdPartyID: d.ThirdPartyID,
		Number:       d.Number,
		Date:         d.Date,
		TotalAmount:  d.TotalAmount,
		TotalCost:    d.TotalCost,
		Lines:        lines,
	}
}

func toInvoiceLineModel(d *invoice.InvoiceLine) *models.InvoiceLine {
	return &models.InvoiceLine{
		BaseModel: models.BaseModel{
			ID: d.ID,
		},
		InvoiceID:   d.InvoiceID,
		ItemID:      d.ItemID,
		Description: d.Description,
		Quantity:    d.Quantity,
		UnitPrice:   d.UnitPrice,
		UnitCost:    d.UnitCost,
		TotalAmount: d.TotalAmount,
		TotalCost:   d.TotalCost,
	}
}

func toInvoiceDomain(m *models.Invoice) *invoice.Invoice {
	lines := make([]invoice.InvoiceLine, len(m.Lines))
	for i, line := range m.Lines {
		lines[i] = *toInvoiceLineDomain(&line)
	}

	domainInvoice := &invoice.Invoice{
		ID:           m.ID,
		ThirdPartyID: m.ThirdPartyID,
		Number:       m.Number,
		Date:         m.Date,
		TotalAmount:  m.TotalAmount,
		TotalCost:    m.TotalCost,
		Lines:        lines,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
	
	// Map ThirdParty if it was loaded
	if m.ThirdParty.ID != uuid.Nil {
		domainInvoice.ThirdParty = toThirdPartyDomain(&m.ThirdParty)
	}

	return domainInvoice
}

func toInvoiceLineDomain(m *models.InvoiceLine) *invoice.InvoiceLine {
	return &invoice.InvoiceLine{
		ID:          m.ID,
		InvoiceID:   m.InvoiceID,
		ItemID:      m.ItemID,
		Description: m.Description,
		Quantity:    m.Quantity,
		UnitPrice:   m.UnitPrice,
		UnitCost:    m.UnitCost,
		TotalAmount: m.TotalAmount,
		TotalCost:   m.TotalCost,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// toThirdPartyDomain is a helper function to convert the ThirdParty model to a domain entity.
// This needs to be defined or imported. Assuming it exists in a relevant package.
func toThirdPartyDomain(m *models.ThirdParty) *thirdparty.ThirdParty {
	return &thirdparty.ThirdParty{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		Type:      thirdparty.ThirdPartyType(m.Type),
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		CreatedBy: m.CreatedBy,
		UpdatedBy: m.UpdatedBy,
	}
}
