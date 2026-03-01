# Flight Search API - Example Requests

This document contains example curl commands for testing the Flight Search API.

## Prerequisites

Make sure the server is running:
```bash
make run
```

The API will be available at `http://localhost:8080`

## 1. Health Check

Check if the API is running:

```bash
curl http://localhost:8080/api/v1/health
```

**Expected Response:**
```json
{
  "service": "flight-search-api",
  "status": "ok",
  "version": "1.0.0"
}
```

## 2. List Providers

Get information about all available flight providers:

```bash
curl http://localhost:8080/api/v1/providers
```

**Expected Response:**
```json
{
  "providers": [
    {
      "name": "Garuda Indonesia",
      "delay": "50ms-100ms",
      "failure_rate": 0
    },
    {
      "name": "Lion Air",
      "delay": "100ms-200ms",
      "failure_rate": 0
    },
    {
      "name": "Batik Air",
      "delay": "200ms-400ms",
      "failure_rate": 0
    },
    {
      "name": "AirAsia",
      "delay": "50ms-150ms",
      "failure_rate": 0.1
    }
  ],
  "total": 4
}
```

## 3. Basic Flight Search

Search for flights without filters:

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d @examples/basic-search.json
```

Or inline:

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "passengers": 1,
    "cabin_class": "economy"
  }'
```

**Expected Response:**
```json
{
  "search_criteria": {
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "passengers": 1,
    "cabin_class": "economy"
  },
  "metadata": {
    "total_results": 13,
    "providers_queried": 4,
    "providers_succeeded": 4,
    "providers_failed": 0,
    "search_time_ms": 350,
    "cache_hit": false
  },
  "flights": [
    {
      "id": "GA400-CGK-DPS-2025-12-15T08:00",
      "provider": "Garuda Indonesia",
      "airline": {
        "name": "Garuda Indonesia",
        "code": "GA"
      },
      "flight_number": "GA400",
      "departure": {
        "airport": "CGK",
        "city": "Jakarta",
        "datetime": "2025-12-15T08:00:00+07:00"
      },
      "arrival": {
        "airport": "DPS",
        "city": "Denpasar",
        "datetime": "2025-12-15T09:50:00+08:00"
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
    // ... more flights
  ]
}
```

## 4. Filtered Flight Search

Search with filters and sorting:

```bash
curl -X POST http://localhost:8080/api/v1/search/filter \
  -H "Content-Type: application/json" \
  -d @examples/filtered-search.json
```

Or inline:

```bash
curl -X POST http://localhost:8080/api/v1/search/filter \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "passengers": 1,
    "cabin_class": "economy",
    "filters": {
      "max_price": 1000000,
      "max_stops": 0,
      "max_duration": 150
    },
    "sort_by": "best_value",
    "limit": 5
  }'
```

### Filter Options

- `min_price` (float): Minimum price in IDR
- `max_price` (float): Maximum price in IDR
- `max_stops` (int): Maximum number of stops (0 = direct only)
- `airlines` ([]string): Filter by airline names or codes (e.g., ["Garuda Indonesia", "QZ"])
- `max_duration` (int): Maximum duration in minutes
- `departure_time` (string): Time window - "morning", "afternoon", "evening", "night"
- `arrival_time` (string): Time window - "morning", "afternoon", "evening", "night"

### Sort Options

- `price_asc`: Sort by price (low to high)
- `price_desc`: Sort by price (high to low)
- `duration_asc`: Sort by duration (shortest first)
- `duration_desc`: Sort by duration (longest first)
- `departure_asc`: Sort by departure time (earliest first)
- `departure_desc`: Sort by departure time (latest first)
- `arrival_asc`: Sort by arrival time (earliest first)
- `arrival_desc`: Sort by arrival time (latest first)
- `best_value`: Sort by best value score (default)

## 5. Advanced Filtering Examples

### Direct Flights Only, Sorted by Price

```bash
curl -X POST http://localhost:8080/api/v1/search/filter \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "passengers": 1,
    "cabin_class": "economy",
    "filters": {
      "max_stops": 0
    },
    "sort_by": "price_asc"
  }'
```

### Specific Airlines, Morning Flights

```bash
curl -X POST http://localhost:8080/api/v1/search/filter \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "passengers": 1,
    "cabin_class": "economy",
    "filters": {
      "airlines": ["Garuda Indonesia", "Batik Air"],
      "departure_time": "morning"
    },
    "sort_by": "best_value"
  }'
```

### Budget Flights (Under 800k, Direct, Short Duration)

```bash
curl -X POST http://localhost:8080/api/v1/search/filter \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "passengers": 1,
    "cabin_class": "economy",
    "filters": {
      "max_price": 800000,
      "max_stops": 0,
      "max_duration": 120
    },
    "sort_by": "best_value",
    "limit": 3
  }'
```

## 6. Cache Statistics

Get cache performance metrics:

```bash
curl http://localhost:8080/api/v1/cache/stats
```

**Expected Response:**
```json
{
  "hit_rate": 45.5,
  "item_count": 10
}
```

## 7. Clear Cache

Clear all cached search results:

```bash
curl -X DELETE http://localhost:8080/api/v1/cache
```

**Expected Response:**
```json
{
  "message": "Cache cleared successfully"
}
```

## Testing Workflow

1. Start the server:
   ```bash
   make run
   ```

2. Test health check:
   ```bash
   curl http://localhost:8080/api/v1/health
   ```

3. Run a basic search:
   ```bash
   curl -X POST http://localhost:8080/api/v1/search \
     -H "Content-Type: application/json" \
     -d @examples/basic-search.json
   ```

4. Run the same search again (should be cached):
   ```bash
   curl -X POST http://localhost:8080/api/v1/search \
     -H "Content-Type: application/json" \
     -d @examples/basic-search.json
   ```
   Check `cache_hit: true` in the response

5. Check cache stats:
   ```bash
   curl http://localhost:8080/api/v1/cache/stats
   ```

6. Try filtered search:
   ```bash
   curl -X POST http://localhost:8080/api/v1/search/filter \
     -H "Content-Type: application/json" \
     -d @examples/filtered-search.json
   ```

## Error Handling

### Invalid Request (400 Bad Request)

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "CGK"
  }'
```

**Response:**
```json
{
  "error": "INVALID_REQUEST",
  "message": "origin and destination cannot be the same"
}
```

### Invalid Date Format

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "15-12-2025",
    "passengers": 1,
    "cabin_class": "economy"
  }'
```

**Response:**
```json
{
  "error": "INVALID_REQUEST",
  "message": "departure_date must be in YYYY-MM-DD format"
}
```
