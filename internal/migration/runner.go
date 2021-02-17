package migration

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
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

type RunnerInterface interface {
	MigrateUp(repo repo.MigrationRepo, m *Migration) error
	MigrateUpSafe(repo repo.MigrationRepo, m *Migration) error
	MigrateDown(repo repo.MigrationRepo, m *Migration) error
	MigrateDownSafe(repo repo.MigrationRepo, m *Migration) error
}

type Runner struct{}

//nolint:dupl // because its lie :)
func (r *Runner) MigrateUp(repo repo.MigrationRepo, m *Migration) error {
	fn := m.UpFn
	log.Warnf("***[NON-TRANSACTIONAL] applying %s", filepath.Base(m.Source))
	start := time.Now()
	if fn != nil {
		err := migrateNoTx(repo, m, start, fn, failedToApplyLogText)
		if err != nil {
			return err
		}

		if err := repo.InsertVersion(m.Version); err != nil {
			duration := time.Since(start)
			log.Errf(failedToApplyLogText, filepath.Base(m.Source), duration.Seconds())

			return errors.Wrap(err, "failed to insert migration version")
		}

		duration := time.Since(start)
		log.Infof("*** applied %s (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())

		return nil
	}

	duration := time.Since(start)
	log.Warnf("*** NOT applied %s (empty fn()) (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())

	return nil
}

//nolint:nestif //im tired
func (r *Runner) MigrateUpSafe(repo repo.MigrationRepo, m *Migration) error {
	fn := m.SafeUpFn

	log.Warnf("***[TRANSACTIONAL] applying %s", filepath.Base(m.Source))
	start := time.Now()
	if fn != nil {
		db, err := repo.GetDB()
		if err != nil {
			duration := time.Since(start)
			log.Errf(failedToApplyLogText, filepath.Base(m.Source), duration.Seconds())

			return errors.Wrap(err, "db not initialized")
		}

		tx, err := db.Begin()
		if err != nil {
			duration := time.Since(start)
			log.Errf(failedToApplyLogText, filepath.Base(m.Source), duration.Seconds())

			return errors.Wrap(err, "failed to begin transaction")
		}

		if err := repo.InsertUnAppliedVersion(m.Version); err != nil {
			return handleInsertUnappliedVersionError(tx, start, m, err)
		}

		// Run Go migration function.
		if err := fn(tx); err != nil {
			return handleGoFuncError(repo, m, tx, start, fn, failedToApplyLogText)
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
		log.Infof("*** applied %s (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())

		return nil
	}

	duration := time.Since(start)
	log.Warnf("*** NOT applied %s (empty fn()) (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())

	return nil
}

func handleInsertUnappliedVersionError(tx *sql.Tx, start time.Time, m *Migration, err error) error {
	txErr := tx.Rollback()
	if txErr != nil {
		duration := time.Since(start)
		log.Warnf(failedToApplyLogText, filepath.Base(m.Source), duration.Seconds())
		log.Warn("This version is currently being applied by another app")

		return errors.Wrap(txErr, "insert unapplied version query tx rollback failed")
	}

	duration := time.Since(start)
	log.Warnf(failedToApplyLogText, filepath.Base(m.Source), duration.Seconds())
	log.Warn("This version is currently being applied by another app")

	return errors.Wrap(err, "gomigrate runner: cant migrate")
}

//nolint:dupl // because its lie :)
func (r *Runner) MigrateDown(repo repo.MigrationRepo, m *Migration) error {
	fn := m.DownFn
	log.Warnf("***[NON-TRANSACTIONAL] reverting %s", filepath.Base(m.Source))
	start := time.Now()
	if fn != nil {
		err := migrateNoTx(repo, m, start, fn, failedToRevertLogText)
		if err != nil {
			return err
		}

		if err := repo.DeleteVersion(m.Version); err != nil {
			duration := time.Since(start)
			log.Errf(failedToRevertLogText, filepath.Base(m.Source), duration.Seconds())

			return errors.Wrap(err, "failed to delete migration version")
		}

		duration := time.Since(start)
		log.Infof("*** reverted %s (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())

		return nil
	}

	duration := time.Since(start)
	log.Warnf("*** NOT reverted %s (empty fn()) (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())

	return nil
}

func migrateNoTx(repo repo.MigrationRepo, m *Migration, start time.Time, fn func(*sql.DB) error, logText string) error {
	db, err := repo.GetDB()
	if err != nil {
		duration := time.Since(start)
		log.Errf(logText, filepath.Base(m.Source), duration.Seconds())

		return errors.Wrap(err, "gomigrate runner: cant migrate")
	}

	if err := fn(db); err != nil {
		duration := time.Since(start)
		log.Errf(logText, filepath.Base(m.Source), duration.Seconds())

		return errors.Wrap(err, "failed to execute go fn()")
	}

	return nil
}

//nolint:nestif // im tired
func (r *Runner) MigrateDownSafe(repo repo.MigrationRepo, m *Migration) error {
	fn := m.SafeDownFn

	log.Warnf("***[TRANSACTIONAL] reverting %s", filepath.Base(m.Source))
	start := time.Now()
	if fn != nil {
		db, err := repo.GetDB()
		if err != nil {
			duration := time.Since(start)
			log.Errf(failedToRevertLogText, filepath.Base(m.Source), duration.Seconds())

			return errors.Wrap(err, "gomigrate runner: cant migrate")
		}

		tx, err := db.Begin()
		if err != nil {
			duration := time.Since(start)
			log.Errf(failedToRevertLogText, filepath.Base(m.Source), duration.Seconds())

			return errors.Wrap(err, "failed to begin transaction")
		}

		if err := repo.LockVersion(m.Version); err != nil {
			return handleLockVersionError(tx, start, m, err)
		}

		// Run Go migration function.
		if err := fn(tx); err != nil {
			return handleGoFuncError(repo, m, tx, start, fn, failedToRevertLogText)
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
		log.Infof("*** reverted %s (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())

		return nil
	}

	duration := time.Since(start)
	log.Warnf("*** NOT reverted %s (empty fn()) (time: %.3f sec.)\n", filepath.Base(m.Source), duration.Seconds())

	return nil
}

func handleUpdateApplyTimeError(repo repo.MigrationRepo, m *Migration, tx driver.Tx, start time.Time) error {
	txErr := tx.Rollback()
	if txErr != nil {
		duration := time.Since(start)
		log.Errf(failedToApplyLogText, filepath.Base(m.Source), duration.Seconds())

		return errors.Wrap(txErr, "update apply time query tx rollback failed")
	}

	db, err := repo.GetDB()
	if err != nil {
		return err
	}

	tx, err = db.Begin()
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

	if err := tx.Commit(); err != nil {
		duration := time.Since(start)
		log.Errf(failedToApplyLogText, filepath.Base(m.Source), duration.Seconds())

		return errors.Wrap(err, "failed to commit delete version transaction")
	}

	duration := time.Since(start)
	log.Errf(failedToApplyLogText, filepath.Base(m.Source), duration.Seconds())

	return errors.New("failed to execute update apply time transaction")
}

func handleGoFuncError(
	repo repo.MigrationRepo,
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

	db, err := repo.GetDB()
	if err != nil {
		return err
	}

	tx, err = db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	if err := repo.DeleteVersion(m.Version); err != nil {
		return handleDeleteVersionError(tx, start, logText, m, err)
	}

	if err := tx.Commit(); err != nil {
		duration := time.Since(start)
		log.Errf(failedToApplyLogText, filepath.Base(m.Source), duration.Seconds())

		return errors.Wrap(err, "failed to commit delete version transaction")
	}

	duration := time.Since(start)
	log.Errf(logText, filepath.Base(m.Source), duration.Seconds())

	return errors.New(fmt.Sprintf("failed to run Go migration function %T", fn))
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

func handleLockVersionError(tx *sql.Tx, start time.Time, m *Migration, err error) error {
	txErr := tx.Rollback()
	if txErr != nil {
		duration := time.Since(start)
		log.Errf(failedToRevertLogText, filepath.Base(m.Source), duration.Seconds())
		log.Warn("This version is currently being reverted by another app")

		return errors.Wrap(txErr, "lock version query tx rollback failed")
	}

	duration := time.Since(start)
	log.Warnf(failedToRevertLogText, filepath.Base(m.Source), duration.Seconds())
	log.Warn("This version is currently being reverted by another app")

	return err
}
