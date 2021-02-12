package config

type AppConfig struct {
	MigrationsPath string `yaml:"gomigrate_migrations_path"`
	MigrationTable string `yaml:"gomigrate_migration_table"`
	Compact        bool   `yaml:"gomigrate_compact"`
	SQLDialect     string `yaml:"gomigrate_sql_dialect"`
	DataSourceName string `yaml:"gomigrate_dsn"`
}
