package action

import (
	"fmt"

	"github.com/tweety53/gomigrate/internal/helpers"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/migration"
	"github.com/tweety53/gomigrate/internal/service"
	errorsInternal "github.com/tweety53/gomigrate/pkg/errors"
)

type MarkAction struct {
	svc *service.MigrationService
}

func NewMarkAction(migrationsSvc *service.MigrationService) *MarkAction {
	return &MarkAction{svc: migrationsSvc}
}

type MarkActionParams struct {
	version string
}

func (p *MarkActionParams) ValidateAndFill(args []string) error {
	if len(args) == 0 {
		return errorsInternal.ErrNotEnoughArgs
	}

	// todo: implement all version formats like in yii/migrate???
	if !helpers.ValidMigrationVersion(args[0]) {
		return errorsInternal.ErrInvalidVersionFormat
	}

	p.version = args[0]

	return nil
}

func (p *MarkActionParams) Get() interface{} {
	return &MarkActionParams{version: p.version}
}

func (a *MarkAction) Run(params interface{}) error {
	p, ok := params.(*MarkActionParams)
	if !ok {
		return errorsInternal.ErrInvalidActionParamsType
	}

	resp := helpers.AskForConfirmation(fmt.Sprintf("Set migration history at %s?", p.version))
	if !resp {
		log.Info("Action was cancelled by user. Nothing has been performed.\n")
		return nil
	}

	// try mark up
	migrations, err := a.svc.GetNewMigrations()
	if err != nil {
		return err
	}

	for i := range migrations {
		if p.version == migrations[i].Version {
			return markUp(i, a, migrations, p)
		}
	}

	// try mark down
	migrationsHistory, err := a.svc.MigrationsRepo.GetMigrationsHistory(0)
	if err != nil {
		return err
	}

	migrations = migration.Convert(migrationsHistory)

	for i := range migrations {
		if p.version == migrations[i].Version {
			return markDown(i, a, migrations, p)
		}
	}

	if p.version == migration.BaseMigrationVersion {
		return markToBaseVersion(migrations, a, p)
	}

	return ErrUnableToFindVersion
}

func markDown(i int, a *MarkAction, migrations migration.Migrations, p *MarkActionParams) error {
	if i != 0 {
		for j := 0; j < i; j++ {
			if err := a.svc.MigrationsRepo.DeleteVersion(migrations[j].Version); err != nil {
				return err
			}
		}
		log.Infof("The migration history is set at %s.\nNo actual migration was performed.\n", p.version)
	} else {
		log.Warnf("Already at '%s'. Nothing needs to be done.\n", p.version)
	}

	return nil
}

func markUp(i int, a *MarkAction, migrations migration.Migrations, p *MarkActionParams) error {
	for j := 0; j <= i; j++ {
		if err := a.svc.MigrationsRepo.InsertVersion(migrations[j].Version); err != nil {
			return err
		}
	}
	log.Infof("The migration history is set at %s.\nNo actual migration was performed.\n", p.version)

	return nil
}

func markToBaseVersion(migrations migration.Migrations, a *MarkAction, p *MarkActionParams) error {
	for i := range migrations {
		if err := a.svc.MigrationsRepo.DeleteVersion(migrations[i].Version); err != nil {
			return err
		}
	}
	log.Infof("The migration history is set at %s.\nNo actual migration was performed.\n", p.version)

	return nil
}
