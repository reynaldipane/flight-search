package services

import (
	"math"
	"sort"

	"github.com/reynaldipane/flight-search/internal/models"
)

// RankingService provides flight ranking and scoring functionality
type RankingService struct{}

// NewRankingService creates a new ranking service
func NewRankingService() *RankingService {
	return &RankingService{}
}

// CalculateBestValueScore calculates a best value score for a flight
// Score ranges from 0-100, higher is better
// Factors: 50% price, 50% convenience
func (s *RankingService) CalculateBestValueScore(flight *models.Flight, allFlights []*models.Flight) float64 {
	if len(allFlights) == 0 {
		return 0
	}

	// Calculate price score (50% weight)
	priceScore := s.calculatePriceScore(flight, allFlights)

	// Calculate convenience score (50% weight)
	convenienceScore := s.calculateConvenienceScore(flight, allFlights)

	// Combined weighted score
	return (priceScore * 0.5) + (convenienceScore * 0.5)
}

// calculatePriceScore scores based on price (lower is better)
// Returns 0-100, where 100 is the cheapest
func (s *RankingService) calculatePriceScore(flight *models.Flight, allFlights []*models.Flight) float64 {
	if len(allFlights) == 0 {
		return 0
	}

	// Find min and max prices
	minPrice := flight.Price.Amount
	maxPrice := flight.Price.Amount

	for _, f := range allFlights {
		if f.Price.Amount < minPrice {
			minPrice = f.Price.Amount
		}
		if f.Price.Amount > maxPrice {
			maxPrice = f.Price.Amount
		}
	}

	// Avoid division by zero
	if maxPrice == minPrice {
		return 100 // All flights same price
	}

	// Normalize: cheaper flights get higher scores
	// If flight is cheapest (minPrice), score = 100
	// If flight is most expensive (maxPrice), score = 0
	score := ((maxPrice - flight.Price.Amount) / (maxPrice - minPrice)) * 100

	return score
}

// calculateConvenienceScore scores based on convenience factors
// Returns 0-100, higher is better
// Factors:
// - Direct flights: +30 points
// - Shorter duration: +30 points (normalized)
// - Preferred time of day (not red-eye): +20 points
// - More available seats: +10 points (normalized)
// - Amenities: +10 points (normalized)
func (s *RankingService) calculateConvenienceScore(flight *models.Flight, allFlights []*models.Flight) float64 {
	score := 0.0

	// 1. Direct flights bonus (30 points)
	if flight.Stops == 0 {
		score += 30
	} else if flight.Stops == 1 {
		score += 15 // Half points for 1 stop
	}
	// 0 points for 2+ stops

	// 2. Duration score (30 points)
	score += s.calculateDurationScore(flight, allFlights) * 30

	// 3. Time of day preference (20 points)
	score += s.calculateTimeScore(flight) * 20

	// 4. Available seats score (10 points)
	score += s.calculateSeatsScore(flight, allFlights) * 10

	// 5. Amenities score (10 points)
	score += s.calculateAmenitiesScore(flight, allFlights) * 10

	return math.Min(score, 100) // Cap at 100
}

// calculateDurationScore returns 0-1 normalized score based on duration
func (s *RankingService) calculateDurationScore(flight *models.Flight, allFlights []*models.Flight) float64 {
	if len(allFlights) == 0 {
		return 0
	}

	minDuration := flight.Duration.TotalMinutes
	maxDuration := flight.Duration.TotalMinutes

	for _, f := range allFlights {
		if f.Duration.TotalMinutes < minDuration {
			minDuration = f.Duration.TotalMinutes
		}
		if f.Duration.TotalMinutes > maxDuration {
			maxDuration = f.Duration.TotalMinutes
		}
	}

	if maxDuration == minDuration {
		return 1.0 // All same duration
	}

	// Shorter flights get higher scores
	return (float64(maxDuration-flight.Duration.TotalMinutes) / float64(maxDuration-minDuration))
}

// calculateTimeScore returns 0-1 score based on time of day
// Penalizes red-eye flights (00:00-05:00)
func (s *RankingService) calculateTimeScore(flight *models.Flight) float64 {
	hour := flight.Departure.DateTime.Hour()

	// Red-eye flights (midnight to 5am): 0 points
	if hour >= 0 && hour < 6 {
		return 0.0
	}

	// Morning flights (6am-12pm): 1.0 points (preferred)
	if hour >= 6 && hour < 12 {
		return 1.0
	}

	// Afternoon flights (12pm-6pm): 0.9 points
	if hour >= 12 && hour < 18 {
		return 0.9
	}

	// Evening flights (6pm-midnight): 0.7 points
	return 0.7
}

// calculateSeatsScore returns 0-1 normalized score based on available seats
func (s *RankingService) calculateSeatsScore(flight *models.Flight, allFlights []*models.Flight) float64 {
	if len(allFlights) == 0 {
		return 0
	}

	minSeats := flight.AvailableSeats
	maxSeats := flight.AvailableSeats

	for _, f := range allFlights {
		if f.AvailableSeats < minSeats {
			minSeats = f.AvailableSeats
		}
		if f.AvailableSeats > maxSeats {
			maxSeats = f.AvailableSeats
		}
	}

	if maxSeats == minSeats {
		return 1.0
	}

	// More seats = higher score
	return float64(flight.AvailableSeats-minSeats) / float64(maxSeats-minSeats)
}

// calculateAmenitiesScore returns 0-1 score based on amenities count
func (s *RankingService) calculateAmenitiesScore(flight *models.Flight, allFlights []*models.Flight) float64 {
	if len(allFlights) == 0 {
		return 0
	}

	flightAmenities := len(flight.Amenities)
	maxAmenities := 0

	for _, f := range allFlights {
		if len(f.Amenities) > maxAmenities {
			maxAmenities = len(f.Amenities)
		}
	}

	if maxAmenities == 0 {
		return 0
	}

	return float64(flightAmenities) / float64(maxAmenities)
}

// RankFlights scores all flights and sorts by best value (highest score first)
func (s *RankingService) RankFlights(flights []*models.Flight) []*models.Flight {
	if len(flights) == 0 {
		return flights
	}

	// Calculate scores for all flights
	scores := make(map[string]float64)
	for _, flight := range flights {
		scores[flight.ID] = s.CalculateBestValueScore(flight, flights)
	}

	// Sort by score (descending)
	sort.Slice(flights, func(i, j int) bool {
		scoreI := scores[flights[i].ID]
		scoreJ := scores[flights[j].ID]
		return scoreI > scoreJ
	})

	return flights
}

// GetTopFlights returns the top N flights by best value score
func (s *RankingService) GetTopFlights(flights []*models.Flight, limit int) []*models.Flight {
	ranked := s.RankFlights(flights)

	if len(ranked) <= limit {
		return ranked
	}

	return ranked[:limit]
}
