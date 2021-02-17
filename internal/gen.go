//go:generate minimock -g -i github.com/tweety53/gomigrate/internal/repo.MigrationRepo -o repo/migration_repo_minimock.go
//go:generate minimock -g -i github.com/tweety53/gomigrate/internal/repo.DBOperationRepo -o repo/db_operation_repo_minimock.go
//go:generate minimock -g -i github.com/tweety53/gomigrate/internal/migration.MigrationsCollectorInterface -o migration/migrations_collector_interface_minimock.go
//go:generate minimock -g -i github.com/tweety53/gomigrate/internal/migration.RunnerInterface -o migration/runner_interface_minimock.go

package internal
