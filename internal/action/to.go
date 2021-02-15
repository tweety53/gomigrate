package action

import (
	"github.com/tweety53/gomigrate/internal/version"
	"strconv"

	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/migration"
	"github.com/tweety53/gomigrate/internal/service"
	errorsInternal "github.com/tweety53/gomigrate/pkg/errors"
)

var ErrUnableToFindVersion = errors.New("unable to find migration with this version")

type ToAction struct {
	svc *service.MigrationService
}

func NewToAction(migrationsSvc *service.MigrationService) *ToAction {
	return &ToAction{svc: migrationsSvc}
}

type ToActionParams struct {
	version string
}

func (p *ToActionParams) ValidateAndFill(args []string) error {
	if len(args) == 0 {
		return errorsInternal.ErrNotEnoughArgs
	}

	// todo: implement all version formats like in yii/migrate???
	if !version.ValidMigrationVersion(args[0]) {
		return errorsInternal.ErrInvalidVersionFormat
	}

	p.version = args[0]

	return nil
}

func (p *ToActionParams) Get() interface{} {
	return &ToActionParams{version: p.version}
}

func (a *ToAction) Run(params interface{}) error {
	p, ok := params.(*ToActionParams)
	if !ok {
		return errorsInternal.ErrInvalidActionParamsType
	}

	// try migrate up
	migrations, err := a.svc.GetNewMigrations()
	if err != nil {
		return err
	}

	for i := range migrations {
		if p.version == migrations[i].Version {
			upAction := NewUpAction(a.svc)
			params := new(UpActionParams)
			if err := params.ValidateAndFill([]string{strconv.Itoa(i + 1)}); err != nil {
				return err
			}
			if err := upAction.Run(params); err != nil {
				return err
			}

			return nil
		}
	}

	// try migrate down
	migrationsHistory, err := a.svc.MigrationsRepo.GetMigrationsHistory(0)
	if err != nil {
		return err
	}

	migrations = migration.Convert(migrationsHistory)

	for i := range migrations {
		if p.version == migrations[i].Version {
			if i != 0 {
				downAction := NewDownAction(a.svc)
				params := new(DownActionParams)
				if err := params.ValidateAndFill([]string{strconv.Itoa(i)}); err != nil {
					return err
				}
				if err := downAction.Run(params); err != nil {
					return err
				}

				return nil
			}

			log.Warnf("Already at '%s'. Nothing needs to be done.\n", p.version)

			return nil
		}
	}

	return ErrUnableToFindVersion
}
