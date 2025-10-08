package relativistic

import (
	"math"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type Transformations struct{}

func NewTransformations() *Transformations {
	return &Transformations{}
}

func (t *Transformations) LorentzTransformation(x, t_coord, v float64) (float64, float64) {
	if v == 0 {
		return x, t_coord
	}

	beta := v / types.SpeedOfLight
	gamma := 1.0 / math.Sqrt(1.0-beta*beta)

	x_prime := gamma * (x - v*t_coord)
	t_prime := gamma * (t_coord - (v*x)/(types.SpeedOfLight*types.SpeedOfLight))

	return x_prime, t_prime
}

func (t *Transformations) InverseLorentzTransformation(x_prime, t_prime, v float64) (float64, float64) {
	if v == 0 {
		return x_prime, t_prime
	}

	beta := v / types.SpeedOfLight
	gamma := 1.0 / math.Sqrt(1.0-beta*beta)

	x := gamma * (x_prime + v*t_prime)
	t_coord := gamma * (t_prime + (v*x_prime)/(types.SpeedOfLight*types.SpeedOfLight))

	return x, t_coord
}

func (t *Transformations) VelocityTransformation(u, v float64) float64 {
	return (u + v) / (1.0 + (u*v)/(types.SpeedOfLight*types.SpeedOfLight))
}

func (t *Transformations) TimeDilationTransformation(properTime, velocity float64) float64 {
	if velocity == 0 {
		return properTime
	}

	beta := velocity / types.SpeedOfLight
	gamma := 1.0 / math.Sqrt(1.0-beta*beta)
	return properTime * gamma
}

func (t *Transformations) LengthContractionTransformation(properLength, velocity float64) float64 {
	if velocity == 0 {
		return properLength
	}

	beta := velocity / types.SpeedOfLight
	gamma := 1.0 / math.Sqrt(1.0-beta*beta)
	return properLength / gamma
}

func (t *Transformations) DopplerTransformation(frequency, relativeVelocity float64, approach bool) float64 {
	beta := relativeVelocity / types.SpeedOfLight

	if approach {
		return frequency * math.Sqrt((1.0+beta)/(1.0-beta))
	} else {
		return frequency * math.Sqrt((1.0-beta)/(1.0+beta))
	}
}

func (t *Transformations) EnergyTransformation(restEnergy, velocity float64) float64 {
	if velocity == 0 {
		return restEnergy
	}

	beta := velocity / types.SpeedOfLight
	gamma := 1.0 / math.Sqrt(1.0-beta*beta)
	return restEnergy * gamma
}

func (t *Transformations) MomentumTransformation(mass, velocity float64) float64 {
	if velocity == 0 {
		return 0
	}

	beta := velocity / types.SpeedOfLight
	gamma := 1.0 / math.Sqrt(1.0-beta*beta)
	return mass * velocity * gamma
}

func (t *Transformations) FourVectorTransformation(x, y, z, t_coord, vx, vy, vz float64) (float64, float64, float64, float64) {
	v := math.Sqrt(vx*vx + vy*vy + vz*vz)
	if v == 0 {
		return x, y, z, t_coord
	}

	beta := v / types.SpeedOfLight
	gamma := 1.0 / math.Sqrt(1.0-beta*beta)

	dotProduct := x*vx + y*vy + z*vz
	factor := (gamma - 1.0) * dotProduct / (v * v)

	x_prime := x + factor*vx - gamma*vx*t_coord
	y_prime := y + factor*vy - gamma*vy*t_coord
	z_prime := z + factor*vz - gamma*vz*t_coord
	t_prime := gamma * (t_coord - dotProduct/(types.SpeedOfLight*types.SpeedOfLight))

	return x_prime, y_prime, z_prime, t_prime
}
