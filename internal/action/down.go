package action

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/helpers"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/migration"
	"github.com/tweety53/gomigrate/internal/repo"
	"github.com/tweety53/gomigrate/internal/service"
	errorsInternal "github.com/tweety53/gomigrate/pkg/errors"
)

var ErrInconsistentMigrationsData = errors.New("migrations data in repo and migration files path inconsistent, please check")

type DownAction struct {
	svc *service.MigrationService
}

func NewDownAction(migrationsSvc *service.MigrationService) *DownAction {
	return &DownAction{svc: migrationsSvc}
}

type DownActionParams struct {
	limit int
}

func (p *DownActionParams) Get() interface{} {
	return &DownActionParams{limit: p.limit}
}

func (p *DownActionParams) ValidateAndFill(args []string) error {
	if len(args) > 0 {
		if args[0] == helpers.LimitAll {
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

func (a *DownAction) Run(params interface{}) error {
	p, ok := params.(*DownActionParams)
	if !ok {
		return errorsInternal.ErrInvalidActionParamsType
	}

	migrationHistoryRecords, err := a.svc.MigrationsRepo.GetMigrationsHistory(p.limit)
	if err != nil {
		return err
	}

	downMigrations := migration.Convert(migrationHistoryRecords)

	if len(downMigrations) == 0 {
		log.Warn("No migration has been done before.\n")

		return nil
	}

	downMigrations, err = a.svc.MigrationsCollector.CollectMigrations(
		a.svc.MigrationsPath,
		migration.GetComparableVersion(downMigrations[0].Version),
		migration.GetComparableVersion(downMigrations[len(downMigrations)-1].Version))
	if err != nil {
		return err
	}

	if len(downMigrations) == 0 {
		return ErrInconsistentMigrationsData
	}

	downMigrations.Reverse()
	n := len(downMigrations)

	log.Warnf("Total %d %s to be reverted:\n", n, helpers.ChooseLogText(n, true))
	log.Infof("%s", downMigrations)

	var reverted int
	for i := range downMigrations {
		r, ok := a.svc.MigrationsRepo.(*repo.MigrationsRepository)
		if !ok {
			return errors.New("MigrationRepo type assertion err")
		}

		if err = downMigrations[i].Down(r); err != nil {
			log.Errf("\n%d from %d %s reverted.\n", reverted, n, helpers.ChooseLogText(reverted, false))

			return err
		}

		reverted++
	}

	log.Infof("\n%d reverted.\n", n)
	log.Info("\nMigrated down successfully.\n")

	return nil
}
