package repositories

import (
	"fmt"
	"log"
	"notification-service/internal/common/models"
	"strings"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

type QueryExecutor interface {
	Exec() error
	Iter() *gocql.Iter
	Scan(...interface{}) error
}

type NotificationRepository struct {
	Session CassandraSession
}

type NotificationRepositoryInterface interface {
	SaveNotification(notification *models.Notification) (*string, error)
	UpdateNotificationByID(notificationID string, updateData *models.Notification) error
}

// Add this line at the top of the file to ensure the concrete type implements the interface
var _ NotificationRepositoryInterface = (*NotificationRepository)(nil)

func NewNotificationRepository(session CassandraSession) NotificationRepositoryInterface {
	return &NotificationRepository{Session: session}
}

func (r *NotificationRepository) SaveNotification(notification *models.Notification) (*string, error) {
	query := `INSERT INTO notifications 
	(id, recipient, channel, primary_service, fallback_service, meta, tags, template_id, primary_retry_count, fallback_retry_count, total_execution_time, notification_status, is_primary_failed, is_fallback_failed, is_moved_to_dlq, tracking_id) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	notificationId := uuid.New().String()

	err := r.Session.Query(query,
		notificationId,
		notification.Recipient,
		notification.Channel,
		notification.PrimaryService,
		notification.FallbackService,
		notification.Meta,
		notification.Tags,
		notification.TemplateID,
		notification.PrimaryRetryCount,
		notification.FallbackRetryCount,
		notification.TotalExecutionTime,
		notification.NotificationStatus,
		notification.IsPrimaryFailed,
		notification.IsFallbackFailed,
		false,
		notification.TrackingId,
	).Exec()

	if err != nil {
		log.Printf("Error saving notification: %v", err)
		return nil, err
	}
	return &notificationId, nil
}

func (r *NotificationRepository) UpdateNotificationByID(notificationID string, updateData *models.Notification) error {
	setClauses := []string{}
	values := []interface{}{}

	if updateData.Meta != nil {
		setClauses = append(setClauses, "meta = ?")
		values = append(values, updateData.Meta)
	}

	if len(updateData.Tags) > 0 {
		setClauses = append(setClauses, "tags = ?")
		values = append(values, updateData.Tags)
	}

	if updateData.PrimaryRetryCount != 0 {
		setClauses = append(setClauses, "primary_retry_count = ?")
		values = append(values, updateData.PrimaryRetryCount)
	}

	if updateData.FallbackRetryCount != 0 {
		setClauses = append(setClauses, "fallback_retry_count = ?")
		values = append(values, updateData.FallbackRetryCount)
	}

	if updateData.TotalExecutionTime != 0 {
		setClauses = append(setClauses, "total_execution_time = ?")
		values = append(values, updateData.TotalExecutionTime)
	}

	if updateData.NotificationStatus != "" {
		setClauses = append(setClauses, "notification_status = ?")
		values = append(values, updateData.NotificationStatus)
	}

	if updateData.ErrorMessage != "" {
		setClauses = append(setClauses, "error_message = ?")
		values = append(values, updateData.ErrorMessage)
	}

	if updateData.IsPrimaryFailed {
		setClauses = append(setClauses, "is_primary_failed = ?")
		values = append(values, updateData.IsPrimaryFailed)
	}

	if updateData.IsFallbackFailed {
		setClauses = append(setClauses, "is_fallback_failed = ?")
		values = append(values, updateData.IsFallbackFailed)
	}

	if updateData.IsMovedToDlq {
		setClauses = append(setClauses, "is_moved_to_dlq = ?")
		values = append(values, updateData.IsMovedToDlq)
	}

	if len(setClauses) == 0 {
		log.Printf("No fields to update for notificationID: %s", notificationID)
		return nil
	}

	query := fmt.Sprintf(`UPDATE notifications SET %s WHERE id = ?`, strings.Join(setClauses, ", "))
	values = append(values, notificationID)

	err := r.Session.Query(query, values...).Exec()
	if err != nil {
		log.Printf("Error updating notification: %v", err)
		return err
	}
	return nil
}
