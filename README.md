# Flight Search & Aggregation System

A flight search aggregation service built in Go that combines results from multiple Indonesian airline providers into a unified API.

## Overview

This system aggregates flight data from 4 major Indonesian airlines (Garuda Indonesia, Lion Air, Batik Air, and AirAsia), normalizes different data formats, and provides advanced filtering, sorting, and ranking capabilities through a RESTful API.

## Features

- **Multi-Provider Aggregation**: Parallel fetching from 4 airline APIs using goroutines
- **Data Normalization**: Handles different JSON formats, timezones, and duration representations
- **Advanced Filtering**: Filter by price, stops, airlines, duration, and departure/arrival times
- **Smart Ranking**: Best value algorithm combining price (50%) and convenience (50%)
- **Caching**: In-memory cache with TTL and performance metrics
- **Robust Error Handling**: Graceful degradation when providers fail
- **Real-time Performance**: Sub-second response times with parallel processing
- **Comprehensive Testing**: 90.8% code coverage

## Architecture

```
┌─────────────────┐
│   HTTP API      │  ← Gin Web Framework
├─────────────────┤
│   Handlers      │  ← Request/Response handling
├─────────────────┤
│   Services      │  ← Business Logic
│  - Aggregator  │     • Parallel provider fetching
│  - Validator   │     • Request/flight validation
│  - Filter      │     • Advanced filtering & sorting
│  - Cache       │     • In-memory caching
│  - Ranking     │     • Best value scoring
├─────────────────┤
│   Providers     │  ← Airline Integrations
│  - Garuda ID   │     • Different JSON formats
│  - Lion Air    │     • Timezone handling
│  - Batik Air   │     • Duration parsing
│  - AirAsia     │     • Failure simulation
└─────────────────┘
```

## Tech Stack

- **Language**: Go 1.25+
- **Web Framework**: Gin
- **Caching**: go-cache
- **Testing**: testify
- **HTTP**: Standard library with context support

## Prerequisites

- Go 1.25 or higher
- Make (optional, for using Makefile commands)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/reynaldipane/flight-search.git
cd flight-search
```

2. Install dependencies:
```bash
make deps
```

Or manually:
```bash
go mod download
go mod tidy
```

## Running the Application

### Using Make

```bash
# Run the server
make run

# Build the binary
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Clean build artifacts
make clean
```

### Manual Commands

```bash
# Run the server directly
go run cmd/server/main.go

# Build the binary
go build -o bin/server cmd/server/main.go

# Run the binary
./bin/server

# Run tests
go test -v ./...
```

The server will start on `http://localhost:8080` by default. You can change the port using the `PORT` environment variable:

```bash
PORT=3000 go run cmd/server/main.go
```

## API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Endpoints

#### 1. Health Check
```bash
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "service": "flight-search-api",
  "version": "1.0.0"
}
```

#### 2. List Providers
```bash
GET /providers
```

**Response:**
```json
{
  "providers": [
    {
      "name": "Garuda Indonesia",
      "delay": "88ms",
      "failure_rate": 0
    },
    {
      "name": "AirAsia",
      "delay": "148ms",
      "failure_rate": 0.1
    }
  ],
  "total": 4
}
```

#### 3. Basic Flight Search
```bash
POST /search
Content-Type: application/json

{
  "origin": "CGK",
  "destination": "DPS",
  "departureDate": "2026-12-15",
  "passengers": 1,
  "cabinClass": "economy"
}
```

**Response:**
```json
{
  "search_criteria": {
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2026-12-15",
    "passengers": 1,
    "cabinClass": "economy"
  },
  "metadata": {
    "total_results": 13,
    "providers_queried": 4,
    "providers_succeeded": 4,
    "providers_failed": 0,
    "search_time_ms": 213,
    "cache_hit": false
  },
  "flights": [
    {
      "id": "GA400-CGK-DPS-2026-12-15T08:00",
      "provider": "Garuda Indonesia",
      "airline": {
        "name": "Garuda Indonesia",
        "code": "GA"
      },
      "flight_number": "GA400",
      "departure": {
        "airport": "CGK",
        "city": "Jakarta",
        "datetime": "2026-12-15T08:00:00+07:00"
      },
      "arrival": {
        "airport": "DPS",
        "city": "Denpasar",
        "datetime": "2026-12-15T09:50:00+08:00"
      },
      "duration": {
        "hours": 1,
        "minutes": 50,
        "total_minutes": 110
      },
      "stops": 0,
      "price": {
        "amount": 1250000,
        "currency": "IDR"
      },
      "available_seats": 45,
      "amenities": ["WiFi", "Meal", "Entertainment"],
      "baggage": {
        "cabin": "7kg",
        "checked": "20kg"
      }
    }
  ]
}
```

#### 4. Advanced Filtered Search
```bash
POST /search/filter
Content-Type: application/json

{
  "origin": "CGK",
  "destination": "DPS",
  "departureDate": "2026-12-15",
  "passengers": 1,
  "cabinClass": "economy",
  "filters": {
    "maxPrice": 1000000,
    "maxStops": 0,
    "maxDuration": 150
  },
  "sortBy": "best_value",
  "limit": 5
}
```

**Filter Options:**
- `minPrice` (float): Minimum price in IDR
- `maxPrice` (float): Maximum price in IDR
- `maxStops` (int): Maximum number of stops (0 = direct only)
- `airlines` ([]string): Filter by airline names or codes
- `maxDuration` (int): Maximum duration in minutes
- `departureTime` (string): "morning", "afternoon", "evening", "night"
- `arrivalTime` (string): "morning", "afternoon", "evening", "night"

**Sort Options:**
- `price_asc`: Price low to high
- `price_desc`: Price high to low
- `duration_asc`: Shortest first
- `duration_desc`: Longest first
- `departure_asc`: Earliest first
- `departure_desc`: Latest first
- `arrival_asc`: Earliest arrival
- `arrival_desc`: Latest arrival
- `best_value`: Best value score (default)

#### 5. Cache Statistics
```bash
GET /cache/stats
```

**Response:**
```json
{
  "hit_rate": 50.0,
  "item_count": 1
}
```

#### 6. Clear Cache
```bash
DELETE /cache
```

**Response:**
```json
{
  "message": "Cache cleared successfully"
}
```

## Example Usage

### Using cURL

1. **Basic Search:**
```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d @examples/basic-search.json
```

2. **Filtered Search:**
```bash
curl -X POST http://localhost:8080/api/v1/search/filter \
  -H "Content-Type: application/json" \
  -d @examples/filtered-search.json
```

3. **Direct Flights Only:**
```bash
curl -X POST http://localhost:8080/api/v1/search/filter \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2026-12-15",
    "passengers": 1,
    "cabinClass": "economy",
    "filters": {
      "maxStops": 0
    },
    "sortBy": "price_asc"
  }'
```

See `examples/API_EXAMPLES.md` for more detailed examples.

## Project Structure

```
flight-search/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── handlers/
│   │   ├── router.go              # Route setup and configuration
│   │   ├── search.go              # Search endpoint handlers
│   │   └── types.go               # Request/response types
│   ├── models/
│   │   ├── filter.go              # Filter and sort models
│   │   ├── flight.go              # Flight data models
│   │   └── search.go              # Search request/response models
│   ├── providers/
│   │   ├── provider.go            # Provider interface
│   │   ├── base/                  # Base provider implementation
│   │   ├── garuda/                # Garuda Indonesia provider
│   │   ├── lionair/               # Lion Air provider
│   │   ├── batikair/              # Batik Air provider
│   │   └── airasia/               # AirAsia provider
│   └── services/
│       ├── aggregator.go          # Provider aggregation service
│       ├── cache.go               # Caching service
│       ├── filter.go              # Filtering and sorting service
│       ├── ranking.go             # Flight ranking service
│       └── validator.go           # Validation service
├── pkg/
│   └── errors/
│       └── errors.go              # Custom error types
├── examples/
│   ├── basic-search.json          # Basic search example
│   ├── filtered-search.json       # Filtered search example
│   └── API_EXAMPLES.md            # Detailed API examples
├── Makefile                        # Build automation
├── go.mod                          # Go module definition
├── go.sum                          # Go module checksums
└── README.md                       # This file
```

## Testing

### Run All Tests
```bash
make test
```

### Run Tests with Coverage
```bash
make test-coverage
```

### Run Specific Test
```bash
go test -v ./internal/services -run TestRankingService
```

### Test Results
- **Total Coverage**: 90.8%
- **All Services**: Fully tested with table-driven tests
- **Providers**: 100% coverage
- **Error Scenarios**: Comprehensive error handling tests

## Best Value Algorithm

The ranking service uses a sophisticated scoring algorithm:

**Final Score = (Price Score × 50%) + (Convenience Score × 50%)**

### Price Score (0-100)
- Cheapest flight gets 100 points
- Most expensive gets 0 points
- Linear interpolation for others

### Convenience Score (0-100)
- **Direct flights**: +30 points
- **One stop**: +15 points
- **Duration**: +30 points (normalized, shorter is better)
- **Time of day**: +20 points
  - Morning (6am-12pm): 100%
  - Afternoon (12pm-6pm): 90%
  - Evening (6pm-12am): 70%
  - Red-eye (12am-6am): 0%
- **Available seats**: +10 points (normalized)
- **Amenities**: +10 points (normalized)

## Performance

- **Parallel Processing**: All providers queried simultaneously using goroutines
- **Response Time**: ~200ms for 4 providers (fastest provider determines speed)
- **Caching**: 5-minute TTL, reduces response time to <5ms on cache hits
- **Failure Handling**: Graceful degradation (continues with successful providers)
- **Timeout**: 10-second context timeout per request

## Error Handling

The API returns appropriate HTTP status codes and descriptive error messages:

- **400 Bad Request**: Invalid input (missing fields, invalid dates, etc.)
- **500 Internal Server Error**: Server-side errors

**Example Error Response:**
```json
{
  "error": "INVALID_REQUEST",
  "message": "departure date cannot be in the past"
}
```

## Provider Details

### Garuda Indonesia
- **Delay**: 50-100ms
- **Failure Rate**: 0%
- **Format**: ISO 8601 timestamps, multi-segment flights

### Lion Air
- **Delay**: 100-200ms
- **Failure Rate**: 0%
- **Format**: Timezone names (Asia/Jakarta), nested structure

### Batik Air
- **Delay**: 200-400ms
- **Failure Rate**: 0%
- **Format**: Human-readable durations ("1h 45m")

### AirAsia
- **Delay**: 50-150ms
- **Failure Rate**: 10%
- **Format**: Decimal hours for duration

## Configuration

### Environment Variables

- `PORT`: Server port (default: 8080)
- `GIN_MODE`: Gin mode - "debug" or "release" (default: debug)

### Cache Configuration

Default cache settings (configurable in `internal/handlers/router.go`):
- **Default TTL**: 5 minutes
- **Cleanup Interval**: 10 minutes

## Development

### Adding a New Provider

1. Create a new provider package in `internal/providers/`
2. Implement the `Provider` interface
3. Add provider to the registry in `internal/providers/provider.go`

Example:
```go
// internal/providers/newairline/newairline.go
package newairline

import (
    "github.com/reynaldipane/flight-search/internal/providers/base"
)

func New() *Provider {
    return &Provider{
        Provider: base.New("New Airline", 100*time.Millisecond, 0.0),
    }
}

func (p *Provider) FetchFlights(ctx context.Context, req *models.SearchRequest) ([]*models.Flight, error) {
    // Implementation
}
```

### Adding New Filters

Add new filter fields to `models.FilterOptions` in `internal/models/filter.go` and implement the filtering logic in `internal/services/filter.go`.

## Troubleshooting

### Port Already in Use
```bash
# Kill process on port 8080
lsof -ti:8080 | xargs kill -9

# Or use a different port
PORT=3000 make run
```

### Tests Failing
```bash
# Clean and rebuild
make clean
make deps
make test
```

### Cache Not Working
```bash
# Clear cache via API
curl -X DELETE http://localhost:8080/api/v1/cache

# Check cache stats
curl http://localhost:8080/api/v1/cache/stats
```
