package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/franzego/stage04/internal/config"
	"github.com/franzego/stage04/internal/handlers"
	"github.com/franzego/stage04/internal/middleware"
	"github.com/franzego/stage04/internal/queue"
	"github.com/franzego/stage04/internal/services"
	"github.com/franzego/stage04/pkg/redis"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func isValidRabbitMQURL(url string) bool {
	// Empty URL
	if url == "" {
		return false
	}

	// Check for mock/placeholder URLs
	lowerURL := strings.ToLower(url)
	invalidIndicators := []string{
		"mock",
		"localhost",
		"127.0.0.1",
		"example",
		"test",
		"fake",
	}

	for _, indicator := range invalidIndicators {
		if strings.Contains(lowerURL, indicator) {
			return false
		}
	}

	// Must start with amqp:// or amqps://
	if !strings.HasPrefix(url, "amqp://") && !strings.HasPrefix(url, "amqps://") {
		return false
	}

	return true
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config", err)
	}
	if cfg.MockServices {
		log.Print("Running in MOCK MODE - external services simulated")
	}

	redisClient := redis.InitRedis(cfg.Redis)

	var rabbitMQClient *queue.RabbitMqClient
	if isValidRabbitMQURL(cfg.RabbitMQ.URL) {
		rabbitMQClient, err = queue.NewRabbitMqService(cfg.RabbitMQ)
		if err != nil {
			log.Print("Failed to connect to RabbitMQ, running in MOCK mode")
			rabbitMQClient = nil // Use nil to indicate mock mode
		} else {
			log.Print("RabbitMQ connected")
		}
	} else {
		log.Print("No valid RabbitMQ URL, running in MOCK mode",
			zap.String("url", cfg.RabbitMQ.URL),
		)
		rabbitMQClient = nil
	}

	if rabbitMQClient != nil {
		defer rabbitMQClient.CloseConnection()
	}

	clientRabbit, err := queue.NewRabbitMqService(cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("failed to connect to rabbitMq")
	}
	defer clientRabbit.CloseConnection()
	userService := services.NewUserServiceClient(cfg.Services.UserServiceURL, cfg.MockServices)
	templateService := services.NewTemplateClient(cfg.Services.TemplateServiceURL, cfg.MockServices)
	notificationHandler := handlers.NewNotificationService(
		clientRabbit,
		redisClient,
		userService,
		templateService,
	)
	healthHandler := handlers.NewHealthHandler(clientRabbit, redisClient, userService, templateService)

	r := gin.Default()
	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())
	{
		api.POST("/notification/email", notificationHandler.SendEmail)
		api.POST("/notification/push", notificationHandler.SendPush)
		api.GET("/notification/status/:id", notificationHandler.GetStatus)

	}

	r.GET("/health", healthHandler.HealthCheck)

	r.GET("/Alive", func(c *gin.Context) {
		// Return JSON response
		c.JSON(http.StatusOK, gin.H{
			"status":  "Alive",
			"service": "api-gateway",
		})
	})

	r.Run()
}
