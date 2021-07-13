package configuration

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
	"github.com/authelia/authelia/internal/utils"
)

// koanfEnvironmentCallback returns a koanf callback to map the environment vars to Configuration keys.
func koanfEnvironmentCallback(keyMap map[string]string, ignoredKeys []string, prefix, delimiter string) func(key, value string) (finalKey string, finalValue interface{}) {
	return func(key, value string) (finalKey string, finalValue interface{}) {
		if k, ok := keyMap[key]; ok {
			return k, value
		}

		if utils.IsStringInSlice(key, ignoredKeys) {
			return "", nil
		}

		formattedKey := strings.TrimPrefix(key, prefix)
		formattedKey = strings.ReplaceAll(strings.ToLower(formattedKey), delimiter, constDelimiter)

		if utils.IsStringInSlice(formattedKey, validator.ValidKeys) {
			return formattedKey, value
		}

		return key, value
	}
}

// koanfEnvironmentSecretsCallback returns a koanf callback to map the environment vars to Configuration keys.
func koanfEnvironmentSecretsCallback(keyMap map[string]string, validator *schema.StructValidator) func(key, value string) (finalKey string, finalValue interface{}) {
	return func(key, value string) (finalKey string, finalValue interface{}) {
		k, ok := keyMap[key]
		if !ok {
			return "", nil
		}

		v, err := loadSecret(value)
		if err != nil {
			validator.Push(fmt.Errorf(errFmtSecretIOIssue, value, k, err))
			return k, ""
		}

		return k, v
	}
}
