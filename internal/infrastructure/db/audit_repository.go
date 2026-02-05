package db

import (
	"context"
	"doligo_001/internal/domain"
	"time"
	"gorm.io/gorm"
)

type auditModel struct {
	ID            string    `gorm:"type:uuid;primaryKey"`
	Timestamp     time.Time `gorm:"not null"`
	UserID        *string   `gorm:"type:uuid"`
	ResourceName  string    `gorm:"not null"`
	ResourceID    string    `gorm:"not null"`
	Action        string    `gorm:"not null"`
	Severity      string    `gorm:"not null;default:INFO"`
	OldValues     []byte    `gorm:"type:jsonb"`
	NewValues     []byte    `gorm:"type:jsonb"`
	CorrelationID string
}

func (auditModel) TableName() string {
	return "audit_logs"
}

type gormAuditRepository struct {
	db *gorm.DB
}

func NewGormAuditRepository(db *gorm.DB) domain.AuditRepository {
	return &gormAuditRepository{db: db}
}

func (r *gormAuditRepository) Save(ctx context.Context, log *domain.AuditLog) error {
	var userIDStr *string
	if log.UserID != nil {
		s := log.UserID.String()
		userIDStr = &s
	}

	model := auditModel{
		ID:            log.ID.String(),
		Timestamp:     log.Timestamp,
		UserID:        userIDStr,
		ResourceName:  log.ResourceName,
		ResourceID:    log.ResourceID,
		Action:        log.Action,
		Severity:      log.Severity,
		OldValues:     []byte(log.OldValues),
		NewValues:     []byte(log.NewValues),
		CorrelationID: log.CorrelationID,
	}

	return r.db.WithContext(ctx).Create(&model).Error
}
