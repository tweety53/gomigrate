package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/tweety53/gomigrate/pkg/config"
	"github.com/tweety53/gomigrate/pkg/errors"
	"github.com/tweety53/gomigrate/pkg/exitcode"
	"github.com/tweety53/gomigrate/pkg/gomigrate"
)

var (
	flags          = flag.NewFlagSet("gomigrate", flag.ExitOnError)
	compact        = flags.Bool("c", false, "indicates whether the console output should be compacted")
	migrationsPath = flags.String("p", "", "the directory containing the migration classes")
	migrationTable = flags.String("t", "", "table name which contains migrations data")
	dataSourceName = flags.String("dsn", "", "full data source name")
	configPath     = flags.String("config", "", "path to gomigrate config file")
	sqlDialect     = flags.String("d", "", "your db sql dialect")

	help = flags.Bool("h", false, "print help")
)

func main() {
	log.Print("gomigrate migration tool\n\n")

	flags.Usage = showUsage

	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	args := flags.Args()
	if len(args) == 0 || *help {
		flags.Usage()

		return
	}

	if len(args) < 1 {
		flags.Usage()

		return
	}

	var appConfig *config.GoMigrateConfig

	if *configPath != "" {
		appConfig, err = config.BuildFromFile(*configPath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		appConfig = config.BuildFromArgs(
			*migrationsPath,
			*migrationTable,
			*compact,
			*sqlDialect,
			*dataSourceName)
	}

	db, err := sql.Open(appConfig.SQLDialect, appConfig.DataSourceName)
	if err != nil {
		log.Fatalf("-dsn=%q: %v\n", appConfig.DataSourceName, err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("gomigrate: database ping err: %v\n", err)
	}

	err = config.Validate(appConfig, db)
	if err != nil {
		log.Printf("%v\n", err)
		shutdown(db, exitcode.Unspecified)
	}

	switch args[0] {
	case "create":
		if err := gomigrate.Run("create", nil, appConfig, args[1:]); err != nil {
			log.Printf("gomigrate error: %v\n", err)
			shutdown(db, errors.ErrorExitCode(err))
		}

		shutdown(db, exitcode.OK)
	case "up", "down", "fresh", "history", "new", "redo", "to", "mark":
		if err := gomigrate.Run(args[0], db, appConfig, args[1:]); err != nil {
			log.Printf("gomigrate error: %v\n", err)
			shutdown(db, errors.ErrorExitCode(err))
		}

		shutdown(db, exitcode.OK)
	}

	shutdown(db, exitcode.OK)
}

func shutdown(db io.Closer, exitCode exitcode.ExitCode) {
	db.Close()
	os.Exit(int(exitCode))
}

//nolint
func showUsage() {
	fmt.Print(usagePrefix)
	flags.PrintDefaults()
	fmt.Print(usageActions)
}

var usagePrefix = `Usage: gomigrate [OPTIONS] ACTION [ACTION PARAMS]

`

var usageActions = `
Actions:
	create [name:string] [type:enum[sql|go,default:go]] [safe:bool,default:true] - Creates a new migration
	  create add_new_table           #create new m000000_000000_add_new_table.go file (will be executed in transaction)
	  create add_new_table go        #create new m000000_000000_add_new_table.go file (will be executed in transaction)
	  create add_new_table go true   #create new m000000_000000_add_new_table.go file (will be executed in transaction)
	  create add_new_table go false  #create new m000000_000000_add_new_table.go file (will be executed without transaction)
	  create add_new_table sql       #create new m000000_000000_add_new_table.sql file (will be executed in transaction)
	  create add_new_table sql true  #create new m000000_000000_add_new_table.sql file (will be executed in transaction)
	  create add_new_table sql false #create new m000000_000000_add_new_table.sql file (will be executed without transaction)

	down [limit:int|all,default:1] - Downgrades the application by reverting old migrations
	  down     #revert last applied migration
	  down 3   #revert last 3 applied migrations
	  down all #revert all applied migrations

	fresh - Truncates the whole database and starts the migration from the beginning

	history [limit:int|all,default:10] - Displays the migration history
	  history     #show last 10 applied versions
	  history 3   #show last 3 applied versions
	  history all #show all applied versions

	mark [version:string] - Modifies the migration history to the specified version
	  mark m000000_000000_add_new_table #modify migrations history to m000000_000000_add_new_table version

	new [limit:int|all,default:10] - Displays the un-applied new migrations
	  new     #show last 10 not applied migrations
	  new 3   #show last 3 not applied migrations
	  new all #show all not applied migrations

	redo [limit:int|all,default:1] - Redoes the last few migrations
	  redo     #redo last applied migration
	  redo 3   #redo last 3 applied migrations
	  redo all #redo all applied migrations

	to [version:string] - Upgrades or downgrades till the specified version
	  to m000000_000000_add_new_table #apply\revert all migrations to m000000_000000_add_new_table version

	up [limit:int,default:0] - Upgrades the application by applying new migrations
	  up   #apply all new migrations
	  up 3 #apply the first 3 new migrations

`
