package storage

import (
	"fmt"
	"sort"

	"github.com/authelia/authelia/internal/utils"
)

func (p *SQLProvider) upgradeCreateTableStatements(tx transaction, statements map[string]string, existingTables []string) error {
	keys := make([]string, 0, len(statements))
	for k := range statements {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, table := range keys {
		if !utils.IsStringInSlice(table, existingTables) {
			_, err := tx.Exec(fmt.Sprintf(statements[table], table))
			if err != nil {
				return fmt.Errorf("unable to create table %s: %v", table, err)
			}
		}
	}

	return nil
}

func (p *SQLProvider) upgradeRunMultipleStatements(tx transaction, statements []string) error {
	for _, statement := range statements {
		_, err := tx.Exec(statement)
		if err != nil {
			return err
		}
	}

	return nil
}

// upgradeFinalize sets the schema version and logs a message, as well as any other future finalization tasks.
func (p *SQLProvider) upgradeFinalize(tx transaction, version SchemaVersion) error {
	_, err := tx.Exec(p.sqlConfigSetValue, "schema", "version", version.ToString())
	if err != nil {
		return err
	}

	p.log.Debugf("%s%d", storageSchemaUpgradeMessage, version)

	return nil
}

// upgradeSchemaToVersion001 upgrades the schema to version 1.
func (p *SQLProvider) upgradeSchemaToVersion001(tx transaction, tables []string) error {
	version := SchemaVersion(1)

	err := p.upgradeCreateTableStatements(tx, p.sqlUpgradesCreateTableStatements[version], tables)
	if err != nil {
		return err
	}

	// Skip mysql create index statements. It doesn't support CREATE INDEX IF NOT EXIST. May be able to work around this with an Index struct.
	if p.name != "mysql" {
		err = p.upgradeRunMultipleStatements(tx, p.sqlUpgradesCreateTableIndexesStatements[1])
		if err != nil {
			return fmt.Errorf("unable to create index: %v", err)
		}
	}

	err = p.upgradeFinalize(tx, version)
	if err != nil {
		return err
	}

	return nil
}
