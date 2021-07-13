package commands

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/configuration"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
	"github.com/authelia/authelia/internal/logging"
)

// cmdWithConfigFlags is used for commands which require access to the configuration to add the flag to the command.
func cmdWithConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceP("config", "c", []string{}, "Configuration files")
	cmd.Flags().String("env.prefix", configuration.DefaultEnvPrefix, "Sets the env prefix for configuration")
}

var config *schema.Configuration

func newCmdWithConfigPreRun(ensureConfigExists, validateKeys, validateConfiguration bool) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, _ []string) {
		logger := logging.Logger()

		prefix, _ := cmd.Flags().GetString("env.prefix")

		configs, err := cmd.Root().Flags().GetStringSlice("config")
		if err != nil {
			logger.Fatalf("Error reading flags: %v", err)
		}

		if ensureConfigExists && len(configs) == 1 {
			created, err := configuration.EnsureConfigurationExists(configs[0])
			if err != nil {
				logger.Fatal(err)
			}

			if created {
				logger.Warnf("Configuration did not exist so a default one has been generated at %s, you will need to configure this", configs[0])
				os.Exit(0)
			}
		}

		var keys []string

		val := schema.NewStructValidator()

		keys, config, err = configuration.Load(val, configuration.NewDefaultSources(configs, prefix, configuration.DefaultEnvDelimiter)...)
		if err != nil {
			logger.Fatalf("Error occurred loading configuration: %v", err)
		}

		if validateKeys {
			validator.ValidateKeys(keys, prefix, val)
		}

		if validateConfiguration {
			validator.ValidateConfiguration(config, val)
		}

		warnings := val.Warnings()
		if len(warnings) != 0 {
			for _, warning := range warnings {
				logger.Warnf("Configuration: %+v", warning)
			}
		}

		errs := val.Errors()
		if len(errs) != 0 {
			for _, err := range errs {
				logger.Errorf("Configuration: %+v", err)
			}

			logger.Fatalf("Can't continue due to the errors loading the configuration")
		}
	}
}
