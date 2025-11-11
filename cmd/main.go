package main

import (
	"log"
	"net/http"

	"github.com/franzego/stage04/internal/config"
	"github.com/franzego/stage04/internal/queue"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config", err)
	}
	clientRabbit := queue.NewRabbitMqService(cfg.RabbitMQ)
	defer clientRabbit.CloseConnection()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		// Return JSON response
		c.JSON(http.StatusOK, gin.H{
			"status":  "Alive",
			"service": "api-gateway",
		})
	})

	r.Run()
}
