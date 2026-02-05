package invoice

import (
	"time"
	"doligo_001/internal/domain/thirdparty"
	"github.com/google/uuid"
)

type Invoice struct {
	ID              uuid.UUID
	ThirdPartyID    uuid.UUID
	ThirdParty      *thirdparty.ThirdParty `gorm:"foreignKey:ThirdPartyID"`
	Number          string
	Date            time.Time
	TotalAmount     float64
	TotalCost       float64
	TotalTax        float64
	Lines           []InvoiceLine
	PDFStatus       string
	PDFUrl          string
	PDFErrorMessage string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedBy       uuid.UUID
	UpdatedBy       uuid.UUID
}

func (i *Invoice) SetCreatedBy(userID uuid.UUID) {
	i.CreatedBy = userID
}

func (i *Invoice) SetUpdatedBy(userID uuid.UUID) {
	i.UpdatedBy = userID
}

type InvoiceLine struct {
	ID          uuid.UUID
	InvoiceID   uuid.UUID
	ItemID      uuid.UUID
	Description string
	Quantity    float64
	UnitPrice   float64
	UnitCost    float64
	TaxRate     float64
	TaxAmount   float64
	NetPrice    float64
	TotalAmount float64
	TotalCost   float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
}
