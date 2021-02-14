package action

import (
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/service"
	errorsInternal "github.com/tweety53/gomigrate/pkg/errors"
	"strconv"
)

type NewAction struct {
	svc *service.MigrationService
}

func NewNewAction(migrationsSvc *service.MigrationService) *NewAction {
	return &NewAction{svc: migrationsSvc}
}

type NewActionParams struct {
	limit int
}

func (p *NewActionParams) ValidateAndFill(args []string) error {
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
		p.limit = 10
	}

	return nil
}

func (p *NewActionParams) Get() interface{} {
	return &NewActionParams{limit: p.limit}
}

func (a *NewAction) Run(params interface{}) error {
	p, ok := params.(*NewActionParams)
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

	n := len(migrations)
	var logText string
	if n == 1 {
		logText = "migration"
	} else {
		logText = "migrations"
	}

	if p.limit > 0 && n > p.limit {
		migrations = migrations[:p.limit]
		log.Warnf("Showing %d out of %d new %s:\n", p.limit, n, logText)
	} else {
		log.Warnf("Found %d new %s:\n", n, logText)
	}

	for _, migration := range migrations {
		log.Printf("\t%s\n", migration.Version)
	}

	return nil
}
