// Package repository provides the GORM-based implementation of the repository
// interfaces defined in the domain layer. It is responsible for translating
// domain entities to and from database models and executing database operations.
package repository

import (
	"context"

	"gorm.io/gorm"
	"github.com/google/uuid"

	"doligo_001/internal/domain/identity"
	"doligo_001/internal/infrastructure/db/models"
)

// GormUserRepository is a GORM implementation of the UserRepository.
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new GormUserRepository.
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// FindByEmail retrieves a user by their email address.
func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (*identity.User, error) {
	var userModel models.User
	if err := r.db.WithContext(ctx).Preload("Roles.Permissions").First(&userModel, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return toUserDomainEntity(&userModel), nil
}

// FindByID retrieves a user by their unique identifier.
func (r *GormUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*identity.User, error) {
	var userModel models.User
	if err := r.db.WithContext(ctx).Preload("Roles.Permissions").First(&userModel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return toUserDomainEntity(&userModel), nil
}

// Create persists a new user to the data store.
func (r *GormUserRepository) Create(ctx context.Context, user *identity.User) error {
	userModel := fromUserDomainEntity(user)
	return r.db.WithContext(ctx).Create(userModel).Error
}

// toUserDomainEntity converts a GORM user model to a domain user entity.
func toUserDomainEntity(model *models.User) *identity.User {
	roles := make([]identity.Role, len(model.Roles))
	for i, roleModel := range model.Roles {
		roles[i] = *toRoleDomainEntity(&roleModel)
	}

	return &identity.User{
		ID:        model.ID,
		FirstName: model.FirstName,
		LastName:  model.LastName,
		Email:     model.Email,
		Password:  model.Password,
		IsActive:  model.IsActive,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Roles:     roles,
	}
}

// fromUserDomainEntity converts a domain user entity to a GORM user model.
func fromUserDomainEntity(entity *identity.User) *models.User {
	roles := make([]models.Role, len(entity.Roles))
	for i, roleEntity := range entity.Roles {
		roles[i] = *fromRoleDomainEntity(&roleEntity)
	}
	return &models.User{
		BaseModel: models.BaseModel{ID: entity.ID},
		FirstName: entity.FirstName,
		LastName:  entity.LastName,
		Email:     entity.Email,
		Password:  entity.Password,
		IsActive:  entity.IsActive,
		Roles:     roles,
	}
}

// toRoleDomainEntity converts a GORM role model to a domain role entity.
func toRoleDomainEntity(model *models.Role) *identity.Role {
	permissions := make([]identity.Permission, len(model.Permissions))
	for i, pModel := range model.Permissions {
		permissions[i] = *toPermissionDomainEntity(&pModel)
	}

	return &identity.Role{
		ID:          model.ID,
		Name:        model.Name,
		Description: model.Description,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		Permissions: permissions,
	}
}

// fromRoleDomainEntity converts a domain role entity to a GORM role model.
func fromRoleDomainEntity(entity *identity.Role) *models.Role {
	permissions := make([]models.Permission, len(entity.Permissions))
	for i, pEntity := range entity.Permissions {
		permissions[i] = *fromPermissionDomainEntity(&pEntity)
	}
	return &models.Role{
		BaseModel:   models.BaseModel{ID: entity.ID},
		Name:        entity.Name,
		Description: entity.Description,
		Permissions: permissions,
	}
}

// toPermissionDomainEntity converts a GORM permission model to a domain permission entity.
func toPermissionDomainEntity(model *models.Permission) *identity.Permission {
	return &identity.Permission{
		ID:          model.ID,
		Name:        model.Name,
		Description: model.Description,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

// fromPermissionDomainEntity converts a domain permission entity to a GORM permission model.
func fromPermissionDomainEntity(entity *identity.Permission) *models.Permission {
	return &models.Permission{
		BaseModel:   models.BaseModel{ID: entity.ID},
		Name:        entity.Name,
		Description: entity.Description,
	}
}
