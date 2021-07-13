package configuration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
	"github.com/authelia/authelia/internal/utils"
)

func TestShouldErrorSecretNotExist(t *testing.T) {
	testReset()

	dir, err := ioutil.TempDir("", "authelia-test-secret-not-exist")
	assert.NoError(t, err)

	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"JWT_SECRET_FILE", filepath.Join(dir, "jwt")))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"DUO_API_SECRET_KEY_FILE", filepath.Join(dir, "duo")))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"SESSION_SECRET_FILE", filepath.Join(dir, "session")))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", filepath.Join(dir, "authentication")))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"NOTIFIER_SMTP_PASSWORD_FILE", filepath.Join(dir, "notifier")))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"SESSION_REDIS_PASSWORD_FILE", filepath.Join(dir, "redis")))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"SESSION_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD_FILE", filepath.Join(dir, "redis-sentinel")))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"STORAGE_MYSQL_PASSWORD_FILE", filepath.Join(dir, "mysql")))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"STORAGE_POSTGRES_PASSWORD_FILE", filepath.Join(dir, "postgres")))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"TLS_KEY_FILE", filepath.Join(dir, "tls")))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY_FILE", filepath.Join(dir, "oidc-key")))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"IDENTITY_PROVIDERS_OIDC_HMAC_SECRET_FILE", filepath.Join(dir, "oidc-hmac")))

	val := schema.NewStructValidator()
	_, _, err = Load(val, NewEnvironmentSource(DefaultEnvPrefix, DefaultEnvDelimiter), NewSecretsSource(DefaultEnvPrefix, DefaultEnvDelimiter))

	assert.NoError(t, err)
	assert.Len(t, val.Warnings(), 0)

	errs := val.Errors()
	require.Len(t, errs, 12)

	sort.Sort(utils.ErrSliceSortAlphabetical(errs))

	errFmt := utils.GetExpectedErrTxt("filenotfound")

	// ignore the errors before this as they are checked by the valdator.
	assert.EqualError(t, errs[0], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "authentication"), "authentication_backend.ldap.password", fmt.Sprintf(errFmt, filepath.Join(dir, "authentication"))))
	assert.EqualError(t, errs[1], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "duo"), "duo_api.secret_key", fmt.Sprintf(errFmt, filepath.Join(dir, "duo"))))
	assert.EqualError(t, errs[2], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "jwt"), "jwt_secret", fmt.Sprintf(errFmt, filepath.Join(dir, "jwt"))))
	assert.EqualError(t, errs[3], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "mysql"), "storage.mysql.password", fmt.Sprintf(errFmt, filepath.Join(dir, "mysql"))))
	assert.EqualError(t, errs[4], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "notifier"), "notifier.smtp.password", fmt.Sprintf(errFmt, filepath.Join(dir, "notifier"))))
	assert.EqualError(t, errs[5], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "oidc-hmac"), "identity_providers.oidc.hmac_secret", fmt.Sprintf(errFmt, filepath.Join(dir, "oidc-hmac"))))
	assert.EqualError(t, errs[6], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "oidc-key"), "identity_providers.oidc.issuer_private_key", fmt.Sprintf(errFmt, filepath.Join(dir, "oidc-key"))))
	assert.EqualError(t, errs[7], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "postgres"), "storage.postgres.password", fmt.Sprintf(errFmt, filepath.Join(dir, "postgres"))))
	assert.EqualError(t, errs[8], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "redis"), "session.redis.password", fmt.Sprintf(errFmt, filepath.Join(dir, "redis"))))
	assert.EqualError(t, errs[9], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "redis-sentinel"), "session.redis.high_availability.sentinel_password", fmt.Sprintf(errFmt, filepath.Join(dir, "redis-sentinel"))))
	assert.EqualError(t, errs[10], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "session"), "session.secret", fmt.Sprintf(errFmt, filepath.Join(dir, "session"))))
	assert.EqualError(t, errs[11], fmt.Sprintf(errFmtSecretIOIssue, filepath.Join(dir, "tls"), "tls_key", fmt.Sprintf(errFmt, filepath.Join(dir, "tls"))))
}

func TestLoadShouldReturnErrWithoutValidator(t *testing.T) {
	_, _, err := Load(nil, NewEnvironmentSource(DefaultEnvPrefix, DefaultEnvDelimiter))
	assert.EqualError(t, err, "no validator provided")
}

func TestLoadShouldReturnErrWithoutSources(t *testing.T) {
	_, _, err := Load(schema.NewStructValidator())
	assert.EqualError(t, err, "no sources provided")
}

func TestShouldHaveNotifier(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)
	assert.NotNil(t, config.Notifier)
}

func TestShouldValidateConfigurationWithEnv(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	val := schema.NewStructValidator()
	_, _, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)
}

func TestShouldNotIgnoreInvalidEnvs(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"SESSION_SECRET", "an env session secret"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"STORAGE_MYSQL_PASSWORD", "an env storage mysql password"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"STORAGE_MYSQL", "a bad env"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"JWT_SECRET", "an env jwt secret"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password"))
	assert.NoError(t, os.Setenv(constSecretEnvLegacyPrefix+"AUTHENTICATION_BACKEND_LDAP_URL", "an env authentication backend ldap password"))

	val := schema.NewStructValidator()
	keys, _, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	assert.Len(t, val.Warnings(), 0)
	require.Len(t, val.Errors(), 1)

	assert.EqualError(t, val.Errors()[0], fmt.Sprintf("configuration environment variable not expected: %sSTORAGE_MYSQL", DefaultEnvPrefix))
}

func TestShouldIgnoreSingleUnderscoreNonSecretEnvs(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"SESSION_SECRET", "an env session secret"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"STORAGE_MYSQL_PASSWORD", "an env storage mysql password"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"JWT_SECRET", "an env jwt secret"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password"))
	assert.NoError(t, os.Setenv(constSecretEnvLegacyPrefix+"AUTHENTICATION_BACKEND_LDAP_URL", "an env authentication backend ldap password"))

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "an env jwt secret", config.JWTSecret)
	assert.Equal(t, "an env session secret", config.Session.Secret)
	assert.Equal(t, "an env storage mysql password", config.Storage.MySQL.Password)
	assert.Equal(t, "an env authentication backend ldap password", config.AuthenticationBackend.LDAP.Password)
	assert.Equal(t, "ldap://127.0.0.1", config.AuthenticationBackend.LDAP.URL)
}

func TestShouldAllowBothLegacyEnvSecretFilesAndNewOnes(t *testing.T) {
	testReset()

	dir, err := ioutil.TempDir("", "authelia-test-secrets")
	assert.NoError(t, err)

	sessionSecret := filepath.Join(dir, "session")
	jwtSecret := filepath.Join(dir, "jwt")
	ldapSecret := filepath.Join(dir, "ldap")
	storageSecret := filepath.Join(dir, "storage")

	assert.NoError(t, testCreateFile(sessionSecret, "a secret session.secret value", 0600))
	assert.NoError(t, testCreateFile(jwtSecret, "a secret jwt_secret value", 0600))
	assert.NoError(t, testCreateFile(ldapSecret, "a secret authentication_backend.ldap.password value", 0600))
	assert.NoError(t, testCreateFile(storageSecret, "a secret storage.mysql.password value", 0600))

	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"SESSION_SECRET_FILE", sessionSecret))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"STORAGE_MYSQL_PASSWORD_FILE", storageSecret))
	assert.NoError(t, os.Setenv(constSecretEnvLegacyPrefix+"JWT_SECRET_FILE", jwtSecret))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", ldapSecret))

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "a secret jwt_secret value", config.JWTSecret)
	assert.Equal(t, "a secret session.secret value", config.Session.Secret)
	assert.Equal(t, "a secret storage.mysql.password value", config.Storage.MySQL.Password)
	assert.Equal(t, "a secret authentication_backend.ldap.password value", config.AuthenticationBackend.LDAP.Password)
}

func TestShouldValidateAndRaiseErrorsOnNormalConfigurationAndSecret(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"SESSION_SECRET", "an env session secret"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"SESSION_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"STORAGE_MYSQL_PASSWORD", "an env storage mysql password"))
	assert.NoError(t, os.Setenv(constSecretEnvLegacyPrefix+"JWT_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "an env authentication backend ldap password"))

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)

	assert.EqualError(t, val.Errors()[0], "secrets: error loading secret into key 'session.secret': it's already defined in other configuration sources")

	assert.Equal(t, "example_secret value", config.JWTSecret)
	assert.Equal(t, "example_secret value", config.Session.Secret)
	assert.Equal(t, "an env storage mysql password", config.Storage.MySQL.Password)
	assert.Equal(t, "an env authentication backend ldap password", config.AuthenticationBackend.LDAP.Password)
}

func TestShouldRaiseIOErrOnUnreadableFile(t *testing.T) {
	if runtime.GOOS == constWindows {
		t.Skip("skipping test due to being on windows")
	}

	testReset()

	dir, err := ioutil.TempDir("", "authelia-conf")
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile(filepath.Join(dir, "myconf.yml"), []byte("server:\n  port: 9091\n"), 0000))

	cfg := filepath.Join(dir, "myconf.yml")

	val := schema.NewStructValidator()
	_, _, err = Load(val, NewYAMLFileSource(cfg))

	assert.NoError(t, err)
	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)
	assert.EqualError(t, val.Errors()[0], fmt.Sprintf("failed to load configuration from yaml file(%s) source: open %s: permission denied", cfg, cfg))
}

func TestShouldValidateConfigurationWithEnvSecrets(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"SESSION_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"STORAGE_MYSQL_PASSWORD_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"JWT_SECRET_FILE", "./test_resources/example_secret"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE", "./test_resources/example_secret"))

	val := schema.NewStructValidator()
	_, config, err := Load(val, NewDefaultSources([]string{"./test_resources/config.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)
	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)

	assert.Equal(t, "example_secret value", config.JWTSecret)
	assert.Equal(t, "example_secret value", config.Session.Secret)
	assert.Equal(t, "example_secret value", config.AuthenticationBackend.LDAP.Password)
	assert.Equal(t, "example_secret value", config.Storage.MySQL.Password)
}

func TestShouldValidateAndRaiseErrorsOnBadConfiguration(t *testing.T) {
	testReset()

	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"SESSION_SECRET", "abc"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"STORAGE_MYSQL_PASSWORD", "abc"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"JWT_SECRET", "abc"))
	assert.NoError(t, os.Setenv(DefaultEnvPrefix+"AUTHENTICATION_BACKEND_LDAP_PASSWORD", "abc"))

	val := schema.NewStructValidator()
	keys, _, err := Load(val, NewDefaultSources([]string{"./test_resources/config_bad_keys.yml"}, DefaultEnvPrefix, DefaultEnvDelimiter)...)

	assert.NoError(t, err)

	validator.ValidateKeys(keys, DefaultEnvPrefix, val)

	require.Len(t, val.Errors(), 2)
	assert.Len(t, val.Warnings(), 0)

	assert.EqualError(t, val.Errors()[0], "configuration key not expected: loggy_file")
	assert.EqualError(t, val.Errors()[1], "invalid configuration key 'logs_level' was replaced by 'log.level'")
}

func TestShouldNotReadConfigurationOnFSAccessDenied(t *testing.T) {
	if runtime.GOOS == constWindows {
		t.Skip("skipping test due to being on windows")
	}

	testReset()

	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	cfg := filepath.Join(dir, "config.yml")
	assert.NoError(t, testCreateFile(filepath.Join(dir, "config.yml"), "port: 9091\n", 0000))

	val := schema.NewStructValidator()
	_, _, err = Load(val, NewYAMLFileSource(cfg))

	assert.NoError(t, err)
	require.Len(t, val.Errors(), 1)

	assert.EqualError(t, val.Errors()[0], fmt.Sprintf("failed to load configuration from yaml file(%s) source: open %s: permission denied", cfg, cfg))
}

func TestShouldNotLoadDirectoryConfiguration(t *testing.T) {
	testReset()

	dir, err := ioutil.TempDir("", "authelia-config")
	assert.NoError(t, err)

	val := schema.NewStructValidator()
	_, _, err = Load(val, NewYAMLFileSource(dir))

	assert.NoError(t, err)
	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)

	expectedErr := fmt.Sprintf(utils.GetExpectedErrTxt("yamlisdir"), dir)
	assert.EqualError(t, val.Errors()[0], fmt.Sprintf("failed to load configuration from yaml file(%s) source: %s", dir, expectedErr))
}

func testReset() {
	testUnsetEnvName("STORAGE_MYSQL")
	testUnsetEnvName("JWT_SECRET")
	testUnsetEnvName("DUO_API_SECRET_KEY")
	testUnsetEnvName("SESSION_SECRET")
	testUnsetEnvName("AUTHENTICATION_BACKEND_LDAP_PASSWORD")
	testUnsetEnvName("AUTHENTICATION_BACKEND_LDAP_URL")
	testUnsetEnvName("NOTIFIER_SMTP_PASSWORD")
	testUnsetEnvName("NOTIFIER_SMTP_PASSWORD")
	testUnsetEnvName("SESSION_REDIS_PASSWORD")
	testUnsetEnvName("SESSION_REDIS_PASSWORD")
	testUnsetEnvName("SESSION_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD")
	testUnsetEnvName("STORAGE_MYSQL_PASSWORD")
	testUnsetEnvName("STORAGE_POSTGRES_PASSWORD")
	testUnsetEnvName("TLS_KEY")
	testUnsetEnvName("PORT")
	testUnsetEnvName("PORT_K8S_EXAMPLE")
	testUnsetEnvName("IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY")
	testUnsetEnvName("IDENTITY_PROVIDERS_OIDC_HMAC_SECRET")
}

func testUnsetEnvName(name string) {
	_ = os.Unsetenv(DefaultEnvPrefix + name)
	_ = os.Unsetenv(constSecretEnvLegacyPrefix + name)
	_ = os.Unsetenv(DefaultEnvPrefix + name + constSecretSuffix)
	_ = os.Unsetenv(constSecretEnvLegacyPrefix + name + constSecretSuffix)
}

func testCreateFile(path, value string, perm os.FileMode) (err error) {
	return os.WriteFile(path, []byte(value), perm)
}
