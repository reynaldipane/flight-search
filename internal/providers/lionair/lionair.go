package lionair

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/reynaldipane/flight-search/internal/models"
	"github.com/reynaldipane/flight-search/internal/providers/base"
)

//go:embed lion_air_search_response.json
var mockData []byte

// Provider implements the Provider interface for Lion Air
type Provider struct {
	*base.Provider
}

// New creates a new Lion Air provider
func New() *Provider {
	return &Provider{
		Provider: base.NewProvider(
			"Lion Air",
			100*time.Millisecond, // min delay
			200*time.Millisecond, // max delay
			0.0,                  // 0% failure rate
		),
	}
}

// response represents the raw JSON response from Lion Air API
type response struct {
	Success bool `json:"success"`
	Data    struct {
		AvailableFlights []flight `json:"available_flights"`
	} `json:"data"`
}

// flight represents a single flight in Lion Air's response format
type flight struct {
	ID         string    `json:"id"`
	Carrier    carrier   `json:"carrier"`
	Route      route     `json:"route"`
	Schedule   schedule  `json:"schedule"`
	FlightTime int       `json:"flight_time"` // in minutes
	IsDirect   bool      `json:"is_direct"`
	StopCount  int       `json:"stop_count,omitempty"`
	Layovers   []layover `json:"layovers,omitempty"`
	Pricing    pricing   `json:"pricing"`
	SeatsLeft  int       `json:"seats_left"`
	PlaneType  string    `json:"plane_type"`
	Services   services  `json:"services"`
}

type carrier struct {
	Name string `json:"name"`
	IATA string `json:"iata"`
}

type route struct {
	From airport `json:"from"`
	To   airport `json:"to"`
}

type airport struct {
	Code string `json:"code"`
	Name string `json:"name"`
	City string `json:"city"`
}

type schedule struct {
	Departure         string `json:"departure"`
	DepartureTimezone string `json:"departure_timezone"`
	Arrival           string `json:"arrival"`
	ArrivalTimezone   string `json:"arrival_timezone"`
}

type layover struct {
	Airport         string `json:"airport"`
	DurationMinutes int    `json:"duration_minutes"`
}

type pricing struct {
	Total    float64 `json:"total"`
	Currency string  `json:"currency"`
	FareType string  `json:"fare_type"`
}

type services struct {
	WifiAvailable    bool             `json:"wifi_available"`
	MealsIncluded    bool             `json:"meals_included"`
	BaggageAllowance baggageAllowance `json:"baggage_allowance"`
}

type baggageAllowance struct {
	Cabin string `json:"cabin"`
	Hold  string `json:"hold"`
}

// FetchFlights fetches and normalizes flights from Lion Air
func (p *Provider) FetchFlights(ctx context.Context, req *models.SearchRequest) ([]*models.Flight, error) {
	// Simulate API delay and potential failure
	if err := p.SimulateDelay(ctx); err != nil {
		return nil, fmt.Errorf("lion air provider failed: %w", err)
	}

	// Parse the mock JSON data
	var resp response
	if err := json.Unmarshal(mockData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse lion air response: %w", err)
	}

	// Normalize flights to our common format
	var flights []*models.Flight
	for _, lf := range resp.Data.AvailableFlights {
		// Filter by search criteria
		if lf.Route.From.Code != req.Origin || lf.Route.To.Code != req.Destination {
			continue
		}

		normalizedFlight, err := p.normalize(&lf)
		if err != nil {
			// Log error but continue with other flights
			continue
		}
		flights = append(flights, normalizedFlight)
	}

	return flights, nil
}

// normalize converts a Lion Air flight to our common Flight model
func (p *Provider) normalize(lf *flight) (*models.Flight, error) {
	// Parse departure time with timezone
	depLoc, err := time.LoadLocation(lf.Schedule.DepartureTimezone)
	if err != nil {
		// Fallback to parsing as-is if timezone loading fails
		depLoc = time.UTC
	}

	// Lion Air uses format: "2025-12-15T05:30:00" (without timezone offset)
	departureTime, err := time.ParseInLocation("2006-01-02T15:04:05", lf.Schedule.Departure, depLoc)
	if err != nil {
		return nil, fmt.Errorf("invalid departure time: %w", err)
	}

	// Parse arrival time with timezone
	arrLoc, err := time.LoadLocation(lf.Schedule.ArrivalTimezone)
	if err != nil {
		arrLoc = time.UTC
	}

	arrivalTime, err := time.ParseInLocation("2006-01-02T15:04:05", lf.Schedule.Arrival, arrLoc)
	if err != nil {
		return nil, fmt.Errorf("invalid arrival time: %w", err)
	}

	// Determine number of stops
	stops := 0
	if !lf.IsDirect {
		stops = lf.StopCount
	}

	// Build amenities list
	var amenities []string
	if lf.Services.WifiAvailable {
		amenities = append(amenities, "wifi")
	}
	if lf.Services.MealsIncluded {
		amenities = append(amenities, "meal")
	}

	// Create normalized flight
	normalizedFlight := &models.Flight{
		ID:       fmt.Sprintf("%s_%s", lf.ID, "LionAir"),
		Provider: p.GetName(),
		Airline: models.Airline{
			Name: lf.Carrier.Name,
			Code: lf.Carrier.IATA,
		},
		FlightNumber: lf.ID,
		Departure: models.Location{
			Airport:   lf.Route.From.Code,
			City:      lf.Route.From.City,
			DateTime:  departureTime,
			Timestamp: departureTime.Unix(),
		},
		Arrival: models.Location{
			Airport:   lf.Route.To.Code,
			City:      lf.Route.To.City,
			DateTime:  arrivalTime,
			Timestamp: arrivalTime.Unix(),
		},
		Duration: models.Duration{
			TotalMinutes: lf.FlightTime,
			Formatted:    formatDuration(lf.FlightTime),
		},
		Stops:          stops,
		Price:          models.Price{Amount: lf.Pricing.Total, Currency: lf.Pricing.Currency},
		AvailableSeats: lf.SeatsLeft,
		CabinClass:     lf.Pricing.FareType,
		Aircraft:       lf.PlaneType,
		Amenities:      amenities,
		Baggage: models.Baggage{
			CarryOn: lf.Services.BaggageAllowance.Cabin,
			Checked: lf.Services.BaggageAllowance.Hold,
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
