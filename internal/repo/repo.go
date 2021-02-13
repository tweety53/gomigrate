package repo

import (
	"database/sql"
	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/migration"
	"github.com/tweety53/gomigrate/internal/sql_dialect"
	"strconv"
	"time"
)

type MigrationsRepo interface {
	GetNewMigrations(db *sql.DB, migrationsPath string) (migration.Migrations, error)
	EnsureDBVersion(db *sql.DB) (string, error)
	TruncateDatabase(db *sql.DB) error
	GetFkRows(db *sql.DB, tableName string) (*sql.Rows, error)
	GetMigrationsHistory(db *sql.DB, limit int) (*sql.Rows, error)
}

func GetNewMigrations(db *sql.DB, migrationsPath string) (migration.Migrations, error) {
	if _, err := GetDBVersion(db); err != nil {
		return nil, err
	}

	rows, err := GetMigrationsHistory(db, 0)
	if err != nil {
		return migration.Migrations{}, errors.Wrap(err, "cannot get migrations history from repo")
	}
	defer rows.Close()

	applied := make(map[string]int)
	for rows.Next() {
		var row migration.MigrationRecord
		if err = rows.Scan(&row.Version, &row.ApplyTime); err != nil {
			return migration.Migrations{}, errors.Wrap(err, "failed to scan row")
		}

		// skip base migration
		if row.Version == migration.BaseMigrationVersion {
			continue
		}

		applied[row.Version] = row.ApplyTime
	}

	allMigrations, err := migration.CollectMigrations(migrationsPath, 0, 0)
	if err != nil {
		return migration.Migrations{}, errors.Wrapf(err, "cannot collect migration files from path: %s", migrationsPath)
	}

	newCnt := len(allMigrations) - len(applied)
	if newCnt < 0 {
		newCnt = 0
	}
	newMigrations := make(migration.Migrations, 0, newCnt)
	for _, row := range allMigrations {
		if _, ok := applied[row.Version]; ok {
			continue
		}

		newMigrations = append(newMigrations, row)
	}

	return newMigrations, err
}

// EnsureDBVersion retrieves the current version for this DB.
// Create and initialize the DB version table if it doesn't exist.
func EnsureDBVersion(db *sql.DB) (string, error) {
	rows, err := GetMigrationsHistory(db, 1)
	if err != nil {
		return "", createVersionTable(db)
	}
	defer rows.Close()

	return "", nil
}

// Create the repo version table
// and insert the initial 0 value into it
func createVersionTable(db *sql.DB) error {
	txn, err := db.Begin()
	if err != nil {
		return err
	}

	d := sql_dialect.GetDialect()

	if _, err := txn.Exec(d.CreateVersionTableSQL()); err != nil {
		txn.Rollback()
		return err
	}

	if _, err := txn.Exec(d.InsertVersionSQL(), migration.BaseMigrationVersion, int(time.Now().Unix())); err != nil {
		txn.Rollback()
		return err
	}

	return txn.Commit()
}

// GetDBVersion is an alias for EnsureDBVersion, but returns -1 in error.
func GetDBVersion(db *sql.DB) (string, error) {
	version, err := EnsureDBVersion(db)
	if err != nil {
		return "", err
	}

	return version, nil
}

func TruncateDatabase(db *sql.DB) error {
	tableNamesRows, err := db.Query(sql_dialect.GetDialect().AllTableNamesSQL())
	if err != nil {
		return err
	}
	defer tableNamesRows.Close()

	tableNames := make([]string, 0, 1024*1024*4)
	// First drop all foreign keys
	for tableNamesRows.Next() {
		var tableName string
		if err = tableNamesRows.Scan(&tableName); err != nil {
			return err
		}

		fkRows, err := GetFkRows(db, tableName)
		if err != nil {
			return err
		}

		for fkRows.Next() {
			var fkName string
			if err = fkRows.Scan(&fkName); err != nil {
				return err
			}

			_, err := db.Exec(sql_dialect.GetDialect().DropFkSQL(tableName, fkName))
			if err != nil {
				log.Errf("Foreign key drop err: %v\n", err)
				return err
			}

			log.Infof("Foreign key %s dropped.", fkName)
		}

		tableNames = append(tableNames, tableName)
	}

	// Then drop the tables
	for _, name := range tableNames {
		//todo: handle repo views errors
		_, err := db.Exec(sql_dialect.GetDialect().DropTableSQL(name))
		if err != nil {
			log.Errf("Cannot drop %s table, err: %v\n", err)
			return err
		}

		log.Infof("Table %s dropped.", name)
	}

	return nil
}

func GetFkRows(db *sql.DB, tableName string) (*sql.Rows, error) {
	fkRows, err := db.Query(sql_dialect.GetDialect().TableForeignKeysSQL(), tableName)
	if err != nil {
		return nil, err
	}
	defer fkRows.Close()

	return fkRows, err
}

func GetMigrationsHistory(db *sql.DB, limit int) (*sql.Rows, error) {
	query := sql_dialect.GetDialect().MigrationsHistorySQL()
	if limit > 0 {
		query += " LIMIT " + strconv.Itoa(limit)
	}

	query += ";"

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	return rows, err
}
