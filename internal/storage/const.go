package storage

import (
	"fmt"
)

const storageSchemaCurrentVersion = SchemaVersion(2)
const storageSchemaUpgradeMessage = "Storage schema upgraded to v"
const storageSchemaUpgradeErrorText = "storage schema upgrade failed at v"

// Keep table names in lower case because some DB does not support upper case.
const userPreferencesTableName = "user_preferences"
const identityVerificationTokensTableName = "identity_verification_tokens"
const totpSecretsTableName = "totp_secrets"
const u2fDeviceHandlesTableName = "u2f_devices"
const duoDevicesTableName = "duo_devices"
const authenticationLogsTableName = "authentication_logs"
const configTableName = "config"

// sqlUpgradeCreateTableStatements is a map of the schema version number, plus a map of the table name and the statement used to create it.
// The statement is fmt.Sprintf'd with the table name as the first argument.
var sqlUpgradeCreateTableStatements = map[SchemaVersion]map[string]string{
	SchemaVersion(1): {
		userPreferencesTableName:            "CREATE TABLE %s (username VARCHAR(100) PRIMARY KEY, second_factor_method VARCHAR(11))",
		identityVerificationTokensTableName: "CREATE TABLE %s (token VARCHAR(512))",
		totpSecretsTableName:                "CREATE TABLE %s (username VARCHAR(100) PRIMARY KEY, secret VARCHAR(64))",
		u2fDeviceHandlesTableName:           "CREATE TABLE %s (username VARCHAR(100) PRIMARY KEY, keyHandle TEXT, publicKey TEXT)",
		authenticationLogsTableName:         "CREATE TABLE %s (username VARCHAR(100), successful BOOL, time INTEGER)",
		configTableName:                     "CREATE TABLE %s (category VARCHAR(32) NOT NULL, key_name VARCHAR(32) NOT NULL, value TEXT, PRIMARY KEY (category, key_name))",
	},
	SchemaVersion(2): {
		duoDevicesTableName: "CREATE TABLE %s (username VARCHAR(100) PRIMARY KEY, device VARCHAR(32), method VARCHAR(16))",
	},
}

// sqlUpgradesCreateTableIndexesStatements is a map of the schema version number, plus a slice of statements to create all of the indexes.
var sqlUpgradesCreateTableIndexesStatements = map[SchemaVersion][]string{
	SchemaVersion(1): {
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS usr_time_idx ON %s (username, time)", authenticationLogsTableName),
	},
}

const unitTestUser = "john"
