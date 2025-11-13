package redis

import (
	"context"
	"log"
	"time"

	"github.com/franzego/stage04/internal/config"
	"github.com/redis/go-redis/v9"
)

func InitRedis(cfg config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         "redis://redis.railway.internal:6379",
		Password:     "GpEMHuxDTvYLZLGRwPHvBvUhrxsiVvka",
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis with addr: %s", cfg.Addr)

	}
	log.Printf("connected to redis successfully on addr: %s", cfg.Addr)
	return client
}
