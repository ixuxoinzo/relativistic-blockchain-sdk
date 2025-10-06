package utils

import (
	"math"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type GeoUtils struct{}

func NewGeoUtils() *GeoUtils {
	return &GeoUtils{}
}

func (gu *GeoUtils) CalculateGreatCircleDistance(pos1, pos2 types.Position) float64 {
	lat1 := gu.DegreesToRadians(pos1.Latitude)
	lon1 := gu.DegreesToRadians(pos1.Longitude)
	lat2 := gu.DegreesToRadians(pos2.Latitude)
	lon2 := gu.DegreesToRadians(pos2.Longitude)

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

func (gu *GeoUtils) CalculateBearing(pos1, pos2 types.Position) float64 {
	lat1 := gu.DegreesToRadians(pos1.Latitude)
	lon1 := gu.DegreesToRadians(pos1.Longitude)
	lat2 := gu.DegreesToRadians(pos2.Latitude)
	lon2 := gu.DegreesToRadians(pos2.Longitude)

	dLon := lon2 - lon1

	y := math.Sin(dLon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) -
		math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLon)

	bearing := math.Atan2(y, x)
	bearing = gu.RadiansToDegrees(bearing)
	
	if bearing < 0 {
		bearing += 360
	}
	
	return bearing
}

func (gu *GeoUtils) CalculateMidpoint(pos1, pos2 types.Position) types.Position {
	lat1 := gu.DegreesToRadians(pos1.Latitude)
	lon1 := gu.DegreesToRadians(pos1.Longitude)
	lat2 := gu.DegreesToRadians(pos2.Latitude)
	lon2 := gu.DegreesToRadians(pos2.Longitude)

	Bx := math.Cos(lat2) * math.Cos(lon2-lon1)
	By := math.Cos(lat2) * math.Sin(lon2-lon1)

	midLat := math.Atan2(
		math.Sin(lat1)+math.Sin(lat2),
		math.Sqrt(math.Pow(math.Cos(lat1)+Bx, 2)+By*By),
	)
	
	midLon := lon1 + math.Atan2(By, math.Cos(lat1)+Bx)

	midAltitude := (pos1.Altitude + pos2.Altitude) / 2

	return types.Position{
		Latitude:  gu.RadiansToDegrees(midLat),
		Longitude: gu.RadiansToDegrees(midLon),
		Altitude:  midAltitude,
	}
}

func (gu *GeoUtils) CalculateDestination(pos types.Position, bearing, distance float64) types.Position {
	lat := gu.DegreesToRadians(pos.Latitude)
	lon := gu.DegreesToRadians(pos.Longitude)
	bearingRad := gu.DegreesToRadians(bearing)

	angularDistance := distance / types.EarthRadius

	destLat := math.Asin(
		math.Sin(lat)*math.Cos(angularDistance) +
			math.Cos(lat)*math.Sin(angularDistance)*math.Cos(bearingRad),
	)

	destLon := lon + math.Atan2(
		math.Sin(bearingRad)*math.Sin(angularDistance)*math.Cos(lat),
		math.Cos(angularDistance)-math.Sin(lat)*math.Sin(destLat),
	)

	return types.Position{
		Latitude:  gu.RadiansToDegrees(destLat),
		Longitude: gu.RadiansToDegrees(destLon),
		Altitude:  pos.Altitude,
	}
}

func (gu *GeoUtils) IsValidPosition(pos types.Position) bool {
	if pos.Latitude < -90 || pos.Latitude > 90 {
		return false
	}
	if pos.Longitude < -180 || pos.Longitude > 180 {
		return false
	}
	if pos.Altitude < 0 {
		return false
	}
	return true
}

func (gu *GeoUtils) NormalizePosition(pos types.Position) types.Position {
	lat := pos.Latitude
	lon := pos.Longitude
	alt := pos.Altitude

	lat = gu.NormalizeLatitude(lat)
	lon = gu.NormalizeLongitude(lon)
	alt = math.Max(0, alt)

	return types.Position{
		Latitude:  lat,
		Longitude: lon,
		Altitude:  alt,
	}
}

func (gu *GeoUtils) NormalizeLatitude(lat float64) float64 {
	for lat < -90 {
		lat += 180
	}
	for lat > 90 {
		lat -= 180
	}
	return lat
}

func (gu *GeoUtils) NormalizeLongitude(lon float64) float64 {
	for lon < -180 {
		lon += 360
	}
	for lon > 180 {
		lon -= 360
	}
	return lon
}

func (gu *GeoUtils) CalculateBounds(positions []types.Position) (types.Position, types.Position) {
	if len(positions) == 0 {
		return types.Position{}, types.Position{}
	}

	minLat := positions[0].Latitude
	maxLat := positions[0].Latitude
	minLon := positions[0].Longitude
	maxLon := positions[0].Longitude
	minAlt := positions[0].Altitude
	maxAlt := positions[0].Altitude

	for _, pos := range positions {
		minLat = math.Min(minLat, pos.Latitude)
		maxLat = math.Max(maxLat, pos.Latitude)
		minLon = math.Min(minLon, pos.Longitude)
		maxLon = math.Max(maxLon, pos.Longitude)
		minAlt = math.Min(minAlt, pos.Altitude)
		maxAlt = math.Max(maxAlt, pos.Altitude)
	}

	return types.Position{
			Latitude:  minLat,
			Longitude: minLon,
			Altitude:  minAlt,
		}, types.Position{
			Latitude:  maxLat,
			Longitude: maxLon,
			Altitude:  maxAlt,
		}
}

func (gu *GeoUtils) CalculateCentroid(positions []types.Position) types.Position {
	if len(positions) == 0 {
		return types.Position{}
	}

	sumLat := 0.0
	sumLon := 0.0
	sumAlt := 0.0

	for _, pos := range positions {
		sumLat += pos.Latitude
		sumLon += pos.Longitude
		sumAlt += pos.Altitude
	}

	count := float64(len(positions))
	return types.Position{
		Latitude:  sumLat / count,
		Longitude: sumLon / count,
		Altitude:  sumAlt / count,
	}
}

func (gu *GeoUtils) DegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

func (gu *GeoUtils) RadiansToDegrees(radians float64) float64 {
	return radians * 180.0 / math.Pi
}

func (gu *GeoUtils) CalculateRegion(positions []types.Position) string {
	if len(positions) == 0 {
		return "unknown"
	}

	centroid := gu.CalculateCentroid(positions)
	
	regions := map[string]struct {
		minLat, maxLat float64
		minLon, maxLon float64
	}{
		"us-east":      {24.0, 50.0, -87.0, -67.0},
		"us-west":      {32.0, 49.0, -125.0, -114.0},
		"eu-west":      {35.0, 60.0, -11.0, 10.0},
		"eu-central":   {45.0, 55.0, 5.0, 15.0},
		"ap-southeast": {-10.0, 10.0, 95.0, 115.0},
		"ap-northeast": {20.0, 45.0, 120.0, 150.0},
	}

	for region, bounds := range regions {
		if centroid.Latitude >= bounds.minLat && centroid.Latitude <= bounds.maxLat &&
			centroid.Longitude >= bounds.minLon && centroid.Longitude <= bounds.maxLon {
			return region
		}
	}

	return "global"
}