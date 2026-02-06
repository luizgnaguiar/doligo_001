// Package dto provides data transfer objects for API communication.
package dto

import (
	"time"
	"github.com/google/uuid"
	"doligo_001/internal/domain/stock"
	"doligo_001/internal/api/sanitizer"
)

// --- Warehouse DTOs ---

type CreateWarehouseRequest struct {
	Name string `json:"name" validate:"required,min=2,max=255"`
}

func (r *CreateWarehouseRequest) Sanitize() {
	r.Name = sanitizer.SanitizeString(r.Name)
}

type UpdateWarehouseRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=255"`
	IsActive bool   `json:"is_active"`
}

func (r *UpdateWarehouseRequest) Sanitize() {
	r.Name = sanitizer.SanitizeString(r.Name)
}

type WarehouseResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewWarehouseResponse(w *stock.Warehouse) *WarehouseResponse {
	return &WarehouseResponse{
		ID:        w.ID,
		Name:      w.Name,
		IsActive:  w.IsActive,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}


// --- Bin DTOs ---

type CreateBinRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=255"`
	WarehouseID string `json:"warehouse_id" validate:"required,uuid"`
}

func (r *CreateBinRequest) Sanitize() {
	r.Name = sanitizer.SanitizeString(r.Name)
}

type UpdateBinRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=255"`
	IsActive bool   `json:"is_active"`
}

func (r *UpdateBinRequest) Sanitize() {
	r.Name = sanitizer.SanitizeString(r.Name)
}

type BinResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	WarehouseID uuid.UUID `json:"warehouse_id"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewBinResponse(b *stock.Bin) *BinResponse {
	return &BinResponse{
		ID:          b.ID,
		Name:        b.Name,
		WarehouseID: b.WarehouseID,
		IsActive:    b.IsActive,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}


// --- Stock Movement DTOs ---

type CreateStockMovementRequest struct {
	ItemID      string  `json:"item_id" validate:"required,uuid"`
	WarehouseID string  `json:"warehouse_id" validate:"required,uuid"`
	BinID       string  `json:"bin_id" validate:"required,uuid"`
	Type        string  `json:"type" validate:"required,oneof=IN OUT"`
	Quantity    float64 `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"omitempty,ge=0"` // Required for IN movements to update CMP
	Reason      string  `json:"reason" validate:"max=255"`
}

func (r *CreateStockMovementRequest) Sanitize() {
	r.Reason = sanitizer.SanitizeString(r.Reason)
}

type StockMovementResponse struct {
	ID          uuid.UUID  `json:"id"`
	ItemID      uuid.UUID  `json:"item_id"`
	WarehouseID uuid.UUID  `json:"warehouse_id"`
	BinID       *uuid.UUID `json:"bin_id,omitempty"`
	Type        string     `json:"type"`
	Quantity    float64    `json:"quantity"`
	Reason      string     `json:"reason"`
	HappenedAt  time.Time  `json:"happened_at"`
	CreatedBy   uuid.UUID  `json:"created_by"`
}

func NewStockMovementResponse(sm *stock.StockMovement) *StockMovementResponse {
	return &StockMovementResponse{
		ID:          sm.ID,
		ItemID:      sm.ItemID,
		WarehouseID: sm.WarehouseID,
		BinID:       sm.BinID,
		Type:        string(sm.Type),
		Quantity:    sm.Quantity,
		Reason:      sm.Reason,
		HappenedAt:  sm.HappenedAt,
		CreatedBy:   sm.CreatedBy,
	}
}
