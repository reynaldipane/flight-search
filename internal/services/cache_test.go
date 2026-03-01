package services

import (
	"testing"
	"time"

	"github.com/reynaldipane/flight-search/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCacheService_SetAndGet(t *testing.T) {
	cache := NewCacheService(5*time.Minute, 10*time.Minute)

	// Create test data
	req := &models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	response := &models.SearchResponse{
		SearchCriteria: *req,
		Metadata: models.ResponseMetadata{
			TotalResults: 10,
		},
		Flights: []*models.Flight{},
	}

	key := cache.GenerateKey(req)

	t.Run("Cache Miss - First Access", func(t *testing.T) {
		_, found := cache.Get(key)
		assert.False(t, found, "Should be cache miss on first access")

		metrics := cache.GetMetrics()
		assert.Equal(t, uint64(0), metrics.hits)
		assert.Equal(t, uint64(1), metrics.misses)
	})

	t.Run("Set Cache", func(t *testing.T) {
		cache.Set(key, response, 0) // 0 means use default expiration
		assert.Equal(t, 1, cache.ItemCount(), "Cache should have 1 item")
	})

	t.Run("Cache Hit - Second Access", func(t *testing.T) {
		cached, found := cache.Get(key)
		assert.True(t, found, "Should be cache hit on second access")
		assert.NotNil(t, cached)
		assert.Equal(t, response.Metadata.TotalResults, cached.Metadata.TotalResults)

		metrics := cache.GetMetrics()
		assert.Equal(t, uint64(1), metrics.hits)
		assert.Equal(t, uint64(1), metrics.misses)
	})
}

func TestCacheService_GenerateKey(t *testing.T) {
	cache := NewCacheService(5*time.Minute, 10*time.Minute)

	t.Run("One Way Trip", func(t *testing.T) {
		req := &models.SearchRequest{
			Origin:        "CGK",
			Destination:   "DPS",
			DepartureDate: "2025-12-15",
			Passengers:    1,
			CabinClass:    "economy",
		}

		key := cache.GenerateKey(req)
		expected := "CGK_DPS_2025-12-15__1_economy"
		assert.Equal(t, expected, key)
	})

	t.Run("Round Trip", func(t *testing.T) {
		returnDate := "2025-12-20"
		req := &models.SearchRequest{
			Origin:        "CGK",
			Destination:   "DPS",
			DepartureDate: "2025-12-15",
			ReturnDate:    &returnDate,
			Passengers:    2,
			CabinClass:    "business",
		}

		key := cache.GenerateKey(req)
		expected := "CGK_DPS_2025-12-15_2025-12-20_2_business"
		assert.Equal(t, expected, key)
	})

	t.Run("Different Requests Different Keys", func(t *testing.T) {
		req1 := &models.SearchRequest{
			Origin:        "CGK",
			Destination:   "DPS",
			DepartureDate: "2025-12-15",
			Passengers:    1,
			CabinClass:    "economy",
		}

		req2 := &models.SearchRequest{
			Origin:        "CGK",
			Destination:   "SUB",
			DepartureDate: "2025-12-15",
			Passengers:    1,
			CabinClass:    "economy",
		}

		key1 := cache.GenerateKey(req1)
		key2 := cache.GenerateKey(req2)

		assert.NotEqual(t, key1, key2, "Different requests should have different keys")
	})
}

func TestCacheService_Delete(t *testing.T) {
	cache := NewCacheService(5*time.Minute, 10*time.Minute)

	req := &models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	response := &models.SearchResponse{
		SearchCriteria: *req,
	}

	key := cache.GenerateKey(req)

	// Set item
	cache.Set(key, response, 0) // 0 means use default expiration
	assert.Equal(t, 1, cache.ItemCount())

	// Get item (should exist)
	_, found := cache.Get(key)
	assert.True(t, found)

	// Delete item
	cache.Delete(key)
	assert.Equal(t, 0, cache.ItemCount())

	// Get item (should not exist)
	_, found = cache.Get(key)
	assert.False(t, found)
}

func TestCacheService_Clear(t *testing.T) {
	cache := NewCacheService(5*time.Minute, 10*time.Minute)

	// Add multiple items
	for i := 1; i <= 5; i++ {
		req := &models.SearchRequest{
			Origin:        "CGK",
			Destination:   "DPS",
			DepartureDate: "2025-12-15",
			Passengers:    i,
			CabinClass:    "economy",
		}

		response := &models.SearchResponse{
			SearchCriteria: *req,
		}

		key := cache.GenerateKey(req)
		cache.Set(key, response, 0) // 0 means use default expiration
	}

	assert.Equal(t, 5, cache.ItemCount(), "Should have 5 items")

	// Clear all
	cache.Clear()
	assert.Equal(t, 0, cache.ItemCount(), "Should have 0 items after clear")
}

func TestCacheService_Expiration(t *testing.T) {
	// Create cache with very short TTL for testing
	cache := NewCacheService(100*time.Millisecond, 50*time.Millisecond)

	req := &models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	response := &models.SearchResponse{
		SearchCriteria: *req,
	}

	key := cache.GenerateKey(req)

	// Set item with short TTL
	cache.Set(key, response, 100*time.Millisecond)

	// Should exist immediately
	_, found := cache.Get(key)
	assert.True(t, found, "Item should exist immediately")

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Should be expired now
	_, found = cache.Get(key)
	assert.False(t, found, "Item should be expired")
}

func TestCacheService_Metrics(t *testing.T) {
	cache := NewCacheService(5*time.Minute, 10*time.Minute)

	req := &models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	response := &models.SearchResponse{
		SearchCriteria: *req,
	}

	key := cache.GenerateKey(req)

	// Initial state
	metrics := cache.GetMetrics()
	assert.Equal(t, uint64(0), metrics.hits)
	assert.Equal(t, uint64(0), metrics.misses)
	assert.Equal(t, 0.0, cache.GetHitRate())

	// Miss
	cache.Get(key)
	metrics = cache.GetMetrics()
	assert.Equal(t, uint64(0), metrics.hits)
	assert.Equal(t, uint64(1), metrics.misses)
	assert.Equal(t, 0.0, cache.GetHitRate())

	// Set and hit
	cache.Set(key, response, 0) // 0 means use default expiration
	cache.Get(key)
	cache.Get(key)
	cache.Get(key)

	metrics = cache.GetMetrics()
	assert.Equal(t, uint64(3), metrics.hits)
	assert.Equal(t, uint64(1), metrics.misses)

	// Hit rate: 3 hits / 4 total = 75%
	assert.Equal(t, 75.0, cache.GetHitRate())
}

func TestCacheService_ConcurrentAccess(t *testing.T) {
	cache := NewCacheService(5*time.Minute, 10*time.Minute)

	req := &models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	response := &models.SearchResponse{
		SearchCriteria: *req,
	}

	key := cache.GenerateKey(req)
	cache.Set(key, response, 0) // 0 means use default expiration

	// Concurrent reads
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_, found := cache.Get(key)
			assert.True(t, found)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	metrics := cache.GetMetrics()
	assert.Equal(t, uint64(10), metrics.hits, "Should handle concurrent reads")
}
