package repository

import (
	"context"
	"fmt"
	"doligo_001/internal/domain/bom"
	"doligo_001/internal/infrastructure/db/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// gormBomRepository is a GORM implementation of the bom.Repository.
type gormBomRepository struct {
	db *gorm.DB
}

// NewGormBomRepository creates a new gormBomRepository.
func NewGormBomRepository(db *gorm.DB) bom.Repository {
	return &gormBomRepository{db: db}
}

// Create creates a new BillOfMaterials in the database.
func (r *gormBomRepository) Create(ctx context.Context, b *bom.BillOfMaterials) error {
	model := fromBomDomainEntity(b)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create BOM: %w", err)
	}
	// Update the domain entity with generated ID and timestamps
	b.ID = model.ID
	b.CreatedAt = model.CreatedAt
	b.UpdatedAt = model.UpdatedAt
	for i, compModel := range model.Components {
		b.Components[i].ID = compModel.ID
		b.Components[i].CreatedAt = compModel.CreatedAt
		b.Components[i].UpdatedAt = compModel.UpdatedAt
	}
	return nil
}

// GetByID retrieves a BillOfMaterials by its ID.
func (r *gormBomRepository) GetByID(ctx context.Context, id uuid.UUID) (*bom.BillOfMaterials, error) {
	var model models.BillOfMaterials
	if err := r.db.WithContext(ctx).Preload("Components").First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get BOM by ID: %w", err)
	}
	return toBomDomainEntity(&model), nil
}

// GetByProductID retrieves a BillOfMaterials by the product it produces.
func (r *gormBomRepository) GetByProductID(ctx context.Context, productID uuid.UUID) (*bom.BillOfMaterials, error) {
	var model models.BillOfMaterials
	if err := r.db.WithContext(ctx).Preload("Components").First(&model, "product_id = ?", productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get BOM by product ID: %w", err)
	}
	return toBomDomainEntity(&model), nil
}

// Update updates an existing BillOfMaterials in the database.
func (r *gormBomRepository) Update(ctx context.Context, b *bom.BillOfMaterials) error {
	model := fromBomDomainEntity(b)
	// For updating associations, it's often better to delete and re-create or manage individually.
	// For simplicity, this example will just update the main BOM fields.
	// A more robust solution would handle component updates (add, remove, modify) explicitly.
	if err := r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(model).Error; err != nil {
		return fmt.Errorf("failed to update BOM: %w", err)
	}
	return nil
}

// Delete deletes a BillOfMaterials by its ID.
func (r *gormBomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.BillOfMaterials{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete BOM: %w", err)
	}
	return nil
}

// List lists all BillOfMaterials.
func (r *gormBomRepository) List(ctx context.Context) ([]*bom.BillOfMaterials, error) {
	var modelsList []models.BillOfMaterials
	if err := r.db.WithContext(ctx).Preload("Components").Find(&modelsList).Error; err != nil {
		return nil, fmt.Errorf("failed to list BOMs: %w", err)
	}
	domainList := make([]*bom.BillOfMaterials, len(modelsList))
	for i, model := range modelsList {
		domainList[i] = toBomDomainEntity(&model)
	}
	return domainList, nil
}


// gormProductionRecordRepository is a GORM implementation of the bom.ProductionRecordRepository.
type gormProductionRecordRepository struct {
	db *gorm.DB
}

// NewGormProductionRecordRepository creates a new gormProductionRecordRepository.
func NewGormProductionRecordRepository(db *gorm.DB) bom.ProductionRecordRepository {
	return &gormProductionRecordRepository{db: db}
}

// Create creates a new ProductionRecord in the database.
func (r *gormProductionRecordRepository) Create(ctx context.Context, pr *bom.ProductionRecord) error {
	model := fromProductionRecordDomainEntity(pr)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create production record: %w", err)
	}
	// Update the domain entity with generated ID
	pr.ID = model.ID
	return nil
}


// --- MAPPING FUNCTIONS ---

func toBomDomainEntity(model *models.BillOfMaterials) *bom.BillOfMaterials {
	if model == nil {
		return nil
	}
	components := make([]bom.BillOfMaterialsComponent, len(model.Components))
	for i, compModel := range model.Components {
		components[i] = *toBomComponentDomainEntity(&compModel)
	}
	return &bom.BillOfMaterials{
		ID:        model.ID,
		ProductID: model.ProductID,
		Name:      model.Name,
		IsActive:  model.IsActive,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		CreatedBy: model.CreatedBy,
		UpdatedBy: model.UpdatedBy,
		Components: components,
	}
}

func fromBomDomainEntity(entity *bom.BillOfMaterials) *models.BillOfMaterials {
	if entity == nil {
		return nil
	}
	components := make([]models.BillOfMaterialsComponent, len(entity.Components))
	for i, compEntity := range entity.Components {
		components[i] = *fromBomComponentDomainEntity(&compEntity)
	}
	return &models.BillOfMaterials{
		BaseModel: models.BaseModel{
			ID:        entity.ID,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			CreatedBy: entity.CreatedBy,
			UpdatedBy: entity.UpdatedBy,
		},
		ProductID: entity.ProductID,
		Name:      entity.Name,
		IsActive:  entity.IsActive,
		Components: components,
	}
}

func toBomComponentDomainEntity(model *models.BillOfMaterialsComponent) *bom.BillOfMaterialsComponent {
	if model == nil {
		return nil
	}
	return &bom.BillOfMaterialsComponent{
		ID:                model.ID,
		BillOfMaterialsID: model.BillOfMaterialsID,
		ComponentItemID:   model.ComponentItemID,
		Quantity:          model.Quantity,
		UnitOfMeasure:     model.UnitOfMeasure,
		IsActive:          model.IsActive,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
		CreatedBy:         model.CreatedBy,
		UpdatedBy:         model.UpdatedBy,
	}
}

func fromBomComponentDomainEntity(entity *bom.BillOfMaterialsComponent) *models.BillOfMaterialsComponent {
	if entity == nil {
		return nil
	}
	return &models.BillOfMaterialsComponent{
		BaseModel: models.BaseModel{
			ID:        entity.ID,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			CreatedBy: entity.CreatedBy,
			UpdatedBy: entity.UpdatedBy,
		},
		BillOfMaterialsID: entity.BillOfMaterialsID,
		ComponentItemID:   entity.ComponentItemID,
		Quantity:          entity.Quantity,
		UnitOfMeasure:     entity.UnitOfMeasure,
		IsActive:          entity.IsActive,
	}
}

func toProductionRecordDomainEntity(model *models.ProductionRecord) *bom.ProductionRecord {
	if model == nil {
		return nil
	}
	return &bom.ProductionRecord{
		ID:                   model.ID,
		BillOfMaterialsID:    model.BillOfMaterialsID,
		ProducedProductID:    model.ProducedProductID,
		ProductionQuantity:   model.ProductionQuantity,
		ActualProductionCost: model.ActualProductionCost,
		WarehouseID:          model.WarehouseID,
		ProducedAt:           model.ProducedAt,
		CreatedBy:            model.CreatedBy,
	}
}

func fromProductionRecordDomainEntity(entity *bom.ProductionRecord) *models.ProductionRecord {
	if entity == nil {
		return nil
	}
	return &models.ProductionRecord{
		ID:                   entity.ID,
		BillOfMaterialsID:    entity.BillOfMaterialsID,
		ProducedProductID:    entity.ProducedProductID,
		ProductionQuantity:   entity.ProductionQuantity,
		ActualProductionCost: entity.ActualProductionCost,
		WarehouseID:          entity.WarehouseID,
		ProducedAt:           entity.ProducedAt,
		CreatedBy:            entity.CreatedBy,
	}
}
