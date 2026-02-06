package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"doligo_001/internal/domain/identity"
	"doligo_001/internal/infrastructure/db/models"
)

// GormPermissionRepository is a GORM implementation of the PermissionRepository.
type GormPermissionRepository struct {
	db *gorm.DB
}

// NewGormPermissionRepository creates a new GormPermissionRepository.
func NewGormPermissionRepository(db *gorm.DB) *GormPermissionRepository {
	return &GormPermissionRepository{db: db}
}

// FindByName retrieves a permission by its name.
func (r *GormPermissionRepository) FindByName(ctx context.Context, name string) (*identity.Permission, error) {
	var pModel models.Permission
	if err := r.db.WithContext(ctx).First(&pModel, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return toPermissionDomainEntity(&pModel), nil
}

// FindByID retrieves a permission by its unique identifier.
func (r *GormPermissionRepository) FindByID(ctx context.Context, id uuid.UUID) (*identity.Permission, error) {
	if id == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}
	var pModel models.Permission
	if err := r.db.WithContext(ctx).First(&pModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return toPermissionDomainEntity(&pModel), nil
}
