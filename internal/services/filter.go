package services

import (
	"sort"
	"strings"

	"github.com/reynaldipane/flight-search/internal/models"
)

// FilterService handles filtering and sorting of flights
type FilterService struct{}

// NewFilterService creates a new filter service
func NewFilterService() *FilterService {
	return &FilterService{}
}

// ApplyFilters filters flights based on the provided options
func (s *FilterService) ApplyFilters(flights []*models.Flight, opts models.FilterOptions) []*models.Flight {
	filtered := make([]*models.Flight, 0, len(flights))

	for _, flight := range flights {
		if s.matchesFilters(flight, opts) {
			filtered = append(filtered, flight)
		}
	}

	return filtered
}

// matchesFilters checks if a flight matches all filter criteria
func (s *FilterService) matchesFilters(flight *models.Flight, opts models.FilterOptions) bool {
	// Filter by price range
	if opts.MinPrice != nil && flight.Price.Amount < *opts.MinPrice {
		return false
	}
	if opts.MaxPrice != nil && flight.Price.Amount > *opts.MaxPrice {
		return false
	}

	// Filter by max stops
	if opts.MaxStops != nil && flight.Stops > *opts.MaxStops {
		return false
	}

	// Filter by airlines
	if len(opts.Airlines) > 0 {
		found := false
		for _, airline := range opts.Airlines {
			if strings.EqualFold(flight.Airline.Name, airline) ||
				strings.EqualFold(flight.Airline.Code, airline) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by max duration
	if opts.MaxDuration != nil && flight.Duration.TotalMinutes > *opts.MaxDuration {
		return false
	}

	// Filter by departure time window
	if opts.DepartureTime != nil {
		if !s.matchesTimeWindow(flight.Departure.DateTime.Hour(), *opts.DepartureTime) {
			return false
		}
	}

	// Filter by arrival time window
	if opts.ArrivalTime != nil {
		if !s.matchesTimeWindow(flight.Arrival.DateTime.Hour(), *opts.ArrivalTime) {
			return false
		}
	}

	return true
}

// matchesTimeWindow checks if an hour matches the specified time window
func (s *FilterService) matchesTimeWindow(hour int, window string) bool {
	switch models.TimeWindow(window) {
	case models.TimeWindowMorning: // 06:00 - 11:59
		return hour >= 6 && hour < 12
	case models.TimeWindowAfternoon: // 12:00 - 17:59
		return hour >= 12 && hour < 18
	case models.TimeWindowEvening: // 18:00 - 23:59
		return hour >= 18 && hour < 24
	case models.TimeWindowNight: // 00:00 - 05:59
		return hour >= 0 && hour < 6
	default:
		return true // Unknown window, don't filter
	}
}

// SortFlights sorts flights based on the specified criteria
func (s *FilterService) SortFlights(flights []*models.Flight, sortBy models.SortBy) {
	switch sortBy {
	case models.SortByPriceAsc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Price.Amount < flights[j].Price.Amount
		})

	case models.SortByPriceDesc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Price.Amount > flights[j].Price.Amount
		})

	case models.SortByDurationAsc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Duration.TotalMinutes < flights[j].Duration.TotalMinutes
		})

	case models.SortByDurationDesc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Duration.TotalMinutes > flights[j].Duration.TotalMinutes
		})

	case models.SortByDepartureAsc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Departure.DateTime.Before(flights[j].Departure.DateTime)
		})

	case models.SortByDepartureDesc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Departure.DateTime.After(flights[j].Departure.DateTime)
		})

	case models.SortByArrivalAsc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Arrival.DateTime.Before(flights[j].Arrival.DateTime)
		})

	case models.SortByArrivalDesc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Arrival.DateTime.After(flights[j].Arrival.DateTime)
		})

	// Note: SortByBestValue will be implemented by RankingService
	}
}

// FilterAndSort applies filters and sorting in one operation
func (s *FilterService) FilterAndSort(flights []*models.Flight, opts models.FilterOptions, sortBy models.SortBy) []*models.Flight {
	// Apply filters first
	filtered := s.ApplyFilters(flights, opts)

	// Then sort
	s.SortFlights(filtered, sortBy)

	return filtered
}
