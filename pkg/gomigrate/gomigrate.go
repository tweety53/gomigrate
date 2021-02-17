package gomigrate

import (
	"database/sql"
	"runtime"

	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/action"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/migration"
	"github.com/tweety53/gomigrate/internal/repo"
	"github.com/tweety53/gomigrate/internal/service"
	"github.com/tweety53/gomigrate/internal/sqldialect"
	"github.com/tweety53/gomigrate/pkg/config"
)

func Run(a string, db *sql.DB, config *config.GoMigrateConfig, args []string) error {
	log.SetVerbose(!config.Compact)

	dialect, err := sqldialect.InitDialect(config.SQLDialect, config.MigrationTable)
	if err != nil {
		return err
	}

	migrationsSvc := service.NewMigrationService(
		db,
		repo.NewMigrationsRepository(db, dialect),
		repo.NewDBOperationsRepository(db, dialect),
		&migration.MigrationsCollector{},
		config.MigrationsPath)

	var (
		act    action.Action
		params action.Params
	)
	switch a {
	case "create":
		act = action.NewCreateAction(config.MigrationsPath)
		params = new(action.CreateActionParams)
	case "down":
		act = action.NewDownAction(migrationsSvc)
		params = new(action.DownActionParams)
	case "fresh":
		act = action.NewFreshAction(migrationsSvc)
		params = new(action.FreshActionParams)
	case "history":
		act = action.NewHistoryAction(migrationsSvc)
		params = new(action.HistoryActionParams)
	case "mark":
		act = action.NewMarkAction(migrationsSvc)
		params = new(action.MarkActionParams)
	case "new":
		act = action.NewNewAction(migrationsSvc)
		params = new(action.NewActionParams)
	case "redo":
		act = action.NewRedoAction(migrationsSvc)
		params = new(action.RedoActionParams)
	case "to":
		act = action.NewToAction(migrationsSvc)
		params = new(action.ToActionParams)
	case "up":
		act = action.NewUpAction(migrationsSvc)
		params = new(action.UpActionParams)
	default:
		return errors.Wrap(errors.New("no such action, run with -h flag to see help"), a)
	}

	if err := params.ValidateAndFill(args); err != nil {
		return err
	}
	if err := act.Run(params); err != nil {
		return err
	}

	return nil
}

func AddSafeMigration(up func(*sql.Tx) error, down func(*sql.Tx) error) {
	_, filename, _, _ := runtime.Caller(1) //nolint:dogsled
	migration.AddSafeNamedMigration(filename, up, down)
}

func AddMigration(up func(*sql.DB) error, down func(*sql.DB) error) {
	_, filename, _, _ := runtime.Caller(1) //nolint:dogsled
	migration.AddNamedMigration(filename, up, down)
}
