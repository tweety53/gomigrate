package action

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/helpers"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/migration"
	"github.com/tweety53/gomigrate/internal/repo"
	"github.com/tweety53/gomigrate/internal/service"
	errorsInternal "github.com/tweety53/gomigrate/pkg/errors"
)

type RedoAction struct {
	svc *service.MigrationService
}

func NewRedoAction(migrationsSvc *service.MigrationService) *RedoAction {
	return &RedoAction{svc: migrationsSvc}
}

type RedoActionParams struct {
	limit int
}

func (p *RedoActionParams) ValidateAndFill(args []string) error {
	if len(args) > 0 {
		if args[0] == helpers.LimitAll {
			p.limit = 0

			return nil
		}

		var err error
		p.limit, err = strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		return nil
	}

	p.limit = 1

	return nil
}

func (p *RedoActionParams) Get() interface{} {
	return &RedoActionParams{limit: p.limit}
}

func (a *RedoAction) Run(params interface{}) error {
	p, ok := params.(*RedoActionParams)
	if !ok {
		return errorsInternal.ErrInvalidActionParamsType
	}

	migrationHistoryRecords, err := a.svc.MigrationsRepo.GetMigrationsHistory(p.limit)
	if err != nil {
		return err
	}

	redoMigrations := migration.Convert(migrationHistoryRecords)

	if len(redoMigrations) == 0 {
		log.Warn("No migration has been done before.\n")

		return nil
	}

	redoMigrations, err = a.svc.MigrationsCollector.CollectMigrations(
		a.svc.MigrationsPath,
		migration.GetComparableVersion(redoMigrations[0].Version),
		migration.GetComparableVersion(redoMigrations[len(redoMigrations)-1].Version))
	if err != nil {
		return err
	}

	if len(redoMigrations) == 0 {
		return ErrInconsistentMigrationsData
	}

	var logText string
	n := len(redoMigrations)

	logText = helpers.ChooseLogText(n, true)
	log.Warnf("Total %d %s to be redone:\n", n, logText)

	log.Println(redoMigrations)

	resp := helpers.AskForConfirmation(fmt.Sprintf("Redo the above %s?", logText))
	if !resp {
		log.Info("Action was cancelled by user. Nothing has been performed.\n")

		return nil
	}

	r, ok := a.svc.MigrationsRepo.(*repo.MigrationsRepository)
	if !ok {
		return errors.New("MigrationRepo type assertion err")
	}

	// reverse for down
	redoMigrations = redoMigrations.Reverse()
	for i := range redoMigrations {
		if err := redoMigrations[i].Down(r, &migration.Runner{}); err != nil {
			log.Err("\nMigration failed. The rest of the migrations are canceled.\n")

			return err
		}
	}

	// reverse for up
	redoMigrations = redoMigrations.Reverse()
	for i := range redoMigrations {
		if err := redoMigrations[i].Up(r, &migration.Runner{}); err != nil {
			log.Err("\nMigration failed. The rest of the migrations are canceled.\n")

			return err
		}
	}

	log.Infof("\n%d %s redone.\n", n, helpers.ChooseLogText(n, false))
	log.Info("\nMigration redone successfully.\n")

	return nil
}
