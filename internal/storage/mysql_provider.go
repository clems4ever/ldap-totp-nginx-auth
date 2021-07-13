package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // Load the MySQL Driver used in the connection string.

	"github.com/authelia/authelia/internal/configuration/schema"
)

// MySQLProvider is a MySQL provider.
type MySQLProvider struct {
	SQLProvider
}

// NewMySQLProvider a MySQL provider.
func NewMySQLProvider(configuration schema.MySQLStorageConfiguration) *MySQLProvider {
	provider := MySQLProvider{
		SQLProvider{
			name: "mysql",

			sqlUpgradesCreateTableStatements: sqlUpgradeCreateTableStatements,

			sqlGetPreferencesByUsername:     fmt.Sprintf("SELECT second_factor_method FROM %s WHERE username=?", userPreferencesTableName),
			sqlUpsertSecondFactorPreference: fmt.Sprintf("REPLACE INTO %s (username, second_factor_method) VALUES (?, ?)", userPreferencesTableName),

			sqlTestIdentityVerificationTokenExistence: fmt.Sprintf("SELECT EXISTS (SELECT * FROM %s WHERE token=?)", identityVerificationTokensTableName),
			sqlInsertIdentityVerificationToken:        fmt.Sprintf("INSERT INTO %s (token) VALUES (?)", identityVerificationTokensTableName),
			sqlDeleteIdentityVerificationToken:        fmt.Sprintf("DELETE FROM %s WHERE token=?", identityVerificationTokensTableName),

			sqlGetTOTPSecretByUsername: fmt.Sprintf("SELECT secret FROM %s WHERE username=?", totpSecretsTableName),
			sqlUpsertTOTPSecret:        fmt.Sprintf("REPLACE INTO %s (username, secret) VALUES (?, ?)", totpSecretsTableName),
			sqlDeleteTOTPSecret:        fmt.Sprintf("DELETE FROM %s WHERE username=?", totpSecretsTableName),

			sqlGetU2FDeviceHandleByUsername: fmt.Sprintf("SELECT keyHandle, publicKey FROM %s WHERE username=?", u2fDeviceHandlesTableName),
			sqlUpsertU2FDeviceHandle:        fmt.Sprintf("REPLACE INTO %s (username, keyHandle, publicKey) VALUES (?, ?, ?)", u2fDeviceHandlesTableName),

			sqlGetDuoDeviceByUsername: fmt.Sprintf("SELECT device FROM %s WHERE username=?", duoDevicesTableName),
			sqlUpsertDuoDevice:        fmt.Sprintf("REPLACE INTO %s (username, device, method) VALUES (?, ?, ?)", duoDevicesTableName),

			sqlInsertAuthenticationLog:     fmt.Sprintf("INSERT INTO %s (username, successful, time) VALUES (?, ?, ?)", authenticationLogsTableName),
			sqlGetLatestAuthenticationLogs: fmt.Sprintf("SELECT successful, time FROM %s WHERE time>? AND username=? ORDER BY time DESC", authenticationLogsTableName),

			sqlGetExistingTables: "SELECT table_name FROM information_schema.tables WHERE table_type='BASE TABLE' AND table_schema=database()",

			sqlConfigSetValue: fmt.Sprintf("REPLACE INTO %s (category, key_name, value) VALUES (?, ?, ?)", configTableName),
			sqlConfigGetValue: fmt.Sprintf("SELECT value FROM %s WHERE category=? AND key_name=?", configTableName),
		},
	}

	provider.sqlUpgradesCreateTableStatements[SchemaVersion(1)][authenticationLogsTableName] = "CREATE TABLE %s (username VARCHAR(100), successful BOOL, time INTEGER, INDEX usr_time_idx (username, time))"

	connectionString := configuration.Username

	if configuration.Password != "" {
		connectionString += fmt.Sprintf(":%s", configuration.Password)
	}

	if connectionString != "" {
		connectionString += "@"
	}

	address := configuration.Host
	if configuration.Port > 0 {
		address += fmt.Sprintf(":%d", configuration.Port)
	}

	connectionString += fmt.Sprintf("tcp(%s)", address)
	if configuration.Database != "" {
		connectionString += fmt.Sprintf("/%s", configuration.Database)
	}

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		provider.log.Fatalf("Unable to connect to SQL database: %v", err)
	}

	if err := provider.initialize(db); err != nil {
		provider.log.Fatalf("Unable to initialize SQL database: %v", err)
	}

	return &provider
}
