package migration

import (
	"database/sql"
	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/log"
)

func assembleUpFnFromStatements(statements []string, useTx bool, m *Migration) {
	if useTx {
		m.UpFn = func(tx *sql.Tx) error {
			for i := range statements {
				log.Debugf("Executing SQL statement: %s\n", clearStatement(statements[i]))
				if _, err := tx.Exec(statements[i]); err != nil {
					log.Err("Rollback transaction")
					tx.Rollback()
					return errors.Wrapf(err, "failed to execute SQL query %q", clearStatement(statements[i]))
				}
			}

			return nil
		}

		return
	}

	m.UpFn = func(tx *sql.Tx) error {
		for i := range statements {
			log.Debugf("Executing SQL statement: %s\n", clearStatement(statements[i]))
			if _, err := tx.Exec(statements[i]); err != nil {
				return errors.Wrapf(err, "failed to execute SQL query %q", clearStatement(statements[i]))
			}
		}

		return nil
	}
}

func assembleDownFnFromStatements(statements []string, useTx bool, m *Migration) {
	if useTx {
		m.DownFn = func(tx *sql.Tx) error {
			for i := range statements {
				log.Debugf("Executing SQL statement: %s\n", clearStatement(statements[i]))
				if _, err := tx.Exec(statements[i]); err != nil {
					log.Err("Rollback transaction")
					tx.Rollback()
					return errors.Wrapf(err, "failed to execute SQL query %q", clearStatement(statements[i]))
				}
			}

			return nil
		}

		return
	}

	m.DownFn = func(tx *sql.Tx) error {
		for i := range statements {
			log.Debugf("Executing SQL statement: %s\n", clearStatement(statements[i]))
			if _, err := tx.Exec(statements[i]); err != nil {
				return errors.Wrapf(err, "failed to execute SQL query %q", clearStatement(statements[i]))
			}
		}

		return nil
	}
}
