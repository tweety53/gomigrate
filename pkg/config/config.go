package config

import (
	"database/sql"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/repo"
	"github.com/tweety53/gomigrate/internal/sqldialect"
	"gopkg.in/yaml.v2"
)

type GoMigrateConfig struct {
	isValid        bool
	Compact        bool   `yaml:"gomigrate_compact"`
	MigrationsPath string `yaml:"gomigrate_migrations_path"`
	MigrationTable string `yaml:"gomigrate_migration_table"`
	SQLDialect     string `yaml:"gomigrate_sql_dialect"`
	DataSourceName string `yaml:"gomigrate_dsn"`
}

func (c *GoMigrateConfig) IsValid() bool {
	return c.isValid
}

func BuildFromFile(f string) (*GoMigrateConfig, error) {
	conf := &GoMigrateConfig{}

	yamlConf, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, errors.Wrap(err, "gomigrate yaml conf get err")
	}

	// expand environment variables
	yamlConf = []byte(os.ExpandEnv(string(yamlConf)))

	err = yaml.Unmarshal(yamlConf, conf)
	if err != nil {
		return nil, errors.Wrap(err, "gomigrate yaml conf unmarshal err")
	}

	return conf, nil
}

func BuildFromArgs(
	migrationsPath string,
	migrationTable string,
	compact bool,
	sqlDialect string,
	dataSourceName string,
) *GoMigrateConfig {
	return &GoMigrateConfig{
		MigrationsPath: migrationsPath,
		MigrationTable: migrationTable,
		Compact:        compact,
		SQLDialect:     sqlDialect,
		DataSourceName: dataSourceName,
	}
}

func Validate(conf *GoMigrateConfig, db *sql.DB) error {
	if _, err := os.Stat(conf.MigrationsPath); err != nil {
		return errors.Wrap(err, "gomigrate config: bad migrations path")
	}

	dialect, err := sqldialect.InitDialect(conf.SQLDialect, conf.MigrationTable)
	if err != nil {
		return errors.Wrap(err, "gomigrate config: unknown sql dialect")
	}

	mRepo := repo.NewMigrationsRepository(db, dialect)
	if _, err := mRepo.EnsureDBVersion(); err != nil {
		return errors.Wrap(err, "gomigrate config: cannot check/create migrations table in DB")
	}
	conf.isValid = true

	return nil
}
