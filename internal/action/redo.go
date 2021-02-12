package action

import (
	"database/sql"
	"fmt"
	"github.com/tweety53/gomigrate/internal/db"
	errors2 "github.com/tweety53/gomigrate/internal/errors"
	"github.com/tweety53/gomigrate/internal/helpers"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/migration"
	"strconv"
)

type RedoAction struct {
	db *sql.DB
}

func NewRedoAction(db *sql.DB) *RedoAction {
	return &RedoAction{db: db}
}

type RedoActionParams struct {
	limit int
}

func (p *RedoActionParams) ValidateAndFill(args []string) error {
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

func (p *RedoActionParams) Get() interface{} {
	return &RedoActionParams{limit: p.limit}
}

func (a *RedoAction) Run(params interface{}) error {
	p, ok := params.(*RedoActionParams)
	if !ok {
		return errors2.ErrInvalidActionParamsType
	}

	migrationsHistory, err := db.GetMigrationsHistory(a.db, p.limit)
	if err != nil {
		return err
	}
	defer migrationsHistory.Close()

	redoMigrations, err := migration.ConvertDbRecordsToMigrationObjects(migrationsHistory)
	if err != nil {
		return err
	}

	if len(redoMigrations) == 0 {
		log.Warn("No migration has been done before.\n")
		return nil
	}

	redoMigrations, err = migration.CollectMigrations(
		"/Users/yuriy.aleksandrov/go/src/gomigrate/migrations",
		migration.GetComparableVersion(redoMigrations[0].Version),
		migration.GetComparableVersion(redoMigrations[len(redoMigrations)-1].Version))
	if len(redoMigrations) == 0 {
		return ErrInconsistentMigrationsData
	}

	n := len(redoMigrations)
	var logText string
	if n == 1 {
		logText = "migration"
	} else {
		logText = "migrations"
	}

	log.Warnf("Total %d %s to be redone:\n", n, logText)
	log.Println(redoMigrations)

	resp := helpers.AskForConfirmation(fmt.Sprintf("Redo the above %s?", logText), false)
	if !resp {
		log.Info("Action was cancelled by user. Nothing has been performed.\n")
		return nil
	}

	// reverse for down
	redoMigrations.Reverse()
	for i := range redoMigrations {
		if err := redoMigrations[i].Down(a.db); err != nil {
			log.Err("\nMigration failed. The rest of the migrations are canceled.\n")
			return err
		}
	}

	// reverse for up
	redoMigrations.Reverse()
	for i := range redoMigrations {
		if err := redoMigrations[i].Up(a.db); err != nil {
			log.Err("\nMigration failed. The rest of the migrations are canceled.\n")
			return err
		}
	}

	if n == 1 {
		logText = "migration was"
	} else {
		logText = "migrations were"
	}

	log.Infof("\n%d %s redone.\n", n, logText)
	log.Info("\nMigration redone successfully.\n")

	return nil
}
