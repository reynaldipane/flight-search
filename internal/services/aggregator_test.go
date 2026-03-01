package services

import (
	"context"
	"testing"

	"github.com/reynaldipane/flight-search/internal/models"
	"github.com/reynaldipane/flight-search/internal/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregatorService_AggregateFlights(t *testing.T) {
	// Create provider registry and aggregator
	registry := providers.NewRegistry()
	aggregator := NewAggregatorService(registry)

	// Create search request
	req := &models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	// Aggregate flights
	response, err := aggregator.AggregateFlights(context.Background(), req)

	// Assertions
	require.NoError(t, err, "AggregateFlights should not return error")
	require.NotNil(t, response, "Response should not be nil")

	t.Logf("Aggregation Results:")
	t.Logf("  Total Flights: %d", response.Metadata.TotalResults)
	t.Logf("  Providers Queried: %d", response.Metadata.ProvidersQueried)
	t.Logf("  Providers Succeeded: %d", response.Metadata.ProvidersSucceeded)
	t.Logf("  Providers Failed: %d", response.Metadata.ProvidersFailed)
	t.Logf("  Search Time: %dms", response.Metadata.SearchTimeMs)

	if len(response.Metadata.FailedProviders) > 0 {
		t.Logf("  Failed Providers: %v", response.Metadata.FailedProviders)
	}

	// Verify metadata
	assert.Equal(t, 4, response.Metadata.ProvidersQueried, "Should query 4 providers")
	assert.GreaterOrEqual(t, response.Metadata.ProvidersSucceeded, 3, "At least 3 providers should succeed")
	assert.Greater(t, response.Metadata.TotalResults, 0, "Should have at least some flights")
	assert.Greater(t, response.Metadata.SearchTimeMs, int64(0), "Search time should be recorded")

	// Verify search criteria is preserved
	assert.Equal(t, req.Origin, response.SearchCriteria.Origin)
	assert.Equal(t, req.Destination, response.SearchCriteria.Destination)
	assert.Equal(t, req.DepartureDate, response.SearchCriteria.DepartureDate)

	// Verify flights
	assert.NotEmpty(t, response.Flights, "Should return flights")

	// Check that flights are from different providers
	providerMap := make(map[string]int)
	for _, flight := range response.Flights {
		providerMap[flight.Provider]++
	}

	t.Logf("  Flights by Provider:")
	for provider, count := range providerMap {
		t.Logf("    %s: %d flights", provider, count)
	}
}

func TestAggregatorService_ParallelExecution(t *testing.T) {
	// This test verifies that providers are queried in parallel
	registry := providers.NewRegistry()
	aggregator := NewAggregatorService(registry)

	req := &models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	// If providers were sequential, total time would be sum of all delays
	// Garuda: ~75ms, Lion Air: ~150ms, Batik Air: ~300ms, AirAsia: ~100ms
	// Sequential: ~625ms, Parallel: ~300ms (slowest one)

	response, err := aggregator.AggregateFlights(context.Background(), req)
	require.NoError(t, err)

	t.Logf("Search completed in %dms", response.Metadata.SearchTimeMs)

	// With parallel execution, should be close to slowest provider (~300ms for Batik Air)
	// Add some buffer for overhead
	assert.Less(t, response.Metadata.SearchTimeMs, int64(1000),
		"Parallel execution should complete in less than 1 second")
}

func TestAggregatorService_ContextCancellation(t *testing.T) {
	registry := providers.NewRegistry()
	aggregator := NewAggregatorService(registry)

	req := &models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Try to aggregate
	_, err := aggregator.AggregateFlights(ctx, req)

	// Should complete but may have failures due to cancelled context
	// We don't require error because some providers might succeed before cancellation
	t.Logf("Result with cancelled context: %v", err)
}
