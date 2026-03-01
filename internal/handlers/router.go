package handlers

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/reynaldipane/flight-search/internal/providers"
	"github.com/reynaldipane/flight-search/internal/services"
)

// SetupRouter creates and configures the Gin router
func SetupRouter() *gin.Engine {
	// Create Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize services
	providerRegistry := providers.NewRegistry()
	aggregator := services.NewAggregatorService(providerRegistry)
	cache := services.NewCacheService(5*time.Minute, 10*time.Minute)
	filter := services.NewFilterService()
	ranking := services.NewRankingService()
	validator := services.NewValidatorService()

	// Initialize handlers
	searchHandler := NewSearchHandler(aggregator, cache, filter, ranking, validator)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", HealthCheck)

		// Flight search routes
		v1.POST("/search", searchHandler.SearchFlights)
		v1.POST("/search/filter", searchHandler.SearchWithFilters)

		// Cache routes
		v1.GET("/cache/stats", searchHandler.GetCacheStats)
		v1.DELETE("/cache", searchHandler.ClearCache)

		// Provider info routes
		v1.GET("/providers", ListProviders(providerRegistry))
	}

	return router
}

// HealthCheck handles GET /health
// @Summary Health check
// @Description Check if the service is running
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "flight-search-api",
		"version": "1.0.0",
	})
}

// ListProviders returns a handler that lists all available providers
// @Summary List providers
// @Description Get information about all flight providers
// @Tags providers
// @Produce json
// @Success 200 {object} ProvidersResponse
// @Router /providers [get]
func ListProviders(registry *providers.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		allProviders := registry.GetAll()
		providerInfo := make([]ProviderInfo, 0, len(allProviders))

		for _, p := range allProviders {
			providerInfo = append(providerInfo, ProviderInfo{
				Name:        p.GetName(),
				Delay:       p.GetDelay().String(),
				FailureRate: p.GetFailureRate(),
			})
		}

		c.JSON(200, ProvidersResponse{
			Providers: providerInfo,
			Total:     len(providerInfo),
		})
	}
}

// ProviderInfo represents provider information
type ProviderInfo struct {
	Name        string  `json:"name"`
	Delay       string  `json:"delay"`
	FailureRate float64 `json:"failure_rate"`
}

// ProvidersResponse represents the list of providers
type ProvidersResponse struct {
	Providers []ProviderInfo `json:"providers"`
	Total     int            `json:"total"`
}
