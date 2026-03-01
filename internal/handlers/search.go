package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reynaldipane/flight-search/internal/models"
	"github.com/reynaldipane/flight-search/internal/services"
	apperrors "github.com/reynaldipane/flight-search/pkg/errors"
)

// SearchHandler handles flight search requests
type SearchHandler struct {
	aggregator *services.AggregatorService
	cache      *services.CacheService
	filter     *services.FilterService
	ranking    *services.RankingService
	validator  *services.ValidatorService
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(
	aggregator *services.AggregatorService,
	cache *services.CacheService,
	filter *services.FilterService,
	ranking *services.RankingService,
	validator *services.ValidatorService,
) *SearchHandler {
	return &SearchHandler{
		aggregator: aggregator,
		cache:      cache,
		filter:     filter,
		ranking:    ranking,
		validator:  validator,
	}
}

// SearchFlights handles POST /search
// @Summary Search for flights
// @Description Search for flights across all providers with optional filtering and sorting
// @Tags flights
// @Accept json
// @Produce json
// @Param request body models.SearchRequest true "Search criteria"
// @Success 200 {object} models.SearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /search [post]
func (h *SearchHandler) SearchFlights(c *gin.Context) {
	var req models.SearchRequest

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Validate search request
	if err := h.validator.ValidateSearchRequest(&req); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.Status, ErrorResponse{
				Error:   appErr.Code,
				Message: appErr.Message,
			})
		} else {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "VALIDATION_ERROR",
				Message: err.Error(),
			})
		}
		return
	}

	// Generate cache key
	cacheKey := h.cache.GenerateKey(&req)

	// Check cache first
	if cached, found := h.cache.Get(cacheKey); found {
		cached.Metadata.CacheHit = true
		c.JSON(http.StatusOK, cached)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Aggregate flights from all providers
	response, err := h.aggregator.AggregateFlights(ctx, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.Status, ErrorResponse{
				Error:   appErr.Code,
				Message: appErr.Message,
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "INTERNAL_ERROR",
				Message: err.Error(),
			})
		}
		return
	}

	// Cache the response (5 minutes TTL)
	h.cache.Set(cacheKey, response, 5*time.Minute)

	// Return response
	c.JSON(http.StatusOK, response)
}

// SearchWithFilters handles POST /search/filter
// @Summary Search for flights with filters
// @Description Search for flights with advanced filtering and sorting options
// @Tags flights
// @Accept json
// @Produce json
// @Param request body SearchFilterRequest true "Search criteria with filters"
// @Success 200 {object} models.SearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /search/filter [post]
func (h *SearchHandler) SearchWithFilters(c *gin.Context) {
	var req SearchFilterRequest

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Validate search request
	if err := h.validator.ValidateSearchRequest(&req.SearchRequest); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.Status, ErrorResponse{
				Error:   appErr.Code,
				Message: appErr.Message,
			})
		} else {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "VALIDATION_ERROR",
				Message: err.Error(),
			})
		}
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Aggregate flights from all providers
	response, err := h.aggregator.AggregateFlights(ctx, &req.SearchRequest)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.JSON(appErr.Status, ErrorResponse{
				Error:   appErr.Code,
				Message: appErr.Message,
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "INTERNAL_ERROR",
				Message: err.Error(),
			})
		}
		return
	}

	// Apply filters and sorting
	filteredFlights := h.filter.ApplyFilters(response.Flights, req.Filters)

	// Sort by requested criteria or best value by default
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = models.SortByBestValue
	}

	// If sorting by best value, use ranking service
	if sortBy == models.SortByBestValue {
		filteredFlights = h.ranking.RankFlights(filteredFlights)
	} else {
		h.filter.SortFlights(filteredFlights, sortBy)
	}

	// Apply limit if specified
	if req.Limit > 0 && len(filteredFlights) > req.Limit {
		filteredFlights = filteredFlights[:req.Limit]
	}

	// Update response with filtered flights
	response.Flights = filteredFlights
	response.Metadata.TotalResults = len(filteredFlights)

	c.JSON(http.StatusOK, response)
}

// GetCacheStats handles GET /cache/stats
// @Summary Get cache statistics
// @Description Retrieve cache performance metrics
// @Tags cache
// @Produce json
// @Success 200 {object} CacheStatsResponse
// @Router /cache/stats [get]
func (h *SearchHandler) GetCacheStats(c *gin.Context) {
	hitRate := h.cache.GetHitRate()
	itemCount := h.cache.ItemCount()

	c.JSON(http.StatusOK, CacheStatsResponse{
		HitRate:   hitRate,
		ItemCount: itemCount,
	})
}

// ClearCache handles DELETE /cache
// @Summary Clear cache
// @Description Clear all cached search results
// @Tags cache
// @Success 200 {object} MessageResponse
// @Router /cache [delete]
func (h *SearchHandler) ClearCache(c *gin.Context) {
	h.cache.Clear()

	c.JSON(http.StatusOK, MessageResponse{
		Message: "Cache cleared successfully",
	})
}
