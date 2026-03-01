package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/reynaldipane/flight-search/internal/models"
	"github.com/reynaldipane/flight-search/internal/providers"
)

// AggregatorService aggregates flight data from multiple providers
type AggregatorService struct {
	registry *providers.Registry
}

// NewAggregatorService creates a new aggregator service
func NewAggregatorService(registry *providers.Registry) *AggregatorService {
	return &AggregatorService{
		registry: registry,
	}
}

// providerResult holds the result from a single provider query
type providerResult struct {
	providerName string
	flights      []*models.Flight
	err          error
	duration     time.Duration
}

// AggregateFlights queries all providers in parallel and aggregates results
func (s *AggregatorService) AggregateFlights(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	startTime := time.Now()

	// Get all providers
	allProviders := s.registry.GetAll()
	if len(allProviders) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Create buffered channel for results
	resultsChan := make(chan providerResult, len(allProviders))

	// Use WaitGroup to coordinate goroutines
	var wg sync.WaitGroup

	// Launch a goroutine for each provider
	for _, provider := range allProviders {
		wg.Add(1)
		go s.fetchFromProvider(ctx, provider, req, resultsChan, &wg)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var allFlights []*models.Flight
	var failedProviders []string
	successCount := 0
	failCount := 0

	for result := range resultsChan {
		if result.err != nil {
			// Provider failed - log it but continue
			failedProviders = append(failedProviders, result.providerName)
			failCount++
			continue
		}

		// Provider succeeded
		successCount++
		allFlights = append(allFlights, result.flights...)
	}

	// Calculate total search time
	searchTime := time.Since(startTime)

	// Build response
	response := &models.SearchResponse{
		SearchCriteria: *req,
		Metadata: models.ResponseMetadata{
			TotalResults:       len(allFlights),
			ProvidersQueried:   len(allProviders),
			ProvidersSucceeded: successCount,
			ProvidersFailed:    failCount,
			FailedProviders:    failedProviders,
			SearchTimeMs:       searchTime.Milliseconds(),
			CacheHit:           false,
		},
		Flights: allFlights,
	}

	return response, nil
}

// fetchFromProvider fetches flights from a single provider
func (s *AggregatorService) fetchFromProvider(
	ctx context.Context,
	provider providers.Provider,
	req *models.SearchRequest,
	resultsChan chan<- providerResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	startTime := time.Now()

	// Create context with timeout for this provider (5 seconds max)
	providerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Fetch flights from provider
	flights, err := provider.FetchFlights(providerCtx, req)
	duration := time.Since(startTime)

	// Send result to channel
	resultsChan <- providerResult{
		providerName: provider.GetName(),
		flights:      flights,
		err:          err,
		duration:     duration,
	}
}
