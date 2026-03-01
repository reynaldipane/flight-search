package models

import "time"

// Flight represents a normalized flight from any provider
type Flight struct {
	ID             string    `json:"id"`              // Format: "{flightNumber}_{provider}"
	Provider       string    `json:"provider"`        // e.g., "Garuda Indonesia"
	Airline        Airline   `json:"airline"`
	FlightNumber   string    `json:"flight_number"`   // e.g., "GA400"
	Departure      Location  `json:"departure"`
	Arrival        Location  `json:"arrival"`
	Duration       Duration  `json:"duration"`
	Stops          int       `json:"stops"`           // 0 = direct, 1+ = with stops
	Price          Price     `json:"price"`
	AvailableSeats int       `json:"available_seats"`
	CabinClass     string    `json:"cabin_class"`     // e.g., "economy"
	Aircraft       string    `json:"aircraft"`        // e.g., "Boeing 737-800", can be empty
	Amenities      []string  `json:"amenities"`       // e.g., ["wifi", "meal"]
	Baggage        Baggage   `json:"baggage"`
	Segments       []Segment `json:"segments,omitempty"` // For multi-leg flights
}

// Airline represents airline information
type Airline struct {
	Name string `json:"name"` // e.g., "Garuda Indonesia"
	Code string `json:"code"` // e.g., "GA"
}

// Location represents airport location and time
type Location struct {
	Airport   string    `json:"airport"`             // e.g., "CGK"
	City      string    `json:"city,omitempty"`      // e.g., "Jakarta"
	DateTime  time.Time `json:"datetime"`            // ISO 8601 with timezone
	Timestamp int64     `json:"timestamp"`           // Unix timestamp
	Terminal  string    `json:"terminal,omitempty"`  // e.g., "3"
}

// Duration represents flight duration
type Duration struct {
	TotalMinutes int    `json:"total_minutes"` // Total duration in minutes
	Formatted    string `json:"formatted"`     // Human-readable: "2h 30m"
}

// Price represents flight price
type Price struct {
	Amount   float64 `json:"amount"`   // e.g., 1250000
	Currency string  `json:"currency"` // e.g., "IDR"
}

// Baggage represents baggage allowance
type Baggage struct {
	CarryOn string `json:"carry_on"` // e.g., "7 kg" or "1 piece"
	Checked string `json:"checked"`  // e.g., "20 kg" or "2 pieces"
}

// Segment represents a single leg of a multi-leg flight (connecting flights)
// Example: CGK → SUB → DPS has 2 segments and 1 stop
// Note: Number of stops = len(Segments) - 1
// We use segments to represent flights with layovers/transfers
type Segment struct {
	FlightNumber    string   `json:"flight_number"`
	Departure       Location `json:"departure"`
	Arrival         Location `json:"arrival"`
	DurationMinutes int      `json:"duration_minutes"`
	LayoverMinutes  int      `json:"layover_minutes,omitempty"` // Time until next segment
}
