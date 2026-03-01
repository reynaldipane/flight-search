package base

import (
	"context"
	"math/rand"
	"time"

	apperrors "github.com/reynaldipane/flight-search/pkg/errors"
)

// Provider contains common functionality for all providers
type Provider struct {
	name        string
	delay       time.Duration
	failureRate float64
}

// NewProvider creates a new base provider
func NewProvider(name string, minDelay, maxDelay time.Duration, failureRate float64) *Provider {
	// Random delay between min and max
	delay := minDelay + time.Duration(rand.Int63n(int64(maxDelay-minDelay)))

	return &Provider{
		name:        name,
		delay:       delay,
		failureRate: failureRate,
	}
}

// GetName returns the provider name
func (b *Provider) GetName() string {
	return b.name
}

// GetDelay returns the simulated delay
func (b *Provider) GetDelay() time.Duration {
	return b.delay
}

// GetFailureRate returns the failure rate
func (b *Provider) GetFailureRate() float64 {
	return b.failureRate
}

// SimulateDelay simulates API latency and potential failure
// Returns error if the simulated failure occurs
func (b *Provider) SimulateDelay(ctx context.Context) error {
	// Check if provider should fail (based on failure rate)
	// Example: if failureRate = 0.1 (10%), and rand.Float64() returns 0.05,
	// then 0.05 < 0.1 is TRUE, so the request fails
	if rand.Float64() < b.failureRate {
		return apperrors.ErrProviderUnavailable
	}

	// Simulate network delay
	select {
	case <-time.After(b.delay):
		return nil
	case <-ctx.Done():
		// Context was cancelled or timed out
		return ctx.Err()
	}
}
