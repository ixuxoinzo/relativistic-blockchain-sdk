package core

import (
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/relativistic"
)

type Calculator struct {
	logger *zap.Logger
}

func NewCalculator(logger *zap.Logger) *Calculator {
	return &Calculator{
		logger: logger,
	}
}

func (c *Calculator) CalculateLightDelay(distance float64) time.Duration {
	delaySeconds := distance / types.SpeedOfLight
	return time.Duration(delaySeconds * float64(time.Second))
}

func (c *Calculator) CalculateNetworkDelay(distance float64, networkFactor float64) time.Duration {
	lightDelay := c.CalculateLightDelay(distance)
	return time.Duration(float64(lightDelay) * networkFactor)
}

func (c *Calculator) CalculateRelativisticDelay(distance float64, velocity float64, networkFactor float64) time.Duration {
    lightDelay := distance / types.SpeedOfLight
    realisticDelay := lightDelay * networkFactor
    realisticDelayDuration := time.Duration(realisticDelay * float64(time.Second))  
    relativisticDelay := relativistic.ApplyTimeDilation(realisticDelayDuration, velocity)
    return relativisticDelay 
}

func (c *Calculator) CalculateGreatCircleDistance(pos1, pos2 types.Position) (float64, error) {
	lat1 := pos1.Latitude * math.Pi / 180
	lon1 := pos1.Longitude * math.Pi / 180
	lat2 := pos2.Latitude * math.Pi / 180
	lon2 := pos2.Longitude * math.Pi / 180

	dLat := lat2 - lat1
	dLon := lon2 - lon1

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	
	cValue := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := types.EarthRadius * cValue

	altDiff := math.Abs(pos1.Altitude - pos2.Altitude)
	totalDistance := math.Sqrt(math.Pow(distance, 2) + math.Pow(altDiff, 2))

	return totalDistance, nil
}

func (c *Calculator) CalculateOptimalBlockTime(delays []time.Duration, safetyFactor float64) time.Duration {
	if len(delays) == 0 {
		return time.Second * 2
	}

	maxDelay := delays[0]
	for _, delay := range delays {
		if delay > maxDelay {
			maxDelay = delay
		}
	}

	optimalTime := time.Duration(float64(maxDelay) * safetyFactor)

	minBlockTime := time.Second * 2
	if optimalTime < minBlockTime {
		optimalTime = minBlockTime
	}

	maxBlockTime := time.Minute * 10
	if optimalTime > maxBlockTime {
		optimalTime = maxBlockTime
	}

	return optimalTime
}

func (c *Calculator) CalculateConfidenceScore(expectedDelay, actualDiff time.Duration, maxAcceptable time.Duration) float64 {
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

func (c *Calculator) CalculateTimeDilationFactor(velocity float64) float64 {
	return relativistic.CalculateLorentzFactor(velocity)
}

func (c *Calculator) CalculateGravitationalTimeDilation(gravityFieldStrength, height float64) float64 {
	if gravityFieldStrength == 0 {
		return 1.0
	}

	c := types.SpeedOfLight
	phi := gravityFieldStrength * height
	return 1.0 + phi/(c*c)
}

func (c *Calculator) CalculateSatelliteDelay(satellitePos, groundStationPos types.Position, satelliteVelocity float64) (time.Duration, error) {
	distance, err := c.CalculateGreatCircleDistance(satellitePos, groundStationPos)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate distance: %w", err)
	}

	lightDelay := distance / types.SpeedOfLight
	relativisticDelay := relativistic.ApplyTimeDilation(lightDelay, satelliteVelocity)
	totalDelay := relativisticDelay * 1.01

	return time.Duration(totalDelay * float64(time.Second)), nil
}

func (c *Calculator) CalculateInterplanetaryDelayWithOrbits(planetA, planetB string, timeOfFlight time.Time) (time.Duration, error) {
	distance, exists := types.PlanetaryDistances[planetA+"-"+planetB]
	if !exists {
		return 0, fmt.Errorf("planetary distance not found for %s-%s", planetA, planetB)
	}

	distanceMeters := distance * 1000
	delaySeconds := distanceMeters / types.SpeedOfLight
	totalDelay := delaySeconds * 1.05

	return time.Duration(totalDelay * float64(time.Second)), nil
}

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

	delayFloats := make([]float64, len(delays))
	for i, delay := range delays {
		delayFloats[i] = float64(delay)
	}

	minDelay := delays[0]
	maxDelay := delays[0]
	sum := time.Duration(0)

	for _, delay := range delays {
		if delay < minDelay {
			minDelay = delay
		}
		if delay > maxDelay {
			maxDelay = delay
		}
		sum += delay
	}

	meanDelay := sum / time.Duration(len(delays))

	var variance float64
	for _, delay := range delays {
		diff := float64(delay - meanDelay)
		variance += diff * diff
	}
	variance /= float64(len(delays))
	stdDev := time.Duration(math.Sqrt(float64(variance)))

	medianDelay := c.calculateMedian(delays)

	return &DelayStatistics{
		MinDelay:    minDelay,
		MaxDelay:    maxDelay,
		MeanDelay:   meanDelay,
		MedianDelay: medianDelay,
		StdDev:      stdDev,
	}
}

func (c *Calculator) calculateMedian(delays []time.Duration) time.Duration {
	sorted := make([]time.Duration, len(delays))
	copy(sorted, delays)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}