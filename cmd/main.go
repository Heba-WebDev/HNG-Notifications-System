package main

import (
	"log"
	"net/http"

	"github.com/franzego/stage04/internal/config"
	"github.com/franzego/stage04/internal/handlers"
	"github.com/franzego/stage04/internal/middleware"
	"github.com/franzego/stage04/internal/queue"
	"github.com/franzego/stage04/internal/services"
	"github.com/franzego/stage04/pkg/redis"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config", err)
	}

	redisClient := redis.InitRedis(cfg.Redis)
	clientRabbit, err := queue.NewRabbitMqService(cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("failed to connect to rabbitMq")
	}
	defer clientRabbit.CloseConnection()
	userService := services.NewUserServiceClient(cfg.Services.UserServiceURL)
	templateService := services.NewTemplateClient(cfg.Services.TemplateServiceURL)
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
