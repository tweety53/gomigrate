package migration

import (
	"database/sql"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/repo"
)

const (
	failedToRevertLogText = "*** failed to revert %s (time: %.3f sec.)\n"
	failedToApplyLogText  = "*** failed to apply %s (time: %.3f sec.)\n"
)

func migrateUpGo(repo *repo.MigrationsRepository, m *Migration) error {
	fn := m.UpFn
	tx, err := repo.DB.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	log.Warnf("*** applying %s", filepath.Base(m.Source))
	start := time.Now()
	if err := repo.InsertUnAppliedVersion(m.Version); err != nil {
		duration := time.Since(start)
		log.Warnf(failedToApplyLogText, filepath.Base(m.Source), duration.Seconds())
		log.Warn("This version is currently being applied by another app")

		return err
	}

	if fn != nil {
		// Run Go migration function.
		if err := fn(tx); err != nil {
			return handleGoFuncError(repo, m, tx, start, fn, failedToApplyLogText)
		}
	}

	if err := repo.UpdateApplyTime(m.Version); err != nil {
		return handleUpdateApplyTimeError(repo, m, tx, start)
	}

	if err := tx.Commit(); err != nil {
		duration := time.Since(start)
		log.Errf(failedToApplyLogText, filepath.Base(m.Source), duration.Seconds())

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

func migrateDownGo(repo *repo.MigrationsRepository, m *Migration) error {
	fn := m.DownFn
	tx, err := repo.DB.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	log.Warnf("*** reverting %s", filepath.Base(m.Source))
	start := time.Now()
	if err := repo.LockVersion(m.Version); err != nil {
		return handleLockVersionError(tx, start, m)
	}

	if fn != nil {
		// Run Go migration function.
		if err := fn(tx); err != nil {
			return handleGoFuncError(repo, m, tx, start, fn, failedToRevertLogText)
		}
	}

	if err := repo.DeleteVersion(m.Version); err != nil {
		return handleDeleteVersionError(tx, start, failedToRevertLogText, m, err)
	}

	if err := tx.Commit(); err != nil {
		duration := time.Since(start)
		log.Errf(failedToRevertLogText, filepath.Base(m.Source), duration.Seconds())

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

func handleUpdateApplyTimeError(repo *repo.MigrationsRepository, m *Migration, tx *sql.Tx, start time.Time) error {
	txErr := tx.Rollback()
	if txErr != nil {
		duration := time.Since(start)
		log.Errf(failedToApplyLogText, filepath.Base(m.Source), duration.Seconds())

		return errors.Wrap(txErr, "update apply time query tx rollback failed")
	}

	tx, err := repo.DB.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin delete version transaction")
	}

	if err := repo.DeleteVersion(m.Version); err != nil {
		txErr := tx.Rollback()
		if txErr != nil {
			duration := time.Since(start)
			log.Errf(failedToApplyLogText, filepath.Base(m.Source), duration.Seconds())

			return errors.Wrap(txErr, "delete version query tx rollback failed")
		}

		duration := time.Since(start)
		log.Errf(failedToApplyLogText, filepath.Base(m.Source), duration.Seconds())

		return errors.Wrap(err, "failed to execute delete version transaction for unapplied version")
	}

	duration := time.Since(start)
	log.Errf(failedToApplyLogText, filepath.Base(m.Source), duration.Seconds())

	return errors.Wrap(err, "failed to execute update apply time transaction")
}

func handleGoFuncError(
	repo *repo.MigrationsRepository,
	m *Migration,
	tx *sql.Tx,
	start time.Time,
	fn func(*sql.Tx) error,
	logText string,
) error {
	txErr := tx.Rollback()
	if txErr != nil {
		duration := time.Since(start)
		log.Errf(logText, filepath.Base(m.Source), duration.Seconds())

		return errors.Wrap(txErr, "failed to rollback failed migration fn() execution")
	}

	tx, err := repo.DB.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	if err := repo.DeleteVersion(m.Version); err != nil {
		return handleDeleteVersionError(tx, start, logText, m, err)
	}

	duration := time.Since(start)
	log.Errf(logText, filepath.Base(m.Source), duration.Seconds())

	return errors.Wrapf(err, "failed to run Go migration function %T", fn)
}

func handleDeleteVersionError(tx *sql.Tx, start time.Time, logText string, m *Migration, err error) error {
	txErr := tx.Rollback()
	if txErr != nil {
		duration := time.Since(start)
		log.Errf(logText, filepath.Base(m.Source), duration.Seconds())

		return errors.Wrap(err, "failed to rollback delete version transaction for unapplied version")
	}

	duration := time.Since(start)
	log.Errf(logText, filepath.Base(m.Source), duration.Seconds())

	return errors.Wrap(err, "failed to execute delete version transaction for unapplied version")
}

func handleLockVersionError(tx *sql.Tx, start time.Time, m *Migration) error {
	txErr := tx.Rollback()
	if txErr != nil {
		duration := time.Since(start)
		log.Errf(failedToRevertLogText, filepath.Base(m.Source), duration.Seconds())

		return errors.Wrap(txErr, "lock version query tx rollback failed")
	}

	duration := time.Since(start)
	log.Warnf(failedToRevertLogText, filepath.Base(m.Source), duration.Seconds())
	log.Warn("This version is currently being reverted by another app")

	return nil
}
