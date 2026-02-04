package usecase_test

import (
	"context"
	"doligo_001/internal/domain"
	"doligo_001/internal/usecase"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) Save(ctx context.Context, log *domain.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func TestAuditService_Log(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	service := usecase.NewAuditService(mockRepo)

	userID := uuid.New()
	resourceName := "item"
	resourceID := uuid.New().String()
	action := "UPDATE"
	oldValues := map[string]interface{}{"price": 100}
	newValues := map[string]interface{}{"price": 120}
	correlationID := "test-corr-id"

	// Expectation
	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.AuditLog")).Return(nil).Run(func(args mock.Arguments) {
		log := args.Get(1).(*domain.AuditLog)
		assert.Equal(t, userID, log.UserID)
		assert.Equal(t, resourceName, log.ResourceName)
		assert.Equal(t, resourceID, log.ResourceID)
		assert.Equal(t, action, log.Action)
		assert.Equal(t, correlationID, log.CorrelationID)
		assert.NotNil(t, log.ID)
		assert.WithinDuration(t, time.Now(), log.Timestamp, 2*time.Second)
	})

	service.Log(context.Background(), userID, resourceName, resourceID, action, oldValues, newValues, correlationID)

	// Since it's async, we need to wait a bit or use a better way to synchronize
	// For this test, we can use a channel or just wait.
	time.Sleep(100 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}
