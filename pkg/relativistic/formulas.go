package relativistic

import (
	"math"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type Formulas struct{}

func NewFormulas() *Formulas {
	return &Formulas{}
}

func (f *Formulas) TimeDilation(properTime, velocity float64) float64 {
	if velocity == 0 {
		return properTime
	}
	
	beta := velocity / types.SpeedOfLight
	gamma := 1.0 / math.Sqrt(1.0 - beta*beta)
	return properTime * gamma
}

func (f *Formulas) LengthContraction(properLength, velocity float64) float64 {
	if velocity == 0 {
		return properLength
	}
	
	beta := velocity / types.SpeedOfLight
	gamma := 1.0 / math.Sqrt(1.0 - beta*beta)
	return properLength / gamma
}

func (f *Formulas) VelocityAddition(u, v float64) float64 {
	return (u + v) / (1.0 + (u*v)/(types.SpeedOfLight*types.SpeedOfLight))
}

func (f *Formulas) RelativisticDoppler(frequency, relativeVelocity float64, approach bool) float64 {
	beta := relativeVelocity / types.SpeedOfLight
	
	if approach {
		return frequency * math.Sqrt((1.0+beta)/(1.0-beta))
	} else {
		return frequency * math.Sqrt((1.0-beta)/(1.0+beta))
	}
}

func (f *Formulas) GravitationalRedshift(frequency, gravitationalPotential float64) float64 {
	c2 := types.SpeedOfLight * types.SpeedOfLight
	return frequency * math.Sqrt(1.0 + 2.0*gravitationalPotential/c2)
}

func (f *Formulas) SchwarzschildMetric(time, radius, mass float64) float64 {
	G := 6.67430e-11
	rs := 2.0 * G * mass / (types.SpeedOfLight * types.SpeedOfLight)
	
	if radius <= rs {
		return 0
	}
	
	return time * math.Sqrt(1.0 - rs/radius)
}

func (f *Formulas) ProperTime(coordinateTime, velocity, gravitationalPotential float64) float64 {
	timeDilation := f.TimeDilation(1.0, velocity)
	gravitationalEffect := f.GravitationalTimeDilation(gravitationalPotential)
	
	return coordinateTime / (timeDilation * gravitationalEffect)
}

func (f *Formulas) GravitationalTimeDilation(gravitationalPotential float64) float64 {
	if gravitationalPotential == 0 {
		return 1.0
	}
	
	c2 := types.SpeedOfLight * types.SpeedOfLight
	return math.Sqrt(1.0 + 2.0*gravitationalPotential/c2)
}

func (f *Formulas) CoordinateDistance(properDistance, velocity float64) float64 {
	return f.LengthContraction(properDistance, velocity)
}

func (f *Formulas) RelativisticMomentum(mass, velocity float64) float64 {
	if velocity == 0 {
		return 0
	}
	
	beta := velocity / types.SpeedOfLight
	gamma := 1.0 / math.Sqrt(1.0 - beta*beta)
	return mass * velocity * gamma
}

func (f *Formulas) RelativisticEnergy(mass, velocity float64) float64 {
	if velocity == 0 {
		return mass * types.SpeedOfLight * types.SpeedOfLight
	}
	
	beta := velocity / types.SpeedOfLight
	gamma := 1.0 / math.Sqrt(1.0 - beta*beta)
	return mass * types.SpeedOfLight * types.SpeedOfLight * gamma
}