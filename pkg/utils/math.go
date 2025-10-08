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
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := percentile * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func (mu *MathUtils) GenerateRandomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func (mu *MathUtils) GenerateRandomInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

func (mu *MathUtils) Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (mu *MathUtils) Lerp(start, end, t float64) float64 {
	return start + (end-start)*t
}

func (mu *MathUtils) Normalize(value, min, max float64) float64 {
	return (value - min) / (max - min)
}

func (mu *MathUtils) CalculateSlope(x1, y1, x2, y2 float64) float64 {
	if x2 == x1 {
		return math.Inf(1)
	}
	return (y2 - y1) / (x2 - x1)
}

func (mu *MathUtils) CalculateIntercept(x, y, slope float64) float64 {
	return y - slope*x
}

func (mu *MathUtils) Round(value float64, precision int) float64 {
	scale := math.Pow(10, float64(precision))
	return math.Round(value*scale) / scale
}

func (mu *MathUtils) CalculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	mean := mu.CalculateMean(values)
	variance := 0.0

	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}

	return variance / float64(len(values))
}

func (mu *MathUtils) CalculateCovariance(x, y []float64) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0
	}

	xMean := mu.CalculateMean(x)
	yMean := mu.CalculateMean(y)
	covariance := 0.0

	for i := 0; i < len(x); i++ {
		covariance += (x[i] - xMean) * (y[i] - yMean)
	}

	return covariance / float64(len(x))
}

func (mu *MathUtils) CalculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0
	}

	covariance := mu.CalculateCovariance(x, y)
	xStdDev := mu.CalculateStandardDeviation(x)
	yStdDev := mu.CalculateStandardDeviation(y)

	if xStdDev == 0 || yStdDev == 0 {
		return 0
	}

	return covariance / (xStdDev * yStdDev)
}
