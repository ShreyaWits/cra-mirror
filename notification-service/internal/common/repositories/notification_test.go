package repositories

import (
	"errors"
	"notification-service/internal/common/models"
	"testing"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockQuery struct {
	mock.Mock
}

func (m *MockQuery) Exec() error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockQuery) Iter() *gocql.Iter {
	args := m.Called()
	return args.Get(0).(*gocql.Iter)
}

func (m *MockQuery) Scan(dest ...interface{}) error {
	args := m.Called(dest...)
	return args.Error(0)
}

func TestSaveNotification_Success(t *testing.T) {
	mockSession := new(MockSession)
	mockQuery := new(MockQuery)

	notification := &models.Notification{
		TemplateID:         "template1",
		Channel:            "email",
		Meta:               map[string]string{"source": "test"},
		Tags:               []string{"tag1"},
		Recipient:          map[string]string{"user_id": "u1"},
		PrimaryService:     "sendgrid",
		FallbackService:    "smtp",
		PrimaryRetryCount:  1,
		FallbackRetryCount: 1,
		TotalExecutionTime: 10000,
		NotificationStatus: "PENDING",
		TrackingId:         "track-123",
	}

	mockSession.On("Query", mock.Anything, mock.Anything).Return(mockQuery)
	mockQuery.On("Exec").Return(nil)

	repo := &NotificationRepository{Session: mockSession}
	id, err := repo.SaveNotification(notification)

	assert.NoError(t, err)
	assert.NotNil(t, id)

	mockSession.AssertExpectations(t)
	mockQuery.AssertExpectations(t)
}

func TestSaveNotification_Failure(t *testing.T) {
	mockSession := new(MockSession)
	mockQuery := new(MockQuery)

	notification := &models.Notification{
		TemplateID: "template1",
		Channel:    "email",
	}

	mockSession.On("Query", mock.Anything, mock.Anything).Return(mockQuery)
	mockQuery.On("Exec").Return(errors.New("insert failed"))

	repo := &NotificationRepository{Session: mockSession}
	id, err := repo.SaveNotification(notification)

	assert.Error(t, err)
	assert.Nil(t, id)

	mockSession.AssertExpectations(t)
	mockQuery.AssertExpectations(t)
}

func TestUpdateNotificationByID(t *testing.T) {
	mockSession := new(MockSession)
	mockQuery := new(MockQuery)

	update := &models.Notification{
		Meta:               map[string]string{"source": "api"},
		Tags:               []string{"x"},
		PrimaryRetryCount:  1,
		FallbackRetryCount: 1,
		TotalExecutionTime: 500,
		NotificationStatus: "SENT",
		ErrorMessage:       "none",
		IsPrimaryFailed:    true,
		IsFallbackFailed:   true,
		IsMovedToDlq:       true,
	}

	mockSession.On("Query", mock.Anything, mock.Anything).Return(mockQuery)
	mockQuery.On("Exec").Return(nil)

	repo := &NotificationRepository{Session: mockSession}
	err := repo.UpdateNotificationByID(uuid.New().String(), update)

	assert.NoError(t, err)
	mockSession.AssertExpectations(t)
	mockQuery.AssertExpectations(t)
}

func TestUpdateNotificationByID_NoFields(t *testing.T) {
	mockSession := new(MockSession)

	repo := &NotificationRepository{Session: mockSession}
	err := repo.UpdateNotificationByID(uuid.New().String(), &models.Notification{})

	assert.NoError(t, err)
}
