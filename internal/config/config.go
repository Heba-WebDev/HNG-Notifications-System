package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	RabbitMQ RabbitMQConfig
	Redis    RedisConfig
	Services ServicesConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Port    string
	Timeout time.Duration
}

type RabbitMQConfig struct {
	URL         string
	EmailQueue  string
	PushQueue   string
	FailedQueue string
	Exchange    string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type ServicesConfig struct {
	UserServiceURL     string
	TemplateServiceURL string
}

type AuthConfig struct {
	JWTSecret string
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.timeout", "10s")
	viper.SetDefault("rabbitmq.exchange", "notifications.direct")
	viper.SetDefault("rabbitmq.email_queue", "email.queue")
	viper.SetDefault("rabbitmq.push_queue", "push.queue")
	viper.SetDefault("rabbitmq.failed_queue", "failed.queue")
	viper.SetDefault("redis.db", 0)

	// Read from environment
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// Config file not found, use environment variables
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
