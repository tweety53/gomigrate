package action

import (
	"strconv"
	"time"

	"github.com/tweety53/gomigrate/internal/helpers"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/service"
	errorsInternal "github.com/tweety53/gomigrate/pkg/errors"
)

type HistoryAction struct {
	svc *service.MigrationService
}

func NewHistoryAction(migrationsSvc *service.MigrationService) *HistoryAction {
	return &HistoryAction{svc: migrationsSvc}
}

type HistoryActionParams struct {
	limit int
}

func (p *HistoryActionParams) ValidateAndFill(args []string) error {
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
		p.limit = 10
	}

	return nil
}

func (p *HistoryActionParams) Get() interface{} {
	return &HistoryActionParams{limit: p.limit}
}

func (a *HistoryAction) Run(params interface{}) error {
	p, ok := params.(*HistoryActionParams)
	if !ok {
		return errorsInternal.ErrInvalidActionParamsType
	}

	migrationRecords, err := a.svc.MigrationsRepo.GetMigrationsHistory(p.limit)
	if err != nil {
		return err
	}

	if len(migrationRecords) == 0 {
		log.Warn("No migration has been done before.\n")

		return nil
	}

	n := len(migrationRecords)
	if p.limit > 0 {
		log.Warnf("Showing the last %d applied %s:\n", n, helpers.ChooseLogText(n, true))
	} else {
		log.Warnf("Total %d %s been applied before:\n", n, helpers.ChooseLogText(n, false))
	}

	const timeFormat = "06-01-02 15:04:05"
	for _, record := range migrationRecords {
		t := time.Unix(int64(record.ApplyTime), 0)
		log.Printf("\t(%s) %s\n", t.Format(timeFormat), record.Version)
	}

	return nil
}
