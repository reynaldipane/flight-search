package handlers

import "github.com/reynaldipane/flight-search/internal/models"

// SearchFilterRequest combines search criteria with filters
type SearchFilterRequest struct {
	models.SearchRequest
	Filters models.FilterOptions `json:"filters,omitempty"`
	SortBy  models.SortBy        `json:"sort_by,omitempty"`
	Limit   int                  `json:"limit,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// MessageResponse represents a generic message response
type MessageResponse struct {
	Message string `json:"message"`
}

// CacheStatsResponse represents cache statistics
type CacheStatsResponse struct {
	HitRate   float64 `json:"hit_rate"`
	ItemCount int     `json:"item_count"`
}
