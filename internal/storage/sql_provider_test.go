package storage

import (
	"database/sql/driver"
	"encoding/base64"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/models"
)

const currentSchemaMockSchemaVersion = "1"

func TestSQLInitializeDatabase(t *testing.T) {
	provider, mock := NewSQLMockProvider()

	rows := sqlmock.NewRows([]string{"name"})
	mock.ExpectQuery(
		"SELECT name FROM sqlite_master WHERE type='table'").
		WillReturnRows(rows)

	mock.ExpectBegin()

	keys := make([]string, 0, len(sqlUpgradeCreateTableStatements[1]))
	for k := range sqlUpgradeCreateTableStatements[1] {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, table := range keys {
		mock.ExpectExec(
			fmt.Sprintf("CREATE TABLE %s .*", table)).
			WillReturnResult(sqlmock.NewResult(0, 0))
	}

	mock.ExpectExec(
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS usr_time_idx ON %s .*", authenticationLogsTableName)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(
		fmt.Sprintf("REPLACE INTO %s \\(category, key_name, value\\) VALUES \\(\\?, \\?, \\?\\)", configTableName)).
		WithArgs("schema", "version", "1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := provider.initialize(provider.db)
	assert.NoError(t, err)
}

func TestSQLUpgradeDatabase(t *testing.T) {
	provider, mock := NewSQLMockProvider()

	mock.ExpectQuery(
		"SELECT name FROM sqlite_master WHERE type='table'").
		WillReturnRows(sqlmock.NewRows([]string{"name"}).
			AddRow(userPreferencesTableName).
			AddRow(identityVerificationTokensTableName).
			AddRow(totpSecretsTableName).
			AddRow(u2fDeviceHandlesTableName).
			AddRow(authenticationLogsTableName))

	mock.ExpectBegin()

	mock.ExpectExec(
		fmt.Sprintf("CREATE TABLE %s .*", configTableName)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS usr_time_idx ON %s .*", authenticationLogsTableName)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(
		fmt.Sprintf("REPLACE INTO %s \\(category, key_name, value\\) VALUES \\(\\?, \\?, \\?\\)", configTableName)).
		WithArgs("schema", "version", "1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := provider.initialize(provider.db)
	assert.NoError(t, err)
}

func TestSQLProviderMethodsAuthenticationLogs(t *testing.T) {
	provider, mock := NewSQLMockProvider()

	mock.ExpectQuery(
		"SELECT name FROM sqlite_master WHERE type='table'").
		WillReturnRows(sqlmock.NewRows([]string{"name"}).
			AddRow(userPreferencesTableName).
			AddRow(identityVerificationTokensTableName).
			AddRow(totpSecretsTableName).
			AddRow(u2fDeviceHandlesTableName).
			AddRow(authenticationLogsTableName).
			AddRow(configTableName))

	args := []driver.Value{"schema", "version"}
	mock.ExpectQuery(
		fmt.Sprintf("SELECT value FROM %s WHERE category=\\? AND key_name=\\?", configTableName)).
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).
			AddRow("1"))

	err := provider.initialize(provider.db)
	assert.NoError(t, err)

	attempts := []models.AuthenticationAttempt{
		{Username: unitTestUser, Successful: true, Time: time.Unix(1577880001, 0)},
		{Username: unitTestUser, Successful: true, Time: time.Unix(1577880002, 0)},
		{Username: unitTestUser, Successful: false, Time: time.Unix(1577880003, 0)},
	}

	rows := sqlmock.NewRows([]string{"successful", "time"})

	for id, attempt := range attempts {
		args = []driver.Value{attempt.Username, attempt.Successful, attempt.Time.Unix()}
		mock.ExpectExec(
			fmt.Sprintf("INSERT INTO %s \\(username, successful, time\\) VALUES \\(\\?, \\?, \\?\\)", authenticationLogsTableName)).
			WithArgs(args...).
			WillReturnResult(sqlmock.NewResult(int64(id), 1))

		err := provider.AppendAuthenticationLog(attempt)
		assert.NoError(t, err)
		rows.AddRow(attempt.Successful, attempt.Time.Unix())
	}

	args = []driver.Value{1577880000, unitTestUser}
	mock.ExpectQuery(
		fmt.Sprintf("SELECT successful, time FROM %s WHERE time>\\? AND username=\\? ORDER BY time DESC", authenticationLogsTableName)).
		WithArgs(args...).
		WillReturnRows(rows)

	after := time.Unix(1577880000, 0)
	results, err := provider.LoadLatestAuthenticationLogs(unitTestUser, after)
	assert.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, unitTestUser, results[0].Username)
	assert.Equal(t, true, results[0].Successful)
	assert.Equal(t, time.Unix(1577880001, 0), results[0].Time)
	assert.Equal(t, unitTestUser, results[1].Username)
	assert.Equal(t, true, results[1].Successful)
	assert.Equal(t, time.Unix(1577880002, 0), results[1].Time)
	assert.Equal(t, unitTestUser, results[2].Username)
	assert.Equal(t, false, results[2].Successful)
	assert.Equal(t, time.Unix(1577880003, 0), results[2].Time)

	// Test Blank Rows.
	mock.ExpectQuery(
		fmt.Sprintf("SELECT successful, time FROM %s WHERE time>\\? AND username=\\? ORDER BY time DESC", authenticationLogsTableName)).
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"successful", "time"}))

	results, err = provider.LoadLatestAuthenticationLogs(unitTestUser, after)
	assert.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestSQLProviderMethodsPreferred(t *testing.T) {
	provider, mock := NewSQLMockProvider()

	mock.ExpectQuery(
		"SELECT name FROM sqlite_master WHERE type='table'").
		WillReturnRows(sqlmock.NewRows([]string{"name"}).
			AddRow(userPreferencesTableName).
			AddRow(identityVerificationTokensTableName).
			AddRow(totpSecretsTableName).
			AddRow(u2fDeviceHandlesTableName).
			AddRow(authenticationLogsTableName).
			AddRow(configTableName))

	args := []driver.Value{"schema", "version"}
	mock.ExpectQuery(
		fmt.Sprintf("SELECT value FROM %s WHERE category=\\? AND key_name=\\?", configTableName)).
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).
			AddRow(currentSchemaMockSchemaVersion))

	err := provider.initialize(provider.db)
	assert.NoError(t, err)

	mock.ExpectExec(
		fmt.Sprintf("REPLACE INTO %s \\(username, second_factor_method\\) VALUES \\(\\?, \\?\\)", userPreferencesTableName)).
		WithArgs(unitTestUser, authentication.TOTP).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = provider.SavePreferred2FAMethod(unitTestUser, authentication.TOTP)
	assert.NoError(t, err)

	mock.ExpectQuery(
		fmt.Sprintf("SELECT second_factor_method FROM %s WHERE username=\\?", userPreferencesTableName)).
		WithArgs(unitTestUser).
		WillReturnRows(sqlmock.NewRows([]string{"second_factor_method"}).AddRow(authentication.TOTP))

	method, err := provider.LoadPreferred2FAMethod(unitTestUser)
	assert.NoError(t, err)
	assert.Equal(t, authentication.TOTP, method)

	// Test Blank Rows.
	mock.ExpectQuery(
		fmt.Sprintf("SELECT second_factor_method FROM %s WHERE username=\\?", userPreferencesTableName)).
		WithArgs(unitTestUser).
		WillReturnRows(sqlmock.NewRows([]string{"second_factor_method"}))

	method, err = provider.LoadPreferred2FAMethod(unitTestUser)
	assert.NoError(t, err)
	assert.Equal(t, "", method)
}

func TestSQLProviderMethodsTOTP(t *testing.T) {
	provider, mock := NewSQLMockProvider()

	mock.ExpectQuery(
		"SELECT name FROM sqlite_master WHERE type='table'").
		WillReturnRows(sqlmock.NewRows([]string{"name"}).
			AddRow(userPreferencesTableName).
			AddRow(identityVerificationTokensTableName).
			AddRow(totpSecretsTableName).
			AddRow(u2fDeviceHandlesTableName).
			AddRow(authenticationLogsTableName).
			AddRow(configTableName))

	args := []driver.Value{"schema", "version"}
	mock.ExpectQuery(
		fmt.Sprintf("SELECT value FROM %s WHERE category=\\? AND key_name=\\?", configTableName)).
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).
			AddRow(currentSchemaMockSchemaVersion))

	err := provider.initialize(provider.db)
	assert.NoError(t, err)

	pretendSecret := "abc123"
	args = []driver.Value{unitTestUser, pretendSecret}
	mock.ExpectExec(
		fmt.Sprintf("REPLACE INTO %s \\(username, secret\\) VALUES \\(\\?, \\?\\)", totpSecretsTableName)).
		WithArgs(args...).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = provider.SaveTOTPSecret(unitTestUser, pretendSecret)
	assert.NoError(t, err)

	args = []driver.Value{unitTestUser}
	mock.ExpectQuery(
		fmt.Sprintf("SELECT secret FROM %s WHERE username=\\?", totpSecretsTableName)).
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"secret"}).AddRow(pretendSecret))

	secret, err := provider.LoadTOTPSecret(unitTestUser)
	assert.NoError(t, err)
	assert.Equal(t, pretendSecret, secret)

	mock.ExpectExec(
		fmt.Sprintf("DELETE FROM %s WHERE username=\\?", totpSecretsTableName)).
		WithArgs(unitTestUser).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = provider.DeleteTOTPSecret(unitTestUser)
	assert.NoError(t, err)

	mock.ExpectQuery(
		fmt.Sprintf("SELECT secret FROM %s WHERE username=\\?", totpSecretsTableName)).
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"secret"}))

	// Test Blank Rows
	secret, err = provider.LoadTOTPSecret(unitTestUser)
	assert.EqualError(t, err, "no TOTP secret registered")
	assert.Equal(t, "", secret)
}

func TestSQLProviderMethodsU2F(t *testing.T) {
	provider, mock := NewSQLMockProvider()

	mock.ExpectQuery(
		"SELECT name FROM sqlite_master WHERE type='table'").
		WillReturnRows(sqlmock.NewRows([]string{"name"}).
			AddRow(userPreferencesTableName).
			AddRow(identityVerificationTokensTableName).
			AddRow(totpSecretsTableName).
			AddRow(u2fDeviceHandlesTableName).
			AddRow(authenticationLogsTableName).
			AddRow(configTableName))

	args := []driver.Value{"schema", "version"}
	mock.ExpectQuery(
		fmt.Sprintf("SELECT value FROM %s WHERE category=\\? AND key_name=\\?", configTableName)).
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).
			AddRow(currentSchemaMockSchemaVersion))

	err := provider.initialize(provider.db)
	assert.NoError(t, err)

	pretendKeyHandle := []byte("abc")
	pretendPublicKey := []byte("123")
	pretendKeyHandleB64 := base64.StdEncoding.EncodeToString(pretendKeyHandle)
	pretendPublicKeyB64 := base64.StdEncoding.EncodeToString(pretendPublicKey)

	args = []driver.Value{unitTestUser, pretendKeyHandleB64, pretendPublicKeyB64}
	mock.ExpectExec(
		fmt.Sprintf("REPLACE INTO %s \\(username, keyHandle, publicKey\\) VALUES \\(\\?, \\?, \\?\\)", u2fDeviceHandlesTableName)).
		WithArgs(args...).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = provider.SaveU2FDeviceHandle(unitTestUser, pretendKeyHandle, pretendPublicKey)
	assert.NoError(t, err)

	args = []driver.Value{unitTestUser}
	mock.ExpectQuery(
		fmt.Sprintf("SELECT keyHandle, publicKey FROM %s WHERE username=\\?", u2fDeviceHandlesTableName)).
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"keyHandle", "publicKey"}).
			AddRow(pretendKeyHandleB64, pretendPublicKeyB64))

	keyHandle, publicKey, err := provider.LoadU2FDeviceHandle(unitTestUser)
	assert.NoError(t, err)
	assert.Equal(t, pretendKeyHandle, keyHandle)
	assert.Equal(t, pretendPublicKey, publicKey)

	// Test Blank Rows.
	mock.ExpectQuery(
		fmt.Sprintf("SELECT keyHandle, publicKey FROM %s WHERE username=\\?", u2fDeviceHandlesTableName)).
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"keyHandle", "publicKey"}))

	keyHandle, publicKey, err = provider.LoadU2FDeviceHandle(unitTestUser)
	assert.EqualError(t, err, "no U2F device handle found")
	assert.Equal(t, []byte(nil), keyHandle)
	assert.Equal(t, []byte(nil), publicKey)
}

func TestSQLProviderMethodsIdentityVerificationTokens(t *testing.T) {
	provider, mock := NewSQLMockProvider()

	mock.ExpectQuery(
		"SELECT name FROM sqlite_master WHERE type='table'").
		WillReturnRows(sqlmock.NewRows([]string{"name"}).
			AddRow(userPreferencesTableName).
			AddRow(identityVerificationTokensTableName).
			AddRow(totpSecretsTableName).
			AddRow(u2fDeviceHandlesTableName).
			AddRow(authenticationLogsTableName).
			AddRow(configTableName))

	args := []driver.Value{"schema", "version"}
	mock.ExpectQuery(
		fmt.Sprintf("SELECT value FROM %s WHERE category=\\? AND key_name=\\?", configTableName)).
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).
			AddRow(currentSchemaMockSchemaVersion))

	err := provider.initialize(provider.db)
	assert.NoError(t, err)

	fakeIdentityVerificationToken := "abc"

	mock.ExpectExec(
		fmt.Sprintf("INSERT INTO %s \\(token\\) VALUES \\(\\?\\)", identityVerificationTokensTableName)).
		WithArgs(fakeIdentityVerificationToken).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = provider.SaveIdentityVerificationToken(fakeIdentityVerificationToken)
	assert.NoError(t, err)

	mock.ExpectQuery(
		fmt.Sprintf("SELECT EXISTS \\(SELECT \\* FROM %s WHERE token=\\?\\)", identityVerificationTokensTableName)).
		WithArgs(fakeIdentityVerificationToken).
		WillReturnRows(sqlmock.NewRows([]string{"EXISTS"}).
			AddRow(true))

	valid, err := provider.FindIdentityVerificationToken(fakeIdentityVerificationToken)
	assert.NoError(t, err)
	assert.True(t, valid)

	mock.ExpectExec(
		fmt.Sprintf("DELETE FROM %s WHERE token=\\?", identityVerificationTokensTableName)).
		WithArgs(fakeIdentityVerificationToken).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = provider.RemoveIdentityVerificationToken(fakeIdentityVerificationToken)
	assert.NoError(t, err)

	mock.ExpectQuery(
		fmt.Sprintf("SELECT EXISTS \\(SELECT \\* FROM %s WHERE token=\\?\\)", identityVerificationTokensTableName)).
		WithArgs(fakeIdentityVerificationToken).
		WillReturnRows(sqlmock.NewRows([]string{"EXISTS"}).
			AddRow(false))

	valid, err = provider.FindIdentityVerificationToken(fakeIdentityVerificationToken)
	assert.NoError(t, err)
	assert.False(t, valid)
}
