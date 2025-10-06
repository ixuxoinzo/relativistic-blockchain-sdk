package relativistic

const (
	SpeedOfLight          = 299792458.0
	GravitationalConstant = 6.67430e-11
	PlanckConstant        = 6.62607015e-34
	BoltzmannConstant     = 1.380649e-23
	ElectronMass          = 9.10938356e-31
	ProtonMass            = 1.6726219e-27
	NeutronMass           = 1.674927498e-27
)

const (
	EarthMass             = 5.9722e24
	EarthRadius           = 6371000.0
	EarthGravity          = 9.80665
	SolarMass             = 1.98847e30
	SolarRadius           = 6.957e8
)

const (
	AstronomicalUnit     = 1.495978707e11
	LightYear            = 9.4607304725808e15
	Parsec               = 3.0856775814913673e16
)

const (
	PlanetEarth   = "earth"
	PlanetMars    = "mars"
	PlanetVenus   = "venus"
	PlanetMercury = "mercury"
	PlanetJupiter = "jupiter"
	PlanetSaturn  = "saturn"
	PlanetUranus  = "uranus"
	PlanetNeptune = "neptune"
)

var PlanetaryConstants = map[string]map[string]float64{
	PlanetEarth: {
		"mass":     EarthMass,
		"radius":   6371000,
		"gravity":  9.80665,
		"rotation": 86164.1,
	},
	PlanetMars: {
		"mass":     6.4171e23,
		"radius":   3389500,
		"gravity":  3.72076,
		"rotation": 88642.65,
	},
	PlanetVenus: {
		"mass":     4.8675e24,
		"radius":   6051800,
		"gravity":  8.87,
		"rotation": -20995200,
	},
	PlanetJupiter: {
		"mass":     1.8982e27,
		"radius":   69911000,
		"gravity":  24.79,
		"rotation": 35730,
	},
}

var PhysicalConstants = map[string]float64{
	"speed_of_light":          SpeedOfLight,
	"gravitational_constant":  GravitationalConstant,
	"planck_constant":         PlanckConstant,
	"boltzmann_constant":      BoltzmannConstant,
	"electron_mass":           ElectronMass,
	"proton_mass":             ProtonMass,
	"neutron_mass":            NeutronMass,
	"avogadro_number":         6.02214076e23,
	"vacuum_permittivity":     8.8541878128e-12,
	"vacuum_permeability":     1.25663706212e-6,
	"reduced_planck_constant": 1.054571817e-34,
}

var ConversionFactors = map[string]float64{
	"meters_to_light_seconds": 1.0 / SpeedOfLight,
	"au_to_meters":            AstronomicalUnit,
	"lightyear_to_meters":     LightYear,
	"parsec_to_meters":        Parsec,
	"seconds_to_years":        1.0 / 31557600,
	"years_to_seconds":        31557600,
}