package invoice

import (
	"time"

	"github.com/google/uuid"
)

type Invoice struct {
	ID              uuid.UUID
	ThirdPartyID    uuid.UUID
	Number          string
	Date            time.Time
	TotalAmount     float64
	TotalCost       float64
	Lines           []InvoiceLine
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
	TotalAmount float64
	TotalCost   float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
}
