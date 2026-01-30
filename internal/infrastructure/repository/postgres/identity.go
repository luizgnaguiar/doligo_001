package postgres

import (
	"context"

	"gorm.io/gorm"

	"doligo_001/internal/domain/identity"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(ctx context.Context, user *identity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *GormUserRepository) GetUserByUsername(ctx context.Context, username string) (*identity.User, error) {
	var user identity.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	return &user, err
}

type GormRoleRepository struct {
	db *gorm.DB
}

func NewGormRoleRepository(db *gorm.DB) *GormRoleRepository {
	return &GormRoleRepository{db: db}
}

func (r *GormRoleRepository) GetRolePermissions(ctx context.Context, roleID int64) ([]*identity.Permission, error) {
	var permissions []*identity.Permission
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("join role_permissions on role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	return permissions, err
}
