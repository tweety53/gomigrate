package migration

import (
	"database/sql"
	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/sql_dialect"
	"path/filepath"
	"time"
)

func migrateUpGo(db *sql.DB, m *Migration) error {
	fn := m.UpFn
	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	log.Warnf("*** applying %s", filepath.Base(m.Source))
	start := time.Now()
	if _, err := tx.Exec(sql_dialect.GetDialect().InsertUnAppliedVersionSQL(), m.Version); err != nil {
		tx.Rollback()

		duration := time.Since(start)
		log.Warnf("*** failed to apply %s (time: %.3f sec.)", filepath.Base(m.Source), duration.Seconds())
		log.Warn("This version is currently being applied by another app")
		return nil
	}

	if fn != nil {
		// Run Go migration function.
		if err := fn(tx); err != nil {
			tx.Rollback()

			tx, err := db.Begin()
			if err != nil {
				return errors.Wrap(err, "failed to begin transaction")
			}

			if _, err := tx.Exec(sql_dialect.GetDialect().DeleteVersionSQL(), m.Version); err != nil {
				tx.Rollback()

				duration := time.Since(start)
				log.Errf("*** failed to apply %s (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())
				return errors.Wrap(err, "failed to execute delete version transaction for unapplied version")
			}

			duration := time.Since(start)
			log.Errf("*** failed to apply %s (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())
			return errors.Wrapf(err, "failed to run Go migration function %T", fn)
		}
	}

	if _, err := tx.Exec(sql_dialect.GetDialect().UpdateApplyTimeSQL(), int(time.Now().Unix()), m.Version); err != nil {
		tx.Rollback()

		tx, err := db.Begin()
		if err != nil {
			return errors.Wrap(err, "failed to begin delete version transaction")
		}

		if _, err := tx.Exec(sql_dialect.GetDialect().DeleteVersionSQL(), m.Version); err != nil {
			tx.Rollback()

			duration := time.Since(start)
			log.Errf("*** failed to apply %s (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())
			return errors.Wrap(err, "failed to execute delete version transaction for unapplied version")
		}

		duration := time.Since(start)
		log.Errf("*** failed to apply %s (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())
		return errors.Wrap(err, "failed to execute update apply time transaction")
	}

	if err := tx.Commit(); err != nil {
		duration := time.Since(start)
		log.Errf("*** failed to apply %s (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())
		return errors.Wrap(err, "failed to commit transaction")
	}

	duration := time.Since(start)
	if fn != nil {
		log.Infof("*** applied %s (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())
	} else {
		log.Warnf("*** NOT applied %s (empty) (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())
	}

	return nil
}

func migrateDownGo(db *sql.DB, m *Migration) error {
	fn := m.DownFn
	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	log.Warnf("*** reverting %s", filepath.Base(m.Source))
	start := time.Now()
	if _, err := tx.Exec(sql_dialect.GetDialect().LockVersionSQL(), m.Version); err != nil {
		tx.Rollback()

		duration := time.Since(start)
		log.Warnf("*** failed to revert %s (time: %.3f sec.)", filepath.Base(m.Source), duration.Seconds())
		log.Warn("This version is currently being reverted by another app")
		return nil
	}

	if fn != nil {
		// Run Go migration function.
		if err := fn(tx); err != nil {
			tx.Rollback()

			duration := time.Since(start)
			log.Errf("*** failed to revert %s (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())
			return errors.Wrapf(err, "failed to run Go migration function %T", fn)
		}
	}

	if _, err := tx.Exec(sql_dialect.GetDialect().DeleteVersionSQL(), m.Version); err != nil {
		tx.Rollback()

		duration := time.Since(start)
		log.Errf("*** failed to revert %s (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())
		return errors.Wrap(err, "failed to execute delete version transaction")
	}

	if err := tx.Commit(); err != nil {
		duration := time.Since(start)
		log.Errf("*** failed to revert %s (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())
		return errors.Wrap(err, "failed to commit transaction")
	}

	duration := time.Since(start)
	if fn != nil {
		log.Infof("*** reverted %s (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())
	} else {
		log.Warnf("*** NOT applied %s (empty) (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())
	}

	return nil
}
