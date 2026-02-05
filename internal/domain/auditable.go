// Package domain contains the core business logic and entities of the application.
// This file defines the Auditable interface and related context utilities,
// establishing a contract for entities that need to track creation and modification
// by user ID.
package domain

import (
	"context"

	"github.com/google/uuid"
)

// Auditable provides a contract for domain entities that are auditable.
// It allows abstracting the logic for setting user IDs during creation and updates.
type Auditable interface {
	SetCreatedBy(userID uuid.UUID)
	SetUpdatedBy(userID uuid.UUID)
}

// contextKey is a private type to prevent collisions with other context keys.
type contextKey string

const (
	// UserIDKey is the key used to store and retrieve the user ID from the context.
	UserIDKey contextKey = "userID"
	// PermissionsKey is the key used to store and retrieve permissions from the context.
	PermissionsKey contextKey = "permissions"
)

// ContextWithUserID returns a new context with the provided user ID.
func ContextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// UserIDFromContext extracts the user ID from the context.
// It returns the zero UUID and false if the user ID is not found.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// ContextWithPermissions returns a new context with the provided permissions.
func ContextWithPermissions(ctx context.Context, permissions []string) context.Context {
	return context.WithValue(ctx, PermissionsKey, permissions)
}

// PermissionsFromContext extracts the permissions from the context.
func PermissionsFromContext(ctx context.Context) ([]string, bool) {
	permissions, ok := ctx.Value(PermissionsKey).([]string)
	return permissions, ok
}
