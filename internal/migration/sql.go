package migration

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/log"
)

//nolint:nestif
func assembleFnFromStatements(statements []string, useTx bool, m *Migration, direction Direction) {
	var fn func(tx *sql.Tx) error
	if useTx {
		fn = func(tx *sql.Tx) error {
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

		if direction == migrationDirectionUp {
			m.UpFn = fn
		} else {
			m.DownFn = fn
		}

		return
	}

	fn = func(tx *sql.Tx) error {
		for i := range statements {
			log.Debugf("Executing SQL statement: %s\n", clearStatement(statements[i]))
			if _, err := tx.Exec(statements[i]); err != nil {
				return errors.Wrapf(err, "failed to execute SQL query %q", clearStatement(statements[i]))
			}
		}

		return nil
	}

	if direction == migrationDirectionUp {
		m.UpFn = fn
	} else {
		m.DownFn = fn
	}
}
