[![go](https://github.com/tweety53/gomigrate/workflows/Go/badge.svg)](https://github.com/tweety53/gomigrate/actions)
[![go](https://github.com/tweety53/gomigrate/workflows/golangci-lint/badge.svg)](https://github.com/tweety53/gomigrate/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/tweety53/gomigrate)](https://goreportcard.com/report/github.com/tweety53/gomigrate)

# gomigrate
Database migrations written in go, with https://github.com/yiisoft/yii2 like API
## Databases supported
* PostgreSQL (dialect: postgres)
## Supported migration file types
* .sql
* .go (WIP)
## CLI usage

###install and run in your system
* ```go get https://github.com/tweety53/gomigrate/cmd/gomigrate```
* run from your ```GOBIN``` path with options:

###run options
* -config string - (only .yaml type supported, env variables expanding supported) 
  * Example: -config /app/config/gomigrate.yaml (copy config from https://github.com/tweety53/gomigrate/blob/master/examples/gomigrate.yaml and update with your actual environment)

OR

* -c bool default: false - indicates whether the console output should be compacted (this is something like verbose:true)
* -p string - the directory containing the migration classes
* -t string - table name which contains migrations data
* -dsn string - full data source name
* -d string - your DB sql dialect (see available [here](#databases-supported))

###and then add action(required) and params(optional, depends on action)
```text
Usage: gomigrate [OPTIONS] ACTION [ACTION PARAMS]

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

```
## Use in your go project as library (WIP)
### Progress check list

- [x] Наличие юнит-тестов на ключевые алгоритмы (core-логику) сервиса. - вохможно есть не все конечно :(
- [x] Наличие валидных Dockerfile и Makefile для сервиса.
- [x] Пайплайн - github actions : go build ... , go test ..., golangci-lint (есть нюанс что предложенный конфиг ломается на текущей версии либы на github actions, запускал с предложенным на локалке)

- [x] Использовать как cli
- [ ] Использовать как библиотеку из кода (отложено из-за непоняток с .go миграциями)
- [ ] Поддержка миграций на Go (запускаются пока только если есть внутри самой программы)
- [x] Поддержка миграций на SQL
- [x] Реализован механизм блокировки на время миграции (на up все вроде бы ок, на down возможно нужно будет пересмотреть некоторые кейсы)
- [x] Реализованы различные способы конфигурирования - yaml конфиг(c expand env) (с указанием пути через флаг -config) или флаги 
- [x] Написаны юнит-тесты - написаны, но не полностью
- [ ] Написаны интеграционные тесты - написаны для create, up, down
- [ ] Тесты адекватны и полностью покрывают фукнционал
- [ ] Понятность и чистота кода - текущий хз, рефакторинг будет по ходу и по окончанию написания юнит и инт. тестов



