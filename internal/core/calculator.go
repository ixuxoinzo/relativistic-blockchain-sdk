package core

import (
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/relativistic"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type Calculator struct {
	logger *zap.Logger
}

func NewCalculator(logger *zap.Logger) *Calculator {
	return &Calculator{
		logger: logger,
	}
}

// ========== BASIC LIGHT & NETWORK DELAYS ==========

func (c *Calculator) CalculateLightDelay(distance float64) time.Duration {
	delaySeconds := distance / types.SpeedOfLight
	return time.Duration(delaySeconds * float64(time.Second))
}

func (c *Calculator) CalculateNetworkDelay(distance float64, networkFactor float64) time.Duration {
	lightDelay := c.CalculateLightDelay(distance)
	return time.Duration(float64(lightDelay) * networkFactor)
}

// ========== RELATIVISTIC & PHYSICAL DELAYS ==========

func (c *Calculator) CalculateRelativisticDelay(distance float64, velocity float64, networkFactor float64) time.Duration {
	lightDelay := distance / types.SpeedOfLight
	realisticDelay := lightDelay * networkFactor
	realisticDelayDuration := time.Duration(realisticDelay * float64(time.Second))
	relativisticDelay := relativistic.ApplyTimeDilation(realisticDelayDuration, velocity)
	return relativisticDelay
}

// ========== GEOGRAPHIC DISTANCES ==========

func (c *Calculator) CalculateGreatCircleDistance(pos1, pos2 types.Position) (float64, error) {
	lat1 := pos1.Latitude * math.Pi / 180
	lon1 := pos1.Longitude * math.Pi / 180
	lat2 := pos2.Latitude * math.Pi / 180
	lon2 := pos2.Longitude * math.Pi / 180

	dLat := lat2 - lat1
	dLon := lon2 - lon1

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)

	cValue := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := types.EarthRadius * cValue

	altDiff := math.Abs(pos1.Altitude - pos2.Altitude)
	totalDistance := math.Sqrt(distance*distance + altDiff*altDiff)
	return totalDistance, nil
}

// ========== BLOCK TIME OPTIMIZATION ==========

func (c *Calculator) CalculateOptimalBlockTime(delays []time.Duration, safetyFactor float64) time.Duration {
	if len(delays) == 0 {
		return 2 * time.Second
	}

	maxDelay := delays[0]
	for _, delay := range delays {
		if delay > maxDelay {
			maxDelay = delay
		}
	}

	optimalTime := time.Duration(float64(maxDelay) * safetyFactor)
	if optimalTime < 2*time.Second {
		optimalTime = 2 * time.Second
	}
	if optimalTime > 10*time.Minute {
		optimalTime = 10 * time.Minute
	}
	return optimalTime
}

// ========== CONFIDENCE SCORING ==========

func (c *Calculator) CalculateConfidenceScore(expectedDelay, actualDiff, maxAcceptable time.Duration) float64 {
	absDiff := actualDiff
	if actualDiff < 0 {
		absDiff = -actualDiff
	}
	if absDiff > maxAcceptable {
		return 0.0
	}
	confidence := 1.0 - (float64(absDiff) / float64(maxAcceptable))
	if confidence < 0 {
		return 0.0
	}
	return confidence
}

// ========== RELATIVISTIC FACTORS ==========

func (c *Calculator) CalculateTimeDilationFactor(velocity float64) float64 {
	return relativistic.CalculateLorentzFactor(velocity)
}

func (c *Calculator) CalculateGravitationalTimeDilation(gravityFieldStrength, height float64) float64 {
	if gravityFieldStrength == 0 {
		return 1.0
	}
	speedOfLight := types.SpeedOfLight
	phi := gravityFieldStrength * height
	return 1.0 + phi/(speedOfLight*speedOfLight)
}

// ========== SATELLITE COMMUNICATION DELAY ==========

func (c *Calculator) CalculateSatelliteDelay(satellitePos, groundStationPos types.Position, satelliteVelocity float64) (time.Duration, error) {
	distance, err := c.CalculateGreatCircleDistance(satellitePos, groundStationPos)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate distance: %w", err)
	}

	lightDelay := distance / types.SpeedOfLight
	relativisticDelay := relativistic.ApplyTimeDilation(time.Duration(lightDelay*float64(time.Second)), satelliteVelocity)
	totalDelay := time.Duration(float64(relativisticDelay) * 1.01)
	return totalDelay, nil
}

// ========== INTERPLANETARY COMMUNICATION DELAY ==========

func (c *Calculator) CalculateInterplanetaryDelayWithOrbits(planetA, planetB string, timeOfFlight time.Time) (time.Duration, error) {
	key := planetA + "-" + planetB
	distance, exists := types.PlanetaryDistances[key]
	if !exists {
		return 0, fmt.Errorf("planetary distance not found for %s", key)
	}
	distanceMeters := distance * 1000
	delaySeconds := distanceMeters / types.SpeedOfLight
	totalDelay := delaySeconds * 1.05
	return time.Duration(totalDelay * float64(time.Second)), nil
}

// ========== STATISTICAL ANALYSIS ==========

type DelayStatistics struct {
	MinDelay    time.Duration
	MaxDelay    time.Duration
	MeanDelay   time.Duration
	MedianDelay time.Duration
	StdDev      time.Duration
}

func (c *Calculator) CalculateDelayStatistics(delays []time.Duration) *DelayStatistics {
	if len(delays) == 0 {
		return &DelayStatistics{}
	}

	minDelay, maxDelay := delays[0], delays[0]
	sum := time.Duration(0)

	for _, d := range delays {
		if d < minDelay {
			minDelay = d
		}
		if d > maxDelay {
			maxDelay = d
		}
		sum += d
	}

	mean := sum / time.Duration(len(delays))
	var variance float64
	for _, d := range delays {
		diff := float64(d - mean)
		variance += diff * diff
	}
	variance /= float64(len(delays))
	stdDev := time.Duration(math.Sqrt(variance))
	median := c.calculateMedian(delays)

	return &DelayStatistics{
		MinDelay:    minDelay,
		MaxDelay:    maxDelay,
		MeanDelay:   mean,
		MedianDelay: median,
		StdDev:      stdDev,
	}
}

func (c *Calculator) calculateMedian(delays []time.Duration) time.Duration {
	n := len(delays)
	if n == 0 {
		return 0
	}
	sorted := make([]time.Duration, n)
	copy(sorted, delays)
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

