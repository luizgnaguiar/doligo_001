// Package item defines the domain model for products and services
// and the repository contract for its persistence.
package item

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ItemType differentiates between storable products and services.
type ItemType string

const (
	Storable ItemType = "STORABLE" // Represents a physical product with stock management.
	Service  ItemType = "SERVICE"  // Represents a service, like man-hours.
)

// Item represents the core entity for a product or service.
// It includes pricing information but no business logic for calculations.
type Item struct {
	ID          uuid.UUID
	Name        string
	Description string
	Type        ItemType
	CostPrice   float64 // Purchase price
	SalePrice   float64 // Selling price
	AverageCost float64 // Calculated average cost - NO CALCULATION IN THIS FASE
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
}

// SetCreatedBy sets the ID of the user who created the entity.
func (i *Item) SetCreatedBy(userID uuid.UUID) {
	i.CreatedBy = userID
}

// SetUpdatedBy sets the ID of the user who last updated the entity.
func (i *Item) SetUpdatedBy(userID uuid.UUID) {
	i.UpdatedBy = userID
}

// Repository defines the contract for data persistence operations for Items.
// It operates purely on Item domain entities.
type Repository interface {
	WithTx(tx *gorm.DB) Repository
	Create(ctx context.Context, item *Item) error
	GetByID(ctx context.Context, id uuid.UUID) (*Item, error)
	Update(ctx context.Context, item *Item) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*Item, error)
}
