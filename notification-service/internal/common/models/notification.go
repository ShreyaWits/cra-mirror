package models

import (
	"github.com/gocql/gocql"
)

type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "PENDING"
	NotificationStatusSuccess NotificationStatus = "SUCCESS"
	NotificationStatusFailed  NotificationStatus = "FAILED"
)

// Recipient represents the contact information for a single recipient.
type Recipient struct {
	UserID         string            `json:"userID" cassandra:"user_id"`                  // Unique user ID
	Email          string            `json:"email" cassandra:"email"`                     // Email address
	Phone          string            `json:"phone" cassandra:"phone"`                     // Phone number
	WhatsappNumber string            `json:"whatsapp_number" cassandra:"whatsapp_number"` // WhatsApp number
	Data           map[string]string `json:"data" cassandra:"data"`                       // Flexible key-value data
}

type Notification struct {
	Id                 string             `json:"id" cassandra:"id"`
	Recipient          map[string]string  `json:"recipient" cassandra:"recipient"`
	Channel            string             `json:"channel" cassandra:"channel"`
	FallbackService    string             `json:"FallbackService" cassandra:"fallback_service"`
	Meta               map[string]string  `json:"meta" cassandra:"meta"`
	Tags               []string           `json:"tags" cassandra:"tags"`
	TemplateID         string             `json:"templateID" cassandra:"template_id"`
	PrimaryRetryCount  int                `json:"primaryRetryCount" cassandra:"primary_retry_count"`
	FallbackRetryCount int                `json:"fallbackRetryCount" cassandra:"fallback_retry_count"`
	TotalExecutionTime int                `json:"totalExecutionTime" cassandra:"total_execution_time"`
	NotificationStatus NotificationStatus `json:"notificationStatus" cassandra:"notification_status"`
	ErrorMessage       string             `json:"errorMessage" cassandra:"error_message"`
	PrimaryService     string             `json:"primaryService" cassandra:"primary_service"`
	IsPrimaryFailed    bool               `json:"isPrimaryFailed" cassandra:"is_primary_failed"`
	IsFallbackFailed   bool               `json:"isFallbackFailed" cassandra:"is_fallback_failed"`
	IsMovedToDlq       bool               `json:"isMovedToDlq" cassandra:"is_moved_to_dlq"`
	TrackingId         string             `json:"trackingId" cassandra:"tracking_id"`
}

// SuccessResponse represents the structure for a successful response.
type SuccessResponse struct {
	Status     string      `json:"status" cassandra:"status"`          // Status of the response
	StatusCode int         `json:"statusCode" cassandra:"status_code"` // HTTP Status Code
	Message    string      `json:"message" cassandra:"message"`        // Response message
	Data       interface{} `json:"data" cassandra:"data"`              // Response data
}

// ErrorResponse represents the structure for an error response.
type ErrorResponse struct {
	Status     string      `json:"status" cassandra:"status"`          // Status of the response
	StatusCode int         `json:"statusCode" cassandra:"status_code"` // HTTP Status Code
	Message    string      `json:"message" cassandra:"message"`        // Response message
	Error      interface{} `json:"error" cassandra:"error"`            // Error details
}

// Cassandra query model for creating the SendNotification table
const createSendNotificationTableQuery = `
CREATE TABLE IF NOT EXISTS notifications (
	id text,
    template_id text,
    recipient frozen<recipient>,
    channel text,
    service text,
    meta map<text, text>,
    tags list<text>,
    total_retry_count int,
    total_execution_time int,
    notification_status text,
    error_message text,
    PRIMARY KEY (template_id)
);
`

// CreateTable initializes the table in Cassandra
func CreateTable(session *gocql.Session) error {
	// Execute the Cassandra query to create the table
	err := session.Query(createSendNotificationTableQuery).Exec()
	if err != nil {
		return err
	}
	return nil
}
