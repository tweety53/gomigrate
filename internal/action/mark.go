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

	//todo: implement all version formats like in yii/migrate???
	if !versionRegex.MatchString(args[0]) {
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

	// try mark up
	migrations, err := a.svc.GetNewMigrations()
	if err != nil {
		return err
	}

	for i := range migrations {
		if p.version == migrations[i].Version {
			resp := helpers.AskForConfirmation(fmt.Sprintf("Set migration history at %s?", p.version))
			if !resp {
				log.Info("Action was cancelled by user. Nothing has been performed.\n")
				return nil
			}

			for j := 0; j <= i; j++ {
				if err := a.svc.MigrationsRepo.InsertVersion(migrations[j].Version); err != nil {
					return err
				}
			}
			log.Infof("The migration history is set at %s.\nNo actual migration was performed.\n", p.version)

			return nil
		}
	}

	// try migrate down
	migrationsHistory, err := a.svc.MigrationsRepo.GetMigrationsHistory(0)
	if err != nil {
		return err
	}

	migrations, err = migration.Convert(migrationsHistory)
	if err != nil {
		return err
	}

	for i := range migrations {
		if p.version == migrations[i].Version {
			if i != 0 {
				resp := helpers.AskForConfirmation(fmt.Sprintf("Set migration history at %s?", p.version))
				if !resp {
					log.Info("Action was cancelled by user. Nothing has been performed.\n")
					return nil
				}
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
	}

	if p.version == migration.BaseMigrationVersion {
		resp := helpers.AskForConfirmation(fmt.Sprintf("Set migration history at %s?", p.version))
		if !resp {
			log.Info("Action was cancelled by user. Nothing has been performed.\n")
			return nil
		}

		for i := range migrations {
			if err := a.svc.MigrationsRepo.DeleteVersion(migrations[i].Version); err != nil {
				return err
			}
		}
		log.Infof("The migration history is set at %s.\nNo actual migration was performed.\n", p.version)

		return nil
	}

	return ErrUnableToFindVersion
}
