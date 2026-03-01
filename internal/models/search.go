package models

// SearchRequest represents the flight search criteria from the user
type SearchRequest struct {
	Origin        string  `json:"origin" binding:"required"`        // e.g., "CGK"
	Destination   string  `json:"destination" binding:"required"`   // e.g., "DPS"
	DepartureDate string  `json:"departureDate" binding:"required"` // e.g., "2025-12-15"
	ReturnDate    *string `json:"returnDate"`                       // Optional, for round trips
	Passengers    int     `json:"passengers" binding:"required,min=1"`
	CabinClass    string  `json:"cabinClass" binding:"required"` // e.g., "economy", "business"
}

// SearchResponse represents the complete response sent back to the user
type SearchResponse struct {
	SearchCriteria SearchRequest    `json:"search_criteria"`
	Metadata       ResponseMetadata `json:"metadata"`
	Flights        []*Flight        `json:"flights"`
}

// ResponseMetadata contains information about the search execution
type ResponseMetadata struct {
	TotalResults       int      `json:"total_results"`
	ProvidersQueried   int      `json:"providers_queried"`
	ProvidersSucceeded int      `json:"providers_succeeded"`
	ProvidersFailed    int      `json:"providers_failed"`
	FailedProviders    []string `json:"failed_providers,omitempty"`
	SearchTimeMs       int64    `json:"search_time_ms"`
	CacheHit           bool     `json:"cache_hit"`
}
