package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

func TestShouldValidateGoodKeys(t *testing.T) {
	configKeys := ValidKeys
	val := schema.NewStructValidator()
	ValidateKeys(configKeys, "AUTHELIA_", val)

	require.Len(t, val.Errors(), 0)
}

func TestShouldNotValidateBadKeys(t *testing.T) {
	configKeys := ValidKeys
	configKeys = append(configKeys, "bad_key")
	configKeys = append(configKeys, "totp.skewy")
	val := schema.NewStructValidator()
	ValidateKeys(configKeys, "AUTHELIA_", val)

	errs := val.Errors()
	require.Len(t, errs, 2)

	assert.EqualError(t, errs[0], "configuration key not expected: bad_key")
	assert.EqualError(t, errs[1], "configuration key not expected: totp.skewy")
}

func TestShouldNotValidateBadEnvKeys(t *testing.T) {
	configKeys := ValidKeys
	configKeys = append(configKeys, "AUTHELIA__BAD_ENV_KEY")
	configKeys = append(configKeys, "AUTHELIA_BAD_ENV_KEY")
	val := schema.NewStructValidator()
	ValidateKeys(configKeys, "AUTHELIA_", val)

	errs := val.Errors()
	require.Len(t, errs, 2)

	assert.EqualError(t, errs[0], "configuration environment variable not expected: AUTHELIA__BAD_ENV_KEY")
	assert.EqualError(t, errs[1], "configuration environment variable not expected: AUTHELIA_BAD_ENV_KEY")
}

func TestAllSpecificErrorKeys(t *testing.T) {
	var configKeys []string //nolint:prealloc // This is because the test is dynamic based on the keys that exist in the map.

	var uniqueValues []string

	// Setup configKeys and uniqueValues expected.
	for key, value := range specificErrorKeys {
		configKeys = append(configKeys, key)

		if !utils.IsStringInSlice(value, uniqueValues) {
			uniqueValues = append(uniqueValues, value)
		}
	}

	val := schema.NewStructValidator()
	ValidateKeys(configKeys, "AUTHELIA_", val)

	errs := val.Errors()

	// Check only unique errors are shown. Require because if we don't the next test panics.
	require.Len(t, errs, len(uniqueValues))

	// Dynamically check all specific errors.
	for i, value := range uniqueValues {
		assert.EqualError(t, errs[i], value)
	}
}

func TestSpecificErrorKeys(t *testing.T) {
	configKeys := []string{
		"notifier.smtp.trusted_cert",
		"google_analytics",
		"authentication_backend.file.password_options.algorithm",
		"authentication_backend.file.password_options.iterations", // This should not show another error since our target for the specific error is password_options.
		"authentication_backend.file.password_hashing.algorithm",
		"authentication_backend.file.hashing.algorithm",
	}

	val := schema.NewStructValidator()
	ValidateKeys(configKeys, "AUTHELIA_", val)

	errs := val.Errors()

	require.Len(t, errs, 5)

	assert.EqualError(t, errs[0], specificErrorKeys["notifier.smtp.trusted_cert"])
	assert.EqualError(t, errs[1], specificErrorKeys["google_analytics"])
	assert.EqualError(t, errs[2], specificErrorKeys["authentication_backend.file.password_options.iterations"])
	assert.EqualError(t, errs[3], specificErrorKeys["authentication_backend.file.password_hashing.algorithm"])
	assert.EqualError(t, errs[4], specificErrorKeys["authentication_backend.file.hashing.algorithm"])
}

func TestReplacedErrors(t *testing.T) {
	configKeys := []string{
		"authentication_backend.ldap.skip_verify",
		"authentication_backend.ldap.minimum_tls_version",
		"notifier.smtp.disable_verify_cert",
		"logs_file_path",
		"logs_level",
	}

	val := schema.NewStructValidator()
	ValidateKeys(configKeys, "AUTHELIA_", val)

	warns := val.Warnings()
	errs := val.Errors()

	assert.Len(t, warns, 0)
	require.Len(t, errs, 5)

	assert.EqualError(t, errs[0], fmt.Sprintf(errFmtReplacedConfigurationKey, "authentication_backend.ldap.skip_verify", "authentication_backend.ldap.tls.skip_verify"))
	assert.EqualError(t, errs[1], fmt.Sprintf(errFmtReplacedConfigurationKey, "authentication_backend.ldap.minimum_tls_version", "authentication_backend.ldap.tls.minimum_version"))
	assert.EqualError(t, errs[2], fmt.Sprintf(errFmtReplacedConfigurationKey, "notifier.smtp.disable_verify_cert", "notifier.smtp.tls.skip_verify"))
	assert.EqualError(t, errs[3], fmt.Sprintf(errFmtReplacedConfigurationKey, "logs_file_path", "log.file_path"))
	assert.EqualError(t, errs[4], fmt.Sprintf(errFmtReplacedConfigurationKey, "logs_level", "log.level"))
}
