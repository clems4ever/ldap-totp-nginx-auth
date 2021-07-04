package configuration

import (
	"errors"
	"fmt"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
)

// NewYAMLFileSource returns a Source configured to load from a specified YAML path. If there is an issue accessing this
// path it also returns an error.
func NewYAMLFileSource(path string) (source *YAMLFileSource) {
	return &YAMLFileSource{
		koanf: koanf.New("."),
		path:  path,
	}
}

// NewYAMLFileSources returns a slice of Source configured to load from specified YAML files.
func NewYAMLFileSources(paths []string) (sources []*YAMLFileSource) {
	for _, path := range paths {
		source := NewYAMLFileSource(path)

		sources = append(sources, source)
	}

	return sources
}

// Name of the Source.
func (s YAMLFileSource) Name() (name string) {
	return fmt.Sprintf("yaml file(%s)", s.path)
}

// Merge the YAMLFileSource koanf.Koanf into the provided one.
func (s *YAMLFileSource) Merge(ko *koanf.Koanf) (err error) {
	return ko.Merge(s.koanf)
}

// Load the Source into the YAMLFileSource koanf.Koanf.
func (s *YAMLFileSource) Load() (err error) {
	if s.path == "" {
		return errors.New("invalid yaml path source configuration")
	}

	return s.koanf.Load(file.Provider(s.path), yaml.Parser())
}

// Validator returns the validator.
func (s *YAMLFileSource) Validator() (validator *schema.StructValidator) {
	return nil
}

// NewEnvironmentSource returns a Source configured to load from environment variables.
func NewEnvironmentSource() (source *EnvironmentSource) {
	return &EnvironmentSource{
		koanf: koanf.New("."),
	}
}

// Name of the Source.
func (s EnvironmentSource) Name() (name string) {
	return "environment"
}

// Merge the EnvironmentSource koanf.Koanf into the provided one.
func (s *EnvironmentSource) Merge(ko *koanf.Koanf) (err error) {
	return ko.Merge(s.koanf)
}

// Load the Source into the EnvironmentSource koanf.Koanf.
func (s *EnvironmentSource) Load() (err error) {
	keyMap, ignoredKeys := getEnvConfigMap(validator.ValidKeys)

	return s.koanf.Load(env.ProviderWithValue(constEnvPrefix, constDelimiter, koanfEnvironmentCallback(keyMap, ignoredKeys)), nil)
}

// Validator returns the validator.
func (s *EnvironmentSource) Validator() (validator *schema.StructValidator) {
	return nil
}

// NewSecretsSource returns a Source configured to load from secrets.
func NewSecretsSource() (source *SecretsSource) {
	return &SecretsSource{
		koanf:     koanf.New("."),
		validator: schema.NewStructValidator(),
	}
}

// Name of the Source.
func (s SecretsSource) Name() (name string) {
	return "secrets"
}

// Merge the SecretsSource koanf.Koanf into the provided one.
func (s *SecretsSource) Merge(ko *koanf.Koanf) (err error) {
	for _, key := range s.koanf.Keys() {
		value, ok := ko.Get(key).(string)

		if ok && value != "" {
			s.validator.Push(fmt.Errorf(errFmtSecretAlreadyDefined, key))
		}
	}

	if !s.validator.HasErrors() {
		s.validator = nil
	}

	return ko.Merge(s.koanf)
}

// Load the Source into the SecretsSource koanf.Koanf.
func (s *SecretsSource) Load() (err error) {
	keyMap := getSecretConfigMap(validator.ValidKeys)

	return s.koanf.Load(env.ProviderWithValue(constEnvPrefixAlt, constDelimiter, koanfEnvironmentSecretsCallback(keyMap, s.validator)), nil)
}

// Validator returns the validator.
func (s *SecretsSource) Validator() (validator *schema.StructValidator) {
	return s.validator
}

// NewDefaultSources returns a slice of Source configured to load from specified YAML files.
func NewDefaultSources(filePaths []string) (sources []Source) {
	fileSources := NewYAMLFileSources(filePaths)
	for _, source := range fileSources {
		sources = append(sources, source)
	}

	sources = append(sources, NewEnvironmentSource())
	sources = append(sources, NewSecretsSource())

	return sources
}