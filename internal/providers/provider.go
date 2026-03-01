package providers

import (
	"context"
	"time"

	"github.com/reynaldipane/flight-search/internal/models"
	"github.com/reynaldipane/flight-search/internal/providers/airasia"
	"github.com/reynaldipane/flight-search/internal/providers/batikair"
	"github.com/reynaldipane/flight-search/internal/providers/garuda"
	"github.com/reynaldipane/flight-search/internal/providers/lionair"
)

// Provider is the interface that all airline providers must implement
type Provider interface {
	// GetName returns the provider's name (e.g., "Garuda Indonesia")
	GetName() string

	// FetchFlights fetches and normalizes flight data from the provider
	// Returns normalized flights or an error
	FetchFlights(ctx context.Context, req *models.SearchRequest) ([]*models.Flight, error)

	// GetDelay returns the simulated API delay for this provider
	GetDelay() time.Duration

	// GetFailureRate returns the probability of API failure (0.0 to 1.0)
	// e.g., 0.1 means 10% chance of failure
	GetFailureRate() float64
}

// Registry holds all registered providers
type Registry struct {
	providers []Provider
}

// NewRegistry creates a new provider registry with all providers
func NewRegistry() *Registry {
	return &Registry{
		providers: []Provider{
			garuda.New(),
			lionair.New(),
			batikair.New(),
			airasia.New(),
		},
	}
}

// GetAll returns all registered providers
func (r *Registry) GetAll() []Provider {
	return r.providers
}

// GetByName returns a provider by name, or nil if not found
func (r *Registry) GetByName(name string) Provider {
	for _, p := range r.providers {
		if p.GetName() == name {
			return p
		}
	}
	return nil
}
