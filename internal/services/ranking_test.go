package services

import (
	"testing"
	"time"

	"github.com/reynaldipane/flight-search/internal/models"
	"github.com/stretchr/testify/assert"
)

func createTestFlightsForRanking() []*models.Flight {
	baseDate := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)

	return []*models.Flight{
		{
			ID:       "R1",
			Provider: "Garuda",
			Airline:  models.Airline{Name: "Garuda Indonesia", Code: "GA"},
			Departure: models.Location{
				Airport:  "CGK",
				DateTime: baseDate.Add(8 * time.Hour), // 08:00 - Morning (1.0 score)
			},
			Arrival: models.Location{
				Airport:  "DPS",
				DateTime: baseDate.Add(10 * time.Hour),
			},
			Duration:       models.Duration{TotalMinutes: 120}, // Medium duration
			Stops:          0,                                  // Direct flight (+30)
			Price:          models.Price{Amount: 1000000, Currency: "IDR"},
			AvailableSeats: 50,
			Amenities:      []string{"WiFi", "Meal", "Entertainment"},
		},
		{
			ID:       "R2",
			Provider: "Lion Air",
			Airline:  models.Airline{Name: "Lion Air", Code: "JT"},
			Departure: models.Location{
				Airport:  "CGK",
				DateTime: baseDate.Add(2 * time.Hour), // 02:00 - Red-eye (0.0 score)
			},
			Arrival: models.Location{
				Airport:  "DPS",
				DateTime: baseDate.Add(6 * time.Hour),
			},
			Duration:       models.Duration{TotalMinutes: 240}, // Long duration
			Stops:          2,                                  // 2 stops (0 points)
			Price:          models.Price{Amount: 500000, Currency: "IDR"},
			AvailableSeats: 20,
			Amenities:      []string{"Meal"},
		},
		{
			ID:       "R3",
			Provider: "AirAsia",
			Airline:  models.Airline{Name: "AirAsia", Code: "QZ"},
			Departure: models.Location{
				Airport:  "CGK",
				DateTime: baseDate.Add(14 * time.Hour), // 14:00 - Afternoon (0.9 score)
			},
			Arrival: models.Location{
				Airport:  "DPS",
				DateTime: baseDate.Add(15*time.Hour + 40*time.Minute),
			},
			Duration:       models.Duration{TotalMinutes: 100}, // Shortest duration
			Stops:          0,                                  // Direct flight (+30)
			Price:          models.Price{Amount: 750000, Currency: "IDR"},
			AvailableSeats: 100,                                // Most seats
			Amenities:      []string{"WiFi", "Meal", "Entertainment", "Lounge", "Priority"},
		},
		{
			ID:       "R4",
			Provider: "Batik Air",
			Airline:  models.Airline{Name: "Batik Air", Code: "ID"},
			Departure: models.Location{
				Airport:  "CGK",
				DateTime: baseDate.Add(20 * time.Hour), // 20:00 - Evening (0.7 score)
			},
			Arrival: models.Location{
				Airport:  "DPS",
				DateTime: baseDate.Add(22 * time.Hour),
			},
			Duration:       models.Duration{TotalMinutes: 120}, // Medium duration
			Stops:          1,                                  // 1 stop (+15)
			Price:          models.Price{Amount: 600000, Currency: "IDR"},
			AvailableSeats: 30,
			Amenities:      []string{"WiFi", "Meal"},
		},
	}
}

func TestRankingService_CalculatePriceScore(t *testing.T) {
	service := NewRankingService()
	flights := createTestFlightsForRanking()

	// Price range: 500000 (R2) to 1000000 (R1)
	// Range: 500000

	t.Run("Cheapest Flight - Maximum Score", func(t *testing.T) {
		score := service.calculatePriceScore(flights[1], flights) // R2: 500000
		assert.Equal(t, 100.0, score, "Cheapest flight should get 100 points")
	})

	t.Run("Most Expensive Flight - Minimum Score", func(t *testing.T) {
		score := service.calculatePriceScore(flights[0], flights) // R1: 1000000
		assert.Equal(t, 0.0, score, "Most expensive flight should get 0 points")
	})

	t.Run("Mid-Range Flight", func(t *testing.T) {
		score := service.calculatePriceScore(flights[2], flights) // R3: 750000
		// (1000000 - 750000) / (1000000 - 500000) * 100 = 250000/500000 * 100 = 50.0
		assert.Equal(t, 50.0, score, "Mid-range flight should get 50 points")
	})

	t.Run("All Same Price", func(t *testing.T) {
		samePriceFlights := []*models.Flight{
			{ID: "F1", Price: models.Price{Amount: 500000}},
			{ID: "F2", Price: models.Price{Amount: 500000}},
		}
		score := service.calculatePriceScore(samePriceFlights[0], samePriceFlights)
		assert.Equal(t, 100.0, score, "When all prices same, should return 100")
	})
}

func TestRankingService_CalculateDurationScore(t *testing.T) {
	service := NewRankingService()
	flights := createTestFlightsForRanking()

	// Duration range: 100 (R3) to 240 (R2)
	// Range: 140 minutes

	t.Run("Shortest Duration - Maximum Score", func(t *testing.T) {
		score := service.calculateDurationScore(flights[2], flights) // R3: 100 min
		assert.Equal(t, 1.0, score, "Shortest flight should get 1.0")
	})

	t.Run("Longest Duration - Minimum Score", func(t *testing.T) {
		score := service.calculateDurationScore(flights[1], flights) // R2: 240 min
		assert.Equal(t, 0.0, score, "Longest flight should get 0.0")
	})

	t.Run("Mid-Range Duration", func(t *testing.T) {
		score := service.calculateDurationScore(flights[0], flights) // R1: 120 min
		// (240 - 120) / (240 - 100) = 120 / 140 = 0.857...
		assert.InDelta(t, 0.857, score, 0.01, "Mid-range duration")
	})

	t.Run("All Same Duration", func(t *testing.T) {
		sameDurationFlights := []*models.Flight{
			{ID: "F1", Duration: models.Duration{TotalMinutes: 120}},
			{ID: "F2", Duration: models.Duration{TotalMinutes: 120}},
		}
		score := service.calculateDurationScore(sameDurationFlights[0], sameDurationFlights)
		assert.Equal(t, 1.0, score, "When all durations same, should return 1.0")
	})
}

func TestRankingService_CalculateTimeScore(t *testing.T) {
	service := NewRankingService()
	baseDate := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)

	t.Run("Morning Flight - Best Score", func(t *testing.T) {
		flight := &models.Flight{
			Departure: models.Location{DateTime: baseDate.Add(8 * time.Hour)}, // 08:00
		}
		score := service.calculateTimeScore(flight)
		assert.Equal(t, 1.0, score, "Morning flights should get 1.0")
	})

	t.Run("Afternoon Flight - Good Score", func(t *testing.T) {
		flight := &models.Flight{
			Departure: models.Location{DateTime: baseDate.Add(14 * time.Hour)}, // 14:00
		}
		score := service.calculateTimeScore(flight)
		assert.Equal(t, 0.9, score, "Afternoon flights should get 0.9")
	})

	t.Run("Evening Flight - Okay Score", func(t *testing.T) {
		flight := &models.Flight{
			Departure: models.Location{DateTime: baseDate.Add(20 * time.Hour)}, // 20:00
		}
		score := service.calculateTimeScore(flight)
		assert.Equal(t, 0.7, score, "Evening flights should get 0.7")
	})

	t.Run("Red-Eye Flight - Worst Score", func(t *testing.T) {
		flight := &models.Flight{
			Departure: models.Location{DateTime: baseDate.Add(2 * time.Hour)}, // 02:00
		}
		score := service.calculateTimeScore(flight)
		assert.Equal(t, 0.0, score, "Red-eye flights should get 0.0")
	})

	t.Run("Boundary Cases", func(t *testing.T) {
		// 05:59 - still red-eye
		flight1 := &models.Flight{
			Departure: models.Location{DateTime: baseDate.Add(5*time.Hour + 59*time.Minute)},
		}
		assert.Equal(t, 0.0, service.calculateTimeScore(flight1))

		// 06:00 - morning starts
		flight2 := &models.Flight{
			Departure: models.Location{DateTime: baseDate.Add(6 * time.Hour)},
		}
		assert.Equal(t, 1.0, service.calculateTimeScore(flight2))

		// 11:59 - still morning
		flight3 := &models.Flight{
			Departure: models.Location{DateTime: baseDate.Add(11*time.Hour + 59*time.Minute)},
		}
		assert.Equal(t, 1.0, service.calculateTimeScore(flight3))

		// 12:00 - afternoon starts
		flight4 := &models.Flight{
			Departure: models.Location{DateTime: baseDate.Add(12 * time.Hour)},
		}
		assert.Equal(t, 0.9, service.calculateTimeScore(flight4))
	})
}

func TestRankingService_CalculateSeatsScore(t *testing.T) {
	service := NewRankingService()
	flights := createTestFlightsForRanking()

	// Seats range: 20 (R2) to 100 (R3)
	// Range: 80

	t.Run("Most Seats - Maximum Score", func(t *testing.T) {
		score := service.calculateSeatsScore(flights[2], flights) // R3: 100 seats
		assert.Equal(t, 1.0, score, "Most seats should get 1.0")
	})

	t.Run("Least Seats - Minimum Score", func(t *testing.T) {
		score := service.calculateSeatsScore(flights[1], flights) // R2: 20 seats
		assert.Equal(t, 0.0, score, "Least seats should get 0.0")
	})

	t.Run("Mid-Range Seats", func(t *testing.T) {
		score := service.calculateSeatsScore(flights[0], flights) // R1: 50 seats
		// (50 - 20) / (100 - 20) = 30 / 80 = 0.375
		assert.Equal(t, 0.375, score, "Mid-range seats")
	})

	t.Run("All Same Seats", func(t *testing.T) {
		sameSeatsFlights := []*models.Flight{
			{ID: "F1", AvailableSeats: 50},
			{ID: "F2", AvailableSeats: 50},
		}
		score := service.calculateSeatsScore(sameSeatsFlights[0], sameSeatsFlights)
		assert.Equal(t, 1.0, score, "When all seats same, should return 1.0")
	})
}

func TestRankingService_CalculateAmenitiesScore(t *testing.T) {
	service := NewRankingService()
	flights := createTestFlightsForRanking()

	// Amenities count: 1 (R2) to 5 (R3)
	// Max: 5

	t.Run("Most Amenities - Maximum Score", func(t *testing.T) {
		score := service.calculateAmenitiesScore(flights[2], flights) // R3: 5 amenities
		assert.Equal(t, 1.0, score, "Most amenities should get 1.0")
	})

	t.Run("Fewest Amenities", func(t *testing.T) {
		score := service.calculateAmenitiesScore(flights[1], flights) // R2: 1 amenity
		assert.Equal(t, 0.2, score, "1 amenity when max is 5 = 0.2")
	})

	t.Run("Mid-Range Amenities", func(t *testing.T) {
		score := service.calculateAmenitiesScore(flights[0], flights) // R1: 3 amenities
		assert.Equal(t, 0.6, score, "3 amenities when max is 5 = 0.6")
	})

	t.Run("No Amenities", func(t *testing.T) {
		noAmenitiesFlights := []*models.Flight{
			{ID: "F1", Amenities: []string{}},
			{ID: "F2", Amenities: []string{}},
		}
		score := service.calculateAmenitiesScore(noAmenitiesFlights[0], noAmenitiesFlights)
		assert.Equal(t, 0.0, score, "No amenities should return 0.0")
	})
}

func TestRankingService_CalculateConvenienceScore(t *testing.T) {
	service := NewRankingService()
	flights := createTestFlightsForRanking()

	t.Run("Direct Flight Morning Best Convenience", func(t *testing.T) {
		// R1: Direct (30) + Morning (20) + other factors
		score := service.calculateConvenienceScore(flights[0], flights)

		// Expected breakdown:
		// - Direct flight: +30
		// - Duration: (240-120)/(240-100) * 30 = 120/140 * 30 = 25.71
		// - Morning time: 1.0 * 20 = 20
		// - Seats: (50-20)/(100-20) * 10 = 30/80 * 10 = 3.75
		// - Amenities: 3/5 * 10 = 6
		// Total: 30 + 25.71 + 20 + 3.75 + 6 = 85.46

		assert.InDelta(t, 85.46, score, 0.1, "Direct morning flight should have high convenience")
		assert.True(t, score > 80, "Should be above 80")
	})

	t.Run("Red-Eye Multi-Stop Worst Convenience", func(t *testing.T) {
		// R2: 2 stops (0) + Red-eye (0) + other factors
		score := service.calculateConvenienceScore(flights[1], flights)

		// Expected breakdown:
		// - 2 stops: 0
		// - Duration: (240-240)/(240-100) * 30 = 0
		// - Red-eye time: 0.0 * 20 = 0
		// - Seats: (20-20)/(100-20) * 10 = 0
		// - Amenities: 1/5 * 10 = 2
		// Total: 0 + 0 + 0 + 0 + 2 = 2

		assert.InDelta(t, 2.0, score, 0.1, "Red-eye multi-stop should have low convenience")
		assert.True(t, score < 10, "Should be very low")
	})

	t.Run("One Stop Gets Half Points", func(t *testing.T) {
		// R4: 1 stop should get 15 points
		score := service.calculateConvenienceScore(flights[3], flights)

		// Should have 15 points from stops
		assert.True(t, score > 15, "Should include 15 points from 1 stop")
		assert.True(t, score > 30, "Should have other convenience factors too")
	})
}

func TestRankingService_CalculateBestValueScore(t *testing.T) {
	service := NewRankingService()
	flights := createTestFlightsForRanking()

	t.Run("Best Value Calculation", func(t *testing.T) {
		// R3 should have good best value: mid-price, direct, short, most amenities
		score := service.CalculateBestValueScore(flights[2], flights)

		// Price score: (1000000 - 750000) / (1000000 - 500000) * 100 = 50
		// Convenience: Direct (30) + Shortest duration (30) + Afternoon (18) + Most seats (10) + Most amenities (10) = ~98
		// Best value: (50 * 0.5) + (98 * 0.5) = 25 + 49 = 74

		assert.True(t, score > 70, "R3 should have high best value score")
		assert.True(t, score < 100, "Should not be perfect 100")
	})

	t.Run("Cheap But Inconvenient", func(t *testing.T) {
		// R2: Cheapest (100 price score) but terrible convenience (2 stops, red-eye, long)
		score := service.CalculateBestValueScore(flights[1], flights)

		// Price score: 100
		// Convenience: Very low (~2)
		// Best value: (100 * 0.5) + (2 * 0.5) = 50 + 1 = 51

		assert.InDelta(t, 51.0, score, 5.0, "Cheap but inconvenient flight")
		assert.True(t, score > 45 && score < 60, "Should be mid-range overall")
	})

	t.Run("Expensive But Convenient", func(t *testing.T) {
		// R1: Most expensive (0 price score) but good convenience
		score := service.CalculateBestValueScore(flights[0], flights)

		// Price score: 0
		// Convenience: High (~85)
		// Best value: (0 * 0.5) + (85 * 0.5) = 0 + 42.5 = 42.5

		assert.InDelta(t, 42.5, score, 5.0, "Expensive but convenient flight")
		assert.True(t, score > 35 && score < 50, "Should be lower mid-range")
	})

	t.Run("50-50 Weight Distribution", func(t *testing.T) {
		// Verify that price and convenience are equally weighted
		for _, flight := range flights {
			score := service.CalculateBestValueScore(flight, flights)
			priceScore := service.calculatePriceScore(flight, flights)
			convenienceScore := service.calculateConvenienceScore(flight, flights)

			expected := (priceScore * 0.5) + (convenienceScore * 0.5)
			assert.Equal(t, expected, score, "Score should be exactly 50-50 weighted")
		}
	})
}

func TestRankingService_RankFlights(t *testing.T) {
	service := NewRankingService()

	t.Run("Rank All Flights", func(t *testing.T) {
		flights := createTestFlightsForRanking()
		ranked := service.RankFlights(flights)

		assert.Len(t, ranked, 4, "Should return all flights")

		// Verify descending order
		for i := 0; i < len(ranked)-1; i++ {
			score1 := service.CalculateBestValueScore(ranked[i], flights)
			score2 := service.CalculateBestValueScore(ranked[i+1], flights)
			assert.GreaterOrEqual(t, score1, score2, "Should be in descending order by score")
		}

		// R3 should likely be first (direct, short, mid-price, most amenities)
		// This is based on the scoring algorithm
		assert.Equal(t, "R3", ranked[0].ID, "R3 should have best value")
	})

	t.Run("Empty Flight List", func(t *testing.T) {
		ranked := service.RankFlights([]*models.Flight{})
		assert.Len(t, ranked, 0, "Should handle empty list")
	})

	t.Run("Single Flight", func(t *testing.T) {
		flights := createTestFlightsForRanking()
		singleFlight := []*models.Flight{flights[0]}
		ranked := service.RankFlights(singleFlight)
		assert.Len(t, ranked, 1, "Should handle single flight")
		assert.Equal(t, "R1", ranked[0].ID, "Should return the same flight")
	})

	t.Run("Does Not Modify Original Slice Order Unexpectedly", func(t *testing.T) {
		flights := createTestFlightsForRanking()
		// Create a copy to verify original is modified (sort.Slice modifies in place)
		flightsCopy := make([]*models.Flight, len(flights))
		copy(flightsCopy, flights)

		ranked := service.RankFlights(flightsCopy)

		// RankFlights modifies the slice in place, so they should be the same
		assert.Equal(t, ranked, flightsCopy, "RankFlights modifies slice in place")
	})
}

func TestRankingService_GetTopFlights(t *testing.T) {
	service := NewRankingService()

	t.Run("Get Top 2 Flights", func(t *testing.T) {
		flights := createTestFlightsForRanking()
		top := service.GetTopFlights(flights, 2)

		assert.Len(t, top, 2, "Should return top 2 flights")

		// Verify they are the best scored
		score1 := service.CalculateBestValueScore(top[0], flights)
		score2 := service.CalculateBestValueScore(top[1], flights)
		assert.GreaterOrEqual(t, score1, score2, "Should be in descending order")
	})

	t.Run("Request More Than Available", func(t *testing.T) {
		flights := createTestFlightsForRanking()
		top := service.GetTopFlights(flights, 10)

		assert.Len(t, top, 4, "Should return all available flights when limit exceeds total")
	})

	t.Run("Get Top 1 Flight", func(t *testing.T) {
		flights := createTestFlightsForRanking()
		top := service.GetTopFlights(flights, 1)

		assert.Len(t, top, 1, "Should return top 1 flight")
		assert.Equal(t, "R3", top[0].ID, "R3 should be the best value flight")
	})

	t.Run("Zero Limit", func(t *testing.T) {
		flights := createTestFlightsForRanking()
		top := service.GetTopFlights(flights, 0)

		assert.Len(t, top, 0, "Should return empty slice for 0 limit")
	})

	t.Run("Empty Flight List", func(t *testing.T) {
		top := service.GetTopFlights([]*models.Flight{}, 5)

		assert.Len(t, top, 0, "Should handle empty list")
	})
}

func TestRankingService_ScoringConsistency(t *testing.T) {
	service := NewRankingService()
	flights := createTestFlightsForRanking()

	t.Run("Same Flight Same Score", func(t *testing.T) {
		score1 := service.CalculateBestValueScore(flights[0], flights)
		score2 := service.CalculateBestValueScore(flights[0], flights)

		assert.Equal(t, score1, score2, "Same flight should get same score")
	})

	t.Run("Score Within Valid Range", func(t *testing.T) {
		for _, flight := range flights {
			score := service.CalculateBestValueScore(flight, flights)
			assert.GreaterOrEqual(t, score, 0.0, "Score should be >= 0")
			assert.LessOrEqual(t, score, 100.0, "Score should be <= 100")
		}
	})

	t.Run("Price Score Within Valid Range", func(t *testing.T) {
		for _, flight := range flights {
			score := service.calculatePriceScore(flight, flights)
			assert.GreaterOrEqual(t, score, 0.0, "Price score should be >= 0")
			assert.LessOrEqual(t, score, 100.0, "Price score should be <= 100")
		}
	})

	t.Run("Convenience Score Within Valid Range", func(t *testing.T) {
		for _, flight := range flights {
			score := service.calculateConvenienceScore(flight, flights)
			assert.GreaterOrEqual(t, score, 0.0, "Convenience score should be >= 0")
			assert.LessOrEqual(t, score, 100.0, "Convenience score should be <= 100")
		}
	})
}
