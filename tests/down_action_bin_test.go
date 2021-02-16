// +build test_integration

package tests

import (
	"log"
	"os"
	"os/exec"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tweety53/gomigrate/internal/helpers"
	"github.com/tweety53/gomigrate/internal/repo"
	"github.com/tweety53/gomigrate/internal/sqldialect"
	"github.com/tweety53/gomigrate/pkg/config"
)

func Test_DownAction(t *testing.T) {
	// prepare
	conf, err := config.BuildFromFile(downActionConfPath)
	if err != nil {
		log.Fatal(err)
	}

	db := getDb(conf)
	defer db.Close()
	dialect, err := sqldialect.InitDialect(conf.SQLDialect, conf.MigrationTable)
	if err != nil {
		log.Fatal(err)
	}

	mRepo := repo.NewMigrationsRepository(db, dialect)
	dboRepo := repo.NewDBOperationsRepository(db, dialect)
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		wantRecords repo.MigrationRecords
		wantTables  []string
	}{
		{
			name: "down action: last",
			args: []string{},
			wantRecords: repo.MigrationRecords{
				&repo.MigrationRecord{
					Version: "m000000_000002_alter_some_table_column",
				},
				&repo.MigrationRecord{
					Version: "m000000_000001_add_some_table",
				},
			},
			wantTables: []string{"some_table"},
		},
		{
			name: "down action: two",
			args: []string{"2"},
			wantRecords: repo.MigrationRecords{
				&repo.MigrationRecord{
					Version: "m000000_000001_add_some_table",
				},
			},
			wantTables: []string{"some_table"},
		},
		{
			name:        "down action: all",
			args:        []string{"all"},
			wantRecords: repo.MigrationRecords{},
			wantTables:  []string{},
		},
		{
			name:        "down action: bad limit arg",
			args:        []string{"kek"},
			wantRecords: repo.MigrationRecords{},
			wantTables:  []string{},
			wantErr:     true,
		},
	}
	for _, test := range tests {

		// cleanup db
		err := dboRepo.TruncateDatabase()
		if err != nil {
			log.Fatal(err)
		}
		err = actionUpAll(downActionConfPath)
		if err != nil {
			log.Fatal(err)
		}

		// run tc
		t.Run(test.name, func(t *testing.T) {
			// prepare and exec cmd
			args := []string{`-config`,
				downActionConfPath,
				`down`}

			args = append(args, test.args...)

			cmd := exec.Command(
				binaryPath,
				args...)

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if test.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			// prepare data for results check
			history, err := mRepo.GetMigrationsHistory(0)
			tables, err := dboRepo.AllTableNames()
			if err != nil {
				log.Fatal(err)
			}
			newTables := make([]string, 0, len(tables))
			for i := range tables {
				if tables[i] == dbSchemaPrefix+conf.MigrationTable {
					continue
				}

				newTables = append(newTables, tables[i])
			}

			for i := range test.wantTables {
				test.wantTables[i] = dbSchemaPrefix + test.wantTables[i]
			}

			// check results
			require.Equal(t, len(test.wantRecords), len(history))
			for i := range history {
				require.Equal(t, test.wantRecords[i].Version, history[i].Version)
				require.Greater(t, history[i].ApplyTime, 0)
			}
			require.Equal(t, test.wantTables, newTables)
		})
	}

	err = dboRepo.TruncateDatabase()
	if err != nil {
		log.Fatal(err)
	}
}

func Test_DownActionParallel(t *testing.T) {
	// prepare
	var (
		successCnt int64
		failCnt    int64
	)

	conf, err := config.BuildFromFile(downActionParallelConfPath)
	if err != nil {
		log.Fatal(err)
	}

	db := getDb(conf)
	defer db.Close()
	dialect, err := sqldialect.InitDialect(conf.SQLDialect, conf.MigrationTable)
	if err != nil {
		log.Fatal(err)
	}

	mRepo := repo.NewMigrationsRepository(db, dialect)
	dboRepo := repo.NewDBOperationsRepository(db, dialect)

	// cleanup db
	err = dboRepo.TruncateDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err = actionUpAll(downActionParallelConfPath)
	if err != nil {
		log.Fatalf("cannot prepare Test_DownActionParallel, err: %v", err)
	}

	wantRecords := repo.MigrationRecords{}

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "down action: all runner 1",
			args: []string{helpers.LimitAll},
		},
		{
			name: "down action: all runner 2",
			args: []string{helpers.LimitAll},
		},
		{
			name: "down action: all runner 3",
			args: []string{helpers.LimitAll},
		},
	}

	t.Run("wrapper for parallel runs waiting", func(t *testing.T) {
		for _, test := range tests {
			test := test
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				args := []string{`-config`,
					downActionParallelConfPath,
					`down`}

				args = append(args, test.args...)

				cmd := exec.Command(
					binaryPath,
					args...)

				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				err := cmd.Run()
				if err != nil {
					atomic.AddInt64(&failCnt, 1)
				} else {
					atomic.AddInt64(&successCnt, 1)
				}
			})
		}
	})

	// prepare data for results check
	history, err := mRepo.GetMigrationsHistory(0)
	require.Equal(t, len(wantRecords), len(history))
	require.Equal(t, int64(1), atomic.LoadInt64(&successCnt))
	require.Equal(t, int64(2), atomic.LoadInt64(&failCnt))
}
