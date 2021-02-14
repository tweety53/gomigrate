package service

import (
	"database/sql"
	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/migration"
	"github.com/tweety53/gomigrate/internal/repo"
)

type MigrationService struct {
	Db                  *sql.DB
	MigrationsRepo      repo.MigrationRepo
	DbOperationRepo     repo.DbOperationRepo
	MigrationsPath      string
	MigrationsCollector migration.MigrationsCollectorInterface
}

func NewMigrationService(
	db *sql.DB,
	mRepo repo.MigrationRepo,
	dboRepo repo.DbOperationRepo,
	migrationsCollector migration.MigrationsCollectorInterface,
	migrationsPath string) *MigrationService {
	return &MigrationService{
		Db:                  db,
		MigrationsRepo:      mRepo,
		DbOperationRepo:     dboRepo,
		MigrationsPath:      migrationsPath,
		MigrationsCollector: migrationsCollector,
	}
}

func (s *MigrationService) GetNewMigrations() (migration.Migrations, error) {
	if _, err := s.MigrationsRepo.GetDBVersion(); err != nil {
		return nil, err
	}

	records, err := s.MigrationsRepo.GetMigrationsHistory(0)
	if err != nil {
		return migration.Migrations{}, errors.Wrap(err, "cannot get migrations history from db")
	}

	applied := make(map[string]int)
	for i := range records {
		// skip base migration
		if records[i].Version == migration.BaseMigrationVersion {
			continue
		}

		applied[records[i].Version] = records[i].ApplyTime
	}

	allMigrations, err := s.MigrationsCollector.CollectMigrations(s.MigrationsPath, 0, 0)
	if err != nil {
		return migration.Migrations{}, errors.Wrapf(err, "cannot collect migration files from path: %s", s.MigrationsPath)
	}

	newCnt := len(allMigrations) - len(applied)
	if newCnt < 0 {
		newCnt = 0
	}
	newMigrations := make(migration.Migrations, 0, newCnt)
	for _, row := range allMigrations {
		if _, ok := applied[row.Version]; ok {
			continue
		}

		newMigrations = append(newMigrations, row)
	}

	return newMigrations, err
}
