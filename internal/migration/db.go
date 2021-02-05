package migration

import (
	"database/sql"
	"github.com/pkg/errors"
	"gomigrate/internal/sql_dialect"
	"time"
)

func GetNewMigrations(db *sql.DB) (Migrations, error) {
	if _, err := GetDBVersion(db); err != nil {
		return nil, err
	}

	rows, err := sql_dialect.GetDialect().GetMigrationsHistory(db, 0)
	if err != nil {
		return Migrations{}, errors.New("cannot get migrations history from dbogar")
	}
	defer rows.Close()

	applied := make(map[string]int)
	for rows.Next() {
		var row MigrationRecord
		if err = rows.Scan(&row.Version, &row.ApplyTime); err != nil {
			return Migrations{}, errors.Wrap(err, "failed to scan row")
		}

		// skip base migration
		if row.Version == "m000000_000000_base" {
			continue
		}

		applied[row.Version] = row.ApplyTime
	}

	allMigrations, err := CollectMigrations("/Users/yuriy.aleksandrov/go/src/gomigrate/migrations", 0, 0)
	if err != nil {
		return Migrations{}, errors.New("cannot collect migr filesogar")
	}

	newCnt := len(allMigrations) - len(applied)
	if newCnt < 0 {
		newCnt = 0
	}
	newMigrations := make(Migrations, 0, newCnt)
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
	rows, err := sql_dialect.GetDialect().GetMigrationsHistory(db, 1)
	if err != nil {
		return "", createVersionTable(db)
	}
	defer rows.Close()

	return "", nil
}

// Create the db version table
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

	if _, err := txn.Exec(d.InsertVersionSQL(), "m000000_000000_base", int(time.Now().Unix())); err != nil {
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
