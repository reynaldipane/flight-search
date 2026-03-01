package services

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/reynaldipane/flight-search/internal/models"
)

// CacheService provides in-memory caching for search results
type CacheService struct {
	cache   *cache.Cache
	mu      sync.RWMutex
	metrics CacheMetrics
}

// CacheMetrics tracks cache performance
type CacheMetrics struct {
	hits   uint64
	misses uint64
}

// NewCacheService creates a new cache service
// defaultTTL: how long items stay in cache (e.g., 5 minutes)
// cleanupInterval: how often to clean up expired items (e.g., 10 minutes)
func NewCacheService(defaultTTL, cleanupInterval time.Duration) *CacheService {
	return &CacheService{
		cache: cache.New(defaultTTL, cleanupInterval),
	}
}

// Get retrieves a search response from cache
func (s *CacheService) Get(key string) (*models.SearchResponse, bool) {
	item, found := s.cache.Get(key)

	if !found {
		atomic.AddUint64(&s.metrics.misses, 1)
		return nil, false
	}

	atomic.AddUint64(&s.metrics.hits, 1)

	if response, ok := item.(*models.SearchResponse); ok {
		return response, true
	}

	// Item exists but wrong type (shouldn't happen)
	return nil, false
}

// Set stores a search response in cache
func (s *CacheService) Set(key string, response *models.SearchResponse, ttl time.Duration) {
	s.cache.Set(key, response, ttl)
}

// Delete removes an item from cache
func (s *CacheService) Delete(key string) {
	s.cache.Delete(key)
}

// Clear removes all items from cache
func (s *CacheService) Clear() {
	s.cache.Flush()
}

// GenerateKey creates a cache key from a search request
func (s *CacheService) GenerateKey(req *models.SearchRequest) string {
	returnDate := ""
	if req.ReturnDate != nil {
		returnDate = *req.ReturnDate
	}

	return fmt.Sprintf("%s_%s_%s_%s_%d_%s",
		req.Origin,
		req.Destination,
		req.DepartureDate,
		returnDate,
		req.Passengers,
		req.CabinClass,
	)
}

// GetMetrics returns current cache metrics
func (s *CacheService) GetMetrics() CacheMetrics {
	return CacheMetrics{
		hits:   atomic.LoadUint64(&s.metrics.hits),
		misses: atomic.LoadUint64(&s.metrics.misses),
	}
}

// GetHitRate returns the cache hit rate as a percentage (0-100)
func (s *CacheService) GetHitRate() float64 {
	metrics := s.GetMetrics()
	total := metrics.hits + metrics.misses

	if total == 0 {
		return 0.0
	}

	return (float64(metrics.hits) / float64(total)) * 100
}

// ItemCount returns the number of items currently in cache
func (s *CacheService) ItemCount() int {
	return s.cache.ItemCount()
}
