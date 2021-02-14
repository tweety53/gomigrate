package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/tweety53/gomigrate/pkg/config"
	"github.com/tweety53/gomigrate/pkg/errors"
	"github.com/tweety53/gomigrate/pkg/exit_code"
	"github.com/tweety53/gomigrate/pkg/gomigrate"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var (
	flags          = flag.NewFlagSet("gomigrate", flag.ExitOnError)
	compact        = flags.Bool("c", false, "indicates whether the console output should be compacted")
	migrationsPath = flags.String("p", "", "the directory containing the migration classes")
	migrationTable = flags.String("t", "", "table name which contains migrations data")
	dataSourceName = flags.String("dsn", "", "full data source name")
	configPath     = flags.String("config", "", "path to gomigrate config file")
	sqlDialect     = flags.String("d", "", "your DB sql dialect")

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

	appConfig := &config.GoMigrateConfig{}

	if *configPath != "" {
		yamlConf, err := ioutil.ReadFile(*configPath)
		if err != nil {
			log.Fatalf("yamlConf.Get err   #%v ", err)
		}

		// expand environment variables
		yamlConf = []byte(os.ExpandEnv(string(yamlConf)))

		err = yaml.Unmarshal(yamlConf, appConfig)
		if err != nil {
			log.Fatalf("Unmarshal err: %v", err)
		}
	} else {
		appConfig = &config.GoMigrateConfig{
			MigrationsPath: *migrationsPath,
			MigrationTable: *migrationTable,
			Compact:        *compact,
			SQLDialect:     *sqlDialect,
			DataSourceName: *dataSourceName,
		}
	}

	dsn, action := appConfig.DataSourceName, args[0]

	db, err := sql.Open(appConfig.SQLDialect, dsn)
	if err != nil {
		log.Fatalf("-dbstring=%q: %v\n", dsn, err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("gomigrate: database ping err: %v\n", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("gomigrate: failed to close DB: %v\n", err)
		}
	}()

	var arguments []string
	if len(args) > 3 {
		arguments = append(arguments, args[3:]...)
	}

	switch action {
	case "create":
		if err := gomigrate.Run("create", nil, appConfig, args[1:]); err != nil {
			log.Printf("gomigrate error: %v\n", err)
			os.Exit(int(errors.ErrorExitCode(err)))
		}

		os.Exit(int(exit_code.ExitCodeOK))
	case "up", "down", "fresh", "history", "new", "redo", "to", "mark":
		if err := gomigrate.Run(action, db, appConfig, args[1:]); err != nil {
			log.Printf("gomigrate error: %v\n", err)
			os.Exit(int(errors.ErrorExitCode(err)))
		}

		os.Exit(int(exit_code.ExitCodeOK))
	}

	os.Exit(int(exit_code.ExitCodeOK))
}

func showUsage() {
	fmt.Print(usagePrefix)
	flags.PrintDefaults()
	fmt.Print(usageActions)
}

var usagePrefix = `Usage: gomigrate [OPTIONS] ACTION [ACTION PARAMS]

`

var usageActions = `
Actions:
	create [name:string] [type:enum[sql|go,default:go]] - Creates a new migration
	  create add_new_table     #create new m000000_000000_add_new_table.go file
	  create add_new_table go  #create new m000000_000000_add_new_table.go file
	  create add_new_table sql #create new m000000_000000_add_new_table.sql file

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
