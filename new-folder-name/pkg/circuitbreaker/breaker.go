package circuitbreaker

import (
	"time"

	"github.com/sony/gobreaker"
)

func NewCircuitBreaker(nameof string) *gobreaker.CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        nameof,
		MaxRequests: 3,
		Interval:    time.Minute,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	}
	return gobreaker.NewCircuitBreaker(settings)
}
