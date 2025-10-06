package utils

import (
	"math"
	"math/rand"
	"time"
)

type MathUtils struct{}

func NewMathUtils() *MathUtils {
	rand.Seed(time.Now().UnixNano())
	return &MathUtils{}
}

func (mu *MathUtils) DegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

func (mu *MathUtils) RadiansToDegrees(radians float64) float64 {
	return radians * 180.0 / math.Pi
}

func (mu *MathUtils) CalculateDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

func (mu *MathUtils) Calculate3DDistance(x1, y1, z1, x2, y2, z2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	dz := z2 - z1
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

func (mu *MathUtils) CalculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (mu *MathUtils) CalculateStandardDeviation(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	mean := mu.CalculateMean(values)
	variance := 0.0
	
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	
	variance /= float64(len(values))
	return math.Sqrt(variance)
}

func (mu *MathUtils) CalculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sorted := make([]float64, len(values))
	copy(sorted, values)
	
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

func (mu *MathUtils) CalculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sorted := make([]float64, len(values))
	copy(sorted, values)
	
	for i := 0; i < len(sorted)-1; i++ {
		for j :=