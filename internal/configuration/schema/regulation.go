package schema

// RegulationConfiguration represents the configuration related to regulation.
type RegulationConfiguration struct {
	MaxRetries int    `koanf:"max_retries"`
	FindTime   string `koanf:"find_time"`
	BanTime    string `koanf:"ban_time"`
}

// DefaultRegulationConfiguration represents default configuration parameters for the regulator.
var DefaultRegulationConfiguration = RegulationConfiguration{
	MaxRetries: 3,
	FindTime:   "2m",
	BanTime:    "5m",
}
