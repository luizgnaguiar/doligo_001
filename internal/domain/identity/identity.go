// Package identity defines the core entities and repository interfaces for user identity,
// authentication, and authorization. It establishes the contracts for how the system
// manages users, roles, and permissions, adhering to Clean Architecture principles.
// The domain layer is kept pure, with no dependencies on external frameworks or
// infrastructure details.
package identity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents the core entity for a system user.
// It contains identification, credentials, and state, but no presentation
// or infrastructure-specific logic.
type User struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     string
	Password  string // This will be a hash, not plaintext
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	Roles     []Role
}

// Role represents a named set of permissions that can be assigned to users.
type Role struct {
	ID          uuid.UUID
	Name        string
	Description string
	Permissions []Permission
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Permission defines a specific action that can be granted to a role.
type Permission struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserRepository defines the contract for data persistence operations for Users.
// It operates purely on User domain entities.
type UserRepository interface {
	// FindByEmail retrieves a user by their email address.
	FindByEmail(ctx context.Context, email string) (*User, error)
	// FindByID retrieves a user by their unique identifier.
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	// Create persists a new user to the data store.
	Create(ctx context.Context, user *User) error
}

// RoleRepository defines the contract for data persistence operations for Roles.
type RoleRepository interface {
	// FindByName retrieves a role by its name.
	FindByName(ctx context.Context, name string) (*Role, error)
	// FindByID retrieves a role by its unique identifier.
	FindByID(ctx context.Context, id uuid.UUID) (*Role, error)
}

// PermissionRepository defines the contract for data persistence operations for Permissions.
type PermissionRepository interface {
	// FindByName retrieves a permission by its name.
	FindByName(ctx context.Context, name string) (*Permission, error)
	// FindByID retrieves a permission by its unique identifier.
	FindByID(ctx context.Context, id uuid.UUID) (*Permission, error)
}
