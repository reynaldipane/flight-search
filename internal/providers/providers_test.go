package providers

import (
	"context"
	"testing"

	"github.com/reynaldipane/flight-search/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllProviders_FetchFlights(t *testing.T) {
	// Create registry with all providers
	registry := NewRegistry()
	providers := registry.GetAll()

	require.Len(t, providers, 4, "Should have 4 providers registered")

	// Create search request
	req := &models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	t.Logf("Testing all providers with search: %s to %s", req.Origin, req.Destination)
	t.Logf("")

	totalFlights := 0
	successfulProviders := 0
	failedProviders := 0

	for _, provider := range providers {
		t.Run(provider.GetName(), func(t *testing.T) {
			t.Logf("Provider: %s", provider.GetName())
			t.Logf("  Configured Delay: %v", provider.GetDelay())
			t.Logf("  Failure Rate: %.0f%%", provider.GetFailureRate()*100)

			// Fetch flights
			flights, err := provider.FetchFlights(context.Background(), req)

			if err != nil {
				// AirAsia might fail due to 10% failure rate
				t.Logf("  Status: FAILED - %v", err)
				failedProviders++
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, flights, "Should return some flights")

			t.Logf("  Status: SUCCESS")
			t.Logf("  Flights Found: %d", len(flights))

			successfulProviders++
			totalFlights += len(flights)

			// Validate each flight
			for i, flight := range flights {
				t.Logf("  Flight %d:", i+1)
				t.Logf("    Number: %s", flight.FlightNumber)
				t.Logf("    Price: %.0f %s", flight.Price.Amount, flight.Price.Currency)
				t.Logf("    Duration: %s", flight.Duration.Formatted)
				t.Logf("    Stops: %d", flight.Stops)

				// Basic validations
				assert.NotEmpty(t, flight.ID)
				assert.Equal(t, provider.GetName(), flight.Provider)
				assert.Equal(t, req.Origin, flight.Departure.Airport)
				assert.NotZero(t, flight.Price.Amount)
				assert.Greater(t, flight.Duration.TotalMinutes, 0)
			}

			t.Logf("")
		})
	}

	t.Logf("==========================================")
	t.Logf("Summary:")
	t.Logf("  Providers Queried: %d", len(providers))
	t.Logf("  Successful: %d", successfulProviders)
	t.Logf("  Failed: %d", failedProviders)
	t.Logf("  Total Flights: %d", totalFlights)
	t.Logf("==========================================")

	// At least 3 providers should succeed (AirAsia might fail)
	assert.GreaterOrEqual(t, successfulProviders, 3, "At least 3 providers should succeed")
}

func TestProviderRegistry_GetByName(t *testing.T) {
	registry := NewRegistry()

	tests := []struct {
		name     string
		expected bool
	}{
		{"Garuda Indonesia", true},
		{"Lion Air", true},
		{"Batik Air", true},
		{"AirAsia", true},
		{"Unknown Airline", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := registry.GetByName(tt.name)

			if tt.expected {
				assert.NotNil(t, provider, "Provider %s should be found", tt.name)
				assert.Equal(t, tt.name, provider.GetName())
			} else {
				assert.Nil(t, provider, "Provider %s should not be found", tt.name)
			}
		})
	}
}

func TestProviderFailureRates(t *testing.T) {
	registry := NewRegistry()

	tests := []struct {
		name         string
		providerName string
		expectedRate float64
	}{
		{"Garuda Indonesia", "Garuda Indonesia", 0.0},
		{"Lion Air", "Lion Air", 0.0},
		{"Batik Air", "Batik Air", 0.0},
		{"AirAsia", "AirAsia", 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := registry.GetByName(tt.providerName)
			require.NotNil(t, provider, "Provider %s should exist", tt.providerName)
			assert.Equal(t, tt.expectedRate, provider.GetFailureRate(),
				"%s should have %.0f%% failure rate", tt.name, tt.expectedRate*100)
		})
	}
}
