package storage

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v4/stdlib" // Load the PostgreSQL Driver used in the connection string.

	"github.com/authelia/authelia/internal/configuration/schema"
)

// PostgreSQLProvider is a PostgreSQL provider.
type PostgreSQLProvider struct {
	SQLProvider
}

// NewPostgreSQLProvider a PostgreSQL provider.
func NewPostgreSQLProvider(configuration schema.PostgreSQLStorageConfiguration) *PostgreSQLProvider {
	provider := PostgreSQLProvider{
		SQLProvider{
			name: "postgres",

			sqlUpgradesCreateTableStatements:        sqlUpgradeCreateTableStatements,
			sqlUpgradesCreateTableIndexesStatements: sqlUpgradesCreateTableIndexesStatements,

			sqlGetPreferencesByUsername:     fmt.Sprintf("SELECT second_factor_method FROM %s WHERE username=$1", userPreferencesTableName),
			sqlUpsertSecondFactorPreference: fmt.Sprintf("INSERT INTO %s (username, second_factor_method) VALUES ($1, $2) ON CONFLICT (username) DO UPDATE SET second_factor_method=$2", userPreferencesTableName),

			sqlTestIdentityVerificationTokenExistence: fmt.Sprintf("SELECT EXISTS (SELECT * FROM %s WHERE token=$1)", identityVerificationTokensTableName),
			sqlInsertIdentityVerificationToken:        fmt.Sprintf("INSERT INTO %s (token) VALUES ($1)", identityVerificationTokensTableName),
			sqlDeleteIdentityVerificationToken:        fmt.Sprintf("DELETE FROM %s WHERE token=$1", identityVerificationTokensTableName),

			sqlGetTOTPSecretByUsername: fmt.Sprintf("SELECT secret FROM %s WHERE username=$1", totpSecretsTableName),
			sqlUpsertTOTPSecret:        fmt.Sprintf("INSERT INTO %s (username, secret) VALUES ($1, $2) ON CONFLICT (username) DO UPDATE SET secret=$2", totpSecretsTableName),
			sqlDeleteTOTPSecret:        fmt.Sprintf("DELETE FROM %s WHERE username=$1", totpSecretsTableName),

			sqlGetU2FDeviceHandleByUsername: fmt.Sprintf("SELECT keyHandle, publicKey FROM %s WHERE username=$1", u2fDeviceHandlesTableName),
			sqlUpsertU2FDeviceHandle:        fmt.Sprintf("INSERT INTO %s (username, keyHandle, publicKey) VALUES ($1, $2, $3) ON CONFLICT (username) DO UPDATE SET keyHandle=$2, publicKey=$3", u2fDeviceHandlesTableName),

			sqlGetDuoDeviceByUsername: fmt.Sprintf("SELECT device, method FROM %s WHERE username=$1", duoDevicesTableName),
			sqlUpsertDuoDevice:        fmt.Sprintf("INSERT INTO %s (username, device, method) VALUES ($1, $2, $3) ON CONFLICT (username) DO UPDATE SET device=$2, method=$3", duoDevicesTableName),

			sqlInsertAuthenticationLog:     fmt.Sprintf("INSERT INTO %s (username, successful, time) VALUES ($1, $2, $3)", authenticationLogsTableName),
			sqlGetLatestAuthenticationLogs: fmt.Sprintf("SELECT successful, time FROM %s WHERE time>$1 AND username=$2 ORDER BY time DESC", authenticationLogsTableName),

			sqlGetExistingTables: "SELECT table_name FROM information_schema.tables WHERE table_type='BASE TABLE' AND table_schema='public'",

			sqlConfigSetValue: fmt.Sprintf("INSERT INTO %s (category, key_name, value) VALUES ($1, $2, $3) ON CONFLICT (category, key_name) DO UPDATE SET value=$3", configTableName),
			sqlConfigGetValue: fmt.Sprintf("SELECT value FROM %s WHERE category=$1 AND key_name=$2", configTableName),
		},
	}

	args := make([]string, 0)
	if configuration.Username != "" {
		args = append(args, fmt.Sprintf("user='%s'", configuration.Username))
	}

	if configuration.Password != "" {
		args = append(args, fmt.Sprintf("password='%s'", configuration.Password))
	}

	if configuration.Host != "" {
		args = append(args, fmt.Sprintf("host=%s", configuration.Host))
	}

	if configuration.Port > 0 {
		args = append(args, fmt.Sprintf("port=%d", configuration.Port))
	}

	if configuration.Database != "" {
		args = append(args, fmt.Sprintf("dbname=%s", configuration.Database))
	}

	if configuration.SSLMode != "" {
		args = append(args, fmt.Sprintf("sslmode=%s", configuration.SSLMode))
	}

	connectionString := strings.Join(args, " ")

	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		provider.log.Fatalf("Unable to connect to SQL database: %v", err)
	}

	if err := provider.initialize(db); err != nil {
		provider.log.Fatalf("Unable to initialize SQL database: %v", err)
	}

	return &provider
}
