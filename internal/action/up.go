package action

import (
	"database/sql"
	"fmt"
	errors2 "gomigrate/internal/errors"
	"gomigrate/internal/log"
	"gomigrate/internal/migration"
	"strconv"
)

type UpAction struct {
	db *sql.DB
}

type UpActionParams struct {
	limit int
}

func (p *UpActionParams) Get() interface{} {
	return &UpActionParams{limit: p.limit}
}

func (p *UpActionParams) ValidateAndFill(args []string) error {
	if len(args) > 0 {
		var err error
		p.limit, err = strconv.Atoi(args[0])
		if err != nil {
			return err
		}
	}

	return nil
}

func NewUpAction(db *sql.DB) *UpAction {
	return &UpAction{db: db}
}

func (a *UpAction) Run(params interface{}) error {
	p, ok := params.(*UpActionParams)
	if !ok {
		return errors2.ErrInvalidActionParamsType
	}

	migrations, err := migration.GetNewMigrations(a.db)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		log.Info("No new migrations found. Your system is up-to-date.\n")
		return nil
	}

	total := len(migrations)
	if p.limit > 0 {
		if p.limit > len(migrations) {
			p.limit = len(migrations)
		}
		migrations = migrations[0:p.limit]
	}

	fmt.Println(migrations)

	var logText string

	n := len(migrations)
	if n == total {
		if n == 1 {
			logText = "migration"
		} else {
			logText = "migrations"
		}

		log.Warnf("Total %d new %s to be applied:\n", n, logText)
	} else {
		if total == 1 {
			logText = "migration"
		} else {
			logText = "migrations"
		}

		log.Warnf("Total %d out of %d new %s to be applied:\n", n, total, logText)
	}

	for i := range migrations {
		log.Infof("\t%s\n", migrations[i])
	}

	var applied int
	for i := range migrations {
		if err = migrations[i].Up(a.db); err != nil {
			if applied == 1 {
				logText = "migration was"
			} else {
				logText = "migrations were"
			}
			log.Errf("\n%d from %d %s applied.\n", applied, n, logText)

			return err
		}

		applied++
	}

	if n == 1 {
		logText = "migration was"
	} else {
		logText = "migrations were"
	}
	log.Infof("\n%d applied.\n", n)
	log.Info("\nMigrated up successfully.\n")

	return nil
}
