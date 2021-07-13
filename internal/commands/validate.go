package commands

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/configuration"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
	"github.com/authelia/authelia/internal/logging"
)

func newValidateConfigCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "validate-config [yaml]",
		Short: "Check a configuration against the internal configuration validation mechanisms",
		Args:  cobra.MinimumNArgs(1),
		Run:   cmdValidateConfigRun,
	}

	return cmd
}

func cmdValidateConfigRun(_ *cobra.Command, args []string) {
	logger := logging.Logger()

	configPath := args[0]
	if _, err := os.Stat(configPath); err != nil {
		logger.Fatalf("Error Loading Configuration: %v\n", err)
	}

	val := schema.NewStructValidator()

	keys, conf, err := configuration.Load(val, configuration.NewYAMLFileSource(configPath))
	if err != nil {
		logger.Fatalf("Error occurred loading configuration: %v", err)
	}

	validator.ValidateKeys(keys, configuration.DefaultEnvPrefix, val)
	validator.ValidateConfiguration(conf, val)

	errs := val.Errors()
	warns := val.Warnings()

	if len(warns) != 0 {
		logger.Warn("Warnings occurred while loading the configuration:")

		for _, warn := range warns {
			logger.Warnf("  %+v", warn)
		}
	}

	if len(errs) != 0 {
		logger.Error("Errors occurred while loading the configuration:")

		for _, err := range errs {
			logger.Errorf("  %+v", err)
		}

		logger.Fatal("Can't continue due to errors")
	}

	log.Println("Configuration parsed successfully without errors.")
}
