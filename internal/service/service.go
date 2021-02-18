package service

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/migration"
	"github.com/tweety53/gomigrate/internal/repo"
)

type MigrationService struct {
	DB                  *sql.DB
	MigrationsRepo      repo.MigrationRepo
	DBOperationRepo     repo.DBOperationRepo
	MigrationsPath      string
	MigrationsCollector migration.MigrationsCollectorInterface
}

func NewMigrationService(
	db *sql.DB,
	mRepo repo.MigrationRepo,
	dboRepo repo.DBOperationRepo,
	migrationsCollector migration.MigrationsCollectorInterface,
	migrationsPath string) *MigrationService {
	return &MigrationService{
		DB:                  db,
		MigrationsRepo:      mRepo,
		DBOperationRepo:     dboRepo,
		MigrationsPath:      migrationsPath,
		MigrationsCollector: migrationsCollector,
	}
}

func (s *MigrationService) GetNewMigrations() (migration.Migrations, error) {
	if _, err := s.MigrationsRepo.GetDBVersion(); err != nil {
		return nil, errors.Wrap(err, "cannot get db version")
	}

	records, err := s.MigrationsRepo.GetMigrationsHistory(0)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get migrations history from db")
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
		return nil, errors.Wrapf(err, "cannot collect migration files from path: %s", s.MigrationsPath)
	}

	newMigrations := make(migration.Migrations, 0, len(allMigrations)-len(applied))
	for _, row := range allMigrations {
		if _, ok := applied[row.Version]; ok {
			continue
		}

		newMigrations = append(newMigrations, row)
	}

	return newMigrations, nil
}
