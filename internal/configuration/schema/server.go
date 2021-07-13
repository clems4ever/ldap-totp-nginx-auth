package schema

// ServerConfiguration represents the configuration of the http server.
type ServerConfiguration struct {
	Host            string `koanf:"host"`
	Port            int    `koanf:"port"`
	TLSCert         string `koanf:"tls_cert"`
	TLSKey          string `koanf:"tls_key"`
	Path            string `koanf:"path"`
	ReadBufferSize  int    `koanf:"read_buffer_size"`
	WriteBufferSize int    `koanf:"write_buffer_size"`
	EnablePprof     bool   `koanf:"enable_endpoint_pprof"`
	EnableExpvars   bool   `koanf:"enable_endpoint_expvars"`
}

// DefaultServerConfiguration represents the default values of the ServerConfiguration.
var DefaultServerConfiguration = ServerConfiguration{
	Host:            "0.0.0.0",
	Port:            9091,
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}
