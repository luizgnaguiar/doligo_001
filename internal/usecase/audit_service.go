package usecase

import (
	"context"
	"doligo_001/internal/domain"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type AuditService interface {
	Log(ctx context.Context, userID uuid.UUID, resourceName, resourceID, action string, oldValues, newValues interface{}, correlationID string)
}

type auditService struct {
	repo domain.AuditRepository
}

func NewAuditService(repo domain.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) Log(ctx context.Context, userID uuid.UUID, resourceName, resourceID, action string, oldValues, newValues interface{}, correlationID string) {
	severity := s.calculateSeverity(resourceName, action, oldValues, newValues)

	// Fire and forget: run in a goroutine
	go func() {
		// Use a background context to ensure it continues even if the request context is canceled
		auditCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		oldJSON, _ := json.Marshal(oldValues)
		newJSON, _ := json.Marshal(newValues)

		var uID *uuid.UUID
		if userID != uuid.Nil {
			uID = &userID
		}

		log := &domain.AuditLog{
			ID:            uuid.New(),
			Timestamp:     time.Now(),
			UserID:        uID,
			ResourceName:  resourceName,
			ResourceID:    resourceID,
			Action:        action,
			Severity:      severity,
			OldValues:     json.RawMessage(oldJSON),
			NewValues:     json.RawMessage(newJSON),
			CorrelationID: correlationID,
		}

		if err := s.repo.Save(auditCtx, log); err != nil {
			slog.Error("Failed to save audit log", "error", err, "resource", resourceName, "id", resourceID)
		}

		// Emit structured log for critical/warning events
		var uIDStr string
		if uID != nil {
			uIDStr = uID.String()
		} else {
			uIDStr = "ANONYMOUS"
		}

		logAttrs := []any{
			slog.String("audit_log_id", log.ID.String()),
			slog.String("user_id", uIDStr),
			slog.String("resource", resourceName),
			slog.String("resource_id", resourceID),
			slog.String("action", action),
			slog.String("correlation_id", correlationID),
			slog.String("severity", severity),
		}

		if severity == "CRITICAL" {
			slog.Error("CRITICAL AUDIT EVENT DETECTED", logAttrs...)
		} else if severity == "WARN" {
			slog.Warn("AUDIT EVENT ALERT", logAttrs...)
		}
	}()
}

func (s *auditService) calculateSeverity(resource, action string, oldV, newV interface{}) string {
	// Critical events defined in specification
	if resource == "identity" && action == "LOGIN_FAILURE" {
		return "CRITICAL"
	}
	if resource == "invoice" && action == "DELETE" {
		return "CRITICAL"
	}
	if resource == "stock" && action == "REVERSAL" {
		return "CRITICAL"
	}

	// Item price change is critical
	if resource == "item" && action == "UPDATE" {
		if oldV != nil && newV != nil {
			oldJSON, _ := json.Marshal(oldV)
			newJSON, _ := json.Marshal(newV)

			var oldMap, newMap map[string]interface{}
			_ = json.Unmarshal(oldJSON, &oldMap)
			_ = json.Unmarshal(newJSON, &newMap)

			// Field names are likely SalePrice and CostPrice based on domain struct
			if oldMap["SalePrice"] != newMap["SalePrice"] || oldMap["CostPrice"] != newMap["CostPrice"] {
				return "CRITICAL"
			}
		}
	}

	// Default warnings for deletions
	if action == "DELETE" {
		return "WARN"
	}

	return "INFO"
}
