package identity

import (
	"context"
	"time"
)

// User represents a user in the system.
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	RoleID       int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
	CreatedBy    *int64
	UpdatedBy    *int64
}

// Role represents a user role.
type Role struct {
	ID   int64
	Name string
}

// Permission represents a permission.
type Permission struct {
	ID   int64
	Name string
}

// Product represents a product in the system.
type Product struct {
	ID        int64
	Name      string
	Price     float64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	CreatedBy int64
	UpdatedBy int64
}

// UserRepository defines the interface for user persistence.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetUserByUsername(ctx context.Context, username string) (*User, error)
}

// RoleRepository defines the interface for role persistence.
type RoleRepository interface {
	GetRolePermissions(ctx context.Context, roleID int64) ([]*Permission, error)
}
