// +build test_integration

package tests

import (
	"log"
	"os"
	"os/exec"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tweety53/gomigrate/internal/repo"
	"github.com/tweety53/gomigrate/internal/sqldialect"
	"github.com/tweety53/gomigrate/pkg/config"
)

func Test_UpAction(t *testing.T) {
	// prepare
	conf, err := config.BuildFromFile(upActionConfPath)
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
			name: "up action: all",
			args: []string{},
			wantRecords: repo.MigrationRecords{
				&repo.MigrationRecord{
					Version: "m000000_000003_add_another_table",
				},
				&repo.MigrationRecord{
					Version: "m000000_000002_alter_some_table_column",
				},
				&repo.MigrationRecord{
					Version: "m000000_000001_add_some_table",
				},
			},
			wantTables: []string{"some_table", "another_table"},
		},
		{
			name: "up action: one",
			args: []string{"1"},
			wantRecords: repo.MigrationRecords{
				&repo.MigrationRecord{
					Version: "m000000_000001_add_some_table",
				},
			},
			wantTables: []string{"some_table"},
		},
		{
			name: "up action: two",
			args: []string{"2"},
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
			name:        "up action: bad limit arg",
			args:        []string{"kek"},
			wantRecords: repo.MigrationRecords{},
			wantTables:  []string{},
			wantErr:     true,
		},
	}
	for _, test := range tests {
		// cleanup db before run
		err := dboRepo.TruncateDatabase()
		if err != nil {
			log.Fatal(err)
		}

		// run tc
		t.Run(test.name, func(t *testing.T) {
			// prepare and exec cmd
			args := []string{`-config`,
				upActionConfPath,
				`up`}

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
			var newTables []string
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
			require.Equal(t, len(history), len(test.wantRecords))
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

func Test_UpActionParallel(t *testing.T) {
	var (
		successCnt int64
		failCnt    int64
	)

	wantRecords := repo.MigrationRecords{
		&repo.MigrationRecord{
			Version: "m000000_000003_add_another_table",
		},
		&repo.MigrationRecord{
			Version: "m000000_000002_alter_some_table_column",
		},
		&repo.MigrationRecord{
			Version: "m000000_000001_add_some_table",
		},
	}

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "up action: all runner 1",
			args: []string{},
		},
		{
			name: "up action: all runner 2",
			args: []string{},
		},
		{
			name: "up action: all runner 3",
			args: []string{},
		},
	}

	t.Run("wrapper for parallel runs waiting", func(t *testing.T) {
		for _, test := range tests {
			test := test
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				args := []string{`-config`,
					upActionParallelConfPath,
					`up`}

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

	conf, err := config.BuildFromFile(upActionParallelConfPath)
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

	history, err := mRepo.GetMigrationsHistory(0)
	require.Equal(t, len(wantRecords), len(history))
	for i := range history {
		require.Equal(t, wantRecords[i].Version, history[i].Version)
		require.Greater(t, history[i].ApplyTime, 0)
	}

	require.Equal(t, int64(1), atomic.LoadInt64(&successCnt))
	require.Equal(t, int64(2), atomic.LoadInt64(&failCnt))

	err = dboRepo.TruncateDatabase()
	if err != nil {
		log.Fatal(err)
	}
}
