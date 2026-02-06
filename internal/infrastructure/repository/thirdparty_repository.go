// Package repository provides the GORM-based implementation of the repository
// interfaces defined in the domain layer.
package repository

import (
	"context"
	"errors"

	"doligo_001/internal/domain/thirdparty"
	"doligo_001/internal/infrastructure/db/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// gormThirdPartyRepository is a GORM implementation of the thirdparty.Repository.
type gormThirdPartyRepository struct {
	db *gorm.DB
}

// NewGormThirdPartyRepository creates a new gormThirdPartyRepository.
func NewGormThirdPartyRepository(db *gorm.DB) thirdparty.Repository {
	return &gormThirdPartyRepository{db: db}
}

// Create persists a new third party to the data store.
func (r *gormThirdPartyRepository) Create(ctx context.Context, tp *thirdparty.ThirdParty) error {
	if tp.CreatedBy == uuid.Nil {
		return errors.New("created_by is required")
	}
	model := fromThirdPartyDomainEntity(tp)
	return r.db.WithContext(ctx).Create(model).Error
}

// GetByID retrieves a third party by their unique identifier.
func (r *gormThirdPartyRepository) GetByID(ctx context.Context, id uuid.UUID) (*thirdparty.ThirdParty, error) {
	var model models.ThirdParty
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return toThirdPartyDomainEntity(&model), nil
}

// Update modifies an existing third party in the data store.
func (r *gormThirdPartyRepository) Update(ctx context.Context, tp *thirdparty.ThirdParty) error {
	if tp.UpdatedBy == uuid.Nil {
		return errors.New("updated_by is required")
	}
	model := fromThirdPartyDomainEntity(tp)
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes a third party from the data store.
func (r *gormThirdPartyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.ThirdParty{}, "id = ?", id).Error
}

// List retrieves all third parties from the data store.
func (r *gormThirdPartyRepository) List(ctx context.Context) ([]*thirdparty.ThirdParty, error) {
	var modelList []models.ThirdParty
	if err := r.db.WithContext(ctx).Find(&modelList).Error; err != nil {
		return nil, err
	}

	domainList := make([]*thirdparty.ThirdParty, len(modelList))
	for i, model := range modelList {
		domainList[i] = toThirdPartyDomainEntity(&model)
	}
	return domainList, nil
}

// toThirdPartyDomainEntity converts a GORM third party model to a domain entity.
func toThirdPartyDomainEntity(model *models.ThirdParty) *thirdparty.ThirdParty {
	return &thirdparty.ThirdParty{
		ID:        model.ID,
		Name:      model.Name,
		Email:     model.Email,
		Type:      thirdparty.ThirdPartyType(model.Type),
		IsActive:  model.IsActive,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		CreatedBy: model.CreatedBy,
		UpdatedBy: model.UpdatedBy,
	}
}

// fromThirdPartyDomainEntity converts a domain third party entity to a GORM model.
func fromThirdPartyDomainEntity(entity *thirdparty.ThirdParty) *models.ThirdParty {
	return &models.ThirdParty{
		BaseModel: models.BaseModel{
			ID:        entity.ID,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			CreatedBy: entity.CreatedBy,
			UpdatedBy: entity.UpdatedBy,
		},
		Name:     entity.Name,
		Email:    entity.Email,
		Type:     string(entity.Type),
		IsActive: entity.IsActive,
	}
}
