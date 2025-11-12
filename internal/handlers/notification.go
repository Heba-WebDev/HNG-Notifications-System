package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/franzego/stage04/internal/models"
	"github.com/franzego/stage04/internal/queue"
	"github.com/franzego/stage04/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type NotificationHandler struct {
	rabbitClient    *queue.RabbitMqClient
	redis           *redis.Client
	userService     *services.UserServiceClient
	templateService *services.TemplateServiceClient
}

func NewNotificationService(
	queue *queue.RabbitMqClient,
	redis *redis.Client,
	userService *services.UserServiceClient,
	templateService *services.TemplateServiceClient,
) *NotificationHandler {
	return &NotificationHandler{
		rabbitClient:    queue,
		redis:           redis,
		userService:     userService,
		templateService: templateService,
	}
}
func (n *NotificationHandler) SendEmail(c *gin.Context) {
	ctx := context.Background()
	correlationID, _ := c.Get("correlation_id")
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
		CorrelationID: correlationID.(string),
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
	correlationID, _ := c.Get("correlation_id")
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
		CorrelationID: correlationID.(string),
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
