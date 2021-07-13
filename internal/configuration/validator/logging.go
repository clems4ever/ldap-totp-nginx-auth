package validator

import (
	"fmt"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateLogging validates the logging configuration.
func ValidateLogging(configuration *schema.Configuration, validator *schema.StructValidator) {
	applyDeprecatedLoggingConfiguration(configuration, validator) // TODO: DEPRECATED LINE. Remove in 4.33.0.

	if configuration.Log.Level == "" {
		configuration.Log.Level = schema.DefaultLoggingConfiguration.Level
	}

	if configuration.Log.Format == "" {
		configuration.Log.Format = schema.DefaultLoggingConfiguration.Format
	}

	if !utils.IsStringInSlice(configuration.Log.Level, validLoggingLevels) {
		validator.Push(fmt.Errorf(errFmtLoggingLevelInvalid, configuration.Log.Level, strings.Join(validLoggingLevels, ", ")))
	}
}

// TODO: DEPRECATED FUNCTION. Remove in 4.33.0.
func applyDeprecatedLoggingConfiguration(configuration *schema.Configuration, validator *schema.StructValidator) {
	if configuration.LogLevel != "" {
		validator.PushWarning(fmt.Errorf(errFmtDeprecatedConfigurationKey, "log_level", "4.33.0", "log.level"))

		if configuration.Log.Level == "" {
			configuration.Log.Level = configuration.LogLevel
		}
	}

	if configuration.LogFormat != "" {
		validator.PushWarning(fmt.Errorf(errFmtDeprecatedConfigurationKey, "log_format", "4.33.0", "log.format"))

		if configuration.Log.Format == "" {
			configuration.Log.Format = configuration.LogFormat
		}
	}

	if configuration.LogFilePath != "" {
		validator.PushWarning(fmt.Errorf(errFmtDeprecatedConfigurationKey, "log_file_path", "4.33.0", "log.file_path"))

		if configuration.Log.FilePath == "" {
			configuration.Log.FilePath = configuration.LogFilePath
		}
	}
}
