package config

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type GoMigrateConfig struct {
	MigrationsPath string `yaml:"gomigrate_migrations_path"`
	MigrationTable string `yaml:"gomigrate_migration_table"`
	Compact        bool   `yaml:"gomigrate_compact"`
	SQLDialect     string `yaml:"gomigrate_sql_dialect"`
	DataSourceName string `yaml:"gomigrate_dsn"`
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
) (*GoMigrateConfig, error) {
	return &GoMigrateConfig{
		MigrationsPath: migrationsPath,
		MigrationTable: migrationTable,
		Compact:        compact,
		SQLDialect:     sqlDialect,
		DataSourceName: dataSourceName,
	}, nil
}
