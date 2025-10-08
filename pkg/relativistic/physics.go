package relativistic

import (
	"math"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type PhysicsEngine struct{}

func NewPhysicsEngine() *PhysicsEngine {
	return &PhysicsEngine{}
}

func (pe *PhysicsEngine) CalculateTimeDilation(properTime, velocity float64) float64 {
	if velocity == 0 {
		return properTime
	}

	lorentzFactor := pe.CalculateLorentzFactor(velocity)
	return properTime * lorentzFactor
}

func (pe *PhysicsEngine) CalculateLorentzFactor(velocity float64) float64 {
	if velocity >= types.SpeedOfLight {
		return math.Inf(1)
	}

	beta := velocity / types.SpeedOfLight
	return 1.0 / math.Sqrt(1.0-beta*beta)
}

func (pe *PhysicsEngine) CalculateLengthContraction(properLength, velocity float64) float64 {
	if velocity == 0 {
		return properLength
	}

	lorentzFactor := pe.CalculateLorentzFactor(velocity)
	return properLength / lorentzFactor
}

func (pe *PhysicsEngine) CalculateRelativisticMomentum(mass, velocity float64) float64 {
	if velocity == 0 {
		return 0
	}

	lorentzFactor := pe.CalculateLorentzFactor(velocity)
	return mass * velocity * lorentzFactor
}

func (pe *PhysicsEngine) CalculateRelativisticEnergy(mass, velocity float64) float64 {
	if velocity == 0 {
		return mass * types.SpeedOfLight * types.SpeedOfLight
	}

	lorentzFactor := pe.CalculateLorentzFactor(velocity)
	return mass * types.SpeedOfLight * types.SpeedOfLight * lorentzFactor
}

func (pe *PhysicsEngine) CalculateGravitationalTimeDilation(gravitationalPotential float64) float64 {
	if gravitationalPotential == 0 {
		return 1.0
	}

	c2 := types.SpeedOfLight * types.SpeedOfLight
	return math.Sqrt(1.0 + 2.0*gravitationalPotential/c2)
}

func (pe *PhysicsEngine) CalculateSchwarzschildRadius(mass float64) float64 {
	G := 6.67430e-11
	return 2.0 * G * mass / (types.SpeedOfLight * types.SpeedOfLight)
}

func (pe *PhysicsEngine) CalculateOrbitalVelocity(centralMass, orbitalRadius float64) float64 {
	G := 6.67430e-11
	return math.Sqrt(G * centralMass / orbitalRadius)
}

func (pe *PhysicsEngine) CalculateEscapeVelocity(mass, radius float64) float64 {
	G := 6.67430e-11
	return math.Sqrt(2.0 * G * mass / radius)
}

func (pe *PhysicsEngine) CalculateDopplerShift(frequency, relativeVelocity float64, approach bool) float64 {
	beta := relativeVelocity / types.SpeedOfLight

	if approach {
		return frequency * math.Sqrt((1.0+beta)/(1.0-beta))
	} else {
		return frequency * math.Sqrt((1.0-beta)/(1.0+beta))
	}
}
