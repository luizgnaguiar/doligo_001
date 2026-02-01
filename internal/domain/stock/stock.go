// Package stock defines the domain models for inventory management,
// including warehouses, bins, stock movements, and the immutable stock ledger.
package stock

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Warehouse represents a physical or logical location where stock is held.
type Warehouse struct {
	ID        uuid.UUID
	Name      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

func (w *Warehouse) SetCreatedBy(userID uuid.UUID) {
	w.CreatedBy = userID
}

func (w *Warehouse) SetUpdatedBy(userID uuid.UUID) {
	w.UpdatedBy = userID
}

// Bin represents a specific storage location within a Warehouse.
type Bin struct {
	ID          uuid.UUID
	WarehouseID uuid.UUID
	Name        string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
}

func (b *Bin) SetCreatedBy(userID uuid.UUID) {
	b.CreatedBy = userID
}

func (b *Bin) SetUpdatedBy(userID uuid.UUID) {
	b.UpdatedBy = userID
}

// MovementType defines the direction of a stock movement.
type MovementType string

const (
	MovementTypeIn  MovementType = "IN"
	MovementTypeOut MovementType = "OUT"
)

// Stock represents the quantity of a specific item in a specific location.
type Stock struct {
    ItemID      uuid.UUID
    WarehouseID uuid.UUID
    BinID       *uuid.UUID // Optional
    Quantity    float64
    UpdatedAt   time.Time
}


// StockMovement represents the record of an item moving into or out of a stock location.
// This is the primary entity for transactional stock operations.
type StockMovement struct {
	ID          uuid.UUID
	ItemID      uuid.UUID
	WarehouseID uuid.UUID
	BinID       *uuid.UUID // Optional, if bin tracking is used
	Type        MovementType
	Quantity    float64
	Reason      string
	HappenedAt  time.Time
	CreatedBy   uuid.UUID
}
// Note: StockMovement is not fully auditable in the sense of CreatedAt/UpdatedAt,
// as it's a point-in-time record. It only has CreatedBy.

func (sm *StockMovement) SetCreatedBy(userID uuid.UUID) {
	sm.CreatedBy = userID
}


// StockLedger is an immutable, append-only record of a stock movement.
// It serves as the ultimate source of truth for all stock history.
type StockLedger struct {
	ID              uuid.UUID
	StockMovementID uuid.UUID
	ItemID          uuid.UUID
	WarehouseID     uuid.UUID
	BinID           *uuid.UUID
	MovementType    MovementType
	QuantityChange  float64
	QuantityBefore  float64
	QuantityAfter   float64
	Reason          string
	HappenedAt      time.Time
	RecordedAt      time.Time
	RecordedBy      uuid.UUID
}


// WarehouseRepository defines the contract for warehouse data persistence.
type WarehouseRepository interface {
	Create(ctx context.Context, warehouse *Warehouse) error
	GetByID(ctx context.Context, id uuid.UUID) (*Warehouse, error)
	Update(ctx context.Context, warehouse *Warehouse) error
	List(ctx context.Context) ([]*Warehouse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// BinRepository defines the contract for bin data persistence.
type BinRepository interface {
	Create(ctx context.Context, bin *Bin) error
	GetByID(ctx context.Context, id uuid.UUID) (*Bin, error)
	Update(ctx context.Context, bin *Bin) error
	ListByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]*Bin, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// StockMovementRepository defines the contract for creating stock movements,
// which must be handled transactionally.
type StockMovementRepository interface {
    Create(ctx context.Context, movement *StockMovement) error
}

// StockLedgerRepository defines the contract for creating immutable ledger entries.
type StockLedgerRepository interface {
	Create(ctx context.Context, entry *StockLedger) error
}

// StockRepository defines the contract for stock-related queries and updates, including pessimistic locking.
type StockRepository interface {
    GetStock(ctx context.Context, itemID, warehouseID uuid.UUID, binID *uuid.UUID) (*Stock, error)
    GetStockForUpdate(ctx context.Context, itemID, warehouseID uuid.UUID, binID *uuid.UUID) (*Stock, error)
    UpsertStock(ctx context.Context, stock *Stock) error
}
