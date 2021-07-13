package schema

// TLSConfig is a representation of the TLS configuration.
type TLSConfig struct {
	MinimumVersion string `koanf:"minimum_version"`
	SkipVerify     bool   `koanf:"skip_verify"`
	ServerName     string `koanf:"server_name"`
}
