package storage

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // Load the SQLite Driver used in the connection string.
)

// SQLiteProvider is a SQLite3 provider.
type SQLiteProvider struct {
	SQLProvider
}

// NewSQLiteProvider constructs a SQLite provider.
func NewSQLiteProvider(path string) *SQLiteProvider {
	provider := SQLiteProvider{
		SQLProvider{
			name: "sqlite",

			sqlUpgradesCreateTableStatements:        sqlUpgradeCreateTableStatements,
			sqlUpgradesCreateTableIndexesStatements: sqlUpgradesCreateTableIndexesStatements,

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

			sqlGetDuoDeviceByUsername: fmt.Sprintf("SELECT device, method FROM %s WHERE username=?", duoDevicesTableName),
			sqlUpsertDuoDevice:        fmt.Sprintf("REPLACE INTO %s (username, device, method) VALUES (?, ?, ?)", duoDevicesTableName),

			sqlInsertAuthenticationLog:     fmt.Sprintf("INSERT INTO %s (username, successful, time) VALUES (?, ?, ?)", authenticationLogsTableName),
			sqlGetLatestAuthenticationLogs: fmt.Sprintf("SELECT successful, time FROM %s WHERE time>? AND username=? ORDER BY time DESC", authenticationLogsTableName),

			sqlGetExistingTables: "SELECT name FROM sqlite_master WHERE type='table'",

			sqlConfigSetValue: fmt.Sprintf("REPLACE INTO %s (category, key_name, value) VALUES (?, ?, ?)", configTableName),
			sqlConfigGetValue: fmt.Sprintf("SELECT value FROM %s WHERE category=? AND key_name=?", configTableName),
		},
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		provider.log.Fatalf("Unable to create SQL database %s: %s", path, err)
	}

	if err := provider.initialize(db); err != nil {
		provider.log.Fatalf("Unable to initialize SQL database %s: %s", path, err)
	}

	return &provider
}
