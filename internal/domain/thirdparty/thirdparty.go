// Package thirdparty defines the domain model for third parties (customers and suppliers)
// and the repository contract for its persistence, following Clean Architecture principles.
package thirdparty

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ThirdPartyType differentiates between customers, suppliers, and other types of third parties.
type ThirdPartyType string

const (
	Customer ThirdPartyType = "CUSTOMER"
	Supplier ThirdPartyType = "SUPPLIER"
)

// ThirdParty represents the core entity for a customer or a supplier.
// It is a pure domain model with no infrastructure-specific details.
type ThirdParty struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Type      ThirdPartyType
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

// SetCreatedBy sets the ID of the user who created the entity.
func (t *ThirdParty) SetCreatedBy(userID uuid.UUID) {
	t.CreatedBy = userID
}

// SetUpdatedBy sets the ID of the user who last updated the entity.
func (t *ThirdParty) SetUpdatedBy(userID uuid.UUID) {
	t.UpdatedBy = userID
}

// Repository defines the contract for data persistence operations for ThirdParties.
// It operates purely on ThirdParty domain entities.
type Repository interface {
	Create(ctx context.Context, thirdParty *ThirdParty) error
	GetByID(ctx context.Context, id uuid.UUID) (*ThirdParty, error)
	Update(ctx context.Context, thirdParty *ThirdParty) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*ThirdParty, error)
}
