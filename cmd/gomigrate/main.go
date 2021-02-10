package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/tweety53/gomigrate/internal/errors"
	"github.com/tweety53/gomigrate/internal/exit_code"
	"github.com/tweety53/gomigrate/internal/gomigrate"
	"log"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/tweety53/gomigrate/migrations"
)

var (
	flags              = flag.NewFlagSet("gomigrate", flag.ExitOnError)
	compact            = flags.Bool("c", false, "indicates whether the console output should be compacted")
	migrationsPath     = flags.String("p", "../migrations", "the directory containing the migration classes")
	migrationTable     = flags.String("t", "migration", "table name which contains migrations data")
	goTemplateFilePath = flags.String("F", "action.GoMigrationTemplate", "file with custom template for .go migrations")

	help = flags.Bool("h", false, "print help")
)

func main() {
	log.Print("gomigrate migration tool\n\n")

	flags.Usage = showUsage

	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if *compact {
		gomigrate.SetCompact(true)
	}
	gomigrate.SetMigrationsPath(*migrationsPath)
	gomigrate.SetMigrationTable(*migrationTable)
	gomigrate.SetGoTemplateFilePath(*goTemplateFilePath)

	args := flags.Args()
	if len(args) == 0 || *help {
		flags.Usage()
		return
	}

	if len(args) < 1 {
		flags.Usage()
		return
	}

	dsn, action := "host=localhost port=5433 user=myuser "+
		"password=mypass dbname=gomigrate_test sslmode=disable", args[0]

	db, err := sql.Open("postgres", dsn)
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
		if err := gomigrate.Run("create", nil, args[1:]); err != nil {
			log.Printf("gomigrate error: %v\n", err)
			os.Exit(int(errors.ErrorExitCode(err)))
		}

		os.Exit(int(exit_code.ExitCodeOK))
	case "up", "down", "fresh", "history", "new", "redo", "to", "mark":
		if err := gomigrate.Run(action, db, args[1:]); err != nil {
			log.Printf("gomigrate error: %v\n", err)
			os.Exit(int(errors.ErrorExitCode(err)))
		}

		os.Exit(int(exit_code.ExitCodeOK))
	}

	os.Exit(0)
}

func showUsage() {
	fmt.Print(usagePrefix)
	flags.PrintDefaults()
	fmt.Print(usageActions)
}

var usagePrefix = `Usage: gomigrate [OPTIONS] ACTION

`

var usageActions = `
Actions:
	create      - Creates a new migration
	down        - Downgrades the application by reverting old migrations
	fresh       - Truncates the whole database and starts the migration from the beginning
	history     - Displays the migration history
	mark        - Modifies the migration history to the specified version
	new         - Displays the un-applied new migrations
	redo        - Redoes the last few migrations
	to          - Upgrades or downgrades till the specified version
	up(default) - Upgrades the application by applying new migrations
`
