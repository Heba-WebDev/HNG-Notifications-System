package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/franzego/stage04/pkg/circuitbreaker"
	"github.com/sony/gobreaker"
)

type UserServiceClient struct {
	baseURL    string
	httpClient *http.Client
	cb         *gobreaker.CircuitBreaker
	mockMode   bool
}

func NewUserServiceClient(baseURL string, mockMode bool) *UserServiceClient {
	return &UserServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		cb:       circuitbreaker.NewCircuitBreaker("user-service"),
		mockMode: mockMode,
	}
}

func (u *UserServiceClient) ValidateUser(ctx context.Context, userID string) (bool, error) {
	// for the mock mode before adding any the other services
	if u.mockMode {
		log.Print("Mock mode enabled: Simulating user validation")
		return true, nil
	}

	result, err := u.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, "GET",
			fmt.Sprintf("%s/users/%s", u.baseURL, userID), nil)
		if err != nil {
			return false, err
		}

		resp, err := u.httpClient.Do(req)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return true, nil
		}
		return false, fmt.Errorf("user not found")
	})

	if err != nil {
		return false, err
	}

	return result.(bool), nil
}
