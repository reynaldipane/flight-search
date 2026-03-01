package services

import (
	"fmt"
	"regexp"
	"time"

	"github.com/reynaldipane/flight-search/internal/models"
)

// ValidatorService validates flight data and search requests
type ValidatorService struct {
	airportCodeRegex *regexp.Regexp
}

// NewValidatorService creates a new validator service
func NewValidatorService() *ValidatorService {
	return &ValidatorService{
		airportCodeRegex: regexp.MustCompile(`^[A-Z]{3}$`),
	}
}

// ValidateFlight validates a single flight's data
func (v *ValidatorService) ValidateFlight(flight *models.Flight) error {
	if flight == nil {
		return fmt.Errorf("flight cannot be nil")
	}

	// Validate required fields
	if flight.ID == "" {
		return fmt.Errorf("flight ID is required")
	}

	if flight.Provider == "" {
		return fmt.Errorf("provider name is required")
	}

	if flight.FlightNumber == "" {
		return fmt.Errorf("flight number is required")
	}

	// Validate airport codes
	if !v.airportCodeRegex.MatchString(flight.Departure.Airport) {
		return fmt.Errorf("invalid departure airport code: %s", flight.Departure.Airport)
	}

	if !v.airportCodeRegex.MatchString(flight.Arrival.Airport) {
		return fmt.Errorf("invalid arrival airport code: %s", flight.Arrival.Airport)
	}

	// Validate times
	if flight.Arrival.DateTime.Before(flight.Departure.DateTime) {
		return fmt.Errorf("arrival time (%s) must be after departure time (%s)",
			flight.Arrival.DateTime, flight.Departure.DateTime)
	}

	if flight.Arrival.DateTime.Equal(flight.Departure.DateTime) {
		return fmt.Errorf("arrival time cannot equal departure time")
	}

	// Validate duration
	if flight.Duration.TotalMinutes <= 0 {
		return fmt.Errorf("duration must be positive, got %d minutes", flight.Duration.TotalMinutes)
	}

	// Check if duration is reasonable (max 24 hours for domestic flights)
	if flight.Duration.TotalMinutes > 1440 {
		return fmt.Errorf("duration seems unreasonable: %d minutes (>24 hours)", flight.Duration.TotalMinutes)
	}

	// Verify duration matches actual time difference (with some tolerance)
	actualDuration := flight.Arrival.DateTime.Sub(flight.Departure.DateTime)
	actualMinutes := int(actualDuration.Minutes())

	// Allow 10% tolerance for duration differences (e.g., timezone handling, rounding)
	tolerance := max(10, flight.Duration.TotalMinutes/10)
	diff := abs(actualMinutes - flight.Duration.TotalMinutes)

	if diff > tolerance {
		return fmt.Errorf("duration mismatch: stated %d minutes, actual %d minutes (diff: %d)",
			flight.Duration.TotalMinutes, actualMinutes, diff)
	}

	// Validate price
	if flight.Price.Amount <= 0 {
		return fmt.Errorf("price must be positive, got %.2f", flight.Price.Amount)
	}

	if flight.Price.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	// Validate stops
	if flight.Stops < 0 {
		return fmt.Errorf("stops cannot be negative, got %d", flight.Stops)
	}

	// Validate available seats
	if flight.AvailableSeats < 0 {
		return fmt.Errorf("available seats cannot be negative, got %d", flight.AvailableSeats)
	}

	return nil
}

// ValidateFlights validates a slice of flights and returns only valid ones
func (v *ValidatorService) ValidateFlights(flights []*models.Flight) []*models.Flight {
	validFlights := make([]*models.Flight, 0, len(flights))

	for _, flight := range flights {
		if err := v.ValidateFlight(flight); err != nil {
			// Skip invalid flights (in production, you'd log this)
			continue
		}
		validFlights = append(validFlights, flight)
	}

	return validFlights
}

// ValidateSearchRequest validates a search request
func (v *ValidatorService) ValidateSearchRequest(req *models.SearchRequest) error {
	if req == nil {
		return fmt.Errorf("search request cannot be nil")
	}

	// Validate origin
	if !v.airportCodeRegex.MatchString(req.Origin) {
		return fmt.Errorf("invalid origin airport code: %s (must be 3 uppercase letters)", req.Origin)
	}

	// Validate destination
	if !v.airportCodeRegex.MatchString(req.Destination) {
		return fmt.Errorf("invalid destination airport code: %s (must be 3 uppercase letters)", req.Destination)
	}

	// Origin and destination must be different
	if req.Origin == req.Destination {
		return fmt.Errorf("origin and destination cannot be the same")
	}

	// Validate departure date format (YYYY-MM-DD)
	departureDate, err := time.Parse("2006-01-02", req.DepartureDate)
	if err != nil {
		return fmt.Errorf("invalid departure date format: %s (expected YYYY-MM-DD)", req.DepartureDate)
	}

	// Departure date should not be in the past (allow today)
	today := time.Now().Truncate(24 * time.Hour)
	if departureDate.Before(today) {
		return fmt.Errorf("departure date cannot be in the past")
	}

	// Validate return date if provided
	if req.ReturnDate != nil && *req.ReturnDate != "" {
		returnDate, err := time.Parse("2006-01-02", *req.ReturnDate)
		if err != nil {
			return fmt.Errorf("invalid return date format: %s (expected YYYY-MM-DD)", *req.ReturnDate)
		}

		// Return date must be after or equal to departure date
		if returnDate.Before(departureDate) {
			return fmt.Errorf("return date must be after departure date")
		}
	}

	// Validate passengers
	if req.Passengers < 1 {
		return fmt.Errorf("passengers must be at least 1, got %d", req.Passengers)
	}

	if req.Passengers > 9 {
		return fmt.Errorf("passengers cannot exceed 9, got %d", req.Passengers)
	}

	// Validate cabin class
	validClasses := map[string]bool{
		"economy":  true,
		"business": true,
		"first":    true,
	}

	if !validClasses[req.CabinClass] {
		return fmt.Errorf("invalid cabin class: %s (must be economy, business, or first)", req.CabinClass)
	}

	return nil
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
