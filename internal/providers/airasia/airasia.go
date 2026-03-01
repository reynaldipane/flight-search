package airasia

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/reynaldipane/flight-search/internal/models"
	"github.com/reynaldipane/flight-search/internal/providers/base"
)

//go:embed airasia_search_response.json
var mockData []byte

// Provider implements the Provider interface for AirAsia
type Provider struct {
	*base.Provider
}

// New creates a new AirAsia provider
func New() *Provider {
	return &Provider{
		Provider: base.NewProvider(
			"AirAsia",
			50*time.Millisecond,  // min delay
			150*time.Millisecond, // max delay
			0.1,                  // 10% failure rate - occasionally fails!
		),
	}
}

// response represents the raw JSON response from AirAsia API
type response struct {
	Status  string   `json:"status"`
	Flights []flight `json:"flights"`
}

// flight represents a single flight in AirAsia's response format
type flight struct {
	FlightCode    string  `json:"flight_code"`
	Airline       string  `json:"airline"`
	FromAirport   string  `json:"from_airport"`
	ToAirport     string  `json:"to_airport"`
	DepartTime    string  `json:"depart_time"`
	ArriveTime    string  `json:"arrive_time"`
	DurationHours float64 `json:"duration_hours"` // Decimal hours (e.g., 1.67)
	DirectFlight  bool    `json:"direct_flight"`
	Stops         []stop  `json:"stops,omitempty"`
	PriceIDR      float64 `json:"price_idr"`
	Seats         int     `json:"seats"`
	CabinClass    string  `json:"cabin_class"`
	BaggageNote   string  `json:"baggage_note"`
}

type stop struct {
	Airport         string `json:"airport"`
	WaitTimeMinutes int    `json:"wait_time_minutes"`
}

// FetchFlights fetches and normalizes flights from AirAsia
func (p *Provider) FetchFlights(ctx context.Context, req *models.SearchRequest) ([]*models.Flight, error) {
	// Simulate API delay and potential failure (10% chance)
	if err := p.SimulateDelay(ctx); err != nil {
		return nil, fmt.Errorf("airasia provider failed: %w", err)
	}

	// Parse the mock JSON data
	var resp response
	if err := json.Unmarshal(mockData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse airasia response: %w", err)
	}

	// Normalize flights to our common format
	var flights []*models.Flight
	for _, af := range resp.Flights {
		// Filter by search criteria
		if af.FromAirport != req.Origin || af.ToAirport != req.Destination {
			continue
		}

		normalizedFlight, err := p.normalize(&af)
		if err != nil {
			// Log error but continue with other flights
			continue
		}
		flights = append(flights, normalizedFlight)
	}

	return flights, nil
}

// normalize converts an AirAsia flight to our common Flight model
func (p *Provider) normalize(af *flight) (*models.Flight, error) {
	// Parse departure time
	departureTime, err := time.Parse(time.RFC3339, af.DepartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid departure time: %w", err)
	}

	// Parse arrival time
	arrivalTime, err := time.Parse(time.RFC3339, af.ArriveTime)
	if err != nil {
		return nil, fmt.Errorf("invalid arrival time: %w", err)
	}

	// Convert duration from decimal hours to minutes
	// Example: 1.67 hours = 100.2 minutes → rounds to 100 minutes
	durationMinutes := int(math.Round(af.DurationHours * 60))

	// Determine number of stops
	stops := 0
	if !af.DirectFlight {
		stops = len(af.Stops)
	}

	// Create normalized flight
	normalizedFlight := &models.Flight{
		ID:       fmt.Sprintf("%s_%s", af.FlightCode, "AirAsia"),
		Provider: p.GetName(),
		Airline: models.Airline{
			Name: af.Airline,
			Code: "QZ", // AirAsia Indonesia code
		},
		FlightNumber: af.FlightCode,
		Departure: models.Location{
			Airport:   af.FromAirport,
			DateTime:  departureTime,
			Timestamp: departureTime.Unix(),
		},
		Arrival: models.Location{
			Airport:   af.ToAirport,
			DateTime:  arrivalTime,
			Timestamp: arrivalTime.Unix(),
		},
		Duration: models.Duration{
			TotalMinutes: durationMinutes,
			Formatted:    formatDuration(durationMinutes),
		},
		Stops:          stops,
		Price:          models.Price{Amount: af.PriceIDR, Currency: "IDR"},
		AvailableSeats: af.Seats,
		CabinClass:     af.CabinClass,
		Aircraft:       "", // Not provided by AirAsia
		Amenities:      []string{}, // Not provided
		Baggage: models.Baggage{
			CarryOn: af.BaggageNote,
			Checked: "",
		},
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
