// Package repository provides the GORM-based implementation of the repository
// interfaces defined in the domain layer.
package repository

import (
	"context"

	"doligo_001/internal/domain/item"
	"doligo_001/internal/infrastructure/db/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// gormItemRepository is a GORM implementation of the item.Repository.
type gormItemRepository struct {
	db *gorm.DB
}

func (r *gormItemRepository) WithTx(tx *gorm.DB) item.Repository {
	return NewGormItemRepository(tx)
}

// NewGormItemRepository creates a new gormItemRepository.
func NewGormItemRepository(db *gorm.DB) item.Repository {
	return &gormItemRepository{db: db}
}

// Create persists a new item to the data store.
func (r *gormItemRepository) Create(ctx context.Context, i *item.Item) error {
	model := fromItemDomainEntity(i)
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByID retrieves an item by its unique identifier.
func (r *gormItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*item.Item, error) {
	var model models.Item
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return toItemDomainEntity(&model), nil
}

// Update modifies an existing item in the data store.
func (r *gormItemRepository) Update(ctx context.Context, i *item.Item) error {
	model := fromItemDomainEntity(i)
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes an item from the data store.
func (r *gormItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Item{}, "id = ?", id).Error
}

// List retrieves all items from the data store.
func (r *gormItemRepository) List(ctx context.Context) ([]*item.Item, error) {
	var modelList []models.Item
	if err := r.db.WithContext(ctx).Find(&modelList).Error; err != nil {
		return nil, err
	}

	domainList := make([]*item.Item, len(modelList))
	for i, model := range modelList {
		domainList[i] = toItemDomainEntity(&model)
	}
	return domainList, nil
}

// toItemDomainEntity converts a GORM item model to a domain entity.
func toItemDomainEntity(model *models.Item) *item.Item {
	return &item.Item{
		ID:          model.ID,
		Name:        model.Name,
		Description: model.Description,
		Type:        item.ItemType(model.Type),
		CostPrice:   model.CostPrice,
		SalePrice:   model.SalePrice,
		AverageCost: model.AverageCost,
		IsActive:    model.IsActive,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		CreatedBy:   model.CreatedBy,
		UpdatedBy:   model.UpdatedBy,
	}
}

// fromItemDomainEntity converts a domain item entity to a GORM model.
func fromItemDomainEntity(entity *item.Item) *models.Item {
	return &models.Item{
		BaseModel: models.BaseModel{
			ID:        entity.ID,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			CreatedBy: entity.CreatedBy,
			UpdatedBy: entity.UpdatedBy,
		},
		Name:        entity.Name,
		Description: entity.Description,
		Type:        string(entity.Type),
		CostPrice:   entity.CostPrice,
		SalePrice:   entity.SalePrice,
		AverageCost: entity.AverageCost,
		IsActive:    entity.IsActive,
	}
}
