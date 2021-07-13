package schema

// DuoAPIConfiguration represents the configuration related to Duo API.
type DuoAPIConfiguration struct {
	Hostname       string `koanf:"hostname"`
	IntegrationKey string `koanf:"integration_key"`
	SecretKey      string `koanf:"secret_key"`
}
