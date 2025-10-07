package relativistic

import (
	"math"
	"time"
)

func ApplyTimeDilation(properTime time.Duration, velocity float64) time.Duration {
	if velocity <= 0 {
		return properTime
	}
	c := 299792458.0
	if velocity >= c {
		return properTime
	}
	lorentz := CalculateLorentzFactor(velocity)
	dilatedTime := float64(properTime) * lorentz
	return time.Duration(dilatedTime)
}

func CalculateLorentzFactor(velocity float64) float64 {
	if velocity <= 0 {
		return 1.0
	}
	c := 299792458.0
	if velocity >= c {
		return 1.0
	}
	beta := velocity / c
	return 1.0 / math.Sqrt(1 - beta*beta)
}

func CalculateRelativisticDelay(distance, velocity float64) time.Duration {
	if velocity <= 0 {
		return 0
	}
	properTime := distance / velocity
	lorentz := CalculateLorentzFactor(velocity)
	dilatedTime := properTime * lorentz
	return time.Duration(dilatedTime * float64(time.Second))
}