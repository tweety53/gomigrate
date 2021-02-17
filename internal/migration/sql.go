package migration

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/log"
)

// dont know how to test this :).
func assembleSafeFnFromStatements(statements []string) func(tx *sql.Tx) error {
	return func(tx *sql.Tx) error {
		for i := range statements {
			log.Debugf("Executing SQL statement: %s\n", clearStatement(statements[i]))
			if _, err := tx.Exec(statements[i]); err != nil {
				log.Err("Rollback transaction")
				txErr := tx.Rollback()
				if txErr != nil {
					return errors.Wrapf(err, "failed to rollback SQL query %q", clearStatement(statements[i]))
				}

				return errors.Wrapf(err, "failed to execute SQL query %q", clearStatement(statements[i]))
			}
		}

		return nil
	}
}

// dont know how to test this :).
func assembleFnFromStatements(statements []string) func(db *sql.DB) error {
	return func(db *sql.DB) error {
		for i := range statements {
			log.Debugf("Executing SQL statement: %s\n", clearStatement(statements[i]))
			if _, err := db.Exec(statements[i]); err != nil {
				return errors.Wrapf(err, "failed to execute SQL query %q", clearStatement(statements[i]))
			}
		}

		return nil
	}
}
