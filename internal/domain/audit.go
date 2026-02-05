package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditLog represents a domain audit event
type AuditLog struct {
	ID            uuid.UUID       `json:"id"`
	Timestamp     time.Time       `json:"timestamp"`
	UserID        *uuid.UUID      `json:"user_id"`
	ResourceName  string          `json:"resource_name"`
	ResourceID    string          `json:"resource_id"`
	Action        string          `json:"action"`
	Severity      string          `json:"severity"`
	OldValues     json.RawMessage `json:"old_values"`
	NewValues     json.RawMessage `json:"new_values"`
	CorrelationID string          `json:"correlation_id"`
}

// AuditRepository defines the contract for persisting audit logs
type AuditRepository interface {
	Save(ctx context.Context, log *AuditLog) error
}
