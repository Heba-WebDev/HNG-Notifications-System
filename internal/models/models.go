package models

import "time"

type NotificationMessage struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"` // "email" or "push"
	UserID        string                 `json:"user_id"`
	TemplateID    string                 `json:"template_id"`
	Variables     map[string]interface{} `json:"variables"`
	Priority      string                 `json:"priority"`
	ScheduledFor  *time.Time             `json:"scheduled_for,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlation_id"`
}
type SendEmailRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	TemplateID string `json:"template_id" binding:"required"`
}

type SendPushRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	TemplateID string `json:"template_id" binding:"required"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message"`
}

type NotificationResponse struct {
	NotificationID string    `json:"notification_id"`
	Status         string    `json:"status"`
	QueuedAt       time.Time `json:"queued_at"`
}
type NotificationStatus struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
