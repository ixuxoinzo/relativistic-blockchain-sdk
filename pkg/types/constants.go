package types

const (
	SpeedOfLight          = 299792458.0
	EarthRadius           = 6371000.0
	NetworkFactor         = 1.5
	ConsensusSafetyFactor = 2.0
	MaxAcceptableDelay    = 5000
)

var Regions = []string{
	"us-east",
	"us-west",
	"eu-west",
	"eu-central",
	"ap-southeast",
	"ap-northeast",
	"sa-east",
	"af-south",
}

var PlanetaryDistances = map[string]float64{
	"earth-mars":    54600000,
	"earth-venus":   41400000,
	"earth-mercury": 91600000,
	"earth-jupiter": 628000000,
	"mars-venus":    120000000,
	"mars-jupiter":  550000000,
}

const (
	ItemTypeBlock       = "block"
	ItemTypeTransaction = "transaction"
)

const (
	RoleAdmin   = "admin"
	RoleUser    = "user"
	RoleNode    = "node"
	RoleMonitor = "monitor"
)

const (
	StatusHealthy   = "healthy"
	StatusDegraded  = "degraded"
	StatusUnhealthy = "unhealthy"
)
