package services

import (
	"testing"
	"time"

	"github.com/reynaldipane/flight-search/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestValidatorService_ValidateFlight(t *testing.T) {
	validator := NewValidatorService()

	// Create a valid flight
	now := time.Now()
	validFlight := &models.Flight{
		ID:           "TEST123_Provider",
		Provider:     "Test Airline",
		FlightNumber: "TEST123",
		Departure: models.Location{
			Airport:  "CGK",
			DateTime: now,
		},
		Arrival: models.Location{
			Airport:  "DPS",
			DateTime: now.Add(2 * time.Hour),
		},
		Duration: models.Duration{
			TotalMinutes: 120,
		},
		Price: models.Price{
			Amount:   1000000,
			Currency: "IDR",
		},
		Stops:          0,
		AvailableSeats: 50,
	}

	t.Run("Valid Flight", func(t *testing.T) {
		err := validator.ValidateFlight(validFlight)
		assert.NoError(t, err, "Valid flight should pass validation")
	})

	t.Run("Nil Flight", func(t *testing.T) {
		err := validator.ValidateFlight(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("Missing ID", func(t *testing.T) {
		flight := *validFlight
		flight.ID = ""
		err := validator.ValidateFlight(&flight)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID is required")
	})

	t.Run("Invalid Airport Code", func(t *testing.T) {
		flight := *validFlight
		flight.Departure.Airport = "INVALID"
		err := validator.ValidateFlight(&flight)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid departure airport code")
	})

	t.Run("Arrival Before Departure", func(t *testing.T) {
		flight := *validFlight
		flight.Arrival.DateTime = flight.Departure.DateTime.Add(-1 * time.Hour)
		err := validator.ValidateFlight(&flight)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be after departure")
	})

	t.Run("Zero Duration", func(t *testing.T) {
		flight := *validFlight
		flight.Duration.TotalMinutes = 0
		err := validator.ValidateFlight(&flight)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duration must be positive")
	})

	t.Run("Unreasonable Duration", func(t *testing.T) {
		flight := *validFlight
		flight.Duration.TotalMinutes = 2000 // More than 24 hours
		err := validator.ValidateFlight(&flight)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unreasonable")
	})

	t.Run("Duration Mismatch", func(t *testing.T) {
		flight := *validFlight
		flight.Duration.TotalMinutes = 60 // Says 1 hour, but actual is 2 hours
		err := validator.ValidateFlight(&flight)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duration mismatch")
	})

	t.Run("Negative Price", func(t *testing.T) {
		flight := *validFlight
		flight.Price.Amount = -100
		err := validator.ValidateFlight(&flight)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "price must be positive")
	})

	t.Run("Negative Stops", func(t *testing.T) {
		flight := *validFlight
		flight.Stops = -1
		err := validator.ValidateFlight(&flight)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stops cannot be negative")
	})
}

func TestValidatorService_ValidateSearchRequest(t *testing.T) {
	validator := NewValidatorService()

	validRequest := &models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02"),
		Passengers:    1,
		CabinClass:    "economy",
	}

	t.Run("Valid Request", func(t *testing.T) {
		err := validator.ValidateSearchRequest(validRequest)
		assert.NoError(t, err, "Valid request should pass validation")
	})

	t.Run("Nil Request", func(t *testing.T) {
		err := validator.ValidateSearchRequest(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("Invalid Origin Code", func(t *testing.T) {
		req := *validRequest
		req.Origin = "INVALID"
		err := validator.ValidateSearchRequest(&req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid origin")
	})

	t.Run("Invalid Destination Code", func(t *testing.T) {
		req := *validRequest
		req.Destination = "XX"
		err := validator.ValidateSearchRequest(&req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid destination")
	})

	t.Run("Same Origin and Destination", func(t *testing.T) {
		req := *validRequest
		req.Destination = req.Origin
		err := validator.ValidateSearchRequest(&req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be the same")
	})

	t.Run("Invalid Date Format", func(t *testing.T) {
		req := *validRequest
		req.DepartureDate = "2025/12/15" // Wrong format
		err := validator.ValidateSearchRequest(&req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid departure date format")
	})

	t.Run("Past Departure Date", func(t *testing.T) {
		req := *validRequest
		req.DepartureDate = "2020-01-01" // Past date
		err := validator.ValidateSearchRequest(&req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be in the past")
	})

	t.Run("Invalid Passengers Count", func(t *testing.T) {
		req := *validRequest
		req.Passengers = 0
		err := validator.ValidateSearchRequest(&req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "passengers must be at least 1")

		req.Passengers = 10
		err = validator.ValidateSearchRequest(&req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot exceed 9")
	})

	t.Run("Invalid Cabin Class", func(t *testing.T) {
		req := *validRequest
		req.CabinClass = "premium"
		err := validator.ValidateSearchRequest(&req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cabin class")
	})

	t.Run("Valid Return Date", func(t *testing.T) {
		req := *validRequest
		returnDate := time.Now().Add(14 * 24 * time.Hour).Format("2006-01-02")
		req.ReturnDate = &returnDate
		err := validator.ValidateSearchRequest(&req)
		assert.NoError(t, err)
	})

	t.Run("Return Date Before Departure", func(t *testing.T) {
		req := *validRequest
		returnDate := time.Now().Add(1 * 24 * time.Hour).Format("2006-01-02")
		req.DepartureDate = time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02")
		req.ReturnDate = &returnDate
		err := validator.ValidateSearchRequest(&req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "return date must be after")
	})
}

func TestValidatorService_ValidateFlights(t *testing.T) {
	validator := NewValidatorService()

	now := time.Now()

	flights := []*models.Flight{
		// Valid flight
		{
			ID:           "VALID1",
			Provider:     "Test",
			FlightNumber: "T1",
			Departure:    models.Location{Airport: "CGK", DateTime: now},
			Arrival:      models.Location{Airport: "DPS", DateTime: now.Add(2 * time.Hour)},
			Duration:     models.Duration{TotalMinutes: 120},
			Price:        models.Price{Amount: 1000, Currency: "IDR"},
			Stops:        0,
			AvailableSeats: 50,
		},
		// Invalid flight (no ID)
		{
			Provider:     "Test",
			FlightNumber: "T2",
			Departure:    models.Location{Airport: "CGK", DateTime: now},
			Arrival:      models.Location{Airport: "DPS", DateTime: now.Add(2 * time.Hour)},
			Duration:     models.Duration{TotalMinutes: 120},
			Price:        models.Price{Amount: 1000, Currency: "IDR"},
		},
		// Valid flight
		{
			ID:           "VALID2",
			Provider:     "Test",
			FlightNumber: "T3",
			Departure:    models.Location{Airport: "CGK", DateTime: now},
			Arrival:      models.Location{Airport: "DPS", DateTime: now.Add(3 * time.Hour)},
			Duration:     models.Duration{TotalMinutes: 180},
			Price:        models.Price{Amount: 2000, Currency: "IDR"},
			Stops:        1,
			AvailableSeats: 30,
		},
	}

	validFlights := validator.ValidateFlights(flights)

	assert.Len(t, validFlights, 2, "Should return only 2 valid flights")
	assert.Equal(t, "VALID1", validFlights[0].ID)
	assert.Equal(t, "VALID2", validFlights[1].ID)
}
