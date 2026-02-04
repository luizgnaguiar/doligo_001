package dto

import (
	"time"

	"github.com/google/uuid"
	"doligo_001/internal/api/sanitizer"
)

type CreateInvoiceRequest struct {
	ThirdPartyID string                `json:"third_party_id" validate:"required,uuid"`
	Number       string                `json:"number" validate:"required"`
	Date         string                `json:"date" validate:"required,datetime=2006-01-02"`
	Lines        []CreateInvoiceLineRequest `json:"lines" validate:"required,min=1"`
}

func (r *CreateInvoiceRequest) Sanitize() {
	r.Number = sanitizer.SanitizeString(r.Number)
	for i := range r.Lines {
		r.Lines[i].Sanitize()
	}
}

type CreateInvoiceLineRequest struct {
	ItemID      string  `json:"item_id" validate:"required,uuid"`
	Description string  `json:"description" validate:"required"`
	Quantity    float64 `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gte=0"`
}

func (r *CreateInvoiceLineRequest) Sanitize() {
	r.Description = sanitizer.SanitizeString(r.Description)
}

type InvoiceResponse struct {
	ID           uuid.UUID           `json:"id"`
	ThirdPartyID uuid.UUID           `json:"third_party_id"`
	Number       string              `json:"number"`
	Date         time.Time           `json:"date"`
	TotalAmount  float64             `json:"total_amount"`
	TotalCost    float64             `json:"total_cost"`
	Lines        []InvoiceLineResponse `json:"lines"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

type InvoiceLineResponse struct {
	ID          uuid.UUID `json:"id"`
	ItemID      uuid.UUID `json:"item_id"`
	Description string    `json:"description"`
	Quantity    float64   `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	UnitCost    float64   `json:"unit_cost"`
	TotalAmount float64   `json:"total_amount"`
	TotalCost   float64   `json:"total_cost"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
