package action

import (
	"database/sql"
	"github.com/pkg/errors"
	errorsInternal "github.com/tweety53/gomigrate/internal/errors"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/migration"
	"github.com/tweety53/gomigrate/internal/repo"
	"regexp"
	"strconv"
)

var ErrUnableToFindVersion = errors.New("unable to find migration with this version")

type ToAction struct {
	db             *sql.DB
	migrationsPath string
}

func NewToAction(db *sql.DB, migrationsPath string) *ToAction {
	return &ToAction{db: db, migrationsPath: migrationsPath}
}

type ToActionParams struct {
	version string
}

var versionRegex = regexp.MustCompile("^(m(\\d{6}_?\\d{6})(\\D*))")

func (p *ToActionParams) ValidateAndFill(args []string) error {
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

func (p *ToActionParams) Get() interface{} {
	return &ToActionParams{version: p.version}
}

func (a *ToAction) Run(params interface{}) error {
	p, ok := params.(*ToActionParams)
	if !ok {
		return errorsInternal.ErrInvalidActionParamsType
	}

	// try migrate up
	migrations, err := repo.GetNewMigrations(a.db, a.migrationsPath)
	if err != nil {
		return err
	}

	for i := range migrations {
		if p.version == migrations[i].Version {
			upAction := NewUpAction(a.db, a.migrationsPath)
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
	migrationsHistory, err := repo.GetMigrationsHistory(a.db, 0)
	if err != nil {
		return err
	}

	migrations, err = migration.ConvertDbRecordsToMigrationObjects(migrationsHistory)
	if err != nil {
		return err
	}

	for i := range migrations {
		if p.version == migrations[i].Version {
			if i != 0 {
				downAction := NewDownAction(a.db, a.migrationsPath)
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
