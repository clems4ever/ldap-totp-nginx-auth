package configuration

import (
	_ "embed" // Embed config.template.yml.
	"fmt"
	"io/ioutil"
	"os"
)

//go:embed config.template.yml
var configTemplate []byte

// Generate creates a config at the given path based on the config template.
func Generate(configPath string) error {
	err := ioutil.WriteFile(configPath, configTemplate, 0600)
	if err != nil {
		return fmt.Errorf("Unable to generate %v: %v", configPath, err)
	}

	return nil
}

// GenerateIfNotFound will create a config at the given path based on the config template, if the file doesn not already exist.
func GenerateIfNotFound(configPath string) (bool, error) {
	_, err := os.Stat(configPath)
	if err != nil {
		err = Generate(configPath)
		return err == nil, err
	}

	return false, nil
}
