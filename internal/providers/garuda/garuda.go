package garuda

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/reynaldipane/flight-search/internal/models"
	"github.com/reynaldipane/flight-search/internal/providers/base"
)

//go:embed garuda_indonesia_search_response.json
var mockData []byte

// Provider implements the Provider interface for Garuda Indonesia
type Provider struct {
	*base.Provider
}

// New creates a new Garuda Indonesia provider
func New() *Provider {
	return &Provider{
		Provider: base.NewProvider(
			"Garuda Indonesia",
			50*time.Millisecond,  // min delay
			100*time.Millisecond, // max delay
			0.0,                  // 0% failure rate - always reliable
		),
	}
}

// response represents the raw JSON response from Garuda API
type response struct {
	Status  string   `json:"status"`
	Flights []flight `json:"flights"`
}

// flight represents a single flight in Garuda's response format
type flight struct {
	FlightID    string    `json:"flight_id"`
	Airline     string    `json:"airline"`
	AirlineCode string    `json:"airline_code"`
	Departure   location  `json:"departure"`
	Arrival     location  `json:"arrival"`
	Duration    int       `json:"duration_minutes"`
	Stops       int       `json:"stops"`
	Aircraft    string    `json:"aircraft"`
	Price       price     `json:"price"`
	Seats       int       `json:"available_seats"`
	FareClass   string    `json:"fare_class"`
	Baggage     baggage   `json:"baggage"`
	Amenities   []string  `json:"amenities,omitempty"`
	Segments    []segment `json:"segments,omitempty"`
}

type location struct {
	Airport  string `json:"airport"`
	City     string `json:"city"`
	Time     string `json:"time"`
	Terminal string `json:"terminal,omitempty"`
}

type price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type baggage struct {
	CarryOn int `json:"carry_on"`
	Checked int `json:"checked"`
}

type segment struct {
	FlightNumber string   `json:"flight_number"`
	Departure    location `json:"departure"`
	Arrival      location `json:"arrival"`
	Duration     int      `json:"duration_minutes"`
	Layover      int      `json:"layover_minutes,omitempty"`
}

// FetchFlights fetches and normalizes flights from Garuda Indonesia
func (p *Provider) FetchFlights(ctx context.Context, req *models.SearchRequest) ([]*models.Flight, error) {
	// Simulate API delay and potential failure
	if err := p.SimulateDelay(ctx); err != nil {
		return nil, fmt.Errorf("garuda provider failed: %w", err)
	}

	// Parse the mock JSON data
	var resp response
	if err := json.Unmarshal(mockData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse garuda response: %w", err)
	}

	// Normalize flights to our common format
	var flights []*models.Flight
	for _, gf := range resp.Flights {
		// Filter by search criteria
		if gf.Departure.Airport != req.Origin ||
			(len(gf.Segments) > 0 && gf.Segments[len(gf.Segments)-1].Arrival.Airport != req.Destination) &&
				(len(gf.Segments) == 0 && gf.Arrival.Airport != req.Destination) {
			continue // Skip flights that don't match origin/destination
		}

		normalizedFlight, err := p.normalize(&gf)
		if err != nil {
			// Log error but continue with other flights
			continue
		}
		flights = append(flights, normalizedFlight)
	}

	return flights, nil
}

// normalize converts a Garuda flight to our common Flight model
func (p *Provider) normalize(gf *flight) (*models.Flight, error) {
	// Parse departure time
	departureTime, err := time.Parse(time.RFC3339, gf.Departure.Time)
	if err != nil {
		return nil, fmt.Errorf("invalid departure time: %w", err)
	}

	// Parse arrival time
	arrivalTime, err := time.Parse(time.RFC3339, gf.Arrival.Time)
	if err != nil {
		return nil, fmt.Errorf("invalid arrival time: %w", err)
	}

	// Create normalized flight
	normalizedFlight := &models.Flight{
		ID:       fmt.Sprintf("%s_%s", gf.FlightID, "Garuda"),
		Provider: p.GetName(),
		Airline: models.Airline{
			Name: gf.Airline,
			Code: gf.AirlineCode,
		},
		FlightNumber: gf.FlightID,
		Departure: models.Location{
			Airport:   gf.Departure.Airport,
			City:      gf.Departure.City,
			DateTime:  departureTime,
			Timestamp: departureTime.Unix(),
			Terminal:  gf.Departure.Terminal,
		},
		Arrival: models.Location{
			Airport:   gf.Arrival.Airport,
			City:      gf.Arrival.City,
			DateTime:  arrivalTime,
			Timestamp: arrivalTime.Unix(),
			Terminal:  gf.Arrival.Terminal,
		},
		Duration: models.Duration{
			TotalMinutes: gf.Duration,
			Formatted:    formatDuration(gf.Duration),
		},
		Stops:          gf.Stops,
		Price:          models.Price{Amount: gf.Price.Amount, Currency: gf.Price.Currency},
		AvailableSeats: gf.Seats,
		CabinClass:     gf.FareClass,
		Aircraft:       gf.Aircraft,
		Amenities:      gf.Amenities,
		Baggage: models.Baggage{
			CarryOn: fmt.Sprintf("%d piece(s)", gf.Baggage.CarryOn),
			Checked: fmt.Sprintf("%d piece(s)", gf.Baggage.Checked),
		},
	}

	// Handle segments if present (for connecting flights)
	if len(gf.Segments) > 0 {
		// Fix the stops count based on segments (data inconsistency handling!)
		normalizedFlight.Stops = len(gf.Segments) - 1

		for _, seg := range gf.Segments {
			depTime, _ := time.Parse(time.RFC3339, seg.Departure.Time)
			arrTime, _ := time.Parse(time.RFC3339, seg.Arrival.Time)

			normalizedSegment := models.Segment{
				FlightNumber: seg.FlightNumber,
				Departure: models.Location{
					Airport:   seg.Departure.Airport,
					DateTime:  depTime,
					Timestamp: depTime.Unix(),
				},
				Arrival: models.Location{
					Airport:   seg.Arrival.Airport,
					DateTime:  arrTime,
					Timestamp: arrTime.Unix(),
				},
				DurationMinutes: seg.Duration,
				LayoverMinutes:  seg.Layover,
			}
			normalizedFlight.Segments = append(normalizedFlight.Segments, normalizedSegment)
		}

		// Update arrival to final destination from last segment
		lastSegment := gf.Segments[len(gf.Segments)-1]
		finalArrivalTime, _ := time.Parse(time.RFC3339, lastSegment.Arrival.Time)
		normalizedFlight.Arrival.Airport = lastSegment.Arrival.Airport
		normalizedFlight.Arrival.DateTime = finalArrivalTime
		normalizedFlight.Arrival.Timestamp = finalArrivalTime.Unix()
	}

	return normalizedFlight, nil
}

// formatDuration converts minutes to human-readable format (e.g., "2h 30m")
func formatDuration(minutes int) string {
	hours := minutes / 60
	mins := minutes % 60

	if hours > 0 && mins > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	} else if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", mins)
}
