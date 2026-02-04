package bom

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrBOMNotFound is returned when a BillOfMaterials is not found.
	ErrBOMNotFound = errors.New("bill of materials not found")
)

// BillOfMaterials represents the definition of how to produce a finished item.
// It consists of components (inputs) and services required.
type BillOfMaterials struct {
	ID         uuid.UUID
	ProductID  uuid.UUID // The finished item produced by this BOM
	Name       string
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	CreatedBy  uuid.UUID
	UpdatedBy  uuid.UUID
	Components []BillOfMaterialsComponent // Inputs needed
}

// BillOfMaterialsComponent represents a single ingredient (item or service) in a BOM.
type BillOfMaterialsComponent struct {
	ID                uuid.UUID
	BillOfMaterialsID uuid.UUID
	ComponentItemID   uuid.UUID // The item (input or service) that is a component
	Quantity          float64   // Quantity of the component needed per unit of ProductID
	UnitOfMeasure     string    // Unit of measure for the quantity (e.g., "kg", "pcs", "hours")
	IsActive          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CreatedBy         uuid.UUID
	UpdatedBy         uuid.UUID
}

// ProductionRecord represents a completed production run based on a BOM.
type ProductionRecord struct {
	ID                    uuid.UUID
	BillOfMaterialsID     uuid.UUID
	ProducedProductID     uuid.UUID // The finished product item ID
	ProductionQuantity    float64
	ActualProductionCost  float64
	WarehouseID           uuid.UUID
	ProducedAt            time.Time
	CreatedBy             uuid.UUID
}

// Ensure BillOfMaterials implements the Auditable interface
func (b *BillOfMaterials) SetCreatedBy(userID uuid.UUID) {
	b.CreatedBy = userID
}

// SetUpdatedBy sets the ID of the user who last updated the entity.
func (b *BillOfMaterials) SetUpdatedBy(userID uuid.UUID) {
	b.UpdatedAt = time.Now()
	b.UpdatedBy = userID
}

// Ensure BillOfMaterialsComponent implements the Auditable interface
func (b *BillOfMaterialsComponent) SetCreatedBy(userID uuid.UUID) {
	b.CreatedBy = userID
}

// SetUpdatedBy sets the ID of the user who last updated the entity.
func (b *BillOfMaterialsComponent) SetUpdatedBy(userID uuid.UUID) {
	b.UpdatedAt = time.Now()
	b.UpdatedBy = userID
}

// SetCreatedBy sets the ID of the user who created the ProductionRecord.
func (pr *ProductionRecord) SetCreatedBy(userID uuid.UUID) {
	pr.CreatedBy = userID
}

// Repository defines the contract for data persistence operations for BillOfMaterials.
type Repository interface {
	WithTx(tx *gorm.DB) Repository
	Create(ctx context.Context, bom *BillOfMaterials) error
	GetByID(ctx context.Context, id uuid.UUID) (*BillOfMaterials, error)
	GetByProductID(ctx context.Context, productID uuid.UUID) (*BillOfMaterials, error)
	Update(ctx context.Context, bom *BillOfMaterials) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*BillOfMaterials, error)
}

// ProductionRecordRepository defines the contract for data persistence operations for ProductionRecords.
type ProductionRecordRepository interface {
	WithTx(tx *gorm.DB) ProductionRecordRepository
	Create(ctx context.Context, record *ProductionRecord) error
}
