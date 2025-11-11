package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/franzego/stage04/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMqClient struct {
	Conn      *amqp.Connection
	Channel   *amqp.Channel
	Config    config.RabbitMQConfig
	Connected bool
}

func NewRabbitMqService(cfg config.RabbitMQConfig) *RabbitMqClient {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		log.Fatal("there was an error connecting to rabbitmq")
	}
	channel, err := conn.Channel()
	if err != nil {
		log.Fatal("could not create a channel")
	}
	return &RabbitMqClient{
		Conn:      conn,
		Channel:   channel,
		Config:    cfg,
		Connected: true,
	}
}
func (r *RabbitMqClient) CloseConnection() {
	r.Channel.Close()
	r.Conn.Close()

}

// set up our exchange
func (r *RabbitMqClient) SetUpExchangeAndQueue() error {
	if err := r.Channel.ExchangeDeclare(
		r.Config.Exchange,
		"direct",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	); err != nil {
		return fmt.Errorf("error in declaring exchange")
	}
	queues := []string{
		r.Config.EmailQueue,
		r.Config.PushQueue,
		r.Config.FailedQueue,
	}
	for _, queueName := range queues {
		if _, err := r.Channel.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("error declaring queue")
		}
		err := r.Channel.QueueBind(
			queueName,
			queueName,
			r.Config.Exchange,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s: %w", queueName, err)
		}
	}
	return nil
}
func (r *RabbitMqClient) Publish(ctx context.Context, routingKey string, message interface{}) error {
	by, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	err = r.Channel.PublishWithContext(
		ctx,
		r.Config.Exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         by,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}
func (r *RabbitMqClient) PublishEmail(ctx context.Context, message interface{}) error {
	return r.Publish(ctx, r.Config.EmailQueue, message)
}
func (r *RabbitMqClient) PublishPushNot(ctx context.Context, message interface{}) error {
	return r.Publish(ctx, r.Config.PushQueue, message)
}
