package models

// FilterOptions represents the criteria for filtering flight results
type FilterOptions struct {
	MinPrice      *float64 `json:"min_price,omitempty"`      // Minimum price in IDR
	MaxPrice      *float64 `json:"max_price,omitempty"`      // Maximum price in IDR
	MaxStops      *int     `json:"max_stops,omitempty"`      // Maximum number of stops (0 = direct only)
	Airlines      []string `json:"airlines,omitempty"`       // Filter by specific airlines
	MaxDuration   *int     `json:"max_duration,omitempty"`   // Maximum duration in minutes
	DepartureTime *string  `json:"departure_time,omitempty"` // Time window: "morning", "afternoon", "evening", "night"
	ArrivalTime   *string  `json:"arrival_time,omitempty"`   // Time window: "morning", "afternoon", "evening", "night"
}

// SortBy represents the sorting option for flights
type SortBy string

const (
	SortByPriceAsc      SortBy = "price_asc"       // Cheapest first
	SortByPriceDesc     SortBy = "price_desc"      // Most expensive first
	SortByDurationAsc   SortBy = "duration_asc"    // Shortest duration first
	SortByDurationDesc  SortBy = "duration_desc"   // Longest duration first
	SortByDepartureAsc  SortBy = "departure_asc"   // Earliest departure first
	SortByDepartureDesc SortBy = "departure_desc"  // Latest departure first
	SortByArrivalAsc    SortBy = "arrival_asc"     // Earliest arrival first
	SortByArrivalDesc   SortBy = "arrival_desc"    // Latest arrival first
	SortByBestValue     SortBy = "best_value"      // Best value algorithm (price + convenience)
)

// TimeWindow represents time of day for filtering
type TimeWindow string

const (
	TimeWindowMorning   TimeWindow = "morning"   // 06:00 - 11:59
	TimeWindowAfternoon TimeWindow = "afternoon" // 12:00 - 17:59
	TimeWindowEvening   TimeWindow = "evening"   // 18:00 - 23:59
	TimeWindowNight     TimeWindow = "night"     // 00:00 - 05:59
)
