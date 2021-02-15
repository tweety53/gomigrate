package action

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/helpers"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/repo"
	"github.com/tweety53/gomigrate/internal/service"
	errorsInternal "github.com/tweety53/gomigrate/pkg/errors"
)

type UpAction struct {
	svc *service.MigrationService
}

func NewUpAction(migrationsSvc *service.MigrationService) *UpAction {
	return &UpAction{svc: migrationsSvc}
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

func (a *UpAction) Run(params interface{}) error {
	p, ok := params.(*UpActionParams)
	if !ok {
		return errorsInternal.ErrInvalidActionParamsType
	}

	migrations, err := a.svc.GetNewMigrations()
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

	var logText string

	n := len(migrations)
	if n == total {
		logText = helpers.ChooseLogText(n, true)
		log.Warnf("Total %d new %s to be applied:\n", n, logText)
	} else {
		logText = helpers.ChooseLogText(total, true)
		log.Warnf("Total %d out of %d new %s to be applied:\n", n, total, logText)
	}

	log.Infof("%s", migrations)

	var applied int
	for i := range migrations {
		r, ok := a.svc.MigrationsRepo.(*repo.MigrationsRepository)
		if !ok {
			return errors.New("MigrationRepo type assertion err")
		}

		if err = migrations[i].Up(r); err != nil {
			log.Errf("\n%d from %d %s applied.\n", applied, n, logText)
			log.Err("\nMigration failed. The rest of the migrations are canceled.\n")
			return err
		}

		logText = helpers.ChooseLogText(applied, false)

		applied++
	}

	log.Infof("\n%d applied.\n", n)
	log.Info("\nMigrated up successfully.\n")

	return nil
}
