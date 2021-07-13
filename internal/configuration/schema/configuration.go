package schema

// Configuration object extracted from YAML configuration file.
type Configuration struct {
	Theme                 string `koanf:"theme"`
	CertificatesDirectory string `koanf:"certificates_directory"`
	JWTSecret             string `koanf:"jwt_secret"`
	DefaultRedirectionURL string `koanf:"default_redirection_url"`

	Host        string `koanf:"host"`          // Deprecated: To be Removed. TODO: Remove in 4.33.0.
	Port        int    `koanf:"port"`          // Deprecated: To be Removed. TODO: Remove in 4.33.0.
	TLSCert     string `koanf:"tls_cert"`      // Deprecated: To be Removed. TODO: Remove in 4.33.0.
	TLSKey      string `koanf:"tls_key"`       // Deprecated: To be Removed. TODO: Remove in 4.33.0.
	LogLevel    string `koanf:"log_level"`     // Deprecated: To be Removed. TODO: Remove in 4.33.0.
	LogFormat   string `koanf:"log_format"`    // Deprecated: To be Removed. TODO: Remove in 4.33.0.
	LogFilePath string `koanf:"log_file_path"` // Deprecated: To be Removed. TODO: Remove in 4.33.0.

	Log                   LogConfiguration                   `koanf:"log"`
	IdentityProviders     IdentityProvidersConfiguration     `koanf:"identity_providers"`
	AuthenticationBackend AuthenticationBackendConfiguration `koanf:"authentication_backend"`
	Session               SessionConfiguration               `koanf:"session"`
	TOTP                  *TOTPConfiguration                 `koanf:"totp"`
	DuoAPI                *DuoAPIConfiguration               `koanf:"duo_api"`
	AccessControl         AccessControlConfiguration         `koanf:"access_control"`
	Regulation            *RegulationConfiguration           `koanf:"regulation"`
	Storage               StorageConfiguration               `koanf:"storage"`
	Notifier              *NotifierConfiguration             `koanf:"notifier"`
	Server                ServerConfiguration                `koanf:"server"`
}
