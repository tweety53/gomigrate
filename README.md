[![go](https://github.com/tweety53/gomigrate/workflows/Go/badge.svg)](https://github.com/tweety53/gomigrate/actions)
[![go](https://github.com/tweety53/gomigrate/workflows/golangci-lint/badge.svg)](https://github.com/tweety53/gomigrate/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/tweety53/gomigrate)](https://goreportcard.com/report/github.com/tweety53/gomigrate)

# gomigrate
___Database migrations written in go
## Databases supported
* PostgreSQL
## Supported migration file types
* .sql
* .go (WIP)
## CLI usage
```text
Usage: gomigrate [OPTIONS] ACTION [ACTION PARAMS]

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

```
## Use in your go project as library (WIP)



