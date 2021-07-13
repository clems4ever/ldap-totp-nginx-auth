package configuration

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
	"github.com/authelia/authelia/internal/logging"
)

// Read a YAML configuration and create a Configuration object out of it.
func Read(configPath string) (*schema.Configuration, []error) {
	logger := logging.Logger()

	if configPath == "" {
		return nil, []error{errors.New("No config file path provided")}
	}

	_, err := os.Stat(configPath)
	if err != nil {
		errs := []error{
			fmt.Errorf("Unable to find config file: %v", configPath),
		}

		return nil, errs
	}

	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, []error{fmt.Errorf("Failed to %v", err)}
	}

	var data interface{}

	err = yaml.Unmarshal(file, &data)
	if err != nil {
		return nil, []error{fmt.Errorf("Error malformed %v", err)}
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Dynamically load the secret env names from the SecretNames map.
	for _, secretName := range validator.SecretNames {
		_ = viper.BindEnv(validator.SecretNameToEnvName(secretName))
	}

	viper.SetConfigFile(configPath)

	_ = viper.ReadInConfig()

	var configuration schema.Configuration

	viper.Unmarshal(&configuration) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

	val := schema.NewStructValidator()
	validator.ValidateSecrets(&configuration, val, viper.GetViper())
	validator.ValidateConfiguration(&configuration, val)
	validator.ValidateKeys(val, viper.AllKeys())

	if val.HasErrors() {
		return nil, val.Errors()
	}

	if val.HasWarnings() {
		for _, warn := range val.Warnings() {
			logger.Warnf(warn.Error())
		}
	}

	return &configuration, nil
}
