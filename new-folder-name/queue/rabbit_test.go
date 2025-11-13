package queue

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockRabbitMQClient struct {
	mock.Mock
}

func (m *MockRabbitMQClient) PublishToEmailQueue(ctx context.Context, message interface{}) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockRabbitMQClient) PublishToPushQueue(ctx context.Context, message interface{}) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockRabbitMQClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}
