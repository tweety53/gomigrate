// +build test_integration

package tests

import (
	"database/sql"
	"log"
	"os"
	"os/exec"

	_ "github.com/lib/pq"
	"github.com/tweety53/gomigrate/pkg/config"
)

const (
	binaryPath     = "/opt/gomigrate"
	dbSchemaPrefix = "public."

	createActionConfPath = "testdata/configs/bin_test_create_gomigrate.yaml"

	upActionConfPath         = "testdata/configs/bin_test_up_gomigrate.yaml"
	upActionParallelConfPath = "testdata/configs/bin_test_up_gomigrate.yaml"

	downActionConfPath         = "testdata/configs/bin_test_down_gomigrate.yaml"
	downActionParallelConfPath = "testdata/configs/bin_test_down_parallel_gomigrate.yaml"
)

func getDb(appConfig *config.GoMigrateConfig) *sql.DB {
	db, err := sql.Open(appConfig.SQLDialect, appConfig.DataSourceName)
	if err != nil {
		log.Fatalf("-dbstring=%q: %v\n", appConfig.DataSourceName, err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("gomigrate: database ping err: %v\n", err)
	}

	return db
}

func actionUpAll(configPath string) error {
	args := []string{`-config`,
		configPath,
		`up`}

	cmd := exec.Command(
		binaryPath,
		args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
