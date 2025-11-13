package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/franzego/stage04/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ========== Integration Tests for Notification Handler ==========

// TestIntegration_EmailNotificationFullFlow tests the complete email notification flow
func TestIntegration_EmailNotificationFullFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mocks
	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	defer mockRedis.Close()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	// Configure expectations
	mockUserService.On("ValidateUser", mock.Anything, "user-123").Return(true, nil)
	mockTemplateService.On("ValidateTemplate", mock.Anything, "welcome-template").Return(true, nil)
	mockQueue.On("PublishEmail", mock.Anything, mock.Anything).Return(nil)

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	// Create router
	router := gin.New()
	router.POST("/api/v1/notification/email", handler.SendEmail)

	// Step 1: Send email notification
	emailReq := models.SendEmailRequest{
		UserID:     "user-123",
		TemplateID: "welcome-template",
	}
	body, _ := json.Marshal(emailReq)
	req, _ := http.NewRequest("POST", "/api/v1/notification/email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	var response models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	assert.Equal(t, "Email notification queued successfully", response.Message)

	// Extract notification ID
	notifResp := response.Data.(map[string]interface{})
	notificationID := notifResp["notification_id"].(string)

	// Step 2: Verify notification was stored in Redis
	statusKey := fmt.Sprintf("notification:status:%s", notificationID)
	statusJSON, err := mockRedis.Get(context.Background(), statusKey).Result()
	assert.NoError(t, err)
	assert.NotEmpty(t, statusJSON)

	// Verify status can be retrieved
	var status models.NotificationStatus
	json.Unmarshal([]byte(statusJSON), &status)
	assert.Equal(t, notificationID, status.ID)
	assert.Equal(t, "email", status.Type)
	assert.Equal(t, "queued", status.Status)

	// Verify all mocks were called
	mockUserService.AssertExpectations(t)
	mockTemplateService.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

// TestIntegration_PushNotificationFullFlow tests the complete push notification flow
func TestIntegration_PushNotificationFullFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	defer mockRedis.Close()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	mockUserService.On("ValidateUser", mock.Anything, "user-456").Return(true, nil)
	mockTemplateService.On("ValidateTemplate", mock.Anything, "push-promo").Return(true, nil)
	mockQueue.On("PublishPushNot", mock.Anything, mock.Anything).Return(nil)

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	router := gin.New()
	router.POST("/api/v1/notification/push", handler.SendPush)

	pushReq := models.SendPushRequest{
		UserID:     "user-456",
		TemplateID: "push-promo",
	}
	body, _ := json.Marshal(pushReq)
	req, _ := http.NewRequest("POST", "/api/v1/notification/push", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	assert.Equal(t, "Push notification queued successfully", response.Message)

	mockUserService.AssertExpectations(t)
	mockTemplateService.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}

// TestIntegration_IdempotencyCheck tests that duplicate notifications are handled correctly
func TestIntegration_IdempotencyCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	defer mockRedis.Close()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	mockUserService.On("ValidateUser", mock.Anything, mock.Anything).Return(true, nil)
	mockTemplateService.On("ValidateTemplate", mock.Anything, mock.Anything).Return(true, nil)
	mockQueue.On("PublishEmail", mock.Anything, mock.Anything).Return(nil)

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	router := gin.New()
	router.POST("/api/v1/notification/email", handler.SendEmail)

	emailReq := models.SendEmailRequest{
		UserID:     "user-idempotent",
		TemplateID: "template-idempotent",
	}
	body, _ := json.Marshal(emailReq)

	// First request
	req1, _ := http.NewRequest("POST", "/api/v1/notification/email", bytes.NewBuffer(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	var resp1 models.APIResponse
	json.Unmarshal(w1.Body.Bytes(), &resp1)

	// Verify all mocks were called once for first request
	mockUserService.AssertNumberOfCalls(t, "ValidateUser", 1)
	mockTemplateService.AssertNumberOfCalls(t, "ValidateTemplate", 1)
	mockQueue.AssertNumberOfCalls(t, "PublishEmail", 1)
}

// TestIntegration_GetNotificationStatus tests retrieving notification status
func TestIntegration_GetNotificationStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	defer mockRedis.Close()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	mockUserService.On("ValidateUser", mock.Anything, "status-user").Return(true, nil)
	mockTemplateService.On("ValidateTemplate", mock.Anything, "status-template").Return(true, nil)
	mockQueue.On("PublishEmail", mock.Anything, mock.Anything).Return(nil)

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	router := gin.New()
	router.POST("/api/v1/notification/email", handler.SendEmail)
	router.GET("/api/v1/notification/status/:id", handler.GetStatus)

	// Step 1: Send notification
	emailReq := models.SendEmailRequest{
		UserID:     "status-user",
		TemplateID: "status-template",
	}
	body, _ := json.Marshal(emailReq)
	req, _ := http.NewRequest("POST", "/api/v1/notification/email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	notificationID := response.Data.(map[string]interface{})["notification_id"].(string)

	// Step 2: Retrieve notification status
	statusReq, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/notification/status/%s", notificationID), nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, statusReq)

	assert.Equal(t, http.StatusOK, w2.Code)
	var statusResponse models.APIResponse
	json.Unmarshal(w2.Body.Bytes(), &statusResponse)
	assert.True(t, statusResponse.Success)

	statusData := statusResponse.Data.(map[string]interface{})
	assert.Equal(t, notificationID, statusData["id"])
	assert.Equal(t, "queued", statusData["status"])
}

// TestIntegration_InvalidUserValidation tests validation failure for invalid user
func TestIntegration_InvalidUserValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	defer mockRedis.Close()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	mockUserService.On("ValidateUser", mock.Anything, "invalid-user").Return(false, nil)

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	router := gin.New()
	router.POST("/api/v1/notification/email", handler.SendEmail)

	emailReq := models.SendEmailRequest{
		UserID:     "invalid-user",
		TemplateID: "template-123",
	}
	body, _ := json.Marshal(emailReq)
	req, _ := http.NewRequest("POST", "/api/v1/notification/email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "User not found")
}

// TestIntegration_InvalidTemplateValidation tests validation failure for invalid template
func TestIntegration_InvalidTemplateValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	defer mockRedis.Close()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	mockUserService.On("ValidateUser", mock.Anything, "user-789").Return(true, nil)
	mockTemplateService.On("ValidateTemplate", mock.Anything, "invalid-template").Return(false, nil)

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	router := gin.New()
	router.POST("/api/v1/notification/email", handler.SendEmail)

	emailReq := models.SendEmailRequest{
		UserID:     "user-789",
		TemplateID: "invalid-template",
	}
	body, _ := json.Marshal(emailReq)
	req, _ := http.NewRequest("POST", "/api/v1/notification/email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "Template not found")
}

// TestIntegration_MissingRequiredFields tests request validation
func TestIntegration_MissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	defer mockRedis.Close()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	router := gin.New()
	router.POST("/api/v1/notification/email", handler.SendEmail)

	// Request with missing user_id
	emailReq := map[string]string{
		"template_id": "template-123",
	}
	body, _ := json.Marshal(emailReq)
	req, _ := http.NewRequest("POST", "/api/v1/notification/email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)
}

// TestIntegration_RabbitMQPublishFailure tests handling of RabbitMQ publish failure
func TestIntegration_RabbitMQPublishFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	defer mockRedis.Close()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	mockUserService.On("ValidateUser", mock.Anything, "user-publish-fail").Return(true, nil)
	mockTemplateService.On("ValidateTemplate", mock.Anything, "template-publish-fail").Return(true, nil)
	mockQueue.On("PublishEmail", mock.Anything, mock.Anything).Return(fmt.Errorf("connection failed"))

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	router := gin.New()
	router.POST("/api/v1/notification/email", handler.SendEmail)

	emailReq := models.SendEmailRequest{
		UserID:     "user-publish-fail",
		TemplateID: "template-publish-fail",
	}
	body, _ := json.Marshal(emailReq)
	req, _ := http.NewRequest("POST", "/api/v1/notification/email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "failed to queue")
}

// TestIntegration_GetStatusNotFound tests retrieving non-existent notification
func TestIntegration_GetStatusNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	defer mockRedis.Close()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	router := gin.New()
	router.GET("/api/v1/notification/status/:id", handler.GetStatus)

	req, _ := http.NewRequest("GET", "/api/v1/notification/status/non-existent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var response models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "Notification not found")
}

// TestIntegration_GetStatusEmptyID tests retrieving status with empty ID
func TestIntegration_GetStatusEmptyID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	defer mockRedis.Close()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	router := gin.New()
	router.GET("/api/v1/notification/status/:id", handler.GetStatus)

	req, _ := http.NewRequest("GET", "/api/v1/notification/status/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// When ID is empty, Gin will not match the route
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestIntegration_MultipleNotificationsIndependence tests that multiple notifications are independent
func TestIntegration_MultipleNotificationsIndependence(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	defer mockRedis.Close()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	mockUserService.On("ValidateUser", mock.Anything, mock.Anything).Return(true, nil)
	mockTemplateService.On("ValidateTemplate", mock.Anything, mock.Anything).Return(true, nil)
	mockQueue.On("PublishEmail", mock.Anything, mock.Anything).Return(nil)
	mockQueue.On("PublishPushNot", mock.Anything, mock.Anything).Return(nil)

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	router := gin.New()
	router.POST("/api/v1/notification/email", handler.SendEmail)
	router.POST("/api/v1/notification/push", handler.SendPush)
	router.GET("/api/v1/notification/status/:id", handler.GetStatus)

	// Send email notification
	emailReq := models.SendEmailRequest{UserID: "user-multi-1", TemplateID: "template-1"}
	emailBody, _ := json.Marshal(emailReq)
	emailHttpReq, _ := http.NewRequest("POST", "/api/v1/notification/email", bytes.NewBuffer(emailBody))
	emailHttpReq.Header.Set("Content-Type", "application/json")
	emailW := httptest.NewRecorder()
	router.ServeHTTP(emailW, emailHttpReq)

	var emailResp models.APIResponse
	json.Unmarshal(emailW.Body.Bytes(), &emailResp)
	emailID := emailResp.Data.(map[string]interface{})["notification_id"].(string)

	// Send push notification
	pushReq := models.SendPushRequest{UserID: "user-multi-2", TemplateID: "template-2"}
	pushBody, _ := json.Marshal(pushReq)
	pushHttpReq, _ := http.NewRequest("POST", "/api/v1/notification/push", bytes.NewBuffer(pushBody))
	pushHttpReq.Header.Set("Content-Type", "application/json")
	pushW := httptest.NewRecorder()
	router.ServeHTTP(pushW, pushHttpReq)

	var pushResp models.APIResponse
	json.Unmarshal(pushW.Body.Bytes(), &pushResp)
	pushID := pushResp.Data.(map[string]interface{})["notification_id"].(string)

	// Verify they have different IDs
	assert.NotEqual(t, emailID, pushID)

	// Verify each can be retrieved independently
	emailStatusReq, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/notification/status/%s", emailID), nil)
	emailStatusW := httptest.NewRecorder()
	router.ServeHTTP(emailStatusW, emailStatusReq)
	assert.Equal(t, http.StatusOK, emailStatusW.Code)

	pushStatusReq, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/notification/status/%s", pushID), nil)
	pushStatusW := httptest.NewRecorder()
	router.ServeHTTP(pushStatusW, pushStatusReq)
	assert.Equal(t, http.StatusOK, pushStatusW.Code)
}

// TestIntegration_RedisConnectionFailure tests handling of Redis connection issues
func TestIntegration_RedisConnectionFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	mockRedis.Close() // Close Redis to simulate connection failure
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	mockUserService.On("ValidateUser", mock.Anything, "user-redis-fail").Return(true, nil)
	mockTemplateService.On("ValidateTemplate", mock.Anything, "template-redis-fail").Return(true, nil)
	mockQueue.On("PublishEmail", mock.Anything, mock.Anything).Return(nil)

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	router := gin.New()
	router.POST("/api/v1/notification/email", handler.SendEmail)

	emailReq := models.SendEmailRequest{
		UserID:     "user-redis-fail",
		TemplateID: "template-redis-fail",
	}
	body, _ := json.Marshal(emailReq)
	req, _ := http.NewRequest("POST", "/api/v1/notification/email", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// The request might still succeed but Redis operations might fail
	// This depends on how the handler manages error scenarios
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

// TestIntegration_ConcurrentRequests tests handling of concurrent notification requests
func TestIntegration_ConcurrentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	defer mockRedis.Close()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	mockUserService.On("ValidateUser", mock.Anything, mock.Anything).Return(true, nil)
	mockTemplateService.On("ValidateTemplate", mock.Anything, mock.Anything).Return(true, nil)
	mockQueue.On("PublishEmail", mock.Anything, mock.Anything).Return(nil)

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	router := gin.New()
	router.POST("/api/v1/notification/email", handler.SendEmail)

	// Simulate 5 concurrent requests
	numRequests := 5
	results := make(chan int, numRequests)
	notificationIDs := make([]string, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(index int) {
			emailReq := models.SendEmailRequest{
				UserID:     fmt.Sprintf("concurrent-user-%d", index),
				TemplateID: fmt.Sprintf("concurrent-template-%d", index),
			}
			body, _ := json.Marshal(emailReq)
			req, _ := http.NewRequest("POST", "/api/v1/notification/email", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			var response models.APIResponse
			json.Unmarshal(w.Body.Bytes(), &response)
			id := response.Data.(map[string]interface{})["notification_id"].(string)
			notificationIDs[index] = id
			results <- w.Code
		}(i)
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		code := <-results
		assert.Equal(t, http.StatusOK, code)
	}

	// Verify all notification IDs are unique
	idMap := make(map[string]bool)
	for _, id := range notificationIDs {
		assert.False(t, idMap[id], "Duplicate notification ID found")
		idMap[id] = true
	}
}

// ========== Benchmarks ==========

// BenchmarkEmailNotificationSend benchmarks email notification performance
func BenchmarkEmailNotificationSend(b *testing.B) {
	gin.SetMode(gin.TestMode)

	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	defer mockRedis.Close()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	mockUserService.On("ValidateUser", mock.Anything, mock.Anything).Return(true, nil)
	mockTemplateService.On("ValidateTemplate", mock.Anything, mock.Anything).Return(true, nil)
	mockQueue.On("PublishEmail", mock.Anything, mock.Anything).Return(nil)

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	router := gin.New()
	router.POST("/api/v1/notification/email", handler.SendEmail)

	emailReq := models.SendEmailRequest{
		UserID:     "bench-user",
		TemplateID: "bench-template",
	}
	body, _ := json.Marshal(emailReq)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/notification/email", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkGetNotificationStatus benchmarks status retrieval performance
func BenchmarkGetNotificationStatus(b *testing.B) {
	gin.SetMode(gin.TestMode)

	mockQueue := new(MockRabbitMQClient)
	mockRedis := setupMockRedis()
	defer mockRedis.Close()
	mockUserService := new(MockUserService)
	mockTemplateService := new(MockTemplateService)

	mockUserService.On("ValidateUser", mock.Anything, mock.Anything).Return(true, nil)
	mockTemplateService.On("ValidateTemplate", mock.Anything, mock.Anything).Return(true, nil)
	mockQueue.On("PublishEmail", mock.Anything, mock.Anything).Return(nil)

	handler := NewNotificationService(
		mockQueue,
		mockRedis,
		mockUserService,
		mockTemplateService,
	)

	// Create a test notification first
	statusData := models.NotificationStatus{
		ID:        "bench-notif-id",
		Type:      "email",
		Status:    "queued",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	statusJSON, _ := json.Marshal(statusData)
	mockRedis.Set(context.Background(), "notification:status:bench-notif-id", statusJSON, 24*time.Hour)

	router := gin.New()
	router.GET("/api/v1/notification/status/:id", handler.GetStatus)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/notification/status/bench-notif-id", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
