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
	// Fire and forget: run in a goroutine
	go func() {
		// Use a background context to ensure it continues even if the request context is canceled
		// However, we might want to inherit some values from ctx if needed (like correlation_id)
		auditCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		oldJSON, _ := json.Marshal(oldValues)
		newJSON, _ := json.Marshal(newValues)

		log := &domain.AuditLog{
			ID:            uuid.New(),
			Timestamp:     time.Now(),
			UserID:        userID,
			ResourceName:  resourceName,
			ResourceID:    resourceID,
			Action:        action,
			OldValues:     json.RawMessage(oldJSON),
			NewValues:     json.RawMessage(newJSON),
			CorrelationID: correlationID,
		}

		if err := s.repo.Save(auditCtx, log); err != nil {
			slog.Error("Failed to save audit log", "error", err, "resource", resourceName, "id", resourceID)
		}
	}()
}
