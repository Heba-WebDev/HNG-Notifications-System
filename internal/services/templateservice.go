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

type TemplateServiceClient struct {
	baseUrl    string
	httpClient *http.Client
	cb         *gobreaker.CircuitBreaker
	mockMode   bool
}

func NewTemplateClient(baseUrl string, mockmode bool) *TemplateServiceClient {
	return &TemplateServiceClient{
		baseUrl: baseUrl,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		cb:       circuitbreaker.NewCircuitBreaker("template-service"),
		mockMode: mockmode,
	}
}
func (t *TemplateServiceClient) ValidateTemplate(ctx context.Context, templateID string) (bool, error) {
	if t.mockMode {
		log.Print("Mock mode enabled: Simulating template validation")
		return true, nil
	}
	result, err := t.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, "GET",
			fmt.Sprintf("%s/templates/%s", t.baseUrl, templateID), nil)
		if err != nil {
			return false, err
		}

		resp, err := t.httpClient.Do(req)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return true, nil
		}
		return false, fmt.Errorf("template not found")
	})

	if err != nil {
		return false, err
	}
	return result.(bool), nil

}
