package batikair

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/reynaldipane/flight-search/internal/models"
	"github.com/reynaldipane/flight-search/internal/providers/base"
)

//go:embed batik_air_search_response.json
var mockData []byte

// Provider implements the Provider interface for Batik Air
type Provider struct {
	*base.Provider
}

// New creates a new Batik Air provider
func New() *Provider {
	return &Provider{
		Provider: base.NewProvider(
			"Batik Air",
			200*time.Millisecond, // min delay
			400*time.Millisecond, // max delay
			0.0,                  // 0% failure rate
		),
	}
}

// response represents the raw JSON response from Batik Air API
type response struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Results []flight `json:"results"`
}

// flight represents a single flight in Batik Air's response format
type flight struct {
	FlightNumber      string       `json:"flightNumber"`
	AirlineName       string       `json:"airlineName"`
	AirlineIATA       string       `json:"airlineIATA"`
	Origin            string       `json:"origin"`
	Destination       string       `json:"destination"`
	DepartureDateTime string       `json:"departureDateTime"`
	ArrivalDateTime   string       `json:"arrivalDateTime"`
	TravelTime        string       `json:"travelTime"` // e.g., "1h 45m"
	NumberOfStops     int          `json:"numberOfStops"`
	Connections       []connection `json:"connections,omitempty"`
	Fare              fare         `json:"fare"`
	SeatsAvailable    int          `json:"seatsAvailable"`
	AircraftModel     string       `json:"aircraftModel"`
	BaggageInfo       string       `json:"baggageInfo"`
	OnboardServices   []string     `json:"onboardServices"`
}

type connection struct {
	StopAirport  string `json:"stopAirport"`
	StopDuration string `json:"stopDuration"` // e.g., "55m"
}

type fare struct {
	BasePrice  float64 `json:"basePrice"`
	Taxes      float64 `json:"taxes"`
	TotalPrice float64 `json:"totalPrice"`
	Currency   string  `json:"currencyCode"`
	Class      string  `json:"class"`
}

// FetchFlights fetches and normalizes flights from Batik Air
func (p *Provider) FetchFlights(ctx context.Context, req *models.SearchRequest) ([]*models.Flight, error) {
	// Simulate API delay and potential failure
	if err := p.SimulateDelay(ctx); err != nil {
		return nil, fmt.Errorf("batik air provider failed: %w", err)
	}

	// Parse the mock JSON data
	var resp response
	if err := json.Unmarshal(mockData, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse batik air response: %w", err)
	}

	// Normalize flights to our common format
	var flights []*models.Flight
	for _, bf := range resp.Results {
		// Filter by search criteria
		if bf.Origin != req.Origin || bf.Destination != req.Destination {
			continue
		}

		normalizedFlight, err := p.normalize(&bf)
		if err != nil {
			// Log error but continue with other flights
			continue
		}
		flights = append(flights, normalizedFlight)
	}

	return flights, nil
}

// normalize converts a Batik Air flight to our common Flight model
func (p *Provider) normalize(bf *flight) (*models.Flight, error) {
	// Parse departure time (format: "2025-12-15T07:15:00+0700")
	departureTime, err := time.Parse("2006-01-02T15:04:05-0700", bf.DepartureDateTime)
	if err != nil {
		return nil, fmt.Errorf("invalid departure time: %w", err)
	}

	// Parse arrival time
	arrivalTime, err := time.Parse("2006-01-02T15:04:05-0700", bf.ArrivalDateTime)
	if err != nil {
		return nil, fmt.Errorf("invalid arrival time: %w", err)
	}

	// Parse duration string (e.g., "1h 45m")
	durationMinutes, err := parseDuration(bf.TravelTime)
	if err != nil {
		return nil, fmt.Errorf("invalid travel time: %w", err)
	}

	// Create normalized flight
	normalizedFlight := &models.Flight{
		ID:       fmt.Sprintf("%s_%s", bf.FlightNumber, "BatikAir"),
		Provider: p.GetName(),
		Airline: models.Airline{
			Name: bf.AirlineName,
			Code: bf.AirlineIATA,
		},
		FlightNumber: bf.FlightNumber,
		Departure: models.Location{
			Airport:   bf.Origin,
			DateTime:  departureTime,
			Timestamp: departureTime.Unix(),
		},
		Arrival: models.Location{
			Airport:   bf.Destination,
			DateTime:  arrivalTime,
			Timestamp: arrivalTime.Unix(),
		},
		Duration: models.Duration{
			TotalMinutes: durationMinutes,
			Formatted:    bf.TravelTime,
		},
		Stops:          bf.NumberOfStops,
		Price:          models.Price{Amount: bf.Fare.TotalPrice, Currency: bf.Fare.Currency},
		AvailableSeats: bf.SeatsAvailable,
		CabinClass:     bf.Fare.Class,
		Aircraft:       bf.AircraftModel,
		Amenities:      bf.OnboardServices,
		Baggage: models.Baggage{
			CarryOn: bf.BaggageInfo, // e.g., "7kg cabin, 20kg checked"
			Checked: "",
		},
	}

	return normalizedFlight, nil
}

// parseDuration parses duration strings like "1h 45m", "2h", "30m" into total minutes
func parseDuration(durationStr string) (int, error) {
	// Regular expression to match hours and/or minutes
	// Matches: "1h 45m", "1h", "45m", "2h 30m"
	re := regexp.MustCompile(`(?:(\d+)h)?(?:\s*(\d+)m)?`)
	matches := re.FindStringSubmatch(durationStr)

	if len(matches) < 3 {
		return 0, fmt.Errorf("invalid duration format: %s", durationStr)
	}

	totalMinutes := 0

	// Parse hours (group 1)
	if matches[1] != "" {
		hours, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0, fmt.Errorf("invalid hours: %w", err)
		}
		totalMinutes += hours * 60
	}

	// Parse minutes (group 2)
	if matches[2] != "" {
		minutes, err := strconv.Atoi(matches[2])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %w", err)
		}
		totalMinutes += minutes
	}

	if totalMinutes == 0 {
		return 0, fmt.Errorf("no valid duration found in: %s", durationStr)
	}

	return totalMinutes, nil
}
