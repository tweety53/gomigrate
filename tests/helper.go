package tests

import (
	"database/sql"
	"github.com/tweety53/gomigrate/pkg/config"
	"log"

	_ "github.com/lib/pq"
)

const binaryPath = "/opt/gomigrate"

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