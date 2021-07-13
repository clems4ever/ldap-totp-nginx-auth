package storage

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/models"
	"github.com/authelia/authelia/internal/utils"
)

// SQLProvider is a storage provider persisting data in a SQL database.
type SQLProvider struct {
	db   *sql.DB
	log  *logrus.Logger
	name string

	sqlUpgradesCreateTableStatements        map[SchemaVersion]map[string]string
	sqlUpgradesCreateTableIndexesStatements map[SchemaVersion][]string

	sqlGetPreferencesByUsername     string
	sqlUpsertSecondFactorPreference string

	sqlTestIdentityVerificationTokenExistence string
	sqlInsertIdentityVerificationToken        string
	sqlDeleteIdentityVerificationToken        string

	sqlGetTOTPSecretByUsername string
	sqlUpsertTOTPSecret        string
	sqlDeleteTOTPSecret        string

	sqlGetU2FDeviceHandleByUsername string
	sqlUpsertU2FDeviceHandle        string

	sqlGetDuoDeviceByUsername string
	sqlUpsertDuoDevice        string

	sqlInsertAuthenticationLog     string
	sqlGetLatestAuthenticationLogs string

	sqlGetExistingTables string

	sqlConfigSetValue string
	sqlConfigGetValue string
}

func (p *SQLProvider) initialize(db *sql.DB) error {
	p.db = db
	p.log = logging.Logger()

	return p.upgrade()
}

func (p *SQLProvider) getSchemaBasicDetails() (version SchemaVersion, tables []string, err error) {
	rows, err := p.db.Query(p.sqlGetExistingTables)
	if err != nil {
		return version, tables, err
	}

	defer rows.Close()

	var table string

	for rows.Next() {
		err := rows.Scan(&table)
		if err != nil {
			return version, tables, err
		}

		tables = append(tables, table)
	}

	if utils.IsStringInSlice(configTableName, tables) {
		rows, err := p.db.Query(p.sqlConfigGetValue, "schema", "version")
		if err != nil {
			return version, tables, err
		}

		for rows.Next() {
			err := rows.Scan(&version)
			if err != nil {
				return version, tables, err
			}
		}
	}

	return version, tables, nil
}

func (p *SQLProvider) upgrade() error {
	p.log.Debug("Storage schema is being checked to verify it is up to date")

	version, tables, err := p.getSchemaBasicDetails()
	if err != nil {
		return err
	}

	if version < storageSchemaCurrentVersion {
		p.log.Debugf("Storage schema is v%d, latest is v%d", version, storageSchemaCurrentVersion)

		tx, err := p.db.Begin()
		if err != nil {
			return err
		}

		switch version {
		case 0:
			err := p.upgradeSchemaToVersion001(tx, tables)
			if err != nil {
				return p.handleUpgradeFailure(tx, 1, err)
			}

			fallthrough
		case 1:
			err := p.upgradeSchemaToVersion002(tx, tables)
			if err != nil {
				return p.handleUpgradeFailure(tx, 2, err)
			}

			fallthrough
		default:
			err := tx.Commit()
			if err != nil {
				return err
			}

			p.log.Infof("Storage schema upgrade to v%d completed", storageSchemaCurrentVersion)
		}
	} else {
		p.log.Debug("Storage schema is up to date")
	}

	return nil
}

func (p *SQLProvider) handleUpgradeFailure(tx *sql.Tx, version SchemaVersion, err error) error {
	rollbackErr := tx.Rollback()
	formattedErr := fmt.Errorf("%s%d: %v", storageSchemaUpgradeErrorText, version, err)

	if rollbackErr != nil {
		return fmt.Errorf("rollback error occurred: %v (inner error %v)", rollbackErr, formattedErr)
	}

	return formattedErr
}

// LoadPreferred2FAMethod load the preferred method for 2FA from the database.
func (p *SQLProvider) LoadPreferred2FAMethod(username string) (string, error) {
	var method string

	rows, err := p.db.Query(p.sqlGetPreferencesByUsername, username)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if !rows.Next() {
		return "", nil
	}

	err = rows.Scan(&method)

	return method, err
}

// SavePreferred2FAMethod save the preferred method for 2FA to the database.
func (p *SQLProvider) SavePreferred2FAMethod(username string, method string) error {
	_, err := p.db.Exec(p.sqlUpsertSecondFactorPreference, username, method)
	return err
}

// FindIdentityVerificationToken look for an identity verification token in the database.
func (p *SQLProvider) FindIdentityVerificationToken(token string) (bool, error) {
	var found bool

	err := p.db.QueryRow(p.sqlTestIdentityVerificationTokenExistence, token).Scan(&found)
	if err != nil {
		return false, err
	}

	return found, nil
}

// SaveIdentityVerificationToken save an identity verification token in the database.
func (p *SQLProvider) SaveIdentityVerificationToken(token string) error {
	_, err := p.db.Exec(p.sqlInsertIdentityVerificationToken, token)
	return err
}

// RemoveIdentityVerificationToken remove an identity verification token from the database.
func (p *SQLProvider) RemoveIdentityVerificationToken(token string) error {
	_, err := p.db.Exec(p.sqlDeleteIdentityVerificationToken, token)
	return err
}

// SaveTOTPSecret save a TOTP secret of a given user in the database.
func (p *SQLProvider) SaveTOTPSecret(username string, secret string) error {
	_, err := p.db.Exec(p.sqlUpsertTOTPSecret, username, secret)
	return err
}

// LoadTOTPSecret load a TOTP secret given a username from the database.
func (p *SQLProvider) LoadTOTPSecret(username string) (string, error) {
	var secret string
	if err := p.db.QueryRow(p.sqlGetTOTPSecretByUsername, username).Scan(&secret); err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNoTOTPSecret
		}

		return "", err
	}

	return secret, nil
}

// DeleteTOTPSecret delete a TOTP secret from the database given a username.
func (p *SQLProvider) DeleteTOTPSecret(username string) error {
	_, err := p.db.Exec(p.sqlDeleteTOTPSecret, username)
	return err
}

// SaveU2FDeviceHandle save a registered U2F device registration blob.
func (p *SQLProvider) SaveU2FDeviceHandle(username string, keyHandle []byte, publicKey []byte) error {
	_, err := p.db.Exec(p.sqlUpsertU2FDeviceHandle,
		username,
		base64.StdEncoding.EncodeToString(keyHandle),
		base64.StdEncoding.EncodeToString(publicKey))

	return err
}

// LoadU2FDeviceHandle load a U2F device registration blob for a given username.
func (p *SQLProvider) LoadU2FDeviceHandle(username string) ([]byte, []byte, error) {
	var keyHandleBase64, publicKeyBase64 string
	if err := p.db.QueryRow(p.sqlGetU2FDeviceHandleByUsername, username).Scan(&keyHandleBase64, &publicKeyBase64); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, ErrNoU2FDeviceHandle
		}

		return nil, nil, err
	}

	keyHandle, err := base64.StdEncoding.DecodeString(keyHandleBase64)

	if err != nil {
		return nil, nil, err
	}

	publicKey, err := base64.StdEncoding.DecodeString(publicKeyBase64)

	if err != nil {
		return nil, nil, err
	}

	return keyHandle, publicKey, nil
}

// SavePreferredDuoDevice save a Duo device and method of a given user.
func (p *SQLProvider) SavePreferredDuoDevice(username string, device string, method string) error {
	_, err := p.db.Exec(p.sqlUpsertDuoDevice, username, device, method)
	return err
}

// LoadPreferredDuoDevice load a Duo device and method given a username.
func (p *SQLProvider) LoadPreferredDuoDevice(username string) (string, string, error) {
	var device string

	var method string

	if err := p.db.QueryRow(p.sqlGetDuoDeviceByUsername, username).Scan(&device, &method); err != nil {
		if err == sql.ErrNoRows {
			return "", "", ErrNoDuoDevice
		}

		return "", "", err
	}

	return device, method, nil
}

// AppendAuthenticationLog append a mark to the authentication log.
func (p *SQLProvider) AppendAuthenticationLog(attempt models.AuthenticationAttempt) error {
	_, err := p.db.Exec(p.sqlInsertAuthenticationLog, attempt.Username, attempt.Successful, attempt.Time.Unix())
	return err
}

// LoadLatestAuthenticationLogs retrieve the latest marks from the authentication log.
func (p *SQLProvider) LoadLatestAuthenticationLogs(username string, fromDate time.Time) ([]models.AuthenticationAttempt, error) {
	var t int64

	rows, err := p.db.Query(p.sqlGetLatestAuthenticationLogs, fromDate.Unix(), username)

	if err != nil {
		return nil, err
	}

	attempts := make([]models.AuthenticationAttempt, 0, 10)

	for rows.Next() {
		attempt := models.AuthenticationAttempt{
			Username: username,
		}
		err = rows.Scan(&attempt.Successful, &t)
		attempt.Time = time.Unix(t, 0)

		if err != nil {
			return nil, err
		}

		attempts = append(attempts, attempt)
	}

	return attempts, nil
}
