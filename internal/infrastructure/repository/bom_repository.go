package repository

import (
	"context"
	"doligo_001/internal/domain/bom"
	"doligo_001/internal/infrastructure/db"
	"doligo_001/internal/infrastructure/db/models"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// gormBomRepository is a GORM implementation of the bom.Repository.
type gormBomRepository struct {
	db            *gorm.DB
	transactioner db.Transactioner
}

// NewGormBomRepository creates a new gormBomRepository.
func NewGormBomRepository(db *gorm.DB, transactioner db.Transactioner) bom.Repository {
	return &gormBomRepository{
		db:            db,
		transactioner: transactioner,
	}
}

// Create creates a new BillOfMaterials in the database.
func (r *gormBomRepository) Create(ctx context.Context, b *bom.BillOfMaterials) error {
	model := fromBomDomainEntity(b)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create BOM: %w", err)
	}
	// Update the domain entity with generated ID and timestamps
	*b = *toBomDomainEntity(model) // Full update to get all generated fields
	return nil
}

// GetByID retrieves a BillOfMaterials by its ID.
func (r *gormBomRepository) GetByID(ctx context.Context, id uuid.UUID) (*bom.BillOfMaterials, error) {
	var model models.BillOfMaterials
	if err := r.db.WithContext(ctx).Preload("Components").First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, bom.ErrBOMNotFound
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
			return nil, bom.ErrBOMNotFound
		}
		return nil, fmt.Errorf("failed to get BOM by product ID: %w", err)
	}
	return toBomDomainEntity(&model), nil
}

// Update updates an existing BillOfMaterials in the database using an explicit transactional approach.
func (r *gormBomRepository) Update(ctx context.Context, b *bom.BillOfMaterials) error {
	return r.transactioner.Transaction(ctx, func(tx *gorm.DB) error {
		// 1. Fetch the existing BOM with its components
		var existingModel models.BillOfMaterials
		if err := tx.Preload("Components").First(&existingModel, "id = ?", b.ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return bom.ErrBOMNotFound
			}
			return fmt.Errorf("failed to find existing BOM for update: %w", err)
		}

		// 2. Map incoming domain entity to a model for processing
		incomingModel := fromBomDomainEntity(b)

		// 3. Differentiate between components to add, update, and remove
		existingCompsMap := make(map[uuid.UUID]models.BillOfMaterialsComponent)
		for _, comp := range existingModel.Components {
			existingCompsMap[comp.ID] = comp
		}

		incomingCompsMap := make(map[uuid.UUID]models.BillOfMaterialsComponent)
		var compsToAdd []models.BillOfMaterialsComponent
		var compsToUpdate []models.BillOfMaterialsComponent

		for _, incomingComp := range incomingModel.Components {
			// If ID is zero, it's a new component
			if incomingComp.ID == uuid.Nil {
				incomingComp.BillOfMaterialsID = b.ID // Ensure association is set
				compsToAdd = append(compsToAdd, incomingComp)
				continue
			}
			incomingCompsMap[incomingComp.ID] = incomingComp
			// If it exists in the old map, it's a potential update
			if _, ok := existingCompsMap[incomingComp.ID]; ok {
				compsToUpdate = append(compsToUpdate, incomingComp)
			}
		}

		var compIDsToRemove []uuid.UUID
		for _, existingComp := range existingModel.Components {
			if _, ok := incomingCompsMap[existingComp.ID]; !ok {
				compIDsToRemove = append(compIDsToRemove, existingComp.ID)
			}
		}

		// 4. Execute DB operations
		// REMOVE
		if len(compIDsToRemove) > 0 {
			if err := tx.Delete(&models.BillOfMaterialsComponent{}, "id IN ?", compIDsToRemove).Error; err != nil {
				return fmt.Errorf("failed to remove components: %w", err)
			}
		}
		// ADD
		if len(compsToAdd) > 0 {
			if err := tx.Create(&compsToAdd).Error; err != nil {
				return fmt.Errorf("failed to add new components: %w", err)
			}
		}
		// MODIFY
		for _, compToUpdate := range compsToUpdate {
			if err := tx.Save(&compToUpdate).Error; err != nil {
				return fmt.Errorf("failed to update component ID %s: %w", compToUpdate.ID, err)
			}
		}

		// 5. Update the parent BOM object
		if err := tx.Omit("Components").Save(incomingModel).Error; err != nil {
			return fmt.Errorf("failed to update BOM header: %w", err)
		}

		return nil
	})
}

// Delete deletes a BillOfMaterials by its ID.
func (r *gormBomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Note: GORM's default delete behavior might not cascade correctly without specific config.
	// A robust implementation should handle deleting associations explicitly in a transaction.
	return r.transactioner.Transaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Where("bill_of_materials_id = ?", id).Delete(&models.BillOfMaterialsComponent{}).Error; err != nil {
			return fmt.Errorf("failed to delete BOM components: %w", err)
		}
		if err := tx.Delete(&models.BillOfMaterials{}, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return bom.ErrBOMNotFound
			}
			return fmt.Errorf("failed to delete BOM: %w", err)
		}
		return nil
	})
}

// List lists all BillOfMaterials.
func (r *gormBomRepository) List(ctx context.Context) ([]*bom.BillOfMaterials, error) {
	var modelsList []models.BillOfMaterials
	if err := r.db.WithContext(ctx).Preload("Components").Find(&modelsList).Error; err != nil {
		return nil, fmt.Errorf("failed to list BOMs: %w", err)
	}
	domainList := make([]*bom.BillOfMaterials, len(modelsList))
	for i := range modelsList {
		domainList[i] = toBomDomainEntity(&modelsList[i])
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
	for i := range model.Components {
		components[i] = *toBomComponentDomainEntity(&model.Components[i])
	}
	return &bom.BillOfMaterials{
		ID:         model.ID,
		ProductID:  model.ProductID,
		Name:       model.Name,
		IsActive:   model.IsActive,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
		CreatedBy:  model.CreatedBy,
		UpdatedBy:  model.UpdatedBy,
		Components: components,
	}
}

func fromBomDomainEntity(entity *bom.BillOfMaterials) *models.BillOfMaterials {
	if entity == nil {
		return nil
	}
	components := make([]models.BillOfMaterialsComponent, len(entity.Components))
	for i := range entity.Components {
		components[i] = *fromBomComponentDomainEntity(&entity.Components[i])
	}
	return &models.BillOfMaterials{
		BaseModel: models.BaseModel{
			ID:        entity.ID,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			CreatedBy: entity.CreatedBy,
			UpdatedBy: entity.UpdatedBy,
		},
		ProductID:  entity.ProductID,
		Name:       entity.Name,
		IsActive:   entity.IsActive,
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
