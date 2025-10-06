package relativistic

import (
	"math"
	"time"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type RelativisticCalculator struct{}

func NewCalculator() *RelativisticCalculator {
	return &RelativisticCalculator{}
}

func (rc *RelativisticCalculator) CalculateLightDelay(distance float64) time.Duration {
	delaySeconds := distance / types.SpeedOfLight
	return time.Duration(delaySeconds * float64(time.Second))
}

func (rc *RelativisticCalculator) CalculateNetworkDelay(distance float64, networkFactor float64) time.Duration {
	lightDelay := rc.CalculateLightDelay(distance)
	return time.Duration(float64(lightDelay) * networkFactor)
}

func (rc *RelativisticCalculator) CalculateRelativisticDelay(distance float64, velocity float64, networkFactor float64) time.Duration {
	lightDelay := distance / types.SpeedOfLight
	realisticDelay := lightDelay * networkFactor
	relativisticDelay := rc.ApplyTimeDilation(realisticDelay, velocity)
	return time.Duration(relativisticDelay * float64(time.Second))
}

func (rc *RelativisticCalculator) ApplyTimeDilation(delay float64, velocity float64) float64 {
	if velocity == 0 {
		return delay
	}
	
	lorentzFactor := rc.CalculateLorentzFactor(velocity)
	return delay * lorentzFactor
}

func (rc *RelativisticCalculator) CalculateLorentzFactor(velocity float64) float64 {
	if velocity >= types.SpeedOfLight {
		return math.Inf(1)
	}
	
	v2 := velocity * velocity
	c2 := types.SpeedOfLight * types.SpeedOfLight
	return 1.0 / math.Sqrt(1.0 - v2/c2)
}

func (rc *RelativisticCalculator) CalculateGravitationalTimeDilation(gravityFieldStrength, height float64) float64 {
	if gravityFieldStrength == 0 {
		return 1.0
	}

	c2 := types.SpeedOfLight * types.SpeedOfLight
	phi := gravityFieldStrength * height
	return 1.0 + phi/c2
}

func (rc *RelativisticCalculator) CalculateCombinedEffects(distance, velocity, gravityFieldStrength, height float64, networkFactor float64) time.Duration {
	lightDelay := distance / types.SpeedOfLight
	networkDelay := lightDelay * networkFactor
	
	lorentzFactor := rc.CalculateLorentzFactor(velocity)
	gravitationalFactor := rc.CalculateGravitationalTimeDilation(gravityFieldStrength, height)
	
	totalDelay := networkDelay * lorentzFactor * gravitationalFactor
	return time.Duration(totalDelay * float64(time.Second))
}

func (rc *RelativisticCalculator) CalculateSatelliteDelay(satellitePos, groundStationPos types.Position, satelliteVelocity float64) (time.Duration, error) {
	distance := rc.CalculateDistance(satellitePos, groundStationPos)
	
	lightDelay := distance / types.SpeedOfLight
	relativisticDelay := rc.ApplyTimeDilation(lightDelay, satelliteVelocity)
	
	totalDelay := relativisticDelay * 1.01

	return time.Duration(totalDelay * float64(time.Second)), nil
}

func (rc *RelativisticCalculator) CalculateDistance(pos1, pos2 types.Position) float64 {
	lat1 := pos1.Latitude * math.Pi / 180
	lon1 := pos1.Longitude * math.Pi / 180
	lat2 := pos2.Latitude * math.Pi / 180
	lon2 := pos2.Longitude * math.Pi / 180

	dLat := lat2 - lat1
	dLon := lon2 - lon1

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := types.EarthRadius * c

	altDiff := math.Abs(pos1.Altitude - pos2.Altitude)
	totalDistance := math.Sqrt(math.Pow(distance, 2) + math.Pow(altDiff, 2))

	return totalDistance
}