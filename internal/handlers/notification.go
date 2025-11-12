package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/franzego/stage04/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type NotificationHandler struct {
	rabbitClient    RabbitClient
	redis           *redis.Client
	userService     UserService
	templateService TemplateService
}

// RabbitClient defines the methods used from the RabbitMq client. Using an
// interface makes testing easier (mocks can implement this).
type RabbitClient interface {
	PublishEmail(ctx context.Context, message interface{}) error
	PublishPushNot(ctx context.Context, message interface{}) error
	IsConnected() bool
}

func NewNotificationService(
	queue RabbitClient,
	redis *redis.Client,
	userService UserService,
	templateService TemplateService,
) *NotificationHandler {
	return &NotificationHandler{
		rabbitClient:    queue,
		redis:           redis,
		userService:     userService,
		templateService: templateService,
	}
}

// UserService defines the subset of methods used from the user service client.
type UserService interface {
	ValidateUser(ctx context.Context, userID string) (bool, error)
}

// TemplateService defines the subset of methods used from the template service client.
type TemplateService interface {
	ValidateTemplate(ctx context.Context, templateID string) (bool, error)
}

func (n *NotificationHandler) SendEmail(c *gin.Context) {
	ctx := context.Background()
	correlationIDVal, _ := c.Get("correlation_id")
	correlationID, _ := correlationIDVal.(string)
	now := time.Now()
	// parse the req
	var req models.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
			Message: "Invalid Request Body",
		})
		return
	}
	notificationID := uuid.New().String()
	isDuplicate, err := n.CheckIdempoteny(ctx, notificationID)
	if err != nil {
		log.Printf("idempotency check failed:%v", err)
	}
	if isDuplicate {
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Error:   err.Error(),
			Message: "Notification Already Processed",
			Data: models.NotificationResponse{
				NotificationID: notificationID,
				Status:         "processing",
				QueuedAt:       now,
			},
		})
		return
	}
	valUser, err := n.userService.ValidateUser(ctx, req.UserID)
	if err != nil || !valUser {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "User not found or unavailable",
			Message: "User not available",
		})
		return
	}
	validTemplate, err := n.templateService.ValidateTemplate(ctx, req.TemplateID)
	if err != nil || !validTemplate {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Template not found or unavailable",
			Message: "Validation failed",
		})
		return
	}
	message := models.NotificationMessage{
		ID:            notificationID,
		Type:          "email",
		UserID:        req.UserID,
		TemplateID:    req.TemplateID,
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
	}
	if err := n.rabbitClient.PublishEmail(ctx, message); err != nil {
		log.Printf("failed to publish email")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "failed to queue notification",
			Message: "Internal Server Error",
		})
		return
	}
	if err := n.storeNotificationStatus(ctx, notificationID, "queued", "email"); err != nil {
		log.Printf("failed to log notification status: %v", err)
	}
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Email notification queued successfully",
		Data: models.NotificationResponse{
			NotificationID: notificationID,
			Status:         "queued",
			QueuedAt:       time.Now(),
		},
	})

}
func (n *NotificationHandler) SendPush(c *gin.Context) {
	ctx := context.Background()
	correlationIDVal, _ := c.Get("correlation_id")
	correlationID, _ := correlationIDVal.(string)
	now := time.Now()
	var req models.SendPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
			Message: "Invalid Request Body",
		})
		return
	}
	notificationID := uuid.New().String()
	isDuplicate, err := n.CheckIdempoteny(ctx, notificationID)
	if err != nil {
		log.Printf("idempotency check failed:%v", err)
	}
	if isDuplicate {
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Error:   err.Error(),
			Message: "Notification Already Processed",
			Data: models.NotificationResponse{
				NotificationID: notificationID,
				Status:         "processing",
				QueuedAt:       now,
			},
		})
		return
	}
	valUser, err := n.userService.ValidateUser(ctx, req.UserID)
	if err != nil || !valUser {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "User not found or unavailable",
			Message: "User not available",
		})
		return
	}
	validTemplate, err := n.templateService.ValidateTemplate(ctx, req.TemplateID)
	if err != nil || !validTemplate {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Template not found or unavailable",
			Message: "Validation failed",
		})
		return
	}
	message := models.NotificationMessage{
		ID:            notificationID,
		Type:          "push",
		UserID:        req.UserID,
		TemplateID:    req.TemplateID,
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
	}
	if err := n.rabbitClient.PublishPushNot(ctx, message); err != nil {
		log.Printf("failed to publish push notification")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "failed to queue push notification",
			Message: "Internal Server Error",
		})
		return
	}
	if err := n.storeNotificationStatus(ctx, notificationID, "queued", "push"); err != nil {
		log.Printf("failed to log push notification status: %v", err)
	}
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Push notification queued successfully",
		Data: models.NotificationResponse{
			NotificationID: notificationID,
			Status:         "queued",
			QueuedAt:       time.Now(),
		},
	})

}
func (n *NotificationHandler) CheckIdempoteny(ctx context.Context, notificationID string) (bool, error) {
	key := fmt.Sprintf("notification:idempotency:%s", notificationID)
	exists, err := n.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, nil
	}
	if exists > 0 {
		return true, nil
	}
	err = n.redis.Set(ctx, key, "processing", 24*time.Hour).Err()
	return false, err

}
func (n *NotificationHandler) storeNotificationStatus(ctx context.Context, notificationID, status, notifType string) error {
	statusData := models.NotificationStatus{
		ID:        notificationID,
		Type:      notifType,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	statusJSON, err := json.Marshal(statusData)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("notification:status:%s", notificationID)
	return n.redis.Set(ctx, key, statusJSON, 24*time.Hour).Err()
}
func (n *NotificationHandler) GetStatus(c *gin.Context) {
	ctx := c.Request.Context()
	notificationID := c.Param("id")

	if notificationID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Notification ID required",
			Message: "Invalid request",
		})
		return
	}

	// Get status from Redis
	statusKey := fmt.Sprintf("notification:status:%s", notificationID)
	statusJSON, err := n.redis.Get(ctx, statusKey).Result()
	if err == redis.Nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "Notification not found",
			Message: "Not found",
		})
		return
	}
	if err != nil {
		log.Print("Failed to get notification status")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to retrieve status",
			Message: "Internal server error",
		})
		return
	}

	var status models.NotificationStatus
	if err := json.Unmarshal([]byte(statusJSON), &status); err != nil {
		log.Print("Failed to unmarshal status")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to parse status",
			Message: "Internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Status retrieved successfully",
		Data:    status,
	})
}
