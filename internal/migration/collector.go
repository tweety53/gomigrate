package migration

import (
	"fmt"
	"os"
	"path/filepath"
)

// CollectMigrations returns all the valid looking migration scripts in the
// migrations folder and go func registry, and key them by version.
func CollectMigrations(dirpath string, current, target int) (Migrations, error) {
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s directory does not exist", dirpath)
	}

	var migrations Migrations

	// SQL migration files.
	sqlMigrationFiles, err := filepath.Glob(dirpath + "/*.sql")
	if err != nil {
		return nil, err
	}
	for _, file := range sqlMigrationFiles {
		v, err := GetVersionFromFileName(file)
		if err != nil {
			return nil, err
		}

		if versionInRange(GetComparableVersion(v), current, target) {
			migration := &Migration{Version: v, Next: "", Previous: "", Source: file}
			migrations = append(migrations, migration)
		}
	}

	// Go migrations registered via AddMigration().
	for _, migration := range registeredMigrations {
		v, err := GetVersionFromFileName(migration.Source)
		if err != nil {
			return nil, err
		}

		if versionInRange(GetComparableVersion(v), current, target) {
			migrations = append(migrations, migration)
		}
	}

	// Go migration files
	goMigrationFiles, err := filepath.Glob(dirpath + "/*.go")
	if err != nil {
		return nil, err
	}
	for _, file := range goMigrationFiles {
		v, err := GetVersionFromFileName(file)
		if err != nil {
			return nil, err
		}

		// Skip migrations already existing migrations registered via goose.AddMigration().
		if _, ok := registeredMigrations[v]; ok {
			continue
		}

		if versionInRange(GetComparableVersion(v), current, target) {
			migration := &Migration{Version: v, Next: "", Previous: "", Source: file, Registered: false}
			migrations = append(migrations, migration)
		}
	}

	migrations = sortAndConnectMigrations(migrations)

	return migrations, nil
}

func versionInRange(v, current, target int) bool {
	if current == target && current == v && target == v {
		return true
	}

	if target == 0 && current == 0 {
		return true
	}

	if target > current {
		return v >= current && v <= target
	}

	if target < current {
		return v <= current && v >= target
	}

	return false
}
