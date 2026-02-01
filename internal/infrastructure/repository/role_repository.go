package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"doligo_001/internal/domain/identity"
	"doligo_001/internal/infrastructure/db/models"
)

// GormRoleRepository is a GORM implementation of the RoleRepository.
type GormRoleRepository struct {
	db *gorm.DB
}

// NewGormRoleRepository creates a new GormRoleRepository.
func NewGormRoleRepository(db *gorm.DB) *GormRoleRepository {
	return &GormRoleRepository{db: db}
}

// FindByName retrieves a role by its name.
func (r *GormRoleRepository) FindByName(ctx context.Context, name string) (*identity.Role, error) {
	var roleModel models.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&roleModel, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return toRoleDomainEntity(&roleModel), nil
}

// FindByID retrieves a role by its unique identifier.
func (r *GormRoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*identity.Role, error) {
	var roleModel models.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&roleModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return toRoleDomainEntity(&roleModel), nil
}
