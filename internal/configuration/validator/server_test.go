package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestShouldSetDefaultServerValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{}

	ValidateServer(config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)

	assert.Equal(t, schema.DefaultServerConfiguration.Host, config.Server.Host)
	assert.Equal(t, schema.DefaultServerConfiguration.Port, config.Server.Port)
	assert.Equal(t, schema.DefaultServerConfiguration.ReadBufferSize, config.Server.ReadBufferSize)
	assert.Equal(t, schema.DefaultServerConfiguration.WriteBufferSize, config.Server.WriteBufferSize)
	assert.Equal(t, schema.DefaultServerConfiguration.TLSKey, config.Server.TLSKey)
	assert.Equal(t, schema.DefaultServerConfiguration.TLSCert, config.Server.TLSCert)
	assert.Equal(t, schema.DefaultServerConfiguration.Path, config.Server.Path)
	assert.Equal(t, schema.DefaultServerConfiguration.EnableExpvars, config.Server.EnableExpvars)
	assert.Equal(t, schema.DefaultServerConfiguration.EnablePprof, config.Server.EnablePprof)
}

// TODO: DEPRECATED TEST. Remove in 4.33.0.
func TestShouldNotOverrideNewValuesWithDeprecatedValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{Host: "123.0.0.1", Port: 9101, TLSKey: "/tmp/key.pem", TLSCert: "/tmp/cert.pem"}
	config.Server.Host = "192.168.0.2"
	config.Server.Port = 80
	config.Server.TLSKey = "/tmp/new/key.pem"
	config.Server.TLSCert = "/tmp/new/cert.pem"

	ValidateServer(config, validator)

	require.Len(t, validator.Errors(), 0)
	require.Len(t, validator.Warnings(), 4)

	assert.EqualError(t, validator.Warnings()[0], fmt.Sprintf(errFmtDeprecatedConfigurationKey, "host", "4.33.0", "server.host"))
	assert.EqualError(t, validator.Warnings()[1], fmt.Sprintf(errFmtDeprecatedConfigurationKey, "port", "4.33.0", "server.port"))
	assert.EqualError(t, validator.Warnings()[2], fmt.Sprintf(errFmtDeprecatedConfigurationKey, "tls_cert", "4.33.0", "server.tls_cert"))
	assert.EqualError(t, validator.Warnings()[3], fmt.Sprintf(errFmtDeprecatedConfigurationKey, "tls_key", "4.33.0", "server.tls_key"))

	assert.Equal(t, "192.168.0.2", config.Server.Host)
	assert.Equal(t, 80, config.Server.Port)
	assert.Equal(t, "/tmp/new/key.pem", config.Server.TLSKey)
	assert.Equal(t, "/tmp/new/cert.pem", config.Server.TLSCert)
}

// TODO: DEPRECATED TEST. Remove in 4.33.0.
func TestShouldSetDeprecatedValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{}

	config.Host = "192.168.0.1"
	config.Port = 80
	config.TLSCert = "/tmp/cert.pem"
	config.TLSKey = "/tmp/key.pem"

	ValidateServer(config, validator)

	assert.Len(t, validator.Errors(), 0)
	require.Len(t, validator.Warnings(), 4)

	assert.Equal(t, "192.168.0.1", config.Server.Host)
	assert.Equal(t, 80, config.Server.Port)
	assert.Equal(t, "/tmp/cert.pem", config.Server.TLSCert)
	assert.Equal(t, "/tmp/key.pem", config.Server.TLSKey)

	assert.EqualError(t, validator.Warnings()[0], fmt.Sprintf(errFmtDeprecatedConfigurationKey, "host", "4.33.0", "server.host"))
	assert.EqualError(t, validator.Warnings()[1], fmt.Sprintf(errFmtDeprecatedConfigurationKey, "port", "4.33.0", "server.port"))
	assert.EqualError(t, validator.Warnings()[2], fmt.Sprintf(errFmtDeprecatedConfigurationKey, "tls_cert", "4.33.0", "server.tls_cert"))
	assert.EqualError(t, validator.Warnings()[3], fmt.Sprintf(errFmtDeprecatedConfigurationKey, "tls_key", "4.33.0", "server.tls_key"))
}

func TestShouldSetDefaultConfig(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{}

	ValidateServer(config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)

	assert.Equal(t, schema.DefaultServerConfiguration.ReadBufferSize, config.Server.ReadBufferSize)
	assert.Equal(t, schema.DefaultServerConfiguration.WriteBufferSize, config.Server.WriteBufferSize)
}

func TestShouldParsePathCorrectly(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Server: schema.ServerConfiguration{
			Path: "apple",
		},
	}

	ValidateServer(config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)

	assert.Equal(t, "/apple", config.Server.Path)
}

func TestShouldRaiseOnNegativeValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Server: schema.ServerConfiguration{
			ReadBufferSize:  -1,
			WriteBufferSize: -1,
		},
	}

	ValidateServer(config, validator)

	require.Len(t, validator.Errors(), 2)

	assert.EqualError(t, validator.Errors()[0], "server read buffer size must be above 0")
	assert.EqualError(t, validator.Errors()[1], "server write buffer size must be above 0")
}

func TestShouldRaiseOnNonAlphanumericCharsInPath(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Server: schema.ServerConfiguration{
			Path: "app le",
		},
	}

	ValidateServer(config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.Error(t, validator.Errors()[0], "server path must only be alpha numeric characters")
}

func TestShouldRaiseOnForwardSlashInPath(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.Configuration{
		Server: schema.ServerConfiguration{
			Path: "app/le",
		},
	}

	ValidateServer(config, validator)

	assert.Len(t, validator.Errors(), 1)

	assert.Error(t, validator.Errors()[0], "server path must not contain any forward slashes")
}
