package validator

import (
	"fmt"
	"os"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateConfiguration and adapt the configuration read from file.
func ValidateConfiguration(configuration *schema.Configuration, validator *schema.StructValidator) {
	if configuration.CertificatesDirectory != "" {
		info, err := os.Stat(configuration.CertificatesDirectory)
		if err != nil {
			validator.Push(fmt.Errorf("Error checking certificate directory: %v", err))
		} else if !info.IsDir() {
			validator.Push(fmt.Errorf("The path %s specified for certificate_directory is not a directory", configuration.CertificatesDirectory))
		}
	}

	if configuration.JWTSecret == "" {
		validator.Push(fmt.Errorf("Provide a JWT secret using \"jwt_secret\" key"))
	}

	if configuration.DefaultRedirectionURL != "" {
		err := utils.IsStringAbsURL(configuration.DefaultRedirectionURL)
		if err != nil {
			validator.Push(fmt.Errorf("Value for \"default_redirection_url\" is invalid: %+v", err))
		}
	}

	ValidateTheme(configuration, validator)

	if configuration.TOTP == nil {
		configuration.TOTP = &schema.DefaultTOTPConfiguration
	}

	ValidateLogging(configuration, validator)

	ValidateTOTP(configuration.TOTP, validator)

	ValidateAuthenticationBackend(&configuration.AuthenticationBackend, validator)

	ValidateAccessControl(&configuration.AccessControl, validator)

	ValidateRules(configuration.AccessControl, validator)

	ValidateSession(&configuration.Session, validator)

	if configuration.Regulation == nil {
		configuration.Regulation = &schema.DefaultRegulationConfiguration
	}

	ValidateRegulation(configuration.Regulation, validator)

	ValidateServer(configuration, validator)

	ValidateStorage(configuration.Storage, validator)

	if configuration.Notifier == nil {
		validator.Push(fmt.Errorf("A notifier configuration must be provided"))
	} else {
		ValidateNotifier(configuration.Notifier, validator)
	}

	ValidateIdentityProviders(&configuration.IdentityProviders, validator)
}
