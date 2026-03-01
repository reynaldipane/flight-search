package services

import (
	"testing"
	"time"

	"github.com/reynaldipane/flight-search/internal/models"
	"github.com/stretchr/testify/assert"
)

func createTestFlights() []*models.Flight {
	// Use fixed date/time for consistent testing
	baseDate := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)

	return []*models.Flight{
		{
			ID:       "F1",
			Provider: "Airline A",
			Airline:  models.Airline{Name: "Garuda Indonesia", Code: "GA"},
			Departure: models.Location{
				Airport:  "CGK",
				DateTime: baseDate.Add(8 * time.Hour), // 08:00 - Morning
			},
			Arrival: models.Location{
				Airport:  "DPS",
				DateTime: baseDate.Add(10 * time.Hour), // 10:00
			},
			Duration: models.Duration{TotalMinutes: 120},
			Stops:    0,
			Price:    models.Price{Amount: 1000000, Currency: "IDR"},
		},
		{
			ID:       "F2",
			Provider: "Airline B",
			Airline:  models.Airline{Name: "Lion Air", Code: "JT"},
			Departure: models.Location{
				Airport:  "CGK",
				DateTime: baseDate.Add(14 * time.Hour), // 14:00 - Afternoon
			},
			Arrival: models.Location{
				Airport:  "DPS",
				DateTime: baseDate.Add(17 * time.Hour), // 17:00
			},
			Duration: models.Duration{TotalMinutes: 180},
			Stops:    1,
			Price:    models.Price{Amount: 750000, Currency: "IDR"},
		},
		{
			ID:       "F3",
			Provider: "Airline C",
			Airline:  models.Airline{Name: "AirAsia", Code: "QZ"},
			Departure: models.Location{
				Airport:  "CGK",
				DateTime: baseDate.Add(20 * time.Hour), // 20:00 - Evening
			},
			Arrival: models.Location{
				Airport:  "DPS",
				DateTime: baseDate.Add(22 * time.Hour), // 22:00
			},
			Duration: models.Duration{TotalMinutes: 100},
			Stops:    0,
			Price:    models.Price{Amount: 500000, Currency: "IDR"},
		},
	}
}

func TestFilterService_ApplyFilters_PriceRange(t *testing.T) {
	service := NewFilterService()
	flights := createTestFlights()

	t.Run("Min Price Filter", func(t *testing.T) {
		minPrice := 600000.0
		opts := models.FilterOptions{MinPrice: &minPrice}

		filtered := service.ApplyFilters(flights, opts)

		assert.Len(t, filtered, 2, "Should filter out flights below min price")
		for _, f := range filtered {
			assert.GreaterOrEqual(t, f.Price.Amount, minPrice)
		}
	})

	t.Run("Max Price Filter", func(t *testing.T) {
		maxPrice := 800000.0
		opts := models.FilterOptions{MaxPrice: &maxPrice}

		filtered := service.ApplyFilters(flights, opts)

		assert.Len(t, filtered, 2, "Should filter out flights above max price")
		for _, f := range filtered {
			assert.LessOrEqual(t, f.Price.Amount, maxPrice)
		}
	})

	t.Run("Price Range Filter", func(t *testing.T) {
		minPrice := 600000.0
		maxPrice := 900000.0
		opts := models.FilterOptions{
			MinPrice: &minPrice,
			MaxPrice: &maxPrice,
		}

		filtered := service.ApplyFilters(flights, opts)

		assert.Len(t, filtered, 1, "Should return only flights in price range")
		assert.Equal(t, "F2", filtered[0].ID)
		assert.Equal(t, 750000.0, filtered[0].Price.Amount)
	})
}

func TestFilterService_ApplyFilters_Stops(t *testing.T) {
	service := NewFilterService()
	flights := createTestFlights()

	t.Run("Direct Flights Only", func(t *testing.T) {
		maxStops := 0
		opts := models.FilterOptions{MaxStops: &maxStops}

		filtered := service.ApplyFilters(flights, opts)

		assert.Len(t, filtered, 2, "Should return only direct flights")
		for _, f := range filtered {
			assert.Equal(t, 0, f.Stops)
		}
	})

	t.Run("Up To One Stop", func(t *testing.T) {
		maxStops := 1
		opts := models.FilterOptions{MaxStops: &maxStops}

		filtered := service.ApplyFilters(flights, opts)

		assert.Len(t, filtered, 3, "Should return all flights (0 or 1 stop)")
	})
}

func TestFilterService_ApplyFilters_Airlines(t *testing.T) {
	service := NewFilterService()
	flights := createTestFlights()

	t.Run("Filter By Airline Name", func(t *testing.T) {
		opts := models.FilterOptions{
			Airlines: []string{"Garuda Indonesia"},
		}

		filtered := service.ApplyFilters(flights, opts)

		assert.Len(t, filtered, 1, "Should return only Garuda flights")
		assert.Equal(t, "Garuda Indonesia", filtered[0].Airline.Name)
	})

	t.Run("Filter By Airline Code", func(t *testing.T) {
		opts := models.FilterOptions{
			Airlines: []string{"QZ"},
		}

		filtered := service.ApplyFilters(flights, opts)

		assert.Len(t, filtered, 1, "Should return only AirAsia flights")
		assert.Equal(t, "QZ", filtered[0].Airline.Code)
	})

	t.Run("Multiple Airlines", func(t *testing.T) {
		opts := models.FilterOptions{
			Airlines: []string{"Garuda Indonesia", "Lion Air"},
		}

		filtered := service.ApplyFilters(flights, opts)

		assert.Len(t, filtered, 2, "Should return Garuda and Lion Air flights")
	})

	t.Run("Case Insensitive", func(t *testing.T) {
		opts := models.FilterOptions{
			Airlines: []string{"garuda indonesia"},
		}

		filtered := service.ApplyFilters(flights, opts)

		assert.Len(t, filtered, 1, "Should match case-insensitively")
	})
}

func TestFilterService_ApplyFilters_Duration(t *testing.T) {
	service := NewFilterService()
	flights := createTestFlights()

	maxDuration := 150
	opts := models.FilterOptions{MaxDuration: &maxDuration}

	filtered := service.ApplyFilters(flights, opts)

	assert.Len(t, filtered, 2, "Should filter out flights longer than max duration")
	for _, f := range filtered {
		assert.LessOrEqual(t, f.Duration.TotalMinutes, maxDuration)
	}
}

func TestFilterService_ApplyFilters_TimeWindow(t *testing.T) {
	service := NewFilterService()
	flights := createTestFlights()

	t.Run("Morning Departures", func(t *testing.T) {
		timeWindow := "morning"
		opts := models.FilterOptions{DepartureTime: &timeWindow}

		filtered := service.ApplyFilters(flights, opts)

		assert.Len(t, filtered, 1, "Should return only morning departures")
		assert.Equal(t, "F1", filtered[0].ID)
	})

	t.Run("Afternoon Departures", func(t *testing.T) {
		timeWindow := "afternoon"
		opts := models.FilterOptions{DepartureTime: &timeWindow}

		filtered := service.ApplyFilters(flights, opts)

		assert.Len(t, filtered, 1, "Should return only afternoon departures")
		assert.Equal(t, "F2", filtered[0].ID)
	})

	t.Run("Evening Departures", func(t *testing.T) {
		timeWindow := "evening"
		opts := models.FilterOptions{DepartureTime: &timeWindow}

		filtered := service.ApplyFilters(flights, opts)

		assert.Len(t, filtered, 1, "Should return only evening departures")
		assert.Equal(t, "F3", filtered[0].ID)
	})
}

func TestFilterService_SortFlights(t *testing.T) {
	service := NewFilterService()

	t.Run("Sort By Price Ascending", func(t *testing.T) {
		flights := createTestFlights()
		service.SortFlights(flights, models.SortByPriceAsc)

		assert.Equal(t, "F3", flights[0].ID, "Cheapest first")
		assert.Equal(t, "F2", flights[1].ID)
		assert.Equal(t, "F1", flights[2].ID, "Most expensive last")
	})

	t.Run("Sort By Price Descending", func(t *testing.T) {
		flights := createTestFlights()
		service.SortFlights(flights, models.SortByPriceDesc)

		assert.Equal(t, "F1", flights[0].ID, "Most expensive first")
		assert.Equal(t, "F2", flights[1].ID)
		assert.Equal(t, "F3", flights[2].ID, "Cheapest last")
	})

	t.Run("Sort By Duration Ascending", func(t *testing.T) {
		flights := createTestFlights()
		service.SortFlights(flights, models.SortByDurationAsc)

		assert.Equal(t, "F3", flights[0].ID, "Shortest first (100 min)")
		assert.Equal(t, "F1", flights[1].ID, "120 min")
		assert.Equal(t, "F2", flights[2].ID, "Longest last (180 min)")
	})

	t.Run("Sort By Duration Descending", func(t *testing.T) {
		flights := createTestFlights()
		service.SortFlights(flights, models.SortByDurationDesc)

		assert.Equal(t, "F2", flights[0].ID, "Longest first")
		assert.Equal(t, "F1", flights[1].ID)
		assert.Equal(t, "F3", flights[2].ID, "Shortest last")
	})

	t.Run("Sort By Departure Time Ascending", func(t *testing.T) {
		flights := createTestFlights()
		service.SortFlights(flights, models.SortByDepartureAsc)

		assert.Equal(t, "F1", flights[0].ID, "Earliest departure")
		assert.Equal(t, "F2", flights[1].ID)
		assert.Equal(t, "F3", flights[2].ID, "Latest departure")
	})
}

func TestFilterService_FilterAndSort(t *testing.T) {
	service := NewFilterService()
	flights := createTestFlights()

	// Filter: direct flights only, price under 1000000
	// Sort: by price ascending
	maxStops := 0
	maxPrice := 1000000.0
	opts := models.FilterOptions{
		MaxStops: &maxStops,
		MaxPrice: &maxPrice,
	}

	result := service.FilterAndSort(flights, opts, models.SortByPriceAsc)

	assert.Len(t, result, 2, "Should return 2 direct flights under price limit")
	assert.Equal(t, "F3", result[0].ID, "Cheapest direct flight first")
	assert.Equal(t, "F1", result[1].ID, "More expensive direct flight second")
}

func TestFilterService_CombinedFilters(t *testing.T) {
	service := NewFilterService()
	flights := createTestFlights()

	// Complex filter: price 400000-800000, max 1 stop, max 150 min
	minPrice := 400000.0
	maxPrice := 800000.0
	maxStops := 1
	maxDuration := 150

	opts := models.FilterOptions{
		MinPrice:    &minPrice,
		MaxPrice:    &maxPrice,
		MaxStops:    &maxStops,
		MaxDuration: &maxDuration,
	}

	filtered := service.ApplyFilters(flights, opts)

	assert.Len(t, filtered, 1, "Should match complex filter criteria")
	assert.Equal(t, "F3", filtered[0].ID)
}
