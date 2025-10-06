package relativistic

import (
	"math"
	"time"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type Utils struct{}

func NewUtils() *Utils {
	return &Utils{}
}

func (u *Utils) DegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

func (u *Utils) RadiansToDegrees(radians float64) float64 {
	return radians * 180.0 / math.Pi
}

func (u *Utils) CalculateGreatCircleDistance(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := u.DegreesToRadians(lat1)
	lon1Rad := u.DegreesToRadians(lon1)
	lat2Rad := u.DegreesToRadians(lat2)
	lon2Rad := u.DegreesToRadians(lon2)

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return types.EarthRadius * c
}

func (u *Utils) CalculateBearing(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := u.DegreesToRadians(lat1)
	lon1Rad := u.DegreesToRadians(lon1)
	lat2Rad := u.DegreesToRadians(lat2)
	lon2Rad := u.DegreesToRadians(lon2)

	dLon := lon2Rad - lon1Rad

	y := math.Sin(dLon) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) -
		math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(dLon)

	bearing := math.Atan2(y, x)
	bearing = u.RadiansToDegrees(bearing)
	
	if bearing < 0 {
		bearing += 360
	}
	
	return bearing
}

func (u *Utils) CalculateMidpoint(lat1, lon1, lat2, lon2 float64) (float64, float64) {
	lat1Rad := u.DegreesToRadians(lat1)
	lon1Rad := u.DegreesToRadians(lon1)
	lat2Rad := u.DegreesToRadians(lat2)
	lon2Rad := u.DegreesToRadians(lon2)

	Bx := math.Cos(lat2Rad) * math.Cos(lon2Rad-lon1Rad)
	By := math.Cos(lat2Rad) * math.Sin(lon2Rad-lon1Rad)

	midLat := math.Atan2(
		math.Sin(lat1Rad)+math.Sin(lat2Rad),
		math.Sqrt(math.Pow(math.Cos(lat1Rad)+Bx, 2)+By*By),
	)
	
	midLon := lon1Rad + math.Atan2(By, math.Cos(lat1Rad)+Bx)

	return u.RadiansToDegrees(midLat), u.RadiansToDegrees(midLon)
}

func (u *Utils) CalculateDestination(lat, lon, bearing, distance float64) (float64, float64) {
	latRad := u.DegreesToRadians(lat)
	lonRad := u.DegreesToRadians(lon)
	bearingRad := u.DegreesToRadians(bearing)

	angularDistance := distance / types.EarthRadius

	destLat := math.Asin(
		math.Sin(latRad)*math.Cos(angularDistance) +
			math.Cos(latRad)*math.Sin(angularDistance)*math.Cos(bearingRad),
	)

	destLon := lonRad + math.Atan2(
		math.Sin(bearingRad)*math.Sin(angularDistance)*math.Cos(latRad),
		math.Cos(angularDistance)-math.Sin(latRad)*math.Sin(destLat),
	)

	return u.RadiansToDegrees(destLat), u.RadiansToDegrees(destLon)
}

func (u *Utils) CalculateSpeed(distance float64, duration time.Duration) float64 {
	return distance / duration.Seconds()
}

func (u *Utils) CalculateDuration(distance, speed float64) time.Duration {
	seconds := distance / speed
	return time.Duration(seconds * float64(time.Second))
}

func (u *Utils) NormalizeAngle(angle float64) float64 {
	angle = math.Mod(angle, 360)
	if angle < 0 {
		angle += 360
	}
	return angle
}

func (u *Utils) Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (u *Utils) Lerp(start, end, t float64) float64 {
	return start + (end-start)*t
}

func (u *Utils) CalculateConfidence(expected, actual, tolerance float64) float64 {
	diff := math.Abs(expected - actual)
	if diff <= tolerance {
		return 1.0 - (diff / tolerance)
	}
	return 0.0
}