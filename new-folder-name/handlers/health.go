package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/franzego/stage04/internal/queue"
	"github.com/franzego/stage04/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	queue           *queue.RabbitMqClient
	redis           *redis.Client
	userService     *services.UserServiceClient
	templateService *services.TemplateServiceClient
}

func NewHealthHandler(
	queue *queue.RabbitMqClient,
	redis *redis.Client,
	userService *services.UserServiceClient,
	templateService *services.TemplateServiceClient,
) *HealthHandler {
	return &HealthHandler{
		queue:           queue,
		redis:           redis,
		userService:     userService,
		templateService: templateService,
	}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checks := make(map[string]string)

	// Check RabbitMQ
	if h.queue.IsConnected() {
		checks["rabbitmq"] = "healthy"
	} else {
		checks["rabbitmq"] = "unhealthy"
	}

	// Check Redis
	if err := h.redis.Ping(ctx).Err(); err == nil {
		checks["redis"] = "healthy"
	} else {
		checks["redis"] = "unhealthy"
	}

	// Check User Service (circuit breaker aware)
	if _, err := h.userService.ValidateUser(ctx, "health-check"); err == nil {
		checks["user_service"] = "healthy"
	} else {
		checks["user_service"] = "degraded"
	}

	// Check Template Service (circuit breaker aware)
	if _, err := h.templateService.ValidateTemplate(ctx, "health-check"); err == nil {
		checks["template_service"] = "healthy"
	} else {
		checks["template_service"] = "degraded"
	}

	// Determine overall status
	overallStatus := "healthy"
	for _, status := range checks {
		if status == "unhealthy" {
			overallStatus = "unhealthy"
			break
		} else if status == "degraded" {
			overallStatus = "degraded"
		}
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"status":    overallStatus,
		"timestamp": time.Now().Format(time.RFC3339),
		"checks":    checks,
		"version":   "1.0.0",
	})
}
