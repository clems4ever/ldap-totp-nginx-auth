package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/commands"
	"github.com/authelia/authelia/internal/configuration"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/notification"
	"github.com/authelia/authelia/internal/oidc"
	"github.com/authelia/authelia/internal/regulation"
	"github.com/authelia/authelia/internal/server"
	"github.com/authelia/authelia/internal/session"
	"github.com/authelia/authelia/internal/storage"
	"github.com/authelia/authelia/internal/utils"
)

var configPathFlag string

//nolint:gocyclo // TODO: Consider refactoring/simplifying, time permitting.
func startServer() {
	logger := logging.Logger()
	created, err := configuration.GenerateIfNotFound(configPathFlag)
	if err != nil {
		logger.Error(err)
	}
	if created {
		logger.Infof("Generated configuration at: %v", configPathFlag)
	}
	config, errs := configuration.Read(configPathFlag)

	if len(errs) > 0 {
		for _, err := range errs {
			logger.Error(err)
		}

		os.Exit(1)
	}

	autheliaCertPool, errs, nonFatalErrs := utils.NewX509CertPool(config.CertificatesDirectory)
	if len(errs) > 0 {
		for _, err := range errs {
			logger.Error(err)
		}

		os.Exit(2)
	}

	if len(nonFatalErrs) > 0 {
		for _, err := range nonFatalErrs {
			logger.Warn(err)
		}
	}

	if err := logging.InitializeLogger(config.Logging.Format, config.Logging.FilePath, config.Logging.KeepStdout); err != nil {
		logger.Fatalf("Cannot initialize logger: %v", err)
	}

	logger.Infof("Authelia %s is starting", utils.Version())

	switch config.Logging.Level {
	case "error":
		logger.Info("Logging severity set to error")
		logging.SetLevel(logrus.ErrorLevel)
	case "warn":
		logger.Info("Logging severity set to warn")
		logging.SetLevel(logrus.WarnLevel)
	case "info":
		logger.Info("Logging severity set to info")
		logging.SetLevel(logrus.InfoLevel)
	case "debug":
		logger.Info("Logging severity set to debug")
		logging.SetLevel(logrus.DebugLevel)
	case "trace":
		logger.Info("Logging severity set to trace")
		logging.SetLevel(logrus.TraceLevel)
	}

	if os.Getenv("ENVIRONMENT") == "dev" {
		logger.Info("===> Authelia is running in development mode. <===")
	}

	var storageProvider storage.Provider

	switch {
	case config.Storage.PostgreSQL != nil:
		storageProvider = storage.NewPostgreSQLProvider(*config.Storage.PostgreSQL)
	case config.Storage.MySQL != nil:
		storageProvider = storage.NewMySQLProvider(*config.Storage.MySQL)
	case config.Storage.Local != nil:
		storageProvider = storage.NewSQLiteProvider(config.Storage.Local.Path)
	default:
		logger.Fatalf("Unrecognized storage backend")
	}

	var (
		userProvider authentication.UserProvider
		err          error
	)

	switch {
	case config.AuthenticationBackend.File != nil:
		userProvider = authentication.NewFileUserProvider(config.AuthenticationBackend.File)
	case config.AuthenticationBackend.LDAP != nil:
		userProvider, err = authentication.NewLDAPUserProvider(config.AuthenticationBackend, autheliaCertPool)
		if err != nil {
			logger.Fatalf("Failed to Check LDAP Authentication Backend: %v", err)
		}
	default:
		logger.Fatalf("Unrecognized authentication backend")
	}

	var notifier notification.Notifier

	switch {
	case config.Notifier.SMTP != nil:
		notifier = notification.NewSMTPNotifier(*config.Notifier.SMTP, autheliaCertPool)
	case config.Notifier.FileSystem != nil:
		notifier = notification.NewFileNotifier(*config.Notifier.FileSystem)
	default:
		logger.Fatalf("Unrecognized notifier")
	}

	if !config.Notifier.DisableStartupCheck {
		_, err = notifier.StartupCheck()
		if err != nil {
			logger.Fatalf("Error during notifier startup check: %s", err)
		}
	}

	clock := utils.RealClock{}
	authorizer := authorization.NewAuthorizer(config)
	sessionProvider := session.NewProvider(config.Session, autheliaCertPool)
	regulator := regulation.NewRegulator(config.Regulation, storageProvider, clock)

	oidcProvider, err := oidc.NewOpenIDConnectProvider(config.IdentityProviders.OIDC)
	if err != nil {
		logger.Fatalf("Error initializing OpenID Connect Provider: %+v", err)
	}

	providers := middlewares.Providers{
		Authorizer:      authorizer,
		UserProvider:    userProvider,
		Regulator:       regulator,
		OpenIDConnect:   oidcProvider,
		StorageProvider: storageProvider,
		Notifier:        notifier,
		SessionProvider: sessionProvider,
	}

	server.StartServer(*config, providers)
}

func main() {
	logger := logging.Logger()

	version := utils.Version()

	rootCmd := &cobra.Command{
		Use: "authelia",
		Run: func(cmd *cobra.Command, args []string) {
			startServer()
		},
		Version: version,
		Short:   fmt.Sprintf("authelia %s", version),
		Long:    fmt.Sprintf(fmtAutheliaLong, version),
	}

	rootCmd.Flags().StringVar(&configPathFlag, "config", "", "Configuration file")

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Show the build of Authelia",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(fmtAutheliaBuild, utils.BuildTag, utils.BuildState, utils.BuildBranch, utils.BuildCommit,
				utils.BuildNumber, runtime.GOOS, runtime.GOARCH, utils.BuildDate, utils.BuildExtra)
		},
	}

	rootCmd.AddCommand(buildCmd, commands.HashPasswordCmd,
		commands.ValidateConfigCmd, commands.CertificatesCmd,
		commands.RSACmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err)
	}
}
