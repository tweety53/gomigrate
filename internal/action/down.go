package action

import (
	"database/sql"
	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/db"
	errors2 "github.com/tweety53/gomigrate/internal/errors"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/migration"
	"strconv"
)

var ErrInconsistentMigrationsData = errors.New("migrations data in db and migration files path inconsistent, please check")

type DownAction struct {
	db *sql.DB
}

type DownActionParams struct {
	limit int
}

func (p *DownActionParams) Get() interface{} {
	return &DownActionParams{limit: p.limit}
}

func (p *DownActionParams) ValidateAndFill(args []string) error {
	if len(args) > 0 {
		if args[0] == "all" {
			p.limit = 0
		} else {
			var err error
			p.limit, err = strconv.Atoi(args[0])
			if err != nil {
				return err
			}
		}
	} else {
		p.limit = 1
	}

	return nil
}

func NewDownAction(db *sql.DB) *DownAction {
	return &DownAction{db: db}
}

func (a *DownAction) Run(params interface{}) error {
	p, ok := params.(*DownActionParams)
	if !ok {
		return errors2.ErrInvalidActionParamsType
	}

	migrationsHistory, err := db.GetMigrationsHistory(a.db, p.limit)
	if err != nil {
		return err
	}

	downMigrations, err := migration.ConvertDbRecordsToMigrationObjects(migrationsHistory)
	if err != nil {
		return err
	}

	if len(downMigrations) == 0 {
		log.Warn("No migration has been done before.\n")
		return nil
	}

	downMigrations, err = migration.CollectMigrations(
		"/Users/yuriy.aleksandrov/go/src/gomigrate/migrations",
		migration.GetComparableVersion(downMigrations[0].Version),
		migration.GetComparableVersion(downMigrations[len(downMigrations)-1].Version))

	if len(downMigrations) == 0 {
		return ErrInconsistentMigrationsData
	}

	downMigrations.Reverse()
	var logText string
	n := len(downMigrations)
	if n == 1 {
		logText = "migration"
	} else {
		logText = "migrations"
	}

	log.Warnf("Total %d %s to be reverted:\n", n, logText)
	log.Infof("%s", downMigrations)

	var reverted int
	for i := range downMigrations {
		if err = downMigrations[i].Down(a.db); err != nil {
			if reverted == 1 {
				logText = "migration was"
			} else {
				logText = "migrations were"
			}
			log.Errf("\n%d from %d %s reverted.\n", reverted, n, logText)

			return err
		}

		reverted++
	}

	if n == 1 {
		logText = "migration was"
	} else {
		logText = "migrations were"
	}
	log.Infof("\n%d reverted.\n", n)
	log.Info("\nMigrated down successfully.\n")

	return nil
}
