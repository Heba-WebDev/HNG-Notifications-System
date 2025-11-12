package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	// tests are in the same package; do not import the package under test
	"github.com/franzego/stage04/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// Mock RabbitMQ Client
type MockRabbitMQClient struct {
	mock.Mock
}

func (m *MockRabbitMQClient) PublishEmail(ctx context.Context, message interface{}) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockRabbitMQClient) PublishPushNot(ctx context.Context, message interface{}) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockRabbitMQClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

// Mock User Service
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) ValidateUser(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

// Mock Template Service
type MockTemplateService struct {
	mock.Mock
}

func (m *MockTemplateService) ValidateTemplate(ctx context.Context, templateID string) (bool, error) {
	args := m.Called(ctx, templateID)
	return args.Bool(0), args.Error(1)
}

func TestSendEmail_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mocks
	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	// Configure mock expectations
	mockUserService.On("ValidateUser", mock.Anything, "user123").Return(true, nil)
	mockTemplateService.On("ValidateTemplate", mock.Anything, "welcome_email").Return(true, nil)
	mockQueue.On("PublishEmail", mock.Anything, mock.Anything).Return(nil)

	// Create handler
	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	// Setup router
	router := gin.New()
	router.POST("/notifications/email", handler.SendEmail)

	// Create request
	reqBody := models.SendEmailRequest{
		UserID:     "user123",
		TemplateID: "welcome_email",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/notifications/email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	assert.Equal(t, "Email notification queued successfully", response.Message)

	// Verify mocks were called
	mockUserService.AssertExpectations(t)
	mockTemplateService.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

func TestSendEmail_InvalidUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	// User validation fails
	mockUserService.On("ValidateUser", mock.Anything, "invalid_user").Return(false, nil)

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	router := gin.New()
	router.POST("/notifications/email", handler.SendEmail)

	reqBody := models.SendEmailRequest{
		UserID:     "invalid_user",
		TemplateID: "welcome_email",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/notifications/email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "User not found")
}

func setupMockRedis() *redis.Client {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	return rdb
}
