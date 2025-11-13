# Testing Guide - HNG Notifications System

## 📋 Overview

This document provides a comprehensive guide to the testing infrastructure for the HNG Notifications System API Gateway.

## 🏗️ Test Structure

```
internal/
├── handlers/
│   ├── notification_test.go          # Existing unit tests
│   ├── integration_test.go           # NEW: Comprehensive integration tests
│   ├── notification.go               # Main notification handler
│   ├── health.go                     # Health check handler
│   └── middleware.go                 # Middleware utilities
├── queue/
│   ├── rabbit_test.go                # RabbitMQ mock tests
│   └── rabbit.go                     # RabbitMQ client
└── ...
```

## 🚀 Quick Start

### Run All Tests
```bash
cd /home/franz/hng/stage04
go test ./... -v
```

### Run Handler Tests Only
```bash
go test ./internal/handlers -v
```

### Run Only Integration Tests
```bash
go test ./internal/handlers -v -run Integration
```

### View Test Coverage
```bash
go test ./internal/handlers -v -cover
```

## 📊 Test Suite Overview

### Integration Tests (13 tests in `integration_test.go`)

| # | Test Name | Purpose | Status |
|---|-----------|---------|--------|
| 1 | `TestIntegration_EmailNotificationFullFlow` | Complete email notification flow | ✅ PASS |
| 2 | `TestIntegration_PushNotificationFullFlow` | Complete push notification flow | ✅ PASS |
| 3 | `TestIntegration_IdempotencyCheck` | Duplicate notification handling | ✅ PASS |
| 4 | `TestIntegration_GetNotificationStatus` | Status retrieval | ✅ PASS |
| 5 | `TestIntegration_InvalidUserValidation` | Invalid user error handling | ✅ PASS |
| 6 | `TestIntegration_InvalidTemplateValidation` | Invalid template error handling | ✅ PASS |
| 7 | `TestIntegration_MissingRequiredFields` | Request validation | ✅ PASS |
| 8 | `TestIntegration_RabbitMQPublishFailure` | RabbitMQ failure handling | ✅ PASS |
| 9 | `TestIntegration_GetStatusNotFound` | 404 handling for missing status | ✅ PASS |
| 10 | `TestIntegration_GetStatusEmptyID` | Empty ID handling | ✅ PASS |
| 11 | `TestIntegration_MultipleNotificationsIndependence` | Multiple notifications isolation | ✅ PASS |
| 12 | `TestIntegration_RedisConnectionFailure` | Redis connection resilience | ✅ PASS |
| 13 | `TestIntegration_ConcurrentRequests` | Concurrent request handling | ✅ PASS |

### Unit Tests (2 tests in `notification_test.go`)

| # | Test Name | Purpose | Status |
|---|-----------|---------|--------|
| 1 | `TestSendEmail_Success` | Successful email send | ✅ PASS |
| 2 | `TestSendEmail_InvalidUser` | Invalid user handling | ✅ PASS |

### Benchmarks (2 benchmarks in `integration_test.go`)

| # | Benchmark Name | Purpose |
|---|----------------|---------|
| 1 | `BenchmarkEmailNotificationSend` | Email notification performance |
| 2 | `BenchmarkGetNotificationStatus` | Status retrieval performance |

## 🧪 Test Categories

### 1. **Happy Path Tests**
Tests successful scenarios where all validations pass and operations complete successfully.
- `TestIntegration_EmailNotificationFullFlow`
- `TestIntegration_PushNotificationFullFlow`
- `TestIntegration_GetNotificationStatus`

### 2. **Validation Tests**
Tests validation logic for requests and data.
- `TestIntegration_InvalidUserValidation`
- `TestIntegration_InvalidTemplateValidation`
- `TestIntegration_MissingRequiredFields`

### 3. **Error Handling Tests**
Tests graceful handling of errors and edge cases.
- `TestIntegration_RabbitMQPublishFailure`
- `TestIntegration_GetStatusNotFound`
- `TestIntegration_RedisConnectionFailure`

### 4. **Concurrency & Isolation Tests**
Tests system behavior under concurrent operations.
- `TestIntegration_IdempotencyCheck`
- `TestIntegration_MultipleNotificationsIndependence`
- `TestIntegration_ConcurrentRequests`

## 📈 Test Metrics

**Current Coverage**: 57.3% of statements in handlers package

### Coverage Breakdown
```
handlers/notification.go  - 57.3% coverage
- SendEmail()            - Covered
- SendPush()             - Covered
- GetStatus()            - Covered
- CheckIdempotency()     - Covered
- storeNotificationStatus() - Covered
```

## 🔧 Test Setup & Infrastructure

### Mock Objects

All tests use mock objects to simulate external dependencies:

```go
// Mock RabbitMQ Client
type MockRabbitMQClient struct {
    mock.Mock
}
func (m *MockRabbitMQClient) PublishEmail(ctx context.Context, message interface{}) error
func (m *MockRabbitMQClient) PublishPushNot(ctx context.Context, message interface{}) error
func (m *MockRabbitMQClient) IsConnected() bool

// Mock User Service
type MockUserService struct {
    mock.Mock
}
func (m *MockUserService) ValidateUser(ctx context.Context, userID string) (bool, error)

// Mock Template Service
type MockTemplateService struct {
    mock.Mock
}
func (m *MockTemplateService) ValidateTemplate(ctx context.Context, templateID string) (bool, error)
```

### Test Utilities

```go
// setupMockRedis() - Creates in-memory Redis instance
// Uses miniredis package for isolated Redis testing
func setupMockRedis() *redis.Client
```

## 💾 Test Dependencies

Add these to `go.mod`:
```
github.com/stretchr/testify v1.11.1      # Assertions & Mocking
github.com/alicebob/miniredis/v2 v2.21.0 # In-memory Redis
github.com/gin-gonic/gin v1.11.0         # HTTP Framework
```

## 🎯 Test Patterns Used

### Pattern 1: Arrange-Act-Assert (AAA)
```go
// Arrange
mockQueue := new(MockRabbitMQClient)
mockQueue.On("PublishEmail", mock.Anything, mock.Anything).Return(nil)

// Act
w := httptest.NewRecorder()
router.ServeHTTP(w, req)

// Assert
assert.Equal(t, http.StatusOK, w.Code)
```

### Pattern 2: Table-Driven Tests
Useful for testing multiple scenarios:
```go
tests := []struct {
    name    string
    userID  string
    wantErr bool
}{
    {"valid user", "user-123", false},
    {"invalid user", "invalid", true},
}
```

### Pattern 3: Mock Expectations
```go
mockUserService.On("ValidateUser", mock.Anything, "user-123").Return(true, nil)
// ... test code ...
mockUserService.AssertExpectations(t)
```

## 🚦 Running Tests in Different Ways

### 1. **Verbose Output**
```bash
go test ./internal/handlers -v
```

### 2. **With Timeout**
```bash
go test ./internal/handlers -v -timeout=30s
```

### 3. **With Race Detection**
```bash
go test ./internal/handlers -race
```

### 4. **Parallel Execution**
```bash
go test ./internal/handlers -parallel 4
```

### 5. **Generate Coverage Report**
```bash
go test ./internal/handlers -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### 6. **Run Specific Tests by Pattern**
```bash
go test -run Email ./internal/handlers -v      # All tests with "Email" in name
go test -run Integration ./internal/handlers -v # All integration tests
```

## 📝 Best Practices

### ✅ DO:
- Use table-driven tests for multiple scenarios
- Mock external dependencies (RabbitMQ, Redis, HTTP calls)
- Test both happy path and error cases
- Use descriptive test names
- Verify mock call expectations
- Clean up resources (defer Close())
- Use assert library for clear failure messages

### ❌ DON'T:
- Don't test the testing framework itself
- Don't make real external calls in unit tests
- Don't use sleep() for synchronization
- Don't ignore test failures
- Don't hardcode values without context
- Don't skip cleanup in tests

## 🔍 Debugging Tests

### Print Debug Information
```bash
go test -v -run TestIntegration_EmailNotificationFullFlow ./internal/handlers
```

### Enable Race Detector
```bash
go test -race ./internal/handlers
```

### Get Detailed Failure Info
```bash
go test -v -failfast ./internal/handlers  # Stop at first failure
```

### Profile Test Performance
```bash
go test -cpuprofile=cpu.prof ./internal/handlers
go tool pprof cpu.prof
```

## 🚀 CI/CD Integration

### GitHub Actions Example
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.25
      - run: go test ./... -v -cover
```

## 📋 Checklist for Adding New Tests

- [ ] Test has clear, descriptive name starting with `Test` or `Benchmark`
- [ ] Test covers specific functionality
- [ ] Arrange-Act-Assert pattern is followed
- [ ] Mocks are properly configured
- [ ] Assertions are specific (not just checking success)
- [ ] Edge cases are tested
- [ ] Cleanup is done (defer statements)
- [ ] Test runs in isolation (no dependencies on other tests)
- [ ] Test is documented with comments
- [ ] Test runs successfully: `go test -v`

## 📚 Resources

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Miniredis GitHub](https://github.com/alicebob/miniredis)
- [Go Best Practices](https://golang.org/doc/effective_go)

## 🆘 Troubleshooting

### Issue: "redis: client is closed"
**Solution**: Expected when testing Redis connection failures. This is handled gracefully.

### Issue: Tests timeout
**Solution**: Increase timeout with `-timeout` flag:
```bash
go test ./internal/handlers -timeout=60s
```

### Issue: Mock not matching
**Solution**: Verify:
1. Mock method names are exact
2. Parameter types match exactly
3. Number of calls matches expectations

### Issue: Intermittent test failures
**Solution**: Could be race condition:
```bash
go test -race ./internal/handlers
```

## 📞 Support

For questions or issues:
1. Check the TEST_SUMMARY.md file
2. Review test examples in integration_test.go
3. Check Go testing documentation
4. Create an issue in the repository

---

**Last Updated**: November 13, 2025
**Go Version**: 1.25.0
**Test Framework**: Go testing + testify
