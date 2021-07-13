package validator

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func newDefaultConfig() schema.Configuration {
	config := schema.Configuration{}
	config.Host = loopback
	config.Port = 9090
	config.Logging.Level = "info"
	config.Logging.Format = "text"
	config.JWTSecret = testJWTSecret
	config.AuthenticationBackend.File = &schema.FileAuthenticationBackendConfiguration{
		Path: "/a/path",
	}
	config.AccessControl = schema.AccessControlConfiguration{
		DefaultPolicy: "two_factor",
	}
	config.Session = schema.SessionConfiguration{
		Domain: "example.com",
		Name:   "authelia_session",
		Secret: "secret",
	}
	config.Storage.Local = &schema.LocalStorageConfiguration{
		Path: "abc",
	}
	config.Notifier = &schema.NotifierConfiguration{
		FileSystem: &schema.FileSystemNotifierConfiguration{
			Filename: "/tmp/file",
		},
	}

	return config
}

func TestShouldNotUpdateConfig(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	ValidateConfiguration(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, 9090, config.Port)
	assert.Equal(t, "info", config.Logging.Level)
}

func TestShouldValidateAndUpdatePort(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.Port = 0

	ValidateConfiguration(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, 9091, config.Port)
}

func TestShouldValidateAndUpdateHost(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.Host = ""

	ValidateConfiguration(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, "0.0.0.0", config.Host)
}

func TestShouldEnsureNotifierConfigIsProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	ValidateConfiguration(&config, validator)
	require.Len(t, validator.Errors(), 0)

	config.Notifier = nil

	ValidateConfiguration(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "A notifier configuration must be provided")
}

func TestShouldAddDefaultAccessControl(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	config.AccessControl.DefaultPolicy = ""
	config.AccessControl.Rules = []schema.ACLRule{
		{
			Policy: "bypass",
			Domains: []string{
				"public.example.com",
			},
		},
	}

	ValidateConfiguration(&config, validator)
	require.Len(t, validator.Errors(), 0)
	assert.NotNil(t, config.AccessControl)
	assert.Equal(t, "deny", config.AccessControl.DefaultPolicy)
}

func TestShouldRaiseErrorWhenTLSCertWithoutKeyIsProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.TLSCert = testTLSCert

	ValidateConfiguration(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "No TLS key provided, please check the \"tls_key\" which has been configured")
}

func TestShouldRaiseErrorWhenTLSKeyWithoutCertIsProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.TLSKey = testTLSKey

	ValidateConfiguration(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "No TLS certificate provided, please check the \"tls_cert\" which has been configured")
}

func TestShouldNotRaiseErrorWhenBothTLSCertificateAndKeyAreProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.TLSCert = testTLSCert
	config.TLSKey = testTLSKey

	ValidateConfiguration(&config, validator)
	require.Len(t, validator.Errors(), 0)
}

func TestShouldRaiseErrorWithUndefinedJWTSecretKey(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.JWTSecret = ""

	ValidateConfiguration(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "Provide a JWT secret using \"jwt_secret\" key")
}

func TestShouldRaiseErrorWithBadDefaultRedirectionURL(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.DefaultRedirectionURL = "bad_default_redirection_url"

	ValidateConfiguration(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "Value for \"default_redirection_url\" is invalid: the url 'bad_default_redirection_url' is not absolute because it doesn't start with a scheme like 'http://' or 'https://'")
}

func TestShouldNotOverrideCertificatesDirectoryAndShouldPassWhenBlank(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	ValidateConfiguration(&config, validator)
	require.Len(t, validator.Errors(), 0)

	require.Equal(t, "", config.CertificatesDirectory)
}

func TestShouldRaiseErrorOnInvalidCertificatesDirectory(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.CertificatesDirectory = "not-a-real-file.go"

	ValidateConfiguration(&config, validator)

	require.Len(t, validator.Errors(), 1)

	if runtime.GOOS == "windows" {
		assert.EqualError(t, validator.Errors()[0], "Error checking certificate directory: CreateFile not-a-real-file.go: The system cannot find the file specified.")
	} else {
		assert.EqualError(t, validator.Errors()[0], "Error checking certificate directory: stat not-a-real-file.go: no such file or directory")
	}

	validator = schema.NewStructValidator()
	config.CertificatesDirectory = "const.go"
	ValidateConfiguration(&config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "The path const.go specified for certificate_directory is not a directory")
}

func TestShouldNotRaiseErrorOnValidCertificatesDirectory(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.CertificatesDirectory = "../../suites/common/ssl"

	ValidateConfiguration(&config, validator)

	require.Len(t, validator.Errors(), 0)
}
