// Package repository provides the GORM-based implementation of the repository
// interfaces defined in the domain layer.
package repository

import (
	"context"

	"doligo_001/internal/domain/stock"
	"doligo_001/internal/infrastructure/db/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// gormWarehouseRepository is a GORM implementation of the stock.WarehouseRepository.
type gormWarehouseRepository struct {
	db *gorm.DB
}

func (r *gormWarehouseRepository) WithTx(tx *gorm.DB) stock.WarehouseRepository {
	return NewGormWarehouseRepository(tx)
}

// NewGormWarehouseRepository creates a new gormWarehouseRepository.
func NewGormWarehouseRepository(db *gorm.DB) stock.WarehouseRepository {
	return &gormWarehouseRepository{db: db}
}

func (r *gormWarehouseRepository) Create(ctx context.Context, w *stock.Warehouse) error {
	model := fromWarehouseDomainEntity(w)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormWarehouseRepository) GetByID(ctx context.Context, id uuid.UUID) (*stock.Warehouse, error) {
	var model models.Warehouse
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return toWarehouseDomainEntity(&model), nil
}

func (r *gormWarehouseRepository) Update(ctx context.Context, w *stock.Warehouse) error {
	model := fromWarehouseDomainEntity(w)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *gormWarehouseRepository) List(ctx context.Context) ([]*stock.Warehouse, error) {
	var modelList []models.Warehouse
	if err := r.db.WithContext(ctx).Find(&modelList).Error; err != nil {
		return nil, err
	}
	domainList := make([]*stock.Warehouse, len(modelList))
	for i, model := range modelList {
		domainList[i] = toWarehouseDomainEntity(&model)
	}
	return domainList, nil
}

func (r *gormWarehouseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Warehouse{}, "id = ?", id).Error
}

// gormBinRepository is a GORM implementation of the stock.BinRepository.
type gormBinRepository struct {
	db *gorm.DB
}

func (r *gormBinRepository) WithTx(tx *gorm.DB) stock.BinRepository {
	return NewGormBinRepository(tx)
}

// NewGormBinRepository creates a new gormBinRepository.
func NewGormBinRepository(db *gorm.DB) stock.BinRepository {
	return &gormBinRepository{db: db}
}

func (r *gormBinRepository) Create(ctx context.Context, b *stock.Bin) error {
	model := fromBinDomainEntity(b)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *gormBinRepository) GetByID(ctx context.Context, id uuid.UUID) (*stock.Bin, error) {
	var model models.Bin
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return toBinDomainEntity(&model), nil
}

func (r *gormBinRepository) Update(ctx context.Context, b *stock.Bin) error {
	model := fromBinDomainEntity(b)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *gormBinRepository) ListByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]*stock.Bin, error) {
	var modelList []models.Bin
	if err := r.db.WithContext(ctx).Where("warehouse_id = ?", warehouseID).Find(&modelList).Error; err != nil {
		return nil, err
	}
	domainList := make([]*stock.Bin, len(modelList))
	for i, model := range modelList {
		domainList[i] = toBinDomainEntity(&model)
	}
	return domainList, nil
}

func (r *gormBinRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Bin{}, "id = ?", id).Error
}

// gormStockRepository is a GORM implementation of the stock.StockRepository.
type gormStockRepository struct {
	db *gorm.DB
}

func (r *gormStockRepository) WithTx(tx *gorm.DB) stock.StockRepository {
	return NewGormStockRepository(tx)
}

// NewGormStockRepository creates a new gormStockRepository.
func NewGormStockRepository(db *gorm.DB) stock.StockRepository {
	return &gormStockRepository{db: db}
}

func (r *gormStockRepository) GetStock(ctx context.Context, itemID, warehouseID uuid.UUID, binID *uuid.UUID) (*stock.Stock, error) {
	var model models.Stock
	query := r.db.WithContext(ctx).Where("item_id = ? AND warehouse_id = ?", itemID, warehouseID)
	if binID != nil {
		query = query.Where("bin_id = ?", *binID)
	} else {
		query = query.Where("bin_id = ?", uuid.Nil)
	}
	if err := query.First(&model).Error; err != nil {
		return nil, err
	}
	return toStockDomainEntity(&model), nil
}

func (r *gormStockRepository) GetStockForUpdate(ctx context.Context, itemID, warehouseID uuid.UUID, binID *uuid.UUID) (*stock.Stock, error) {
	var model models.Stock
	query := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("item_id = ? AND warehouse_id = ?", itemID, warehouseID)
	if binID != nil {
		query = query.Where("bin_id = ?", *binID)
	} else {
		query = query.Where("bin_id = ?", uuid.Nil)
	}
	if err := query.First(&model).Error; err != nil {
		return nil, err
	}
	return toStockDomainEntity(&model), nil
}

func (r *gormStockRepository) UpsertStock(ctx context.Context, s *stock.Stock) error {
	model := fromStockDomainEntity(s)
	// Use Clauses(clause.OnConflict) to perform an upsert.
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "item_id"}, {Name: "warehouse_id"}, {Name: "bin_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"quantity", "updated_at"}),
	}).Create(model).Error
}

// gormStockMovementRepository is a GORM implementation of the stock.StockMovementRepository.
type gormStockMovementRepository struct {
	db *gorm.DB
}

func (r *gormStockMovementRepository) WithTx(tx *gorm.DB) stock.StockMovementRepository {
	return NewGormStockMovementRepository(tx)
}

// NewGormStockMovementRepository creates a new gormStockMovementRepository.
func NewGormStockMovementRepository(db *gorm.DB) stock.StockMovementRepository {
	return &gormStockMovementRepository{db: db}
}

func (r *gormStockMovementRepository) Create(ctx context.Context, sm *stock.StockMovement) error {
	model := fromStockMovementDomainEntity(sm)
	return r.db.WithContext(ctx).Create(model).Error
}

// gormStockLedgerRepository is a GORM implementation of the stock.StockLedgerRepository.
type gormStockLedgerRepository struct {
	db *gorm.DB
}

func (r *gormStockLedgerRepository) WithTx(tx *gorm.DB) stock.StockLedgerRepository {
	return NewGormStockLedgerRepository(tx)
}

// NewGormStockLedgerRepository creates a new gormStockLedgerRepository.
func NewGormStockLedgerRepository(db *gorm.DB) stock.StockLedgerRepository {
	return &gormStockLedgerRepository{db: db}
}

func (r *gormStockLedgerRepository) Create(ctx context.Context, sl *stock.StockLedger) error {
	model := fromStockLedgerDomainEntity(sl)
	return r.db.WithContext(ctx).Create(model).Error
}


// --- MAPPING FUNCTIONS ---

func toWarehouseDomainEntity(model *models.Warehouse) *stock.Warehouse {
	return &stock.Warehouse{
		ID:        model.ID,
		Name:      model.Name,
		IsActive:  model.IsActive,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		CreatedBy: model.CreatedBy,
		UpdatedBy: model.UpdatedBy,
	}
}

func fromWarehouseDomainEntity(entity *stock.Warehouse) *models.Warehouse {
	return &models.Warehouse{
		BaseModel: models.BaseModel{
			ID:        entity.ID,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			CreatedBy: entity.CreatedBy,
			UpdatedBy: entity.UpdatedBy,
		},
		Name:     entity.Name,
		IsActive: entity.IsActive,
	}
}

func toBinDomainEntity(model *models.Bin) *stock.Bin {
	return &stock.Bin{
		ID:          model.ID,
		WarehouseID: model.WarehouseID,
		Name:        model.Name,
		IsActive:    model.IsActive,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		CreatedBy:   model.CreatedBy,
		UpdatedBy:   model.UpdatedBy,
	}
}

func fromBinDomainEntity(entity *stock.Bin) *models.Bin {
	return &models.Bin{
		BaseModel: models.BaseModel{
			ID:        entity.ID,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			CreatedBy: entity.CreatedBy,
			UpdatedBy: entity.UpdatedBy,
		},
		WarehouseID: entity.WarehouseID,
		Name:        entity.Name,
		IsActive:    entity.IsActive,
	}
}

func toStockDomainEntity(model *models.Stock) *stock.Stock {
	return &stock.Stock{
		ItemID:      model.ItemID,
		WarehouseID: model.WarehouseID,
		BinID:       &model.BinID,
		Quantity:    model.Quantity,
		UpdatedAt:   model.UpdatedAt,
	}
}

func fromStockDomainEntity(entity *stock.Stock) *models.Stock {
    binID := uuid.Nil
    if entity.BinID != nil {
        binID = *entity.BinID
    }
	return &models.Stock{
		ItemID:      entity.ItemID,
		WarehouseID: entity.WarehouseID,
		BinID:       binID,
		Quantity:    entity.Quantity,
		UpdatedAt:   entity.UpdatedAt,
	}
}

func toStockMovementDomainEntity(model *models.StockMovement) *stock.StockMovement {
	return &stock.StockMovement{
		ID:          model.ID,
		ItemID:      model.ItemID,
		WarehouseID: model.WarehouseID,
		BinID:       model.BinID,
		Type:        stock.MovementType(model.Type),
		Quantity:    model.Quantity,
		Reason:      model.Reason,
		HappenedAt:  model.HappenedAt,
		CreatedBy:   model.CreatedBy,
	}
}

func fromStockMovementDomainEntity(entity *stock.StockMovement) *models.StockMovement {
	return &models.StockMovement{
		ID:          entity.ID,
		ItemID:      entity.ItemID,
		WarehouseID: entity.WarehouseID,
		BinID:       entity.BinID,
		Type:        string(entity.Type),
		Quantity:    entity.Quantity,
		Reason:      entity.Reason,
		HappenedAt:  entity.HappenedAt,
		CreatedBy:   entity.CreatedBy,
	}
}

func toStockLedgerDomainEntity(model *models.StockLedger) *stock.StockLedger {
	return &stock.StockLedger{
		ID:              model.ID,
		StockMovementID: model.StockMovementID,
		ItemID:          model.ItemID,
		WarehouseID:     model.WarehouseID,
		BinID:           model.BinID,
		MovementType:    stock.MovementType(model.MovementType),
		QuantityChange:  model.QuantityChange,
		QuantityBefore:  model.QuantityBefore,
		QuantityAfter:   model.QuantityAfter,
		Reason:          model.Reason,
		HappenedAt:      model.HappenedAt,
		RecordedAt:      model.RecordedAt,
		RecordedBy:      model.RecordedBy,
	}
}

func fromStockLedgerDomainEntity(entity *stock.StockLedger) *models.StockLedger {
	return &models.StockLedger{
		ID:              entity.ID,
		StockMovementID: entity.StockMovementID,
		ItemID:          entity.ItemID,
		WarehouseID:     entity.WarehouseID,
		BinID:           entity.BinID,
		MovementType:    string(entity.MovementType),
		QuantityChange:  entity.QuantityChange,
		QuantityBefore:  entity.QuantityBefore,
		QuantityAfter:   entity.QuantityAfter,
		Reason:          entity.Reason,
		HappenedAt:      entity.HappenedAt,
		RecordedAt:      entity.RecordedAt,
		RecordedBy:      entity.RecordedBy,
	}
}
