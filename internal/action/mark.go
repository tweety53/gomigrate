package action

import (
	"database/sql"
	"fmt"
	"github.com/tweety53/gomigrate/internal/db"
	errorsInternal "github.com/tweety53/gomigrate/internal/errors"
	"github.com/tweety53/gomigrate/internal/helpers"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/migration"
	"github.com/tweety53/gomigrate/internal/sql_dialect"
	"time"
)

type MarkAction struct {
	db *sql.DB
}

func NewMarkAction(db *sql.DB) *MarkAction {
	return &MarkAction{db: db}
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
	migrations, err := db.GetNewMigrations(a.db)
	if err != nil {
		return err
	}

	for i := range migrations {
		if p.version == migrations[i].Version {
			resp := helpers.AskForConfirmation(fmt.Sprintf("Set migration history at %s?", p.version), false)
			if !resp {
				log.Info("Action was cancelled by user. Nothing has been performed.\n")
				return nil
			}

			for j := 0; j <= i; j++ {
				if _, err := a.db.Exec(sql_dialect.GetDialect().InsertVersionSQL(), migrations[j].Version, int(time.Now().Unix())); err != nil {
					return err
				}
			}
			log.Infof("The migration history is set at %s.\nNo actual migration was performed.\n", p.version)

			return nil
		}
	}

	// try migrate down
	migrationsHistory, err := db.GetMigrationsHistory(a.db, 0)
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
				resp := helpers.AskForConfirmation(fmt.Sprintf("Set migration history at %s?", p.version), false)
				if !resp {
					log.Info("Action was cancelled by user. Nothing has been performed.\n")
					return nil
				}
				for j := 0; j < i; j++ {
					if _, err := a.db.Exec(sql_dialect.GetDialect().DeleteVersionSQL(), migrations[j].Version); err != nil {
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
		resp := helpers.AskForConfirmation(fmt.Sprintf("Set migration history at %s?", p.version), false)
		if !resp {
			log.Info("Action was cancelled by user. Nothing has been performed.\n")
			return nil
		}

		for i := range migrations {
			if _, err := a.db.Exec(sql_dialect.GetDialect().DeleteVersionSQL(), migrations[i].Version); err != nil {
				return err
			}
		}
		log.Infof("The migration history is set at %s.\nNo actual migration was performed.\n", p.version)

		return nil
	}

	return ErrUnableToFindVersion
}
