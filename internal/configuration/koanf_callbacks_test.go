package configuration

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestKoanfEnvironmentCallback(t *testing.T) {
	var (
		key   string
		value interface{}
	)

	keyMap := map[string]string{
		DefaultEnvPrefix + "KEY_EXAMPLE_UNDERSCORE": "key.example_underscore",
	}
	ignoredKeys := []string{constSecretEnvLegacyPrefix + "SOME_SECRET"}

	callback := koanfEnvironmentCallback(keyMap, ignoredKeys, DefaultEnvPrefix, DefaultEnvDelimiter)

	key, value = callback(DefaultEnvPrefix+"KEY_EXAMPLE_UNDERSCORE", "value")
	assert.Equal(t, "key.example_underscore", key)
	assert.Equal(t, "value", value)

	key, value = callback(DefaultEnvPrefix+"KEY_EXAMPLE", "value")
	assert.Equal(t, DefaultEnvPrefix+"KEY_EXAMPLE", key)
	assert.Equal(t, "value", value)

	key, value = callback(DefaultEnvPrefix+"THEME", "value")
	assert.Equal(t, "theme", key)
	assert.Equal(t, "value", value)

	key, value = callback(constSecretEnvLegacyPrefix+"SOME_SECRET", "value")
	assert.Equal(t, "", key)
	assert.Nil(t, value)
}

func TestKoanfSecretCallbackWithValidSecrets(t *testing.T) {
	var (
		key   string
		value interface{}
	)

	keyMap := map[string]string{
		"AUTHELIA__JWT_SECRET":                  "jwt_secret",
		"AUTHELIA_JWT_SECRET":                   "jwt_secret",
		"AUTHELIA_FAKE_KEY":                     "fake_key",
		"AUTHELIA__FAKE_KEY":                    "fake_key",
		"AUTHELIA_STORAGE_MYSQL_FAKE_PASSWORD":  "storage.mysql.fake_password",
		"AUTHELIA__STORAGE_MYSQL_FAKE_PASSWORD": "storage.mysql.fake_password",
	}

	dir, err := ioutil.TempDir("", "authelia-test-callbacks")
	assert.NoError(t, err)

	secretOne := filepath.Join(dir, "secert_one")
	secretTwo := filepath.Join(dir, "secret_two")

	assert.NoError(t, testCreateFile(secretOne, "value one", 0600))
	assert.NoError(t, testCreateFile(secretTwo, "value two", 0600))

	val := schema.NewStructValidator()

	callback := koanfEnvironmentSecretsCallback(keyMap, val)

	key, value = callback("AUTHELIA_FAKE_KEY", secretOne)
	assert.Equal(t, "fake_key", key)
	assert.Equal(t, "value one", value)

	key, value = callback("AUTHELIA__STORAGE_MYSQL_FAKE_PASSWORD", secretTwo)
	assert.Equal(t, "storage.mysql.fake_password", key)
	assert.Equal(t, "value two", value)
}

func TestKoanfSecretCallbackShouldIgnoreUndetectedSecrets(t *testing.T) {
	keyMap := map[string]string{
		"AUTHELIA__JWT_SECRET": "jwt_secret",
		"AUTHELIA_JWT_SECRET":  "jwt_secret",
	}

	val := schema.NewStructValidator()

	callback := koanfEnvironmentSecretsCallback(keyMap, val)

	key, value := callback("AUTHELIA__SESSION_DOMAIN", "/tmp/not-a-path")
	assert.Equal(t, "", key)
	assert.Nil(t, value)

	assert.Len(t, val.Errors(), 0)
	assert.Len(t, val.Warnings(), 0)
}

func TestKoanfSecretCallbackShouldErrorOnFSError(t *testing.T) {
	if runtime.GOOS == constWindows {
		t.Skip("skipping test due to being on windows")
	}

	keyMap := map[string]string{
		"AUTHELIA__THEME": "theme",
		"AUTHELIA_THEME":  "theme",
	}

	dir, err := ioutil.TempDir("", "authelia-test-callbacks")
	assert.NoError(t, err)

	secret := filepath.Join(dir, "inaccessible")

	assert.NoError(t, testCreateFile(secret, "secret", 0000))

	val := schema.NewStructValidator()

	callback := koanfEnvironmentSecretsCallback(keyMap, val)

	key, value := callback("AUTHELIA_THEME", secret)
	assert.Equal(t, "theme", key)
	assert.Equal(t, "", value)

	require.Len(t, val.Errors(), 1)
	assert.Len(t, val.Warnings(), 0)
	assert.EqualError(t, val.Errors()[0], fmt.Sprintf(errFmtSecretIOIssue, secret, "theme", fmt.Sprintf("open %s: permission denied", secret)))
}
