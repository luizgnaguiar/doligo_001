// Package dto provides data transfer objects for API communication.
package dto

import (
	"time"
	"github.com/google/uuid"
	"doligo_001/internal/domain/item"
	"doligo_001/internal/api/sanitizer"
)

// CreateItemRequest defines the structure for creating a new item.
type CreateItemRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=255"`
	Description string  `json:"description"`
	Type        string  `json:"type" validate:"required,oneof=STORABLE SERVICE"`
	CostPrice   float64 `json:"cost_price" validate:"gte=0"`
	SalePrice   float64 `json:"sale_price" validate:"gte=0"`
}

func (r *CreateItemRequest) Sanitize() {
	r.Name = sanitizer.SanitizeString(r.Name)
	r.Description = sanitizer.SanitizeString(r.Description)
}

// UpdateItemRequest defines the structure for updating an existing item.
type UpdateItemRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=255"`
	Description string  `json:"description"`
	Type        string  `json:"type" validate:"required,oneof=STORABLE SERVICE"`
	CostPrice   float64 `json:"cost_price" validate:"gte=0"`
	SalePrice   float64 `json:"sale_price" validate:"gte=0"`
	IsActive    bool    `json:"is_active"`
}

func (r *UpdateItemRequest) Sanitize() {
	r.Name = sanitizer.SanitizeString(r.Name)
	r.Description = sanitizer.SanitizeString(r.Description)
}

// ItemResponse defines the structure for an item response.
type ItemResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	CostPrice   float64   `json:"cost_price"`
	SalePrice   float64   `json:"sale_price"`
	AverageCost float64   `json:"average_cost"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   uuid.UUID `json:"created_by"`
	UpdatedBy   uuid.UUID `json:"updated_by"`
}

// NewItemResponse creates a response DTO from a domain entity.
func NewItemResponse(i *item.Item) *ItemResponse {
	return &ItemResponse{
		ID:          i.ID,
		Name:        i.Name,
		Description: i.Description,
		Type:        string(i.Type),
		CostPrice:   i.CostPrice,
		SalePrice:   i.SalePrice,
		AverageCost: i.AverageCost,
		IsActive:    i.IsActive,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
		CreatedBy:   i.CreatedBy,
		UpdatedBy:   i.UpdatedBy,
	}
}
