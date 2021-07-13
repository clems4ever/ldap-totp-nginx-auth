package configuration

import (
	"errors"
)

// DefaultEnvPrefix is the default environment prefix.
const DefaultEnvPrefix = "AUTHELIA__"

// DefaultEnvDelimiter is the default environment delimiter.
const DefaultEnvDelimiter = "_"

const (
	constSecretEnvLegacyPrefix = "AUTHELIA_"

	constSecretSuffix = "_FILE"

	constDelimiter = "."

	constWindows = "windows"
)

const (
	errFmtSecretAlreadyDefined = "secrets: error loading secret into key '%s': it's already defined in other " +
		"configuration sources"
	errFmtSecretIOIssue         = "secrets: error loading secret path %s into key '%s': %v"
	errFmtGenerateConfiguration = "error occurred generating configuration: %+v"
)

var secretSuffixes = []string{"key", "secret", "password", "token"}
var errNoSources = errors.New("no sources provided")
var errNoValidator = errors.New("no validator provided")
